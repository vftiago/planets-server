package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"planets-server/internal/auth"
	"planets-server/internal/database"
	"planets-server/internal/middleware"
	"planets-server/internal/models"
	"planets-server/internal/server"
	"planets-server/internal/utils"

	"github.com/joho/godotenv"
)

func initLogger() {
	var handler slog.Handler
	
	if utils.GetEnv("ENVIRONMENT", "development") == "production" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}
	
	slog.SetDefault(slog.New(handler))
}

func main() {
	// Initialize logger
	initLogger()
	logger := slog.With("component", "main")
	
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logger.Debug("No .env file found, using system environment variables")
	}

	// Initialize OAuth configuration
	auth.InitOAuth()
	logger.Info("OAuth configuration initialized")
	
	// Connect to database
	db, err := database.Connect()
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("Database connection established")

	// Run migrations
	logger.Info("Running database migrations...")
	if err := db.RunMigrations(); err != nil {
		logger.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}
	logger.Info("Migrations completed successfully")

	// Initialize repositories and services
	playerRepo := models.NewPlayerRepository(db.DB)
	oauthService := auth.NewOAuthService(playerRepo)
	logger.Info("Services initialized")

	// Setup middleware
	corsMiddleware := middleware.SetupCORS()
	frontendURL := utils.GetEnv("FRONTEND_URL", "http://localhost:3000")
	logger.Info("CORS configured", "allowed_origin", frontendURL)

	// Setup routes
	routes := server.NewRoutes(db, playerRepo, oauthService)
	mux := routes.Setup()
	handler := corsMiddleware.Handler(mux)

	// Configure and start server
	port := utils.GetEnv("PORT", "8080")
	if port[0] != ':' {
		port = ":" + port
	}
	
	httpServer := &http.Server{
		Addr:         port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting Planets! server", "port", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "error", err, "port", port)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	logger.Info("Shutting down server...")
	
	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}
	
	logger.Info("Server exited gracefully")
}
