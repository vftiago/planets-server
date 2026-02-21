package game

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"hash/fnv"
	mathrand "math/rand"

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

func (s *Service) CreateGame(ctx context.Context, config GameConfig) (*Game, error) {
	tx, err := s.gameRepo.db.BeginTx(ctx)
	if err != nil {
		return nil, errors.WrapInternal("failed to begin transaction for game creation", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	name, err := generateGameName()
	if err != nil {
		return nil, errors.WrapInternal("failed to generate game name", err)
	}

	seed := config.Seed
	if seed == "" {
		var err error
		seed, err = generateSeed()
		if err != nil {
			return nil, errors.WrapInternal("failed to generate seed", err)
		}
	} else if len(seed) < 3 || len(seed) > 32 {
		return nil, errors.Validation("seed must be between 3 and 32 characters")
	}

	seedInt := hashSeed(seed)

	game, err := s.gameRepo.CreateGame(ctx, name, seed, config, tx)
	if err != nil {
		return nil, errors.WrapInternal("failed to create game", err)
	}

	rng := mathrand.New(mathrand.NewSource(seedInt))

	err = s.generateUniverse(ctx, game.ID, config, rng, tx)
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

func (s *Service) DeleteGame(ctx context.Context, gameID int) error {
	return s.gameRepo.DeleteGame(ctx, gameID)
}

func generateGameName() (string, error) {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func generateSeed() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func hashSeed(seed string) int64 {
	h := fnv.New64a()
	h.Write([]byte(seed))
	return int64(h.Sum64())
}

func (s *Service) generateUniverse(ctx context.Context, gameID int, config GameConfig, rng *mathrand.Rand, tx *database.Tx) error {
	// Create the universe entity (level 0, root of spatial hierarchy)
	universeIDs, err := s.spatialService.GenerateEntities(
		ctx,
		gameID,
		[]*int{nil},
		spatial.EntityTypeUniverse,
		1,
		tx,
	)
	if err != nil {
		return errors.WrapInternal("failed to create universe entity", err)
	}

	universeID := universeIDs[0]

	if err := s.gameRepo.SetUniverseID(ctx, gameID, universeID, tx); err != nil {
		return errors.WrapInternal("failed to link universe to game", err)
	}

	// Generate spatial hierarchy: galaxies → sectors → systems
	plan := config.BuildGenerationPlan()
	currentLevelIDs := universeIDs

	for _, level := range plan {
		if err := ctx.Err(); err != nil {
			return errors.WrapInternal("universe generation cancelled", err)
		}

		parentIDs := make([]*int, len(currentLevelIDs))
		for j, id := range currentLevelIDs {
			idCopy := id
			parentIDs[j] = &idCopy
		}

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
		rng,
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
