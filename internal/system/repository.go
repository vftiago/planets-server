package system

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
	logger.Debug("Initializing system repository")

	return &Repository{
		db:     db,
		logger: logger,
	}
}

func (r *Repository) CreateSystem(sectorID, systemX, systemY int, name string) (*System, error) {
	logger := r.logger.With(
		"component", "system_repository",
		"operation", "create_system",
		"sector_id", sectorID,
		"coordinates", fmt.Sprintf("(%d,%d)", systemX, systemY),
	)
	logger.Debug("Creating system")

	query := `
		INSERT INTO systems (sector_id, system_x, system_y, name, planet_count)
		VALUES ($1, $2, $3, $4, 0)
		RETURNING id, sector_id, system_x, system_y, name, planet_count, created_at, updated_at
	`

	var system System
	err := r.db.QueryRow(query, sectorID, systemX, systemY, name).Scan(
		&system.ID,
		&system.SectorID,
		&system.SystemX,
		&system.SystemY,
		&system.Name,
		&system.PlanetCount,
		&system.CreatedAt,
		&system.UpdatedAt,
	)

	if err != nil {
		logger.Error("Failed to create system", "error", err)
		return nil, fmt.Errorf("failed to create system: %w", err)
	}

	logger.Debug("System created successfully", "system_id", system.ID)
	return &system, nil
}

func (r *Repository) GetSystemsBySectorID(sectorID int) ([]System, error) {
	r.logger.Debug("Getting systems by sector ID")

	query := `
		SELECT id, sector_id, system_x, system_y, name, planet_count, created_at, updated_at
		FROM systems
		WHERE sector_id = $1
		ORDER BY system_x, system_y
	`

	rows, err := r.db.Query(query, sectorID)
	if err != nil {
		r.logger.Error("Failed to query systems", "error", err)
		return nil, fmt.Errorf("failed to query systems: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			r.logger.Error("Failed to close rows", "error", err)
		}
	}()

	var systems []System
	for rows.Next() {
		var system System
		err := rows.Scan(
			&system.ID,
			&system.SectorID,
			&system.SystemX,
			&system.SystemY,
			&system.Name,
			&system.PlanetCount,
			&system.CreatedAt,
			&system.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan system row", "error", err)
			return nil, fmt.Errorf("failed to scan system: %w", err)
		}
		systems = append(systems, system)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error during rows iteration", "error", err)
		return nil, fmt.Errorf("error iterating systems: %w", err)
	}

	r.logger.Debug("Systems retrieved", "count", len(systems))
	return systems, nil
}
