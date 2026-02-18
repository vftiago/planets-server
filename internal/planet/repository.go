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

const planetColumns = `id, system_id, planet_index, name, type, size, population, max_population, owner_id, created_at, updated_at`

func (r *Repository) scanPlanet(scanner interface{ Scan(...any) error }) (Planet, error) {
	var p Planet
	err := scanner.Scan(
		&p.ID, &p.SystemID, &p.PlanetIndex, &p.Name, &p.Type,
		&p.Size, &p.Population, &p.MaxPopulation, &p.OwnerID, &p.CreatedAt, &p.UpdatedAt,
	)
	return p, err
}

func (r *Repository) GetBySystemID(ctx context.Context, systemID int) ([]Planet, error) {
	query := `SELECT ` + planetColumns + ` FROM planets WHERE system_id = $1 ORDER BY planet_index`

	rows, err := r.db.QueryContext(ctx, query, systemID)
	if err != nil {
		return nil, errors.WrapInternal("failed to query planets by system", err)
	}
	defer func() { _ = rows.Close() }()

	var planets []Planet
	for rows.Next() {
		planet, err := r.scanPlanet(rows)
		if err != nil {
			return nil, errors.WrapInternal("failed to scan planet", err)
		}
		planets = append(planets, planet)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.WrapInternal("error iterating planets", err)
	}

	return planets, nil
}
