package universe

import (
	"time"
)

type Galaxy struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	SectorCount int       `json:"sector_count"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

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

type PlanetType string

const (
	PlanetTypeBarren     	PlanetType = "barren"
	PlanetTypeTerrestrial PlanetType = "terrestrial"
	PlanetTypeGasGiant   	PlanetType = "gas_giant"
	PlanetTypeIce        	PlanetType = "ice"
	PlanetTypeVolcanic   	PlanetType = "volcanic"
)

type Planet struct {
	ID           int        `json:"id"`
	SystemID     int        `json:"system_id"`
	PlanetIndex  int        `json:"planet_index"`
	Name         string     `json:"name"`
	Type         PlanetType `json:"type"`
	Size         int        `json:"size"`
	Population   int64      `json:"population"`
	MaxPopulation int64     `json:"max_population"`
	OwnerID      *int       `json:"owner_id"`
	IsHomeworld  bool       `json:"is_homeworld"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type UniverseStats struct {
	GalaxyCount         int `json:"galaxy_count"`
	TotalSectors        int `json:"total_sectors"`
	TotalSystems        int `json:"total_systems"`
	TotalPlanets        int `json:"total_planets"`
	InhabitedPlanets    int `json:"inhabited_planets"`
	HomeworldCount      int `json:"homeworld_count"`
	AvgPlanetsPerSystem float64 `json:"avg_planets_per_system"`
}
