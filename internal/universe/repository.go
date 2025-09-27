package universe

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
	logger.Debug("Initializing universe repository")

	return &Repository{
		db:     db,
		logger: logger,
	}
}

func (r *Repository) CreateUniverse(universe *Universe) error {
	query := `
		INSERT INTO universes (name, description, galaxy_count, sector_count, system_count, planet_count)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(
		query,
		universe.Name,
		universe.Description,
		universe.GalaxyCount,
		universe.SectorCount,
		universe.SystemCount,
		universe.PlanetCount,
	).Scan(&universe.ID, &universe.CreatedAt, &universe.UpdatedAt)

	if err != nil {
		r.logger.Error("Failed to create universe", "error", err)
		return fmt.Errorf("failed to create universe: %w", err)
	}

	return nil
}

func (r *Repository) GetUniverse(id int) (*Universe, error) {
	query := `
		SELECT id, name, description, galaxy_count, sector_count, system_count, planet_count, created_at, updated_at
		FROM universes
		WHERE id = $1`

	universe := &Universe{}
	err := r.db.QueryRow(query, id).Scan(
		&universe.ID,
		&universe.Name,
		&universe.Description,
		&universe.GalaxyCount,
		&universe.SectorCount,
		&universe.SystemCount,
		&universe.PlanetCount,
		&universe.CreatedAt,
		&universe.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("universe not found")
		}
		r.logger.Error("Failed to get universe", "id", id, "error", err)
		return nil, fmt.Errorf("failed to get universe: %w", err)
	}

	return universe, nil
}

// ListUniverses retrieves all universes
func (r *Repository) ListUniverses() ([]*Universe, error) {
	query := `
		SELECT id, name, description, galaxy_count, sector_count, system_count, planet_count, created_at, updated_at
		FROM universes
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		r.logger.Error("Failed to list universes", "error", err)
		return nil, fmt.Errorf("failed to list universes: %w", err)
	}
	defer rows.Close()

	var universes []*Universe
	for rows.Next() {
		universe := &Universe{}
		err := rows.Scan(
			&universe.ID,
			&universe.Name,
			&universe.Description,
			&universe.GalaxyCount,
			&universe.SectorCount,
			&universe.SystemCount,
			&universe.PlanetCount,
			&universe.CreatedAt,
			&universe.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("Failed to scan universe", "error", err)
			return nil, fmt.Errorf("failed to scan universe: %w", err)
		}
		universes = append(universes, universe)
	}

	return universes, nil
}

// DeleteUniverse deletes a universe and all its associated data
func (r *Repository) DeleteUniverse(id int) error {
	query := `DELETE FROM universes WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		r.logger.Error("Failed to delete universe", "id", id, "error", err)
		return fmt.Errorf("failed to delete universe: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("universe not found")
	}

	return nil
}

// UpdateUniverseCounts updates the counts for a universe after generation
func (r *Repository) UpdateUniverseCounts(universeID int, galaxyCount, sectorCount, systemCount, planetCount int) error {
	query := `
		UPDATE universes 
		SET galaxy_count = $2, sector_count = $3, system_count = $4, planet_count = $5, updated_at = NOW()
		WHERE id = $1`

	_, err := r.db.Exec(query, universeID, galaxyCount, sectorCount, systemCount, planetCount)
	if err != nil {
		r.logger.Error("Failed to update universe counts", "universe_id", universeID, "error", err)
		return fmt.Errorf("failed to update universe counts: %w", err)
	}

	return nil
}
