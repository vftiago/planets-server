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

func (r *Repository) CreateGalaxy(gameID, galaxyX, galaxyY int, name string) error {
	logger := r.logger.With(
		"component", "galaxy_repository",
		"operation", "create_game_galaxy",
		"game_id", gameID,
		"name", name,
	)
	logger.Info("Creating game galaxy")

	query := `
		INSERT INTO galaxies (game_id, galaxy_x, galaxy_y, name, sector_count)
		VALUES ($1, $2, $3, $4, 0)
		RETURNING id, created_at, updated_at
	`

	var galaxy Galaxy
	err := r.db.QueryRow(
		query,
		gameID,
		galaxyX,
		galaxyY,
		name,
	).Scan(&galaxy.ID, &galaxy.CreatedAt, &galaxy.UpdatedAt)

	return err
}

func (r *Repository) GetGalaxiesByGameID(gameID int) ([]Galaxy, error) {
	logger := r.logger.With("component", "galaxy_repository", "operation", "get_galaxies_by_game", "game_id", gameID)
	logger.Debug("Getting galaxies by game ID")

	query := `
		SELECT id, game_id, name, galaxy_x, galaxy_y, sector_count, created_at, updated_at
		FROM galaxies
		WHERE game_id = $1
		ORDER BY galaxy_x, galaxy_y
	`

	rows, err := r.db.Query(query, gameID)
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
			&galaxy.GameID,
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
		SELECT id, game_id, name, galaxy_x, galaxy_y, sector_count, created_at, updated_at
		FROM galaxies
		WHERE id = $1
	`

	var galaxy Galaxy
	err := r.db.QueryRow(query, galaxyID).Scan(
		&galaxy.ID,
		&galaxy.GameID,
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
