package middleware

import (
	"planets-server/internal/utils"

	"github.com/rs/cors"
)

func SetupCORS() *cors.Cors {
	allowedOrigins := []string{
		utils.GetEnv("FRONTEND_URL", "http://localhost:3000"),
	}

	if prodOrigin := utils.GetEnv("PROD_FRONTEND_URL", ""); prodOrigin != "" {
		allowedOrigins = append(allowedOrigins, prodOrigin)
	}

	return cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		Debug:            utils.GetEnv("CORS_DEBUG", "") == "true",
	})
}
