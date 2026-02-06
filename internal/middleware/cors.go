package middleware

import (
	"log/slog"
	"net/http"
	"planets-server/internal/shared/config"

	"github.com/rs/cors"
)

type CORSMiddleware struct {
	*cors.Cors
}

func NewCORS() *CORSMiddleware {
	cfg := config.GlobalConfig
	logger := slog.With("component", "cors", "operation", "setup")
	logger.Debug("Setting up CORS middleware")

	allowedOrigins := []string{cfg.Frontend.URL}

	corsConfig := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		Debug:            cfg.Frontend.CORSDebug,
	})

	logger.Info("CORS middleware configured",
		"allowed_origins", allowedOrigins,
		"allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		"allow_credentials", true,
		"debug_mode", cfg.Frontend.CORSDebug,
	)

	if cfg.Frontend.CORSDebug {
		logger.Debug("CORS debug mode enabled - will log CORS request details")
	}

	return &CORSMiddleware{corsConfig}
}

func (c *CORSMiddleware) Middleware(h http.Handler) http.Handler {
	return c.Cors.Handler(h)
}
