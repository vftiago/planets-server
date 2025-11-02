package planet

import (
	"context"
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
	logger.Debug("Initializing planet repository")

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
func (r *Repository) CreatePlanetsBatch(ctx context.Context, planets []BatchInsertRequest, tx *database.Tx) ([]Planet, error) {
	if len(planets) == 0 {
		return []Planet{}, nil
	}

	exec := r.getExecutor(tx)

	logger := r.logger.With(
		"component", "planet_repository",
		"operation", "create_planets_batch",
		"count", len(planets),
	)
	logger.Debug("Creating planets in batch")

	// Convert planets to JSON
	planetsJSON, err := json.Marshal(planets)
	if err != nil {
		logger.Error("Failed to marshal planets to JSON", "error", err)
		return nil, fmt.Errorf("failed to marshal planets: %w", err)
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
		FROM json_array_elements($1::json) AS data
		RETURNING id, system_id, planet_index, name, type, size, population, max_population, owner_id, created_at, updated_at`

	rows, err := exec.QueryContext(ctx, query, string(planetsJSON))
	if err != nil {
		logger.Error("Failed to batch create planets", "error", err)
		return nil, fmt.Errorf("failed to batch create planets: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Error("Failed to close rows", "error", err)
		}
	}()

	var createdPlanets []Planet
	for rows.Next() {
		var planet Planet
		err := rows.Scan(
			&planet.ID,
			&planet.SystemID,
			&planet.PlanetIndex,
			&planet.Name,
			&planet.Type,
			&planet.Size,
			&planet.Population,
			&planet.MaxPopulation,
			&planet.OwnerID,
			&planet.CreatedAt,
			&planet.UpdatedAt,
		)
		if err != nil {
			logger.Error("Failed to scan planet row", "error", err)
			return nil, fmt.Errorf("failed to scan planet: %w", err)
		}
		createdPlanets = append(createdPlanets, planet)
	}

	if err := rows.Err(); err != nil {
		logger.Error("Error during rows iteration", "error", err)
		return nil, fmt.Errorf("error iterating planets: %w", err)
	}

	logger.Info("Planets batch created successfully", "count", len(createdPlanets))
	return createdPlanets, nil
}
