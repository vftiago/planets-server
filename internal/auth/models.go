package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	PlayerID int    `json:"player_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

type PlayerAuthProvider struct {
	ID             int       `json:"id"`
	PlayerID       int       `json:"player_id"`
	Provider       string    `json:"provider"`
	ProviderUserID *string   `json:"provider_user_id"`
	ProviderEmail  *string   `json:"provider_email"`
	CreatedAt      time.Time `json:"created_at"`
}
