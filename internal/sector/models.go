package sector

import (
	"time"
)

type Sector struct {
	ID          int       `json:"id"`
	GalaxyID    int       `json:"galaxy_id"`
	SectorX     int       `json:"sector_x"`
	SectorY     int       `json:"sector_y"`
	Name        string    `json:"name"`
	SystemCount int       `json:"system_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
