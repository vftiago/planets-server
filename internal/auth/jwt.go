package auth

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"planets-server/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	PlayerID int    `json:"player_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

func getJWTSecret() (string, error) {
	logger := slog.With("component", "jwt", "operation", "get_secret")
	
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		logger.Error("JWT_SECRET environment variable is required but not set")
		return "", fmt.Errorf("JWT_SECRET environment variable is required but not set")
	}
	if len(secret) < 32 {
		logger.Error("JWT_SECRET is too short", "min_length", 32)
		return "", fmt.Errorf("JWT_SECRET must be at least 32 characters long")
	}
	
	logger.Debug("JWT secret validated")
	return secret, nil
}

func GenerateJWT(player *models.Player) (string, error) {
	logger := slog.With(
		"component", "jwt", 
		"operation", "generate",
		"player_id", player.ID,
		"username", player.Username,
	)
	logger.Debug("Generating JWT token for player")
	
	secret, err := getJWTSecret()
	if err != nil {
		logger.Error("Failed to get JWT secret", "error", err)
		return "", fmt.Errorf("cannot generate JWT: %w", err)
	}
	
	expiresAt := time.Now().Add(24 * time.Hour)
	claims := Claims{
		PlayerID: player.ID,
		Username: player.Username,
		Email:    player.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("player_%d", player.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		logger.Error("Failed to sign JWT token", "error", err)
		return "", fmt.Errorf("failed to sign JWT token: %w", err)
	}
	
	logger.Debug("JWT token generated successfully", "expires_at", expiresAt)
	return tokenString, nil
}

func ValidateJWT(tokenString string) (*Claims, error) {
	logger := slog.With("component", "jwt", "operation", "validate")
	logger.Debug("Validating JWT token")
	
	secret, err := getJWTSecret()
	if err != nil {
		logger.Error("Failed to get JWT secret for validation", "error", err)
		return nil, fmt.Errorf("cannot validate JWT: %w", err)
	}
	
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			logger.Error("Unexpected JWT signing method", "method", token.Header["alg"])
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		logger.Warn("JWT token validation failed", "error", err)
		return nil, fmt.Errorf("invalid JWT token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		logger.Debug("JWT token validated successfully", 
			"player_id", claims.PlayerID,
			"username", claims.Username,
			"expires_at", claims.ExpiresAt.Time)
		return claims, nil
	}

	logger.Error("JWT token claims are invalid")
	return nil, fmt.Errorf("invalid token claims")
}
