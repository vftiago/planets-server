package middleware

import (
	"log/slog"
	"net/http"
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
			logger.Warn("No user context found in admin middleware")
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		if claims.Role != "admin" {
			logger.Warn("Non-admin user attempted to access admin endpoint",
				"player_id", claims.PlayerID,
				"username", claims.Username,
				"role", claims.Role)
			http.Error(w, "Admin access required", http.StatusForbidden)
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
