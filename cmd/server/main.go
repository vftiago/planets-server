package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"planets-server/internal/auth"
	"planets-server/internal/game"
	"planets-server/internal/middleware"
	"planets-server/internal/planet"
	"planets-server/internal/player"
	"planets-server/internal/server"
	"planets-server/internal/shared/config"
	"planets-server/internal/shared/database"
	"planets-server/internal/shared/logger"
	"planets-server/internal/shared/redis"
	"planets-server/internal/spatial"
)

func main() {
	if err := config.Init(); err != nil {
		slog.Error("Failed to initialize configuration", "error", err)
		os.Exit(1)
	}

	cfg := config.GlobalConfig

	logger.Init()

	logger := slog.With("component", "main")
	logger.Info("Starting Planets! server",
		"environment", cfg.Server.Environment,
		"port", cfg.Server.Port,
	)

	redisClient, err := initRedis()
	if err != nil {
		logger.Error("Failed to initialize Redis", "error", err)
		os.Exit(1)
	}
	defer func() {
		if redisClient != nil {
			if err := redisClient.Close(); err != nil {
				logger.Error("Failed to close Redis connection", "error", err)
			}
		}
	}()

	auth.InitStateManager(redisClient)

	oauthConfig := initOAuth()

	db, err := initDatabase()
	if err != nil {
		logger.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("Failed to close database connection", "error", err)
		}
	}()

	if err := db.RunMigrations(); err != nil {
		logger.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	authRepo := auth.NewRepository(db)
	playerRepo := player.NewRepository(db)
	spatialRepo := spatial.NewRepository(db)
	planetRepo := planet.NewRepository(db)

	authService := auth.NewService(authRepo)
	playerService := player.NewService(playerRepo)
	spatialService := spatial.NewService(spatialRepo)
	planetService := planet.NewService(planetRepo)

	gameRepo := game.NewRepository(db)
	gameService := game.NewService(gameRepo, spatialService, planetService)

	cors := initCORS()
	rateLimiter := initRateLimiter()

	routes := server.NewRoutes(db, playerService, authService, gameService, spatialService, planetService, oauthConfig, logger)
	mux := routes.Setup()

	var handler http.Handler = mux
	handler = rateLimiter.Middleware(handler)
	handler = cors.Middleware(handler)

	httpServer := createHTTPServer(handler)

	go startServer(httpServer, logger)

	waitForShutdown(httpServer, logger)
}

func initRedis() (*redis.Client, error) {
	cfg := config.GlobalConfig
	logger := slog.With("component", "redis", "operation", "init")
	logger.Debug("Initializing Redis connection")

	if !cfg.Redis.Enabled {
		logger.Info("Redis disabled, OAuth state will use in-memory storage")
		return nil, nil
	}

	redisClient, err := redis.Connect()
	if err != nil {
		logger.Error("Failed to connect to Redis", "error", err)
		return nil, err
	}

	logger.Info("Redis initialized successfully",
		"url_provided", cfg.Redis.URL != "",
		"host", cfg.Redis.Host)

	return redisClient, nil
}

func initOAuth() *auth.OAuthConfig {
	cfg := config.GlobalConfig
	logger := slog.With("component", "oauth", "operation", "init")
	logger.Debug("Initializing OAuth configurations")

	oauthConfig := auth.InitOAuth()

	logger.Info("OAuth configuration completed",
		"google_configured", cfg.GoogleOAuthConfigured(),
		"github_configured", cfg.GitHubOAuthConfigured(),
	)

	return oauthConfig
}

func initDatabase() (*database.DB, error) {
	cfg := config.GlobalConfig
	logger := slog.With("component", "database", "operation", "init")
	logger.Debug("Connecting to database")

	db, err := database.Connect()
	if err != nil {
		return nil, err
	}

	logger.Info("Database connection established",
		"host", cfg.Database.Host,
		"database", cfg.Database.Name,
	)

	return db, nil
}

func initCORS() *middleware.CORSMiddleware {
	return middleware.NewCORS()
}

func initRateLimiter() *middleware.RateLimiter {
	cfg := config.GlobalConfig
	logger := slog.With("component", "rate_limit", "operation", "init")
	logger.Debug("Setting up rate limiting middleware")

	rateLimitConfig := middleware.RateLimitConfig{
		RequestsPerSecond: cfg.RateLimit.RequestsPerSecond,
		BurstSize:         cfg.RateLimit.BurstSize,
		TrustProxy:        cfg.RateLimit.TrustProxy,
	}

	rateLimiter := middleware.NewRateLimiter(rateLimitConfig)

	logger.Info("Rate limiting middleware configured",
		"requests_per_second", rateLimitConfig.RequestsPerSecond,
		"burst_size", rateLimitConfig.BurstSize,
	)

	return rateLimiter
}

func createHTTPServer(handler http.Handler) *http.Server {
	cfg := config.GlobalConfig
	port := cfg.Server.Port
	if port[0] != ':' {
		port = ":" + port
	}

	return &http.Server{
		Addr:           port,
		Handler:        handler,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		IdleTimeout:    cfg.Server.IdleTimeout,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}
}

func startServer(server *http.Server, logger *slog.Logger) {
	logger.Info("HTTP server starting", "addr", server.Addr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("Server failed to start", "error", err, "addr", server.Addr)
		os.Exit(1)
	}
}

func waitForShutdown(server *http.Server, logger *slog.Logger) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), config.GlobalConfig.Server.WriteTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("Server exited gracefully")
}
