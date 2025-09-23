package player

import (
	"time"
)

type PlayerRole string

const (
	PlayerRoleUser  PlayerRole = "user"
	PlayerRoleAdmin PlayerRole = "admin"
)

type Player struct {
	ID          int        `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	DisplayName string     `json:"display_name"`
	AvatarURL   *string    `json:"avatar_url"`
	Role        PlayerRole `json:"role"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (r PlayerRole) String() string {
	return string(r)
}

func (r PlayerRole) IsValid() bool {
	return r == PlayerRoleUser || r == PlayerRoleAdmin
}

func ParsePlayerRole(s string) PlayerRole {
	switch s {
	case "admin":
		return PlayerRoleAdmin
	case "user":
		return PlayerRoleUser
	default:
		return PlayerRoleUser
	}
}
