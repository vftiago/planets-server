package middleware

import (
	"context"
	"net/http"
	"planets-server/internal/auth"
)

type contextKey string

const UserContextKey contextKey = "user"

func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth_token")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		claims, err := auth.ValidateJWT(cookie.Value)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add user info to request context
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
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
