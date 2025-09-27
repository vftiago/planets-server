package sector

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
	logger.Debug("Initializing sector repository")

	return &Repository{
		db:     db,
		logger: logger,
	}
}

func (r *Repository) CreateSector(galaxyID, sectorX, sectorY int, name string) (*Sector, error) {
	logger := slog.With(
		"component", "sector_repository",
		"operation", "create_sector",
		"galaxy_id", galaxyID,
		"coordinates", fmt.Sprintf("(%d,%d)", sectorX, sectorY),
	)
	logger.Debug("Creating sector")

	query := `
		INSERT INTO sectors (galaxy_id, sector_x, sector_y, name, system_count)
		VALUES ($1, $2, $3, $4, 0)
		RETURNING id, galaxy_id, sector_x, sector_y, name, system_count, created_at, updated_at
	`

	var sector Sector
	err := r.db.QueryRow(query, galaxyID, sectorX, sectorY, name).Scan(
		&sector.ID,
		&sector.GalaxyID,
		&sector.SectorX,
		&sector.SectorY,
		&sector.Name,
		&sector.SystemCount,
		&sector.CreatedAt,
		&sector.UpdatedAt,
	)

	if err != nil {
		logger.Error("Failed to create sector", "error", err)
		return nil, fmt.Errorf("failed to create sector: %w", err)
	}

	logger.Debug("Sector created successfully", "sector_id", sector.ID)
	return &sector, nil
}

func (r *Repository) GetSectorsByGalaxyID(galaxyID int) ([]Sector, error) {
	logger := slog.With("component", "sector_repository", "operation", "get_sectors_by_galaxy", "galaxy_id", galaxyID)
	logger.Debug("Getting sectors by galaxy ID")

	query := `
		SELECT id, galaxy_id, sector_x, sector_y, name, system_count, created_at, updated_at
		FROM sectors
		WHERE galaxy_id = $1
		ORDER BY sector_x, sector_y
	`

	rows, err := r.db.Query(query, galaxyID)
	if err != nil {
		logger.Error("Failed to query sectors", "error", err)
		return nil, fmt.Errorf("failed to query sectors: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Error("Failed to close rows", "error", err)
		}
	}()

	var sectors []Sector
	for rows.Next() {
		var sector Sector
		err := rows.Scan(
			&sector.ID,
			&sector.GalaxyID,
			&sector.SectorX,
			&sector.SectorY,
			&sector.Name,
			&sector.SystemCount,
			&sector.CreatedAt,
			&sector.UpdatedAt,
		)
		if err != nil {
			logger.Error("Failed to scan sector row", "error", err)
			return nil, fmt.Errorf("failed to scan sector: %w", err)
		}
		sectors = append(sectors, sector)
	}

	if err := rows.Err(); err != nil {
		logger.Error("Error during rows iteration", "error", err)
		return nil, fmt.Errorf("error iterating sectors: %w", err)
	}

	logger.Debug("Sectors retrieved", "count", len(sectors))
	return sectors, nil
}
