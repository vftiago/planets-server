package galaxy

import (
	"time"
)

type Galaxy struct {
	ID          int       `json:"id"`
	GameID      int       `json:"game_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	SectorCount int       `json:"sector_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
