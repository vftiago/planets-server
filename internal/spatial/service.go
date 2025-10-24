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
	var entities []SpatialEntity
	
	for x := 0; x < entitiesPerSide; x++ {
		for y := 0; y < entitiesPerSide; y++ {
			if len(entities) >= count {
				break
			}
			
			name := names[nameIndex%len(names)]
			nameIndex++
			
			entity, err := s.CreateEntity(gameID, parentID, entityType, x, y, name, tx)
			if err != nil {
				logger.Error("Failed to create entity", "error", err, "coordinates", fmt.Sprintf("(%d,%d)", x, y))
				return nil, fmt.Errorf("failed to create %s at (%d,%d): %w", entityType, x, y, err)
			}
			entities = append(entities, *entity)
		}
		if len(entities) >= count {
			break
		}
	}
	
	logger.Info("Entities generated", "count", len(entities))
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
