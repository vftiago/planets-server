package game

import (
	"context"
	"database/sql"
	"planets-server/internal/shared/database"
	"planets-server/internal/shared/errors"
	"time"
)

type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) getExecutor(tx *database.Tx) database.Executor {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *Repository) CreateGame(ctx context.Context, name string, seed string, config GameConfig, tx *database.Tx) (*Game, error) {
	exec := r.getExecutor(tx)

	query := `
		INSERT INTO games (name, seed, status, current_turn, max_players, turn_interval_hours)
		VALUES ($1, $2, 'creating', 0, $3, $4)
		RETURNING id, name, seed, planet_count, status, current_turn, max_players, turn_interval_hours, next_turn_at, created_at, updated_at
	`

	var game Game
	err := exec.QueryRowContext(ctx, query, name, seed, config.MaxPlayers, config.TurnIntervalHours).Scan(
		&game.ID,
		&game.Name,
		&game.Seed,
		&game.PlanetCount,
		&game.Status,
		&game.CurrentTurn,
		&game.MaxPlayers,
		&game.TurnIntervalHours,
		&game.NextTurnAt,
		&game.CreatedAt,
		&game.UpdatedAt,
	)

	if err != nil {
		return nil, errors.WrapInternal("failed to create game", err)
	}

	return &game, nil
}

func (r *Repository) GetGameByID(ctx context.Context, gameID int) (*Game, error) {
	query := `
		SELECT id, name, seed, universe_id, planet_count, status, current_turn, max_players, turn_interval_hours, next_turn_at, created_at, updated_at
		FROM games
		WHERE id = $1
	`

	var game Game
	err := r.db.QueryRowContext(ctx, query, gameID).Scan(
		&game.ID,
		&game.Name,
		&game.Seed,
		&game.UniverseID,
		&game.PlanetCount,
		&game.Status,
		&game.CurrentTurn,
		&game.MaxPlayers,
		&game.TurnIntervalHours,
		&game.NextTurnAt,
		&game.CreatedAt,
		&game.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFoundf("game not found with id: %d", gameID)
		}
		return nil, errors.WrapInternal("failed to get game by id", err)
	}

	return &game, nil
}

func (r *Repository) GetAllGames(ctx context.Context) ([]Game, error) {
	query := `
		SELECT id, name, seed, universe_id, planet_count, status, current_turn, max_players, turn_interval_hours, next_turn_at, created_at, updated_at
		FROM games
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.WrapInternal("failed to query games", err)
	}
	defer func() { _ = rows.Close() }()

	var games []Game
	for rows.Next() {
		var game Game
		err := rows.Scan(
			&game.ID,
			&game.Name,
			&game.Seed,
			&game.UniverseID,
			&game.PlanetCount,
			&game.Status,
			&game.CurrentTurn,
			&game.MaxPlayers,
			&game.TurnIntervalHours,
			&game.NextTurnAt,
			&game.CreatedAt,
			&game.UpdatedAt,
		)
		if err != nil {
			return nil, errors.WrapInternal("failed to scan game", err)
		}
		games = append(games, game)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.WrapInternal("error iterating games", err)
	}

	return games, nil
}

func (r *Repository) ActivateGame(ctx context.Context, gameID int, tx *database.Tx) error {
	exec := r.getExecutor(tx)

	nextTurnAt := time.Now().Add(1 * time.Hour).Truncate(time.Hour)

	query := `
		UPDATE games
		SET status = 'active', current_turn = 1, next_turn_at = $1
		WHERE id = $2 AND status = 'creating'
	`

	result, err := exec.ExecContext(ctx, query, nextTurnAt, gameID)
	if err != nil {
		return errors.WrapInternal("failed to activate game", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.WrapInternal("failed to get rows affected after activation", err)
	}

	if rowsAffected == 0 {
		return errors.Conflictf("game not found or not ready for activation (id: %d)", gameID)
	}

	return nil
}

func (r *Repository) GetGameStats(ctx context.Context, gameID int) (*GameStats, error) {
	query := `
		SELECT
			g.id,
			g.name,
			g.status,
			g.current_turn,
			COALESCE(player_count.count, 0) as player_count,
			g.max_players,
			g.next_turn_at,
			g.planet_count
		FROM games g
		LEFT JOIN (
			SELECT game_id, COUNT(*) as count
			FROM game_players
			WHERE game_id = $1
		) player_count ON g.id = player_count.game_id
		WHERE g.id = $1
	`

	var stats GameStats
	err := r.db.QueryRowContext(ctx, query, gameID).Scan(
		&stats.ID,
		&stats.Name,
		&stats.Status,
		&stats.CurrentTurn,
		&stats.PlayerCount,
		&stats.MaxPlayers,
		&stats.NextTurnAt,
		&stats.PlanetCount,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFoundf("game not found with id: %d", gameID)
		}
		return nil, errors.WrapInternal("failed to get game stats", err)
	}

	return &stats, nil
}

func (r *Repository) DeleteGame(ctx context.Context, gameID int) error {
	query := `DELETE FROM games WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, gameID)
	if err != nil {
		return errors.WrapInternal("failed to delete game", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.WrapInternal("failed to get rows affected after deletion", err)
	}

	if rowsAffected == 0 {
		return errors.NotFoundf("game not found with id: %d", gameID)
	}

	return nil
}

func (r *Repository) SetUniverseID(ctx context.Context, gameID int, universeID int, tx *database.Tx) error {
	exec := r.getExecutor(tx)

	query := `UPDATE games SET universe_id = $2 WHERE id = $1`
	_, err := exec.ExecContext(ctx, query, gameID, universeID)
	if err != nil {
		return errors.WrapInternal("failed to set universe_id on game", err)
	}

	return nil
}

func (r *Repository) UpdateGameCounts(ctx context.Context, gameID int, planetCount int, tx *database.Tx) error {
	exec := r.getExecutor(tx)

	query := `
		UPDATE games
		SET planet_count = $2, updated_at = NOW()
		WHERE id = $1`

	result, err := exec.ExecContext(ctx, query, gameID, planetCount)
	if err != nil {
		return errors.WrapInternal("failed to update game counts", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.WrapInternal("failed to get rows affected after count update", err)
	}

	if rowsAffected == 0 {
		return errors.NotFoundf("game not found with id: %d", gameID)
	}

	return nil
}
