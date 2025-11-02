package game

import (
	"context"

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

func (s *Service) CreateGame(ctx context.Context, config GameConfig, universeConfig UniverseConfig) (*Game, error) {
	// Development constraint: only one game allowed at a time
	// TODO: Remove this when implementing multi-game support
	if err := s.gameRepo.DeleteAllGames(ctx); err != nil {
		return nil, errors.WrapInternal("failed to delete existing games", err)
	}

	tx, err := s.gameRepo.db.BeginTxContext(ctx)
	if err != nil {
		return nil, errors.WrapInternal("failed to begin transaction for game creation", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	game, err := s.gameRepo.CreateGame(ctx, config, tx)
	if err != nil {
		return nil, errors.WrapInternal("failed to create game", err)
	}

	err = s.generateUniverse(ctx, game.ID, universeConfig, tx)
	if err != nil {
		return nil, errors.WrapInternal("failed to generate universe", err)
	}

	if err := s.gameRepo.ActivateGame(ctx, game.ID, tx); err != nil {
		return nil, errors.WrapInternal("failed to activate game", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.WrapInternal("failed to commit game creation transaction", err)
	}

	updatedGame, err := s.gameRepo.GetGameByID(ctx, game.ID)
	if err != nil {
		return nil, errors.WrapInternal("failed to reload game after creation", err)
	}

	return updatedGame, nil
}

func (s *Service) GetAllGames(ctx context.Context) ([]Game, error) {
	return s.gameRepo.GetAllGames(ctx)
}

func (s *Service) GetGameStats(ctx context.Context, gameID int) (*GameStats, error) {
	return s.gameRepo.GetGameStats(ctx, gameID)
}

func (s *Service) generateUniverse(ctx context.Context, gameID int, config UniverseConfig, tx *database.Tx) error {
	plan := config.BuildGenerationPlan()

	var currentLevelIDs []int

	for i, level := range plan {
		if err := ctx.Err(); err != nil {
			return errors.WrapInternal("universe generation cancelled", err)
		}

		var parentIDs []*int
		if i == 0 {
			// For first level parent_id is NULL
			parentIDs = []*int{nil}
		} else {
			// For subsequent levels, use IDs from previous level as parents
			parentIDs = make([]*int, len(currentLevelIDs))
			for j, id := range currentLevelIDs {
				idCopy := id
				parentIDs[j] = &idCopy
			}
		}

		var err error
		currentLevelIDs, err = s.spatialService.GenerateEntities(
			ctx,
			gameID,
			parentIDs,
			level.EntityType,
			level.Count,
			tx,
		)
		if err != nil {
			return errors.WrapInternal("failed to generate spatial entities", err)
		}
	}

	// Final level IDs are system IDs for planet generation
	systemIDs := currentLevelIDs

	totalPlanets, err := s.planetService.GeneratePlanets(
		ctx,
		systemIDs,
		config.MinPlanetsPerSystem,
		config.MaxPlanetsPerSystem,
		tx,
	)
	if err != nil {
		return errors.WrapInternal("failed to generate planets", err)
	}

	err = s.gameRepo.UpdateGameCounts(ctx, gameID, totalPlanets, tx)
	if err != nil {
		return errors.WrapInternal("failed to update game counts", err)
	}

	return nil
}
