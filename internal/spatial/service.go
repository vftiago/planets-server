package spatial

import (
	"context"
	"math"
	"planets-server/internal/shared/database"
	"planets-server/internal/shared/errors"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// GenerateEntities generates entities for one or more parent entities in a single batch operation
// Returns only the IDs of created entities to minimize memory usage
func (s *Service) GenerateEntities(ctx context.Context, gameID int, parentIDs []*int, entityType EntityType, countPerParent int, tx *database.Tx) ([]int, error) {
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
			return nil, errors.WrapInternal("spatial entity generation cancelled", err)
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
		return nil, errors.WrapInternal("failed to batch create spatial entities", err)
	}

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
