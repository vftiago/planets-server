package galaxy

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
	logger.Debug("Initializing galaxy repository")

	return &Repository{
		db:     db,
		logger: logger,
	}
}

func (r *Repository) CreateGalaxy(universeID, galaxyX, galaxyY int, name string) error {
	logger := r.logger.With(
		"component", "galaxy_repository",
		"operation", "create_universe_galaxy",
		"universe_id", universeID,
		"name", name,
	)
	logger.Info("Creating universe galaxy")

	query := `
		INSERT INTO galaxies (universe_id, galaxy_x, galaxy_y, name, sector_count)
		VALUES ($1, $2, $3, $4, 0)
		RETURNING id, created_at, updated_at
	`

	var galaxy Galaxy
	err := r.db.QueryRow(
		query,
		universeID,
		galaxyX,
		galaxyY,
		name,
	).Scan(&galaxy.ID, &galaxy.CreatedAt, &galaxy.UpdatedAt)

	return err
}

func (r *Repository) GetGalaxiesByUniverseID(universeID int) ([]Galaxy, error) {
	logger := r.logger.With("component", "galaxy_repository", "operation", "get_galaxies_by_universe", "universe_id", universeID)
	logger.Debug("Getting galaxies by universe ID")

	query := `
		SELECT id, universe_id, name, galaxy_x, galaxy_y, sector_count, created_at, updated_at
		FROM galaxies
		WHERE universe_id = $1
		ORDER BY galaxy_x, galaxy_y
	`

	rows, err := r.db.Query(query, universeID)
	if err != nil {
		logger.Error("Failed to query galaxies", "error", err)
		return nil, fmt.Errorf("failed to query galaxies: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Error("Failed to close rows", "error", err)
		}
	}()

	var galaxies []Galaxy
	for rows.Next() {
		var galaxy Galaxy
		err := rows.Scan(
			&galaxy.ID,
			&galaxy.UniverseID,
			&galaxy.Name,
			&galaxy.GalaxyX,
			&galaxy.GalaxyY,
			&galaxy.SectorCount,
			&galaxy.CreatedAt,
			&galaxy.UpdatedAt,
		)
		if err != nil {
			logger.Error("Failed to scan galaxy row", "error", err)
			return nil, fmt.Errorf("failed to scan galaxy: %w", err)
		}
		galaxies = append(galaxies, galaxy)
	}

	if err := rows.Err(); err != nil {
		logger.Error("Error during rows iteration", "error", err)
		return nil, fmt.Errorf("error iterating galaxies: %w", err)
	}

	logger.Debug("Galaxies retrieved", "count", len(galaxies))
	return galaxies, nil
}

func (r *Repository) GetGalaxyByID(galaxyID int) (*Galaxy, error) {
	logger := r.logger.With("component", "galaxy_repository", "operation", "get_galaxy", "galaxy_id", galaxyID)
	logger.Debug("Getting galaxy by ID")

	query := `
		SELECT id, universe_id, name, galaxy_x, galaxy_y, sector_count, created_at, updated_at
		FROM galaxies
		WHERE id = $1
	`

	var galaxy Galaxy
	err := r.db.QueryRow(query, galaxyID).Scan(
		&galaxy.ID,
		&galaxy.UniverseID,
		&galaxy.Name,
		&galaxy.GalaxyX,
		&galaxy.GalaxyY,
		&galaxy.SectorCount,
		&galaxy.CreatedAt,
		&galaxy.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			logger.Debug("Galaxy not found")
			return nil, nil
		}
		logger.Error("Database error getting galaxy", "error", err)
		return nil, fmt.Errorf("database error: %w", err)
	}

	logger.Debug("Galaxy retrieved", "name", galaxy.Name)
	return &galaxy, nil
}
