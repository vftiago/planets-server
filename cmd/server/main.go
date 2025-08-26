package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"planets-server/internal/auth"
	"planets-server/internal/middleware"
	"planets-server/internal/models"
	"planets-server/internal/server"
	"planets-server/internal/shared/config"
	"planets-server/internal/shared/database"
	"planets-server/internal/shared/logger"
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

	oauthConfig := initOAuth()
	
	db, err := initDatabase()
	if err != nil {
		logger.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.RunMigrations(); err != nil {
		logger.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	playerRepo := models.NewPlayerRepository(db.DB)

	corsMiddleware := initCORS()

	routes := server.NewRoutes(db, playerRepo, oauthConfig)
	mux := routes.Setup()
	handler := corsMiddleware.Handler(mux)

	httpServer := createHTTPServer(handler)

	go startServer(httpServer, logger)
	
	waitForShutdown(httpServer, logger)
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
	cfg := config.GlobalConfig
	logger := slog.With("component", "cors", "operation", "setup")
	logger.Debug("Setting up CORS middleware")

	corsMiddleware := middleware.SetupCORS()

	logger.Info("CORS middleware configured",
		"allowed_origins", []string{cfg.Frontend.URL},
		"debug_mode", cfg.Frontend.CORSDebug,
	)

	return corsMiddleware
}

func createHTTPServer(handler http.Handler) *http.Server {
	cfg := config.GlobalConfig
	port := cfg.Server.Port
	if port[0] != ':' {
		port = ":" + port
	}

	return &http.Server{
		Addr:         port,
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
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
