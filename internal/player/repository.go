package player

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"planets-server/internal/shared/config"
	"planets-server/internal/shared/database"
)

type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) *Repository {
	logger := slog.With("component", "player_repository", "operation", "init")
	logger.Debug("Initializing player repository")
	return &Repository{db: db}
}

func (r *Repository) GetPlayerCount(ctx context.Context) (int, error) {
	logger := slog.With("component", "player_repository", "operation", "get_count")
	logger.Debug("Getting total player count")

	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM players").Scan(&count)
	if err != nil {
		logger.Error("Failed to get player count", "error", err)
		return 0, fmt.Errorf("failed to get player count: %w", err)
	}

	logger.Debug("Player count retrieved", "count", count)
	return count, nil
}

func (r *Repository) GetAllPlayers(ctx context.Context) ([]Player, error) {
	logger := slog.With("component", "player_repository", "operation", "get_all")
	logger.Debug("Retrieving all players")

	query := `
		SELECT id, username, email, display_name, avatar_url, role, created_at, updated_at
		FROM players
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		logger.Error("Failed to query players", "error", err)
		return nil, fmt.Errorf("failed to query players: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Error("Failed to close rows", "error", err)
		}
	}()

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
			logger.Error("Failed to scan player row", "error", err)
			return nil, fmt.Errorf("failed to scan player: %w", err)
		}
		player.Role = ParsePlayerRole(roleStr)
		players = append(players, player)
	}

	if err := rows.Err(); err != nil {
		logger.Error("Error during rows iteration", "error", err)
		return nil, fmt.Errorf("error iterating players: %w", err)
	}

	logger.Debug("Players retrieved successfully", "count", len(players))
	return players, nil
}

func (r *Repository) CreatePlayer(ctx context.Context, username, email, displayName string, avatarURL *string) (*Player, error) {
	logger := slog.With(
		"component", "player_repository",
		"operation", "create",
		"username", username,
		"email", email,
	)
	logger.Info("Creating new player")

	role := r.determinePlayerRole(email)
	if role == PlayerRoleAdmin {
		logger.Info("Creating player with admin role", "email", email)
	}

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
		logger.Error("Failed to create player", "error", err)
		return nil, fmt.Errorf("failed to create player: %w", err)
	}

	player.Role = ParsePlayerRole(roleStr)

	logger.Info("Player created successfully",
		"player_id", player.ID,
		"username", player.Username,
		"role", player.Role)
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
	logger := slog.With(
		"component", "player_repository",
		"operation", "find_by_email",
		"email", email,
	)
	logger.Debug("Finding player by email")

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
			logger.Debug("No player found with email")
			return nil, nil
		}
		logger.Error("Database error finding player by email", "error", err)
		return nil, fmt.Errorf("database error: %w", err)
	}

	player.Role = ParsePlayerRole(roleStr)
	logger.Debug("Found player by email", "player_id", player.ID, "role", player.Role)
	return &player, nil
}

func (r *Repository) GetPlayerByID(ctx context.Context, id int) (*Player, error) {
	logger := slog.With(
		"component", "player_repository",
		"operation", "get_by_id",
		"player_id", id,
	)
	logger.Debug("Getting player by ID")

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
			logger.Debug("No player found with ID")
			return nil, nil
		}
		logger.Error("Database error getting player by ID", "error", err)
		return nil, fmt.Errorf("database error: %w", err)
	}

	player.Role = ParsePlayerRole(roleStr)
	logger.Debug("Found player by ID", "username", player.Username, "role", player.Role)
	return &player, nil
}

func (r *Repository) UpdatePlayerRole(ctx context.Context, playerID int, role PlayerRole) error {
	logger := slog.With(
		"component", "player_repository",
		"operation", "update_role",
		"player_id", playerID,
		"role", role,
	)
	logger.Info("Updating player role")

	if !role.IsValid() {
		return fmt.Errorf("invalid role: %s", role)
	}

	query := `UPDATE players SET role = $1 WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, role.String(), playerID)
	if err != nil {
		logger.Error("Failed to update player role", "error", err)
		return fmt.Errorf("failed to update player role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Failed to get rows affected", "error", err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		logger.Warn("Player not found for role update")
		return fmt.Errorf("player not found")
	}

	logger.Info("Player role updated successfully")
	return nil
}
