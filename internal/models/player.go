package models

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

type Player struct {
	ID          int       `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PlayerAuthProvider struct {
	ID             int       `json:"id"`
	PlayerID       int       `json:"player_id"`
	Provider       string    `json:"provider"`
	ProviderUserID *string   `json:"provider_user_id"`
	ProviderEmail  *string   `json:"provider_email"`
	CreatedAt      time.Time `json:"created_at"`
}

type PlayerRepository struct {
	db *sql.DB
}

func NewPlayerRepository(db *sql.DB) *PlayerRepository {
	logger := slog.With("component", "player_repository", "operation", "init")
	logger.Debug("Initializing player repository")
	return &PlayerRepository{db: db}
}

func (r *PlayerRepository) GetPlayerCount() (int, error) {
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

func (r *PlayerRepository) GetAllPlayers() ([]Player, error) {
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

func (r *PlayerRepository) CreatePlayer(username, email, displayName string, avatarURL *string) (*Player, error) {
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

func (r *PlayerRepository) FindOrCreatePlayerByOAuth(provider, providerUserID, email, displayName string, avatarURL *string) (*Player, error) {
	logger := slog.With(
		"component", "player_repository",
		"operation", "find_or_create_oauth",
		"provider", provider,
		"email", email,
		"provider_user_id", providerUserID,
	)
	logger.Debug("Finding or creating player by OAuth")

	// Check if this exact provider + userID combo already exists
	logger.Debug("Checking for existing OAuth player")
	query := `
		SELECT p.id, p.username, p.email, p.display_name, p.avatar_url, p.created_at, p.updated_at
		FROM players p
		JOIN player_auth_providers pap ON p.id = pap.player_id
		WHERE pap.provider = $1 AND pap.provider_user_id = $2
	`
	
	var player Player
	err := r.db.QueryRow(query, provider, providerUserID).Scan(
		&player.ID, &player.Username, &player.Email, &player.DisplayName,
		&player.AvatarURL, &player.CreatedAt, &player.UpdatedAt,
	)
	
	if err == nil {
		logger.Info("Found existing OAuth player", "player_id", player.ID, "username", player.Username)
		return &player, nil
	}
	
	if err != sql.ErrNoRows {
		logger.Error("Database error checking for OAuth player", "error", err)
		return nil, fmt.Errorf("database error: %w", err)
	}
	
	// Check if a player with this email already exists (account linking)
	logger.Debug("Checking for existing player by email for account linking")
	emailQuery := `
		SELECT id, username, email, display_name, avatar_url, created_at, updated_at
		FROM players
		WHERE email = $1
	`
	
	err = r.db.QueryRow(emailQuery, email).Scan(
		&player.ID, &player.Username, &player.Email, &player.DisplayName,
		&player.AvatarURL, &player.CreatedAt, &player.UpdatedAt,
	)
	
	if err == nil {
		// Link the OAuth provider to existing player
		logger.Info("Linking OAuth provider to existing player", "player_id", player.ID)
		authQuery := `
			INSERT INTO player_auth_providers (player_id, provider, provider_user_id, provider_email)
			VALUES ($1, $2, $3, $4)
		`
		
		_, err = r.db.Exec(authQuery, player.ID, provider, providerUserID, email)
		if err != nil {
			logger.Error("Failed to link auth provider", "error", err, "player_id", player.ID)
			return nil, fmt.Errorf("failed to link auth provider: %w", err)
		}
		
		logger.Info("Successfully linked OAuth provider to existing player", "player_id", player.ID)
		return &player, nil
	}
	
	if err != sql.ErrNoRows {
		logger.Error("Database error checking for player by email", "error", err)
		return nil, fmt.Errorf("database error: %w", err)
	}
	
	// No existing player found, create new one
	logger.Info("Creating new player with OAuth provider")
	tx, err := r.db.Begin()
	if err != nil {
		logger.Error("Failed to begin transaction", "error", err)
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	username := generateUsernameFromEmail(email)
	logger.Debug("Generated username from email", "username", username, "email", email)
	
	playerQuery := `
		INSERT INTO players (username, email, display_name, avatar_url)
		VALUES ($1, $2, $3, $4)
		RETURNING id, username, email, display_name, avatar_url, created_at, updated_at
	`
	
	err = tx.QueryRow(playerQuery, username, email, displayName, avatarURL).Scan(
		&player.ID, &player.Username, &player.Email, &player.DisplayName,
		&player.AvatarURL, &player.CreatedAt, &player.UpdatedAt,
	)
	if err != nil {
		logger.Error("Failed to insert new player", "error", err)
		return nil, fmt.Errorf("failed to create player: %w", err)
	}
	
	// Link the OAuth provider
	authQuery := `
		INSERT INTO player_auth_providers (player_id, provider, provider_user_id, provider_email)
		VALUES ($1, $2, $3, $4)
	`
	
	_, err = tx.Exec(authQuery, player.ID, provider, providerUserID, email)
	if err != nil {
		logger.Error("Failed to insert auth provider", "error", err, "player_id", player.ID)
		return nil, fmt.Errorf("failed to create auth provider: %w", err)
	}
	
	if err = tx.Commit(); err != nil {
		logger.Error("Failed to commit transaction", "error", err)
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	logger.Info("Successfully created new player with OAuth", 
		"player_id", player.ID, 
		"username", player.Username,
		"provider", provider)
	
	return &player, nil
}

func generateUsernameFromEmail(email string) string {
	// Extract username part from email
	if idx := strings.Index(email, "@"); idx > 0 {
		return email[:idx]
	}
	return "player"
}
