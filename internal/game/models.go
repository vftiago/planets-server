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
	UniverseID        *int       `json:"universe_id"`
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
	UniverseID        int    `json:"universe_id"`
	MaxPlayers        int    `json:"max_players"`
	TurnIntervalHours int    `json:"turn_interval_hours"`
}

type GameStats struct {
	ID          int        `json:"id"`
	UniverseID  *int       `json:"universe_id"`
	Name        string     `json:"name"`
	Status      GameStatus `json:"status"`
	CurrentTurn int        `json:"current_turn"`
	PlayerCount int        `json:"player_count"`
	MaxPlayers  int        `json:"max_players"`
	NextTurnAt  *time.Time `json:"next_turn_at"`
}
