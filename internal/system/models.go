package system

import (
	"time"
)

type System struct {
	ID          int       `json:"id"`
	SectorID    int       `json:"sector_id"`
	SystemX     int       `json:"system_x"`
	SystemY     int       `json:"system_y"`
	Name        string    `json:"name"`
	PlanetCount int       `json:"planet_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
