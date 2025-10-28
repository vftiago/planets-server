package game

import (
	"context"
	"fmt"

	"planets-server/internal/planet"
	"planets-server/internal/shared/database"
	"planets-server/internal/shared/errors"
	"planets-server/internal/spatial"
)

type Service struct {
	gameRepo       *Repository
	spatialService *spatial.Service
	planetService  *planet.Service
}

func NewService(
	gameRepo *Repository,
	spatialService *spatial.Service,
	planetService *planet.Service,
) *Service {
	return &Service{
		gameRepo:       gameRepo,
		spatialService: spatialService,
		planetService:  planetService,
	}
}

// CreateGame creates a new game with an integrated universe
func (s *Service) CreateGame(ctx context.Context, config GameConfig, universeConfig UniverseConfig) (*Game, error) {
	if err := s.deleteExistingGames(ctx); err != nil {
		return nil, fmt.Errorf("failed to delete existing games: %w", err)
	}

	// BEGIN TRANSACTION
	tx, err := s.gameRepo.db.BeginTxContext(ctx)
	if err != nil {
		return nil, errors.WrapInternal("failed to begin transaction for game creation", err)
	}

	// Ensure rollback on any error
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Create the game within transaction
	game, err := s.gameRepo.CreateGame(ctx, config, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	// Generate the universe content within transaction
	err = s.generateUniverse(ctx, game.ID, universeConfig, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate universe: %w", err)
	}

	// Activate the game within transaction
	if err := s.gameRepo.ActivateGame(ctx, game.ID, tx); err != nil {
		return nil, fmt.Errorf("failed to activate game: %w", err)
	}

	// COMMIT TRANSACTION
	if err := tx.Commit(); err != nil {
		return nil, errors.WrapInternal("failed to commit game creation transaction", err)
	}

	// Reload the game with updated counts and status (no transaction needed)
	updatedGame, err := s.gameRepo.GetGameByID(ctx, game.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload game after creation: %w", err)
	}

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
	return s.gameRepo.DeleteGame(ctx, gameID)
}

// deleteExistingGames deletes all existing games to maintain single game design
func (s *Service) deleteExistingGames(ctx context.Context) error {
	existingGames, err := s.gameRepo.GetAllGames(ctx)
	if err != nil {
		return fmt.Errorf("failed to get existing games: %w", err)
	}

	if len(existingGames) == 0 {
		return nil
	}

	for _, game := range existingGames {
		// Check for context cancellation
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("game deletion cancelled: %w", err)
		}

		if err := s.gameRepo.DeleteGame(ctx, game.ID); err != nil {
			return fmt.Errorf("failed to delete existing game %d: %w", game.ID, err)
		}
	}

	return nil
}

// generateUniverse orchestrates the generation of universe content for the game
func (s *Service) generateUniverse(ctx context.Context, gameID int, config UniverseConfig, tx *database.Tx) error {
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

	return nil
}

// generateSectorsForGalaxies generates all sectors for multiple galaxies in a single batch operation
func (s *Service) generateSectorsForGalaxies(ctx context.Context, gameID int, galaxies []spatial.SpatialEntity, sectorsPerGalaxy int, tx *database.Tx) ([]spatial.SpatialEntity, error) {
	if len(galaxies) == 0 {
		return []spatial.SpatialEntity{}, nil
	}

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

	// Generate ALL systems for ALL sectors in a single batch insert
	systems, err := s.spatialService.GenerateEntitiesForMultipleParents(ctx, gameID, sectors, spatial.EntityTypeSystem, systemsPerSector, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to batch generate systems: %w", err)
	}

	return systems, nil
}
