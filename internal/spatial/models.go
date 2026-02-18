package spatial

import (
	"time"
)

type EntityType string

const (
	EntityTypeUniverse EntityType = "universe"
	EntityTypeGalaxy   EntityType = "galaxy"
	EntityTypeSector   EntityType = "sector"
	EntityTypeSystem   EntityType = "system"
)

var EntityLevels = map[EntityType]int{
	EntityTypeUniverse: 0,
	EntityTypeGalaxy:   1,
	EntityTypeSector:   2,
	EntityTypeSystem:   3,
}

type SpatialEntity struct {
	ID          int        `json:"id"`
	GameID      int        `json:"game_id"`
	ParentID    *int       `json:"parent_id"`
	EntityType  EntityType `json:"entity_type"`
	Level       int        `json:"level"`
	XCoord      int        `json:"x_coord"`
	YCoord      int        `json:"y_coord"`
	Name       string     `json:"name"`
	ChildCount int        `json:"child_count"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Convenience type aliases
type Universe = SpatialEntity
type Galaxy = SpatialEntity
type Sector = SpatialEntity
type System = SpatialEntity
