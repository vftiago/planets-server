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
	ID                int        `json:"id"`
	Name              string     `json:"name"`
	Description       string     `json:"description"`
	Status            GameStatus `json:"status"`
	CurrentTurn       int        `json:"current_turn"`
	MaxPlayers        int        `json:"max_players"`
	TurnIntervalHours int        `json:"turn_interval_hours"`
	NextTurnAt        *time.Time `json:"next_turn_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type GameConfig struct {
	Name              string `json:"name"`
	Description       string `json:"description"`
	MaxPlayers        int    `json:"max_players"`
	TurnIntervalHours int    `json:"turn_interval_hours"`
	UniverseConfig    UniverseConfig `json:"universe_config"`
}

type UniverseConfig struct {
	GalaxyName          string `json:"galaxy_name"`
	SectorCount         int    `json:"sector_count"`
	SystemsPerSector    int    `json:"systems_per_sector"`
	MinPlanetsPerSystem int    `json:"min_planets_per_system"`
	MaxPlanetsPerSystem int    `json:"max_planets_per_system"`
}

type GameStats struct {
	ID              int     `json:"id"`
	Name            string  `json:"name"`
	Status          GameStatus `json:"status"`
	CurrentTurn     int     `json:"current_turn"`
	PlayerCount     int     `json:"player_count"`
	MaxPlayers      int     `json:"max_players"`
	SectorCount     int     `json:"sector_count"`
	SystemCount     int     `json:"system_count"`
	PlanetCount     int     `json:"planet_count"`
	InhabitedPlanets int    `json:"inhabited_planets"`
	NextTurnAt      *time.Time `json:"next_turn_at"`
}

type GenerationProgress struct {
	GameID          int       `json:"game_id"`
	Stage           string    `json:"stage"`
	Progress        float64   `json:"progress"`
	SectorsCreated  int       `json:"sectors_created"`
	SystemsCreated  int       `json:"systems_created"`
	PlanetsCreated  int       `json:"planets_created"`
	StartedAt       time.Time `json:"started_at"`
	EstimatedEnd    *time.Time `json:"estimated_end"`
}
