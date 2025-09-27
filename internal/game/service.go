package game

import (
	"fmt"
	"log/slog"
	"time"

	"planets-server/internal/universe"
)

type Service struct {
	gameRepo     *Repository
	universeRepo *universe.Repository
	logger       *slog.Logger
}

func NewService(
	gameRepo *Repository,
	universeRepo *universe.Repository,
	logger *slog.Logger,
) *Service {
	return &Service{
		gameRepo:     gameRepo,
		universeRepo: universeRepo,
		logger:       logger,
	}
}

// CreateGame creates a new game that references an existing universe
func (s *Service) CreateGame(config GameConfig) (*Game, error) {
	logger := s.logger.With("component", "game_service", "operation", "create_game",
		"name", config.Name, "universe_id", config.UniverseID)
	logger.Info("Creating new game")

	// Verify the universe exists
	universe, err := s.universeRepo.GetUniverse(config.UniverseID)
	if err != nil {
		logger.Error("Failed to get universe", "error", err)
		return nil, fmt.Errorf("universe not found: %w", err)
	}

	logger.Info("Using existing universe", "universe_name", universe.Name,
		"sectors", universe.SectorCount, "systems", universe.SystemCount, "planets", universe.PlanetCount)

	// Delete existing games (single game design)
	if err := s.deleteExistingGames(); err != nil {
		logger.Error("Failed to delete existing games", "error", err)
		return nil, fmt.Errorf("failed to delete existing games: %w", err)
	}

	// Create the game
	game, err := s.gameRepo.CreateGame(config)
	if err != nil {
		logger.Error("Failed to create game", "error", err)
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	logger.Info("Game created successfully", "game_id", game.ID)

	// Activate the game
	if err := s.gameRepo.ActivateGame(game.ID); err != nil {
		logger.Error("Failed to activate game", "error", err)
		return nil, fmt.Errorf("failed to activate game: %w", err)
	}

	// Update the game object with activation details
	game.Status = GameStatusActive
	game.CurrentTurn = 1
	nextTurn := time.Now().Add(1 * time.Hour).Truncate(time.Hour)
	game.NextTurnAt = &nextTurn

	logger.Info("Game created and activated successfully",
		"game_id", game.ID,
		"name", game.Name,
		"universe_id", universe.ID,
		"next_turn_at", nextTurn)

	return game, nil
}

// GetAllGames retrieves all games
func (s *Service) GetAllGames() ([]Game, error) {
	return s.gameRepo.GetAllGames()
}

// GetGameStats retrieves game statistics
func (s *Service) GetGameStats(gameID int) (*GameStats, error) {
	return s.gameRepo.GetGameStats(gameID)
}

// DeleteGame deletes a game and all related data
func (s *Service) DeleteGame(gameID int) error {
	logger := s.logger.With("component", "game_service", "operation", "delete_game", "game_id", gameID)
	logger.Info("Deleting game and all related data")

	return s.gameRepo.DeleteGame(gameID)
}

// deleteExistingGames deletes all existing games to maintain single game design
func (s *Service) deleteExistingGames() error {
	logger := s.logger.With("component", "game_service", "operation", "delete_existing_games")
	logger.Info("Deleting all existing games to maintain single universe")

	existingGames, err := s.gameRepo.GetAllGames()
	if err != nil {
		logger.Error("Failed to get existing games", "error", err)
		return fmt.Errorf("failed to get existing games: %w", err)
	}

	if len(existingGames) == 0 {
		logger.Debug("No existing games to delete")
		return nil
	}

	logger.Info("Found existing games to delete", "count", len(existingGames))

	for _, game := range existingGames {
		if err := s.gameRepo.DeleteGame(game.ID); err != nil {
			logger.Error("Failed to delete existing game", "error", err, "game_id", game.ID)
			return fmt.Errorf("failed to delete existing game %d: %w", game.ID, err)
		}
		logger.Debug("Deleted existing game", "game_id", game.ID, "name", game.Name)
	}

	logger.Info("Successfully deleted all existing games", "deleted_count", len(existingGames))
	return nil
}
