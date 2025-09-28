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

func (r *Repository) GetEntitiesByParent(parentID *int, entityType EntityType) ([]SpatialEntity, error) {
	logger := r.logger.With(
		"component", "spatial_repository",
		"operation", "get_entities_by_parent",
		"entity_type", entityType,
	)
	logger.Debug("Getting entities by parent")

	var query string
	var args []interface{}

	if parentID == nil {
		query = `
			SELECT id, parent_id, entity_type, level, x_coord, y_coord, 
				   name, description, child_count, created_at, updated_at
			FROM spatial_entities 
			WHERE parent_id IS NULL AND entity_type = $1
			ORDER BY x_coord, y_coord`
		args = []interface{}{entityType}
	} else {
		query = `
			SELECT id, parent_id, entity_type, level, x_coord, y_coord, 
				   name, description, child_count, created_at, updated_at
			FROM spatial_entities 
			WHERE parent_id = $1 AND entity_type = $2
			ORDER BY x_coord, y_coord`
		args = []interface{}{parentID, entityType}
		logger = logger.With("parent_id", *parentID)
	}

	rows, err := r.db.Query(query, args...)
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
		var parentIDVal sql.NullInt64

		err := rows.Scan(
			&entity.ID,
			&entity.GameID,
			&parentIDVal,
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

		if parentIDVal.Valid {
			parentID := int(parentIDVal.Int64)
			entity.ParentID = &parentID
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

func (r *Repository) CreateEntity(parentID *int, entityType EntityType, level int, x, y int, name, description string) (*SpatialEntity, error) {
	logger := r.logger.With(
		"component", "spatial_repository",
		"operation", "create_entity",
		"entity_type", entityType,
		"level", level,
		"coordinates", fmt.Sprintf("(%d,%d)", x, y),
	)
	logger.Debug("Creating spatial entity")

	query := `
		INSERT INTO spatial_entities (parent_id, entity_type, level, x_coord, y_coord, name, description, child_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 0)
		RETURNING id, parent_id, entity_type, level, x_coord, y_coord, name, description, child_count, created_at, updated_at`

	var entity SpatialEntity
	var descriptionVal sql.NullString
	var parentIDVal sql.NullInt64

	err := r.db.QueryRow(query, parentID, entityType, level, x, y, name, description).Scan(
		&entity.ID,
		&parentIDVal,
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

	if parentIDVal.Valid {
		parentID := int(parentIDVal.Int64)
		entity.ParentID = &parentID
	}
	if descriptionVal.Valid {
		entity.Description = descriptionVal.String
	}

	logger.Debug("Spatial entity created successfully", "entity_id", entity.ID)
	return &entity, nil
}
