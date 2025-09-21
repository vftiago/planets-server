package planet

import (
	"database/sql"
	"fmt"
	"log/slog"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	logger := slog.With("component", "planet_repository", "operation", "init")
	logger.Debug("Initializing planet repository")
	return &Repository{db: db}
}

func (r *Repository) CreatePlanet(systemID, planetIndex int, name string, planetType PlanetType, size int, maxPopulation int64) (*Planet, error) {
	logger := slog.With(
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
		&planet.IsHomeworld,
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

func (r *Repository) CreatePlanetsBatch(planets []Planet) error {
	logger := slog.With("component", "planet_repository", "operation", "create_planets_batch")
	logger.Debug("Creating planets in batch", "count", len(planets))

	if len(planets) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		logger.Error("Failed to begin transaction", "error", err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err.Error() != "sql: transaction has already been committed or rolled back" {
			logger.Error("Failed to rollback transaction", "error", err)
		}
	}()

	stmt, err := tx.Prepare(`
		INSERT INTO planets (system_id, planet_index, name, type, size, population, max_population, owner_id, is_homeworld)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`)
	if err != nil {
		logger.Error("Failed to prepare statement", "error", err)
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, planet := range planets {
		_, err := stmt.Exec(
			planet.SystemID,
			planet.PlanetIndex,
			planet.Name,
			planet.Type,
			planet.Size,
			planet.Population,
			planet.MaxPopulation,
			planet.OwnerID,
			planet.IsHomeworld,
		)
		if err != nil {
			logger.Error("Failed to insert planet", "error", err, "planet_name", planet.Name)
			return fmt.Errorf("failed to insert planet %s: %w", planet.Name, err)
		}
	}

	if err := tx.Commit(); err != nil {
		logger.Error("Failed to commit transaction", "error", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	logger.Info("Planets created successfully", "count", len(planets))
	return nil
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
			&planet.IsHomeworld,
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
