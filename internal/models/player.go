package models

import (
	"database/sql"
	"fmt"
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
	return &PlayerRepository{db: db}
}

func (r *PlayerRepository) GetPlayerCount() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM players").Scan(&count)
	return count, err
}

func (r *PlayerRepository) GetAllPlayers() ([]Player, error) {
	query := `
		SELECT id, username, email, display_name, avatar_url, created_at, updated_at
		FROM players
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
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
			return nil, err
		}
		players = append(players, player)
	}
	
	return players, rows.Err()
}

func (r *PlayerRepository) CreatePlayer(username, email, displayName string, avatarURL *string) (*Player, error) {
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
	
	return &player, err
}

func (r *PlayerRepository) FindOrCreatePlayerByOAuth(provider, providerUserID, email, displayName string, avatarURL *string) (*Player, error) {
	// Check if this exact provider + userID combo already exists
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
		return &player, nil
	}
	
	if err != sql.ErrNoRows {
		return nil, err // Database error
	}
	
	// Check if a player with this email already exists (account linking)
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
		authQuery := `
			INSERT INTO player_auth_providers (player_id, provider, provider_user_id, provider_email)
			VALUES ($1, $2, $3, $4)
		`
		
		_, err = r.db.Exec(authQuery, player.ID, provider, providerUserID, email)
		if err != nil {
			return nil, fmt.Errorf("failed to link auth provider: %w", err)
		}
		
		return &player, nil
	}
	
	if err != sql.ErrNoRows {
		return nil, err // Database error
	}
	
	// No existing player found, create new one	
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	
	username := generateUsernameFromEmail(email)
	
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
		return nil, err
	}
	
	authQuery := `
		INSERT INTO player_auth_providers (player_id, provider, provider_user_id, provider_email)
		VALUES ($1, $2, $3, $4)
	`
	
	_, err = tx.Exec(authQuery, player.ID, provider, providerUserID, email)
	if err != nil {
		return nil, err
	}
	
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	
	return &player, nil
}

func generateUsernameFromEmail(email string) string {
	// Extract username part from email
	if idx := strings.Index(email, "@"); idx > 0 {
		return email[:idx]
	}
	return "player"
}
