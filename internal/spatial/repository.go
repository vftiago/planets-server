package spatial

import (
	"database/sql"
	"fmt"
	"log/slog"
)

type Repository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewRepository(db *sql.DB, logger *slog.Logger) *Repository {
	logger.Debug("Initializing spatial repository")
	return &Repository{
		db:     db,
		logger: logger,
	}
}

func (r *Repository) GetEntitiesByParent(parentID int, entityType EntityType) ([]SpatialEntity, error) {
	logger := r.logger.With(
		"component", "spatial_repository",
		"operation", "get_entities_by_parent",
		"entity_type", entityType,
		"parent_id", parentID,
	)
	logger.Debug("Getting entities by parent")

	query := `
		SELECT id, game_id, parent_id, entity_type, level, x_coord, y_coord, 
		       name, description, child_count, created_at, updated_at
		FROM spatial_entities 
		WHERE parent_id = $1 AND entity_type = $2
		ORDER BY x_coord, y_coord`

	rows, err := r.db.Query(query, parentID, entityType)
	if err != nil {
		logger.Error("Failed to query spatial entities", "error", err)
		return nil, fmt.Errorf("failed to query spatial entities: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Error("Failed to close rows", "error", err)
		}
	}()

	var entities []SpatialEntity
	for rows.Next() {
		var entity SpatialEntity
		var description sql.NullString

		err := rows.Scan(
			&entity.ID,
			&entity.GameID,
			&entity.ParentID,
			&entity.EntityType,
			&entity.Level,
			&entity.XCoord,
			&entity.YCoord,
			&entity.Name,
			&description,
			&entity.ChildCount,
			&entity.CreatedAt,
			&entity.UpdatedAt,
		)
		if err != nil {
			logger.Error("Failed to scan spatial entity row", "error", err)
			return nil, fmt.Errorf("failed to scan spatial entity: %w", err)
		}

		if description.Valid {
			entity.Description = description.String
		}

		entities = append(entities, entity)
	}

	if err := rows.Err(); err != nil {
		logger.Error("Error during rows iteration", "error", err)
		return nil, fmt.Errorf("error iterating spatial entities: %w", err)
	}

	logger.Debug("Spatial entities retrieved", "count", len(entities))
	return entities, nil
}

func (r *Repository) CreateEntity(gameID, parentID int, entityType EntityType, level int, x, y int, name, description string) (*SpatialEntity, error) {
	logger := r.logger.With(
		"component", "spatial_repository",
		"operation", "create_entity",
		"entity_type", entityType,
		"level", level,
		"game_id", gameID,
		"parent_id", parentID,
		"coordinates", fmt.Sprintf("(%d,%d)", x, y),
	)
	logger.Debug("Creating spatial entity")

	query := `
		INSERT INTO spatial_entities (game_id, parent_id, entity_type, level, x_coord, y_coord, name, description, child_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 0)
		RETURNING id, game_id, parent_id, entity_type, level, x_coord, y_coord, name, description, child_count, created_at, updated_at`

	var entity SpatialEntity
	var descriptionVal sql.NullString

	err := r.db.QueryRow(query, gameID, parentID, entityType, level, x, y, name, description).Scan(
		&entity.ID,
		&entity.GameID,
		&entity.ParentID,
		&entity.EntityType,
		&entity.Level,
		&entity.XCoord,
		&entity.YCoord,
		&entity.Name,
		&descriptionVal,
		&entity.ChildCount,
		&entity.CreatedAt,
		&entity.UpdatedAt,
	)

	if err != nil {
		logger.Error("Failed to create spatial entity", "error", err)
		return nil, fmt.Errorf("failed to create spatial entity: %w", err)
	}

	if descriptionVal.Valid {
		entity.Description = descriptionVal.String
	}

	logger.Debug("Spatial entity created successfully", "entity_id", entity.ID)
	return &entity, nil
}
