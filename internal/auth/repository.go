package auth

import (
	"context"
	"database/sql"
	"planets-server/internal/shared/database"
	"planets-server/internal/shared/errors"
)

type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateAuthProvider(ctx context.Context, playerID int, provider, providerUserID, providerEmail string) error {
	query := `
		INSERT INTO player_auth_providers (player_id, provider, provider_user_id, provider_email)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.ExecContext(ctx, query, playerID, provider, providerUserID, providerEmail)
	if err != nil {
		return errors.WrapInternal("failed to create auth provider", err)
	}

	return nil
}

func (r *Repository) FindPlayerByAuthProvider(ctx context.Context, provider, providerUserID string) (int, error) {
	query := `
		SELECT player_id
		FROM player_auth_providers
		WHERE provider = $1 AND provider_user_id = $2
	`

	var playerID int
	err := r.db.QueryRowContext(ctx, query, provider, providerUserID).Scan(&playerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.NotFoundf("player not found for auth provider: %s", provider)
		}
		return 0, errors.WrapInternal("failed to find player by auth provider", err)
	}

	return playerID, nil
}
