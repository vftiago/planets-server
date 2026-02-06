package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"planets-server/internal/auth"
	"planets-server/internal/shared/errors"
	"planets-server/internal/shared/response"
)

type contextKey string

const UserContextKey contextKey = "user"

func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := slog.With(
			"middleware", "jwt",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
		)
		logger.Debug("Processing JWT authentication")

		// Get auth token from cookie
		cookie, err := r.Cookie("auth_token")
		if err != nil {
			response.Error(w, r, logger, errors.Unauthorized("authentication required"))
			return
		}

		// Validate JWT token
		claims, err := auth.ValidateJWT(cookie.Value)
		if err != nil {
			response.Error(w, r, logger, errors.Unauthorized("invalid token"))
			return
		}

		// Add user info to request context
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		logger.Debug("JWT authentication successful",
			"player_id", claims.PlayerID,
			"username", claims.Username)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Helper to get user from context
func GetUserFromContext(r *http.Request) *auth.Claims {
	if claims, ok := r.Context().Value(UserContextKey).(*auth.Claims); ok {
		return claims
	}
	return nil
}
