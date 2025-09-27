package galaxy

import (
	"time"
)

type Galaxy struct {
	ID          int       `json:"id"`
	UniverseID  int       `json:"universe_id"`
	Name        string    `json:"name"`
	GalaxyX     int       `json:"galaxy_x"`
	GalaxyY     int       `json:"galaxy_y"`
	SectorCount int       `json:"sector_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
