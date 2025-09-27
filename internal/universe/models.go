package universe

import (
	"time"
)

// Universe represents the root container for all spatial data
type Universe struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	GalaxyCount int       `json:"galaxy_count"`
	SectorCount int       `json:"sector_count"`
	SystemCount int       `json:"system_count"`
	PlanetCount int       `json:"planet_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UniverseConfig contains configuration for universe generation
type UniverseConfig struct {
	GalaxyCount         int `json:"galaxy_count"`
	SectorsPerGalaxy    int `json:"sector_count"`
	SystemsPerSector    int `json:"systems_per_sector"`
	MinPlanetsPerSystem int `json:"min_planets_per_system"`
	MaxPlanetsPerSystem int `json:"max_planets_per_system"`
}
