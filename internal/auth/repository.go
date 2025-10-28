package auth

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"planets-server/internal/shared/database"
)

type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) *Repository {
	logger := slog.With("component", "auth_repository", "operation", "init")
	logger.Debug("Initializing auth repository")
	return &Repository{db: db}
}

func (r *Repository) CreateAuthProvider(ctx context.Context, playerID int, provider, providerUserID, providerEmail string) error {
	logger := slog.With(
		"component", "auth_repository",
		"operation", "create_auth_provider",
		"player_id", playerID,
		"provider", provider,
	)
	logger.Debug("Creating auth provider record")

	query := `
		INSERT INTO player_auth_providers (player_id, provider, provider_user_id, provider_email)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.ExecContext(ctx, query, playerID, provider, providerUserID, providerEmail)
	if err != nil {
		logger.Error("Failed to create auth provider", "error", err)
		return fmt.Errorf("failed to create auth provider: %w", err)
	}

	logger.Debug("Auth provider created successfully")
	return nil
}

func (r *Repository) FindPlayerByAuthProvider(ctx context.Context, provider, providerUserID string) (int, error) {
	logger := slog.With(
		"component", "auth_repository",
		"operation", "find_player_by_auth",
		"provider", provider,
	)
	logger.Debug("Finding player by auth provider")

	query := `
		SELECT player_id
		FROM player_auth_providers
		WHERE provider = $1 AND provider_user_id = $2
	`

	var playerID int
	err := r.db.QueryRowContext(ctx, query, provider, providerUserID).Scan(&playerID)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Debug("No player found for auth provider")
			return 0, nil
		}
		logger.Error("Database error finding player by auth provider", "error", err)
		return 0, fmt.Errorf("database error: %w", err)
	}

	logger.Debug("Found player for auth provider", "player_id", playerID)
	return playerID, nil
}
