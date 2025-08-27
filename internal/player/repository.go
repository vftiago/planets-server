package player

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	logger := slog.With("component", "player_repository", "operation", "init")
	logger.Debug("Initializing player repository")
	return &Repository{db: db}
}

func (r *Repository) GetPlayerCount() (int, error) {
	logger := slog.With("component", "player_repository", "operation", "get_count")
	logger.Debug("Getting total player count")

	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM players").Scan(&count)
	if err != nil {
		logger.Error("Failed to get player count", "error", err)
		return 0, fmt.Errorf("failed to get player count: %w", err)
	}

	logger.Debug("Player count retrieved", "count", count)
	return count, nil
}

func (r *Repository) GetAllPlayers() ([]Player, error) {
	logger := slog.With("component", "player_repository", "operation", "get_all")
	logger.Debug("Retrieving all players")

	query := `
		SELECT id, username, email, display_name, avatar_url, created_at, updated_at
		FROM players
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		logger.Error("Failed to query players", "error", err)
		return nil, fmt.Errorf("failed to query players: %w", err)
	}
	defer rows.Close()

	var players []Player
	for rows.Next() {
		var player Player
		err := rows.Scan(
			&player.ID,
			&player.Username,
			&player.Email,
			&player.DisplayName,
			&player.AvatarURL,
			&player.CreatedAt,
			&player.UpdatedAt,
		)
		if err != nil {
			logger.Error("Failed to scan player row", "error", err)
			return nil, fmt.Errorf("failed to scan player: %w", err)
		}
		players = append(players, player)
	}

	if err := rows.Err(); err != nil {
		logger.Error("Error during rows iteration", "error", err)
		return nil, fmt.Errorf("error iterating players: %w", err)
	}

	logger.Debug("Players retrieved successfully", "count", len(players))
	return players, nil
}

func (r *Repository) CreatePlayer(username, email, displayName string, avatarURL *string) (*Player, error) {
	logger := slog.With(
		"component", "player_repository",
		"operation", "create",
		"username", username,
		"email", email,
	)
	logger.Info("Creating new player")

	query := `
		INSERT INTO players (username, email, display_name, avatar_url)
		VALUES ($1, $2, $3, $4)
		RETURNING id, username, email, display_name, avatar_url, created_at, updated_at
	`

	var player Player
	err := r.db.QueryRow(query, username, email, displayName, avatarURL).Scan(
		&player.ID,
		&player.Username,
		&player.Email,
		&player.DisplayName,
		&player.AvatarURL,
		&player.CreatedAt,
		&player.UpdatedAt,
	)

	if err != nil {
		logger.Error("Failed to create player", "error", err)
		return nil, fmt.Errorf("failed to create player: %w", err)
	}

	logger.Info("Player created successfully", "player_id", player.ID, "username", player.Username)
	return &player, nil
}

func (r *Repository) FindPlayerByEmail(email string) (*Player, error) {
	logger := slog.With(
		"component", "player_repository",
		"operation", "find_by_email",
		"email", email,
	)
	logger.Debug("Finding player by email")

	query := `
		SELECT id, username, email, display_name, avatar_url, created_at, updated_at
		FROM players
		WHERE email = $1
	`

	var player Player
	err := r.db.QueryRow(query, email).Scan(
		&player.ID,
		&player.Username,
		&player.Email,
		&player.DisplayName,
		&player.AvatarURL,
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

	logger.Debug("Found player by email", "player_id", player.ID)
	return &player, nil
}

func (r *Repository) GetPlayerByID(id int) (*Player, error) {
	logger := slog.With(
		"component", "player_repository",
		"operation", "get_by_id",
		"player_id", id,
	)
	logger.Debug("Getting player by ID")

	query := `
		SELECT id, username, email, display_name, avatar_url, created_at, updated_at
		FROM players
		WHERE id = $1
	`

	var player Player
	err := r.db.QueryRow(query, id).Scan(
		&player.ID,
		&player.Username,
		&player.Email,
		&player.DisplayName,
		&player.AvatarURL,
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

	logger.Debug("Found player by ID", "username", player.Username)
	return &player, nil
}

func generateUsernameFromEmail(email string) string {
	if idx := strings.Index(email, "@"); idx > 0 {
		return email[:idx]
	}
	return "player"
}