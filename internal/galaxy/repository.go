package galaxy

import (
	"database/sql"
	"fmt"
	"log/slog"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	logger := slog.With("component", "galaxy_repository", "operation", "init")
	logger.Debug("Initializing galaxy repository")
	return &Repository{db: db}
}

func (r *Repository) CreateGalaxy(gameID int, name, description string, sectorCount int) (*Galaxy, error) {
	logger := slog.With(
		"component", "galaxy_repository",
		"operation", "create_galaxy",
		"game_id", gameID,
		"name", name,
		"sector_count", sectorCount,
	)
	logger.Info("Creating galaxy")

	query := `
		INSERT INTO galaxies (game_id, name, description, sector_count)
		VALUES ($1, $2, $3, $4)
		RETURNING id, game_id, name, description, sector_count, created_at, updated_at
	`

	var galaxy Galaxy
	err := r.db.QueryRow(query, gameID, name, description, sectorCount).Scan(
		&galaxy.ID,
		&galaxy.GameID,
		&galaxy.Name,
		&galaxy.Description,
		&galaxy.SectorCount,
		&galaxy.CreatedAt,
		&galaxy.UpdatedAt,
	)

	if err != nil {
		logger.Error("Failed to create galaxy", "error", err)
		return nil, fmt.Errorf("failed to create galaxy: %w", err)
	}

	logger.Info("Galaxy created successfully", "galaxy_id", galaxy.ID)
	return &galaxy, nil
}

func (r *Repository) GetGalaxyByGameID(gameID int) (*Galaxy, error) {
	logger := slog.With("component", "galaxy_repository", "operation", "get_galaxy_by_game", "game_id", gameID)
	logger.Debug("Getting galaxy by game ID")

	query := `
		SELECT id, game_id, name, description, sector_count, created_at, updated_at
		FROM galaxies
		WHERE game_id = $1
	`

	var galaxy Galaxy
	err := r.db.QueryRow(query, gameID).Scan(
		&galaxy.ID,
		&galaxy.GameID,
		&galaxy.Name,
		&galaxy.Description,
		&galaxy.SectorCount,
		&galaxy.CreatedAt,
		&galaxy.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			logger.Debug("Galaxy not found for game")
			return nil, nil
		}
		logger.Error("Database error getting galaxy", "error", err)
		return nil, fmt.Errorf("database error: %w", err)
	}

	logger.Debug("Galaxy retrieved", "galaxy_id", galaxy.ID, "name", galaxy.Name)
	return &galaxy, nil
}

func (r *Repository) GetGalaxyByID(galaxyID int) (*Galaxy, error) {
	logger := slog.With("component", "galaxy_repository", "operation", "get_galaxy", "galaxy_id", galaxyID)
	logger.Debug("Getting galaxy by ID")

	query := `
		SELECT id, game_id, name, description, sector_count, created_at, updated_at
		FROM galaxies
		WHERE id = $1
	`

	var galaxy Galaxy
	err := r.db.QueryRow(query, galaxyID).Scan(
		&galaxy.ID,
		&galaxy.GameID,
		&galaxy.Name,
		&galaxy.Description,
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
