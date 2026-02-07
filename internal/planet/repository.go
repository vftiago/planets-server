package planet

import (
	"context"
	"encoding/json"
	"planets-server/internal/shared/database"
	"planets-server/internal/shared/errors"
)

type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) getExecutor(tx *database.Tx) database.Executor {
	if tx != nil {
		return tx
	}
	return r.db
}

// BatchInsertRequest represents a single planet to be inserted in a batch
type BatchInsertRequest struct {
	SystemID      int
	PlanetIndex   int
	Name          string
	Type          PlanetType
	Size          int
	MaxPopulation int64
}

// CreatePlanetsBatch creates multiple planets in a single database operation using JSON
func (r *Repository) CreatePlanetsBatch(ctx context.Context, planets []BatchInsertRequest, tx *database.Tx) (int, error) {
	if len(planets) == 0 {
		return 0, nil
	}

	exec := r.getExecutor(tx)

	// Convert planets to JSON
	planetsJSON, err := json.Marshal(planets)
	if err != nil {
		return 0, errors.WrapInternal("failed to marshal planets", err)
	}

	query := `
		INSERT INTO planets (system_id, planet_index, name, type, size, population, max_population, owner_id)
		SELECT
			(data->>'SystemID')::integer,
			(data->>'PlanetIndex')::integer,
			data->>'Name',
			(data->>'Type')::planet_type,
			(data->>'Size')::integer,
			0,
			(data->>'MaxPopulation')::bigint,
			NULL
		FROM json_array_elements($1::json) AS data`

	result, err := exec.ExecContext(ctx, query, string(planetsJSON))
	if err != nil {
		return 0, errors.WrapInternal("failed to batch create planets", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, errors.WrapInternal("failed to get rows affected", err)
	}

	return int(count), nil
}
