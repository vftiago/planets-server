package auth

import (
	"fmt"
	"time"

	"planets-server/internal/models"
	"planets-server/internal/utils"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	PlayerID int    `json:"player_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

func GenerateJWT(player *models.Player) (string, error) {
	secret := utils.GetEnv("JWT_SECRET", "dev-secret-key-change-in-production")
	
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
	secret := utils.GetEnv("JWT_SECRET", "dev-secret-key-change-in-production")
	
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
