package auth

import (
	"fmt"
	"log/slog"
	"time"

	"planets-server/internal/shared/config"

	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	logger := slog.With("component", "auth_service", "operation", "init")
	logger.Debug("Initializing auth service")
	return &Service{repo: repo}
}

func (s *Service) GenerateJWT(playerID int, username, email string) (string, error) {
	cfg := config.GlobalConfig
	logger := slog.With(
		"component", "auth_service",
		"operation", "generate_jwt",
		"player_id", playerID,
		"username", username,
	)
	logger.Debug("Generating JWT token for player")

	expiresAt := time.Now().Add(cfg.Auth.TokenExpiration)
	claims := Claims{
		PlayerID: playerID,
		Username: username,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("player_%d", playerID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.Auth.JWTSecret))
	if err != nil {
		logger.Error("Failed to sign JWT token", "error", err)
		return "", fmt.Errorf("failed to sign JWT token: %w", err)
	}

	logger.Debug("JWT token generated successfully", "expires_at", expiresAt)
	return tokenString, nil
}

func (s *Service) ValidateJWT(tokenString string) (*Claims, error) {
	cfg := config.GlobalConfig
	logger := slog.With("component", "auth_service", "operation", "validate_jwt")
	logger.Debug("Validating JWT token")

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			logger.Error("Unexpected JWT signing method", "method", token.Header["alg"])
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.Auth.JWTSecret), nil
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

func (s *Service) CreateAuthProvider(playerID int, provider, providerUserID, providerEmail string) error {
	return s.repo.CreateAuthProvider(playerID, provider, providerUserID, providerEmail)
}

func (s *Service) FindPlayerByAuthProvider(provider, providerUserID string) (int, error) {
	return s.repo.FindPlayerByAuthProvider(provider, providerUserID)
}