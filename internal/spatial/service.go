package spatial

import (
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

func (s *Service) GetEntitiesByParent(parentID int, entityType EntityType) ([]SpatialEntity, error) {
	return s.repo.GetEntitiesByParent(parentID, entityType)
}

func (s *Service) CreateEntity(gameID, parentID int, entityType EntityType, x, y int, name string, tx *database.Tx) (*SpatialEntity, error) {
	level := EntityLevels[entityType]
	return s.repo.CreateEntity(gameID, parentID, entityType, level, x, y, name, "", tx)
}

func (s *Service) GetGalaxiesByGame(gameID int) ([]SpatialEntity, error) {
	return s.GetEntitiesByParent(gameID, EntityTypeGalaxy)
}

func (s *Service) GetSectorsByGalaxy(galaxyID int) ([]SpatialEntity, error) {
	return s.GetEntitiesByParent(galaxyID, EntityTypeSector)
}

func (s *Service) GetSystemsBySector(sectorID int) ([]SpatialEntity, error) {
	return s.GetEntitiesByParent(sectorID, EntityTypeSystem)
}

func (s *Service) GenerateEntities(gameID, parentID int, entityType EntityType, count int, tx *database.Tx) ([]SpatialEntity, error) {
	logger := s.logger.With(
		"operation", "generate_entities",
		"type", entityType,
		"count", count,
		"game_id", gameID,
		"parent_id", parentID,
	)
	logger.Debug("Generating spatial entities")

	entitiesPerSide := int(math.Sqrt(float64(count)))
	if entitiesPerSide*entitiesPerSide != count {
		entitiesPerSide = int(math.Ceil(math.Sqrt(float64(count))))
	}

	names := s.generateNames(entityType)
	nameIndex := 0
	level := EntityLevels[entityType]

	// Prepare all entities upfront for batch insert
	var batchRequests []BatchInsertRequest

	for x := 0; x < entitiesPerSide; x++ {
		for y := 0; y < entitiesPerSide; y++ {
			if len(batchRequests) >= count {
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
		}
		if len(batchRequests) >= count {
			break
		}
	}

	// Perform batch insert
	entities, err := s.repo.CreateEntitiesBatch(batchRequests, tx)
	if err != nil {
		logger.Error("Failed to batch create entities", "error", err)
		return nil, fmt.Errorf("failed to batch create %s: %w", entityType, err)
	}

	logger.Info("Entities generated", "count", len(entities))
	return entities, nil
}

// GenerateEntitiesForMultipleParents generates entities for multiple parent entities in a single batch operation
func (s *Service) GenerateEntitiesForMultipleParents(gameID int, parents []SpatialEntity, entityType EntityType, countPerParent int, tx *database.Tx) ([]SpatialEntity, error) {
	logger := s.logger.With(
		"operation", "generate_entities_for_multiple_parents",
		"type", entityType,
		"parent_count", len(parents),
		"count_per_parent", countPerParent,
		"game_id", gameID,
	)
	logger.Debug("Generating spatial entities for multiple parents")

	if len(parents) == 0 {
		return []SpatialEntity{}, nil
	}

	entitiesPerSide := int(math.Sqrt(float64(countPerParent)))
	if entitiesPerSide*entitiesPerSide != countPerParent {
		entitiesPerSide = int(math.Ceil(math.Sqrt(float64(countPerParent))))
	}

	names := s.generateNames(entityType)
	level := EntityLevels[entityType]

	// Prepare all entities for all parents upfront for batch insert
	var batchRequests []BatchInsertRequest

	for _, parent := range parents {
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
					ParentID:    parent.ID,
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
	entities, err := s.repo.CreateEntitiesBatch(batchRequests, tx)
	if err != nil {
		logger.Error("Failed to batch create entities for multiple parents", "error", err)
		return nil, fmt.Errorf("failed to batch create %s for multiple parents: %w", entityType, err)
	}

	logger.Info("Entities generated for multiple parents",
		"total_count", len(entities),
		"parent_count", len(parents),
		"avg_per_parent", len(entities)/len(parents))
	return entities, nil
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
