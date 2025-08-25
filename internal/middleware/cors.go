package middleware

import (
	"log/slog"
	"planets-server/internal/utils"

	"github.com/rs/cors"
)

func SetupCORS() *cors.Cors {
	logger := slog.With("component", "cors", "operation", "setup")
	logger.Debug("Setting up CORS middleware")

	// Get allowed origins
	allowedOrigins := []string{
		utils.GetEnv("FRONTEND_URL", "http://localhost:3000"),
	}
	
	// Check debug mode
	debugMode := utils.GetEnv("CORS_DEBUG", "") == "true"

	corsConfig := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		Debug:            debugMode,
	})

	logger.Info("CORS middleware configured",
		"allowed_origins", allowedOrigins,
		"allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		"allow_credentials", true,
		"debug_mode", debugMode,
	)

	if debugMode {
		logger.Debug("CORS debug mode enabled - will log CORS request details")
	}

	return corsConfig
}
