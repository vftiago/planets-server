package spatial

import (
	"context"
	"fmt"
	"log/slog"
	"planets-server/internal/shared/database"

	"github.com/lib/pq"
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
// Returns only the IDs of created entities to minimize memory usage
func (r *Repository) CreateEntitiesBatch(ctx context.Context, entities []BatchInsertRequest, tx *database.Tx) ([]int, error) {
	if len(entities) == 0 {
		return []int{}, nil
	}

	exec := r.getExecutor(tx)

	logger := r.logger.With(
		"component", "spatial_repository",
		"operation", "create_entities_batch",
		"count", len(entities),
	)
	logger.Debug("Creating spatial entities in batch")

	// Build arrays for each column
	gameIDs := make([]int, len(entities))
	parentIDs := make([]*int, len(entities))
	entityTypes := make([]string, len(entities))
	levels := make([]int, len(entities))
	xCoords := make([]int, len(entities))
	yCoords := make([]int, len(entities))
	names := make([]string, len(entities))
	descriptions := make([]string, len(entities))

	for i, entity := range entities {
		gameIDs[i] = entity.GameID
		parentIDs[i] = entity.ParentID
		entityTypes[i] = string(entity.EntityType)
		levels[i] = entity.Level
		xCoords[i] = entity.XCoord
		yCoords[i] = entity.YCoord
		names[i] = entity.Name
		descriptions[i] = entity.Description
	}

	query := `
		INSERT INTO spatial_entities (game_id, parent_id, entity_type, level, x_coord, y_coord, name, description, child_count)
		SELECT
			unnest($1::int[]),
			unnest($2::int[]),
			unnest($3::entity_type[]),
			unnest($4::int[]),
			unnest($5::int[]),
			unnest($6::int[]),
			unnest($7::text[]),
			unnest($8::text[]),
			0
		RETURNING id`

	rows, err := exec.QueryContext(ctx, query,
		pq.Array(gameIDs),
		pq.Array(parentIDs),
		pq.Array(entityTypes),
		pq.Array(levels),
		pq.Array(xCoords),
		pq.Array(yCoords),
		pq.Array(names),
		pq.Array(descriptions),
	)
	if err != nil {
		logger.Error("Failed to batch create spatial entities", "error", err)
		return nil, fmt.Errorf("failed to batch create spatial entities: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Error("Failed to close rows", "error", err)
		}
	}()

	var entityIDs []int
	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			logger.Error("Failed to scan entity ID", "error", err)
			return nil, fmt.Errorf("failed to scan entity ID: %w", err)
		}
		entityIDs = append(entityIDs, id)
	}

	if err := rows.Err(); err != nil {
		logger.Error("Error during rows iteration", "error", err)
		return nil, fmt.Errorf("error iterating entity IDs: %w", err)
	}

	logger.Info("Spatial entities batch created successfully", "count", len(entityIDs))
	return entityIDs, nil
}
