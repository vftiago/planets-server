package middleware

import (
	"log/slog"
	"net/http"
	"planets-server/internal/shared/errors"
	"planets-server/internal/shared/response"
)

func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := slog.With(
			"middleware", "admin",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
		)
		logger.Debug("Processing admin authorization")

		claims := GetUserFromContext(r)
		if claims == nil {
			response.Error(w, r, logger, errors.Unauthorized("authentication required"))
			return
		}

		if claims.Role != "admin" {
			logger.Warn("Non-admin user attempted to access admin endpoint",
				"player_id", claims.PlayerID,
				"username", claims.Username,
				"role", claims.Role)
			response.Error(w, r, logger, errors.Forbidden("admin access required"))
			return
		}

		logger.Debug("Admin authorization successful",
			"player_id", claims.PlayerID,
			"username", claims.Username)

		next.ServeHTTP(w, r)
	})
}

func RequireAdmin(next http.Handler) http.Handler {
	return JWTMiddleware(AdminMiddleware(next))
}
