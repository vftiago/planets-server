package player

import (
	"context"
	"database/sql"
	"planets-server/internal/shared/config"
	"planets-server/internal/shared/database"
	"planets-server/internal/shared/errors"
)

type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetPlayerCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM players").Scan(&count)
	if err != nil {
		return 0, errors.WrapInternal("failed to get player count", err)
	}
	return count, nil
}

func (r *Repository) GetAllPlayers(ctx context.Context) ([]Player, error) {
	query := `
		SELECT id, username, email, display_name, avatar_url, role, created_at, updated_at
		FROM players
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.WrapInternal("failed to query players", err)
	}
	defer func() { _ = rows.Close() }()

	var players []Player
	for rows.Next() {
		var player Player
		var roleStr string
		err := rows.Scan(
			&player.ID,
			&player.Username,
			&player.Email,
			&player.DisplayName,
			&player.AvatarURL,
			&roleStr,
			&player.CreatedAt,
			&player.UpdatedAt,
		)
		if err != nil {
			return nil, errors.WrapInternal("failed to scan player", err)
		}
		player.Role = ParsePlayerRole(roleStr)
		players = append(players, player)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.WrapInternal("error iterating players", err)
	}

	return players, nil
}

func (r *Repository) CreatePlayer(ctx context.Context, username, email, displayName string, avatarURL *string) (*Player, error) {
	role := r.determinePlayerRole(email)

	query := `
		INSERT INTO players (username, email, display_name, avatar_url, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, username, email, display_name, avatar_url, role, created_at, updated_at
	`

	var player Player
	var roleStr string
	err := r.db.QueryRowContext(ctx, query, username, email, displayName, avatarURL, role.String()).Scan(
		&player.ID,
		&player.Username,
		&player.Email,
		&player.DisplayName,
		&player.AvatarURL,
		&roleStr,
		&player.CreatedAt,
		&player.UpdatedAt,
	)

	if err != nil {
		return nil, errors.WrapInternal("failed to create player", err)
	}

	player.Role = ParsePlayerRole(roleStr)
	return &player, nil
}

func (r *Repository) determinePlayerRole(email string) PlayerRole {
	cfg := config.GlobalConfig
	if cfg != nil && email == cfg.Admin.Email {
		return PlayerRoleAdmin
	}
	return PlayerRoleUser
}

func (r *Repository) FindPlayerByEmail(ctx context.Context, email string) (*Player, error) {
	query := `
		SELECT id, username, email, display_name, avatar_url, role, created_at, updated_at
		FROM players
		WHERE email = $1
	`

	var player Player
	var roleStr string
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&player.ID,
		&player.Username,
		&player.Email,
		&player.DisplayName,
		&player.AvatarURL,
		&roleStr,
		&player.CreatedAt,
		&player.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFoundf("player not found with email: %s", email)
		}
		return nil, errors.WrapInternal("failed to find player by email", err)
	}

	player.Role = ParsePlayerRole(roleStr)
	return &player, nil
}

func (r *Repository) GetPlayerByID(ctx context.Context, id int) (*Player, error) {
	query := `
		SELECT id, username, email, display_name, avatar_url, role, created_at, updated_at
		FROM players
		WHERE id = $1
	`

	var player Player
	var roleStr string
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&player.ID,
		&player.Username,
		&player.Email,
		&player.DisplayName,
		&player.AvatarURL,
		&roleStr,
		&player.CreatedAt,
		&player.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFoundf("player not found with id: %d", id)
		}
		return nil, errors.WrapInternal("failed to get player by id", err)
	}

	player.Role = ParsePlayerRole(roleStr)
	return &player, nil
}

func (r *Repository) UpdatePlayerRole(ctx context.Context, playerID int, role PlayerRole) error {
	if !role.IsValid() {
		return errors.Validationf("invalid role: %s", role)
	}

	query := `UPDATE players SET role = $1 WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, role.String(), playerID)
	if err != nil {
		return errors.WrapInternal("failed to update player role", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.WrapInternal("failed to get rows affected after role update", err)
	}

	if rowsAffected == 0 {
		return errors.NotFoundf("player not found with id: %d", playerID)
	}

	return nil
}
