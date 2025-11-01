package spatial

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"planets-server/internal/shared/database"
)

type Repository struct {
	db     *database.DB
	logger *slog.Logger
}

func NewRepository(db *database.DB, logger *slog.Logger) *Repository {
	logger.Debug("Initializing spatial repository")
	return &Repository{
		db:     db,
		logger: logger,
	}
}

func (r *Repository) getExecutor(tx *database.Tx) database.Executor {
	if tx != nil {
		return tx
	}
	return r.db
}

// BatchInsertRequest represents a single entity to be inserted in a batch
type BatchInsertRequest struct {
	GameID      int
	ParentID    *int
	EntityType  EntityType
	Level       int
	XCoord      int
	YCoord      int
	Name        string
	Description string
}

// CreateEntitiesBatch creates multiple spatial entities in a single database operation using JSON
func (r *Repository) CreateEntitiesBatch(ctx context.Context, entities []BatchInsertRequest, tx *database.Tx) ([]SpatialEntity, error) {
	if len(entities) == 0 {
		return []SpatialEntity{}, nil
	}

	exec := r.getExecutor(tx)

	logger := r.logger.With(
		"component", "spatial_repository",
		"operation", "create_entities_batch",
		"count", len(entities),
	)
	logger.Debug("Creating spatial entities in batch")

	// Convert entities to JSON
	entitiesJSON, err := json.Marshal(entities)
	if err != nil {
		logger.Error("Failed to marshal entities to JSON", "error", err)
		return nil, fmt.Errorf("failed to marshal entities: %w", err)
	}

	query := `
		INSERT INTO spatial_entities (game_id, parent_id, entity_type, level, x_coord, y_coord, name, description, child_count)
		SELECT
			(data->>'GameID')::integer,
			(data->>'ParentID')::integer,
			(data->>'EntityType')::entity_type,
			(data->>'Level')::integer,
			(data->>'XCoord')::integer,
			(data->>'YCoord')::integer,
			data->>'Name',
			data->>'Description',
			0
		FROM json_array_elements($1::json) AS data
		RETURNING id, game_id, parent_id, entity_type, level, x_coord, y_coord, name, description, child_count, created_at, updated_at`

	rows, err := exec.QueryContext(ctx, query, string(entitiesJSON))
	if err != nil {
		logger.Error("Failed to batch create spatial entities", "error", err)
		return nil, fmt.Errorf("failed to batch create spatial entities: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Error("Failed to close rows", "error", err)
		}
	}()

	var createdEntities []SpatialEntity
	for rows.Next() {
		var entity SpatialEntity
		var descriptionVal sql.NullString
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
			&descriptionVal,
			&entity.ChildCount,
			&entity.CreatedAt,
			&entity.UpdatedAt,
		)
		if err != nil {
			logger.Error("Failed to scan spatial entity row", "error", err)
			return nil, fmt.Errorf("failed to scan spatial entity: %w", err)
		}

		if descriptionVal.Valid {
			entity.Description = descriptionVal.String
		}

		if parentIDVal.Valid {
			parentID := int(parentIDVal.Int64)
			entity.ParentID = &parentID
		}

		createdEntities = append(createdEntities, entity)
	}

	if err := rows.Err(); err != nil {
		logger.Error("Error during rows iteration", "error", err)
		return nil, fmt.Errorf("error iterating spatial entities: %w", err)
	}

	logger.Info("Spatial entities batch created successfully", "count", len(createdEntities))
	return createdEntities, nil
}
