package planet

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
	logger.Debug("Initializing planet repository")

	return &Repository{
		db:     db,
		logger: logger,
	}
}

func (r *Repository) CreatePlanet(systemID, planetIndex int, name string, planetType PlanetType, size int, maxPopulation int64) (*Planet, error) {
	logger := r.logger.With(
		"component", "planet_repository",
		"operation", "create_planet",
		"system_id", systemID,
		"planet_index", planetIndex,
		"type", planetType,
	)
	logger.Debug("Creating planet")

	query := `
		INSERT INTO planets (system_id, planet_index, name, type, size, population, max_population, owner_id, is_homeworld)
		VALUES ($1, $2, $3, $4, $5, 0, $6, NULL, false)
		RETURNING id, system_id, planet_index, name, type, size, population, max_population, owner_id, is_homeworld, created_at, updated_at
	`

	var planet Planet
	err := r.db.QueryRow(query, systemID, planetIndex, name, planetType, size, maxPopulation).Scan(
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
		logger.Error("Failed to create planet", "error", err)
		return nil, fmt.Errorf("failed to create planet: %w", err)
	}

	logger.Debug("Planet created successfully", "planet_id", planet.ID)
	return &planet, nil
}

func (r *Repository) GetPlanetsBySystemID(systemID int) ([]Planet, error) {
	logger := slog.With("component", "planet_repository", "operation", "get_planets_by_system", "system_id", systemID)
	logger.Debug("Getting planets by system ID")

	query := `
		SELECT id, system_id, planet_index, name, type, size, population, max_population, owner_id, is_homeworld, created_at, updated_at
		FROM planets
		WHERE system_id = $1
		ORDER BY planet_index
	`

	rows, err := r.db.Query(query, systemID)
	if err != nil {
		logger.Error("Failed to query planets", "error", err)
		return nil, fmt.Errorf("failed to query planets: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Error("Failed to close rows", "error", err)
		}
	}()

	var planets []Planet
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
		planets = append(planets, planet)
	}

	if err := rows.Err(); err != nil {
		logger.Error("Error during rows iteration", "error", err)
		return nil, fmt.Errorf("error iterating planets: %w", err)
	}

	logger.Debug("Planets retrieved", "count", len(planets))
	return planets, nil
}
