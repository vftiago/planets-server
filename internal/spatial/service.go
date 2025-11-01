package spatial

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"planets-server/internal/shared/database"
)

type Service struct {
	repo   *Repository
	logger *slog.Logger
}

func NewService(repo *Repository, logger *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// GenerateEntities generates entities for one or more parent entities in a single batch operation
// Returns only the IDs of created entities to minimize memory usage
func (s *Service) GenerateEntities(ctx context.Context, gameID int, parentIDs []*int, entityType EntityType, countPerParent int, tx *database.Tx) ([]int, error) {
	logger := s.logger.With(
		"operation", "generate_entities",
		"type", entityType,
		"parent_count", len(parentIDs),
		"count_per_parent", countPerParent,
		"game_id", gameID,
	)
	logger.Debug("Generating spatial entities")

	if len(parentIDs) == 0 {
		return []int{}, nil
	}

	entitiesPerSide := int(math.Sqrt(float64(countPerParent)))
	if entitiesPerSide*entitiesPerSide != countPerParent {
		entitiesPerSide = int(math.Ceil(math.Sqrt(float64(countPerParent))))
	}

	names := s.generateNames(entityType)
	level := EntityLevels[entityType]

	// Prepare all entities for all parents upfront for batch insert
	var batchRequests []BatchInsertRequest

	for _, parentID := range parentIDs {
		// Check for context cancellation
		if err := ctx.Err(); err != nil {
			logger.Warn("Context cancelled during entity generation", "error", err)
			return nil, fmt.Errorf("entity generation cancelled: %w", err)
		}

		nameIndex := 0
		entityCount := 0

		for x := 0; x < entitiesPerSide; x++ {
			for y := 0; y < entitiesPerSide; y++ {
				if entityCount >= countPerParent {
					break
				}

				name := names[nameIndex%len(names)]
				nameIndex++

				batchRequests = append(batchRequests, BatchInsertRequest{
					GameID:      gameID,
					ParentID:    parentID,
					EntityType:  entityType,
					Level:       level,
					XCoord:      x,
					YCoord:      y,
					Name:        name,
					Description: "",
				})

				entityCount++
			}
			if entityCount >= countPerParent {
				break
			}
		}
	}

	// Perform single batch insert for all entities across all parents
	entityIDs, err := s.repo.CreateEntitiesBatch(ctx, batchRequests, tx)
	if err != nil {
		logger.Error("Failed to batch create entities", "error", err)
		return nil, fmt.Errorf("failed to batch create %s: %w", entityType, err)
	}

	logger.Info("Entities generated",
		"total_count", len(entityIDs),
		"parent_count", len(parentIDs),
		"avg_per_parent", len(entityIDs)/len(parentIDs))
	return entityIDs, nil
}

func (s *Service) generateNames(entityType EntityType) []string {
	switch entityType {
	case EntityTypeGalaxy:
		return []string{"Andromeda", "Milky Way", "Centaurus", "Pegasus", "Cygnus", "Draco"}
	case EntityTypeSector:
		return []string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon", "Zeta", "Eta", "Theta"}
	case EntityTypeSystem:
		return []string{"Altair", "Vega", "Sirius", "Arcturus", "Capella", "Rigel", "Procyon"}
	default:
		return []string{"Entity-1", "Entity-2", "Entity-3"}
	}
}
