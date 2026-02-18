package game

import (
	"planets-server/internal/spatial"
	"time"
)

type GameStatus string

const (
	GameStatusCreating  GameStatus = "creating"
	GameStatusActive    GameStatus = "active"
	GameStatusPaused    GameStatus = "paused"
	GameStatusCompleted GameStatus = "completed"
)

type Game struct {
	ID                int        `json:"id"`
	Name              string     `json:"name"`
	Seed              string     `json:"seed"`
	UniverseID        *int       `json:"universe_id"`
	PlanetCount       int        `json:"planet_count"`
	Status            GameStatus `json:"status"`
	CurrentTurn       int        `json:"current_turn"`
	MaxPlayers        int        `json:"max_players"`
	TurnIntervalHours int        `json:"turn_interval_hours"`
	NextTurnAt        *time.Time `json:"next_turn_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type GameConfig struct {
	Seed                string `json:"seed,omitempty"`
	MaxPlayers          int    `json:"max_players"`
	TurnIntervalHours   int    `json:"turn_interval_hours"`
	GalaxyCount         int `json:"galaxy_count"`
	SectorsPerGalaxy    int `json:"sectors_per_galaxy"`
	SystemsPerSector    int `json:"systems_per_sector"`
	MinPlanetsPerSystem int `json:"min_planets_per_system"`
	MaxPlanetsPerSystem int `json:"max_planets_per_system"`
}

type GameStats struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	Status      GameStatus `json:"status"`
	CurrentTurn int        `json:"current_turn"`
	PlayerCount int        `json:"player_count"`
	MaxPlayers  int        `json:"max_players"`
	NextTurnAt  *time.Time `json:"next_turn_at"`
	PlanetCount int        `json:"planet_count"`
}

type SpatialLevel struct {
	EntityType spatial.EntityType
	Count      int
}

func (c GameConfig) BuildGenerationPlan() []SpatialLevel {
	return []SpatialLevel{
		{EntityType: spatial.EntityTypeGalaxy, Count: c.GalaxyCount},
		{EntityType: spatial.EntityTypeSector, Count: c.SectorsPerGalaxy},
		{EntityType: spatial.EntityTypeSystem, Count: c.SystemsPerSector},
	}
}
