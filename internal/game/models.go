package game

import (
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
	ID                  int        `json:"id"`
	Name                string     `json:"name"`
	Description         string     `json:"description"`
	GalaxyCount         int        `json:"galaxy_count"`
	SectorCount         int        `json:"sector_count"`
	SystemCount         int        `json:"system_count"`
	PlanetCount         int        `json:"planet_count"`
	Status              GameStatus `json:"status"`
	CurrentTurn         int        `json:"current_turn"`
	MaxPlayers          int        `json:"max_players"`
	TurnIntervalHours   int        `json:"turn_interval_hours"`
	NextTurnAt          *time.Time `json:"next_turn_at"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type GameConfig struct {
	Name                string `json:"name"`
	Description         string `json:"description"`
	MaxPlayers          int    `json:"max_players"`
	TurnIntervalHours   int    `json:"turn_interval_hours"`
}

type GameStats struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	Status      GameStatus `json:"status"`
	CurrentTurn int        `json:"current_turn"`
	PlayerCount int        `json:"player_count"`
	MaxPlayers  int        `json:"max_players"`
	NextTurnAt  *time.Time `json:"next_turn_at"`
	GalaxyCount int        `json:"galaxy_count"`
	SectorCount int        `json:"sector_count"`
	SystemCount int        `json:"system_count"`
	PlanetCount int        `json:"planet_count"`
}

type UniverseConfig struct {
	GalaxyCount         int `json:"galaxy_count"`
	SectorsPerGalaxy    int `json:"sectors_per_galaxy"`
	SystemsPerSector    int `json:"systems_per_sector"`
	MinPlanetsPerSystem int `json:"min_planets_per_system"`
	MaxPlanetsPerSystem int `json:"max_planets_per_system"`
}
