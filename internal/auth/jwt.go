package auth

import (
	"fmt"
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
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", fmt.Errorf("JWT_SECRET environment variable is required but not set")
	}
	if len(secret) < 32 {
		return "", fmt.Errorf("JWT_SECRET must be at least 32 characters long for security")
	}
	return secret, nil
}

func GenerateJWT(player *models.Player) (string, error) {
	secret, err := getJWTSecret()
	if err != nil {
		return "", fmt.Errorf("cannot generate JWT: %w", err)
	}
	
	claims := Claims{
		PlayerID: player.ID,
		Username: player.Username,
		Email:    player.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("player_%d", player.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateJWT(tokenString string) (*Claims, error) {
	secret, err := getJWTSecret()
	if err != nil {
		return nil, fmt.Errorf("cannot validate JWT: %w", err)
	}
	
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
