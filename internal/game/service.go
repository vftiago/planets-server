package game

import (
	"context"
	"fmt"
	"log/slog"

	"planets-server/internal/planet"
	"planets-server/internal/shared/database"
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
func (s *Service) CreateGame(ctx context.Context, config GameConfig, universeConfig UniverseConfig) (*Game, error) {
	logger := s.logger.With("component", "game_service", "operation", "create_game", "name", config.Name)
	logger.Info("Creating new game with transaction")

	if err := s.deleteExistingGames(ctx); err != nil {
		logger.Error("Failed to delete existing games", "error", err)
		return nil, fmt.Errorf("failed to delete existing games: %w", err)
	}

	// BEGIN TRANSACTION
	tx, err := s.gameRepo.db.BeginTxContext(ctx)
	if err != nil {
		logger.Error("Failed to begin transaction", "error", err)
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure rollback on any error
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				logger.Error("Failed to rollback transaction", "error", rbErr)
			}
		}
	}()

	// Create the game within transaction
	game, err := s.gameRepo.CreateGame(ctx, config, tx)
	if err != nil {
		logger.Error("Failed to create game", "error", err)
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	logger.Info("Game created successfully", "game_id", game.ID)

	// Generate the universe content within transaction
	err = s.generateUniverse(ctx, game.ID, universeConfig, tx)
	if err != nil {
		logger.Error("Failed to generate universe", "game_id", game.ID, "error", err)
		return nil, fmt.Errorf("failed to generate universe: %w", err)
	}

	// Activate the game within transaction
	if err := s.gameRepo.ActivateGame(ctx, game.ID, tx); err != nil {
		logger.Error("Failed to activate game", "error", err)
		return nil, fmt.Errorf("failed to activate game: %w", err)
	}

	// COMMIT TRANSACTION
	if err := tx.Commit(); err != nil {
		logger.Error("Failed to commit transaction", "error", err)
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	logger.Info("Transaction committed successfully")

	// Reload the game with updated counts and status (no transaction needed)
	updatedGame, err := s.gameRepo.GetGameByID(ctx, game.ID)
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
func (s *Service) GetAllGames(ctx context.Context) ([]Game, error) {
	return s.gameRepo.GetAllGames(ctx)
}

// GetGameStats retrieves game statistics
func (s *Service) GetGameStats(ctx context.Context, gameID int) (*GameStats, error) {
	return s.gameRepo.GetGameStats(ctx, gameID)
}

// DeleteGame deletes a game and all related data
func (s *Service) DeleteGame(ctx context.Context, gameID int) error {
	logger := s.logger.With("component", "game_service", "operation", "delete_game", "game_id", gameID)
	logger.Info("Deleting game and all related data")

	return s.gameRepo.DeleteGame(ctx, gameID)
}

// deleteExistingGames deletes all existing games to maintain single game design
func (s *Service) deleteExistingGames(ctx context.Context) error {
	logger := s.logger.With("component", "game_service", "operation", "delete_existing_games")
	logger.Info("Deleting all existing games to maintain single universe")

	existingGames, err := s.gameRepo.GetAllGames(ctx)
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
		// Check for context cancellation
		if err := ctx.Err(); err != nil {
			logger.Warn("Context cancelled during game deletion", "error", err)
			return fmt.Errorf("game deletion cancelled: %w", err)
		}

		if err := s.gameRepo.DeleteGame(ctx, game.ID); err != nil {
			logger.Error("Failed to delete existing game", "error", err, "game_id", game.ID)
			return fmt.Errorf("failed to delete existing game %d: %w", game.ID, err)
		}
		logger.Debug("Deleted existing game", "game_id", game.ID, "name", game.Name)
	}

	logger.Info("Successfully deleted all existing games", "deleted_count", len(existingGames))
	return nil
}

// generateUniverse orchestrates the generation of universe content for the game
func (s *Service) generateUniverse(ctx context.Context, gameID int, config UniverseConfig, tx *database.Tx) error {
	s.logger.Info("Starting universe generation", "game_id", gameID)

	// Level 1: Generate all galaxies in one batch
	galaxies, err := s.spatialService.GenerateEntities(ctx, gameID, gameID, spatial.EntityTypeGalaxy, config.GalaxyCount, tx)
	if err != nil {
		return fmt.Errorf("failed to generate galaxies: %w", err)
	}

	// Level 2: Generate all sectors for all galaxies in one batch
	sectors, err := s.generateSectorsForGalaxies(ctx, gameID, galaxies, config.SectorsPerGalaxy, tx)
	if err != nil {
		return fmt.Errorf("failed to generate sectors: %w", err)
	}

	// Level 3: Generate all systems for all sectors in one batch
	systems, err := s.generateSystemsForSectors(ctx, gameID, sectors, config.SystemsPerSector, tx)
	if err != nil {
		return fmt.Errorf("failed to generate systems: %w", err)
	}

	// Level 4: Generate all planets for all systems in one batch
	systemIDs := make([]int, len(systems))
	for i, system := range systems {
		systemIDs[i] = system.ID
	}

	totalPlanets, err := s.planetService.GeneratePlanetsForSystems(ctx, systemIDs, config.MinPlanetsPerSystem, config.MaxPlanetsPerSystem, tx)
	if err != nil {
		return fmt.Errorf("failed to generate planets: %w", err)
	}

	// Update the game with final counts
	err = s.gameRepo.UpdateGameCounts(ctx, gameID, totalPlanets, tx)
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

// generateSectorsForGalaxies generates all sectors for multiple galaxies in a single batch operation
func (s *Service) generateSectorsForGalaxies(ctx context.Context, gameID int, galaxies []spatial.SpatialEntity, sectorsPerGalaxy int, tx *database.Tx) ([]spatial.SpatialEntity, error) {
	if len(galaxies) == 0 {
		return []spatial.SpatialEntity{}, nil
	}

	s.logger.Info("Generating sectors for all galaxies in single batch", "galaxy_count", len(galaxies), "sectors_per_galaxy", sectorsPerGalaxy)

	// Generate ALL sectors for ALL galaxies in a single batch insert
	sectors, err := s.spatialService.GenerateEntitiesForMultipleParents(ctx, gameID, galaxies, spatial.EntityTypeSector, sectorsPerGalaxy, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to batch generate sectors: %w", err)
	}

	return sectors, nil
}

// generateSystemsForSectors generates all systems for multiple sectors in a single batch operation
func (s *Service) generateSystemsForSectors(ctx context.Context, gameID int, sectors []spatial.SpatialEntity, systemsPerSector int, tx *database.Tx) ([]spatial.SpatialEntity, error) {
	if len(sectors) == 0 {
		return []spatial.SpatialEntity{}, nil
	}

	s.logger.Info("Generating systems for all sectors in single batch", "sector_count", len(sectors), "systems_per_sector", systemsPerSector)

	// Generate ALL systems for ALL sectors in a single batch insert
	systems, err := s.spatialService.GenerateEntitiesForMultipleParents(ctx, gameID, sectors, spatial.EntityTypeSystem, systemsPerSector, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to batch generate systems: %w", err)
	}

	return systems, nil
}
