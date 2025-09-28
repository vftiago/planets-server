package game

import (
	"fmt"
	"log/slog"

	"planets-server/internal/planet"
	"planets-server/internal/spatial"
)

type Service struct {
	gameRepo       *Repository
	spatialService *spatial.Service
	planetService  *planet.Service
	logger         *slog.Logger
}

func NewService(
	gameRepo *Repository,
	spatialService *spatial.Service,
	planetService *planet.Service,
	logger *slog.Logger,
) *Service {
	return &Service{
		gameRepo:       gameRepo,
		spatialService: spatialService,
		planetService:  planetService,
		logger:         logger,
	}
}

// CreateGame creates a new game with an integrated universe
func (s *Service) CreateGame(config GameConfig, universeConfig UniverseConfig) (*Game, error) {
	logger := s.logger.With("component", "game_service", "operation", "create_game", "name", config.Name)
	logger.Info("Creating new game")

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

	// Generate the universe content
	err = s.generateUniverse(game.ID, universeConfig)
	if err != nil {
		logger.Error("Failed to generate universe", "game_id", game.ID, "error", err)
		// Clean up the game if generation failed
		s.gameRepo.DeleteGame(game.ID)
		return nil, fmt.Errorf("failed to generate universe: %w", err)
	}

	// Activate the game
	if err := s.gameRepo.ActivateGame(game.ID); err != nil {
		logger.Error("Failed to activate game", "error", err)
		return nil, fmt.Errorf("failed to activate game: %w", err)
	}

	// Reload the game with updated counts and status
	updatedGame, err := s.gameRepo.GetGameByID(game.ID)
	if err != nil {
		logger.Error("Failed to reload game", "error", err)
		return nil, fmt.Errorf("failed to reload game: %w", err)
	}

	logger.Info("Game created and activated successfully",
		"game_id", updatedGame.ID,
		"name", updatedGame.Name,
		"planets", updatedGame.PlanetCount)

	return updatedGame, nil
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

// generateUniverse orchestrates the generation of universe content for the game
func (s *Service) generateUniverse(gameID int, config UniverseConfig) error {
	s.logger.Info("Starting universe generation", "game_id", gameID)

	// Level 1: Generate all galaxies (parent_id = game_id)
	galaxies, err := s.spatialService.GenerateEntities(gameID, gameID, spatial.EntityTypeGalaxy, config.GalaxyCount)
	if err != nil {
		return fmt.Errorf("failed to generate galaxies: %w", err)
	}

	// Level 2: Generate all sectors for all galaxies
	var sectors []spatial.SpatialEntity
	for _, galaxy := range galaxies {
		galaxySectors, err := s.spatialService.GenerateEntities(gameID, galaxy.ID, spatial.EntityTypeSector, config.SectorsPerGalaxy)
		if err != nil {
			return fmt.Errorf("failed to generate sectors for galaxy %d: %w", galaxy.ID, err)
		}
		sectors = append(sectors, galaxySectors...)
	}

	// Level 3: Generate all systems for all sectors
	var systems []spatial.SpatialEntity
	for _, sector := range sectors {
		sectorSystems, err := s.spatialService.GenerateEntities(gameID, sector.ID, spatial.EntityTypeSystem, config.SystemsPerSector)
		if err != nil {
			return fmt.Errorf("failed to generate systems for sector %d: %w", sector.ID, err)
		}
		systems = append(systems, sectorSystems...)
	}

	// Level 4: Generate all planets for all systems
	var totalPlanets int
	for _, system := range systems {
		planetCount, err := s.planetService.GeneratePlanets(system.ID, config.MinPlanetsPerSystem, config.MaxPlanetsPerSystem)
		if err != nil {
			return fmt.Errorf("failed to generate planets for system %d: %w", system.ID, err)
		}
		totalPlanets += planetCount
	}

	// Update the game with final counts
	err = s.gameRepo.UpdateGameCounts(gameID, totalPlanets)
	if err != nil {
		return fmt.Errorf("failed to update game counts: %w", err)
	}

	s.logger.Info("Universe generation completed",
		"game_id", gameID,
		"galaxies", len(galaxies),
		"sectors", len(sectors),
		"systems", len(systems),
		"planets", totalPlanets)

	return nil
}
