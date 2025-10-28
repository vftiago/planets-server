package game

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"planets-server/internal/shared/database"
	"time"
)

type Repository struct {
	db     *database.DB
	logger *slog.Logger
}

func NewRepository(db *database.DB, logger *slog.Logger) *Repository {
	logger.Debug("Initializing game repository")

	return &Repository{
		db:     db,
		logger: logger,
	}
}

func (r *Repository) getExecutor(tx *database.Tx) database.Executor {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *Repository) CreateGame(ctx context.Context, config GameConfig, tx *database.Tx) (*Game, error) {
	exec := r.getExecutor(tx)

	logger := r.logger.With(
		"component", "game_repository",
		"operation", "create_game",
		"name", config.Name,
		"max_players", config.MaxPlayers,
	)
	logger.Info("Creating new game")

	query := `
		INSERT INTO games (name, description, universe_name, universe_description, status, current_turn, max_players, turn_interval_hours)
		VALUES ($1, $2, $3, $4, 'creating', 0, $5, $6)
		RETURNING id, name, description, universe_name, universe_description, planet_count, status, current_turn, max_players, turn_interval_hours, next_turn_at, created_at, updated_at
	`

	var game Game
	err := exec.QueryRowContext(ctx, query, config.Name, config.Description, config.UniverseName, config.UniverseDescription, config.MaxPlayers, config.TurnIntervalHours).Scan(
		&game.ID,
		&game.Name,
		&game.Description,
		&game.UniverseName,
		&game.UniverseDescription,
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
		logger.Error("Failed to create game", "error", err)
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	logger.Info("Game created successfully", "game_id", game.ID)
	return &game, nil
}

func (r *Repository) GetGameByID(ctx context.Context, gameID int) (*Game, error) {
	logger := slog.With("component", "game_repository", "operation", "get_game", "game_id", gameID)
	logger.Debug("Getting game by ID")

	query := `
		SELECT id, name, description, universe_name, universe_description, planet_count, status, current_turn, max_players, turn_interval_hours, next_turn_at, created_at, updated_at
		FROM games
		WHERE id = $1
	`

	var game Game
	err := r.db.QueryRowContext(ctx, query, gameID).Scan(
		&game.ID,
		&game.Name,
		&game.Description,
		&game.UniverseName,
		&game.UniverseDescription,
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
			logger.Debug("Game not found")
			return nil, nil
		}
		logger.Error("Database error getting game", "error", err)
		return nil, fmt.Errorf("database error: %w", err)
	}

	logger.Debug("Game retrieved", "name", game.Name, "status", game.Status)
	return &game, nil
}

func (r *Repository) GetAllGames(ctx context.Context) ([]Game, error) {
	logger := slog.With("component", "game_repository", "operation", "get_all_games")
	logger.Debug("Getting all games")

	query := `
		SELECT id, name, description, universe_name, universe_description, planet_count, status, current_turn, max_players, turn_interval_hours, next_turn_at, created_at, updated_at
		FROM games
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		logger.Error("Failed to query games", "error", err)
		return nil, fmt.Errorf("failed to query games: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Error("Failed to close rows", "error", err)
		}
	}()

	var games []Game
	for rows.Next() {
		var game Game
		err := rows.Scan(
			&game.ID,
			&game.Name,
			&game.Description,
			&game.UniverseName,
			&game.UniverseDescription,
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
			logger.Error("Failed to scan game row", "error", err)
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}
		games = append(games, game)
	}

	if err := rows.Err(); err != nil {
		logger.Error("Error during rows iteration", "error", err)
		return nil, fmt.Errorf("error iterating games: %w", err)
	}

	logger.Debug("Games retrieved", "count", len(games))
	return games, nil
}

func (r *Repository) UpdateGameStatus(ctx context.Context, gameID int, status GameStatus) error {
	logger := slog.With("component", "game_repository", "operation", "update_status", "game_id", gameID, "status", status)
	logger.Debug("Updating game status")

	query := `UPDATE games SET status = $1 WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, status, gameID)
	if err != nil {
		logger.Error("Failed to update game status", "error", err)
		return fmt.Errorf("failed to update game status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Failed to get rows affected", "error", err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		logger.Warn("Game not found for status update")
		return fmt.Errorf("game not found")
	}

	logger.Debug("Game status updated")
	return nil
}

func (r *Repository) ActivateGame(ctx context.Context, gameID int, tx *database.Tx) error {
	exec := r.getExecutor(tx)

	logger := slog.With("component", "game_repository", "operation", "activate_game", "game_id", gameID)
	logger.Info("Activating game")

	nextTurnAt := time.Now().Add(1 * time.Hour).Truncate(time.Hour)

	query := `
		UPDATE games
		SET status = 'active', current_turn = 1, next_turn_at = $1
		WHERE id = $2 AND status = 'creating'
	`

	result, err := exec.ExecContext(ctx, query, nextTurnAt, gameID)
	if err != nil {
		logger.Error("Failed to activate game", "error", err)
		return fmt.Errorf("failed to activate game: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Failed to get rows affected", "error", err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		logger.Warn("Game not found or not in creating status")
		return fmt.Errorf("game not found or not ready for activation")
	}

	logger.Info("Game activated successfully", "next_turn_at", nextTurnAt)
	return nil
}

func (r *Repository) GetGameStats(ctx context.Context, gameID int) (*GameStats, error) {
	logger := slog.With("component", "game_repository", "operation", "get_game_stats", "game_id", gameID)
	logger.Debug("Getting game statistics")

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
			logger.Debug("Game not found for stats")
			return nil, nil
		}
		logger.Error("Failed to get game stats", "error", err)
		return nil, fmt.Errorf("failed to get game stats: %w", err)
	}

	logger.Debug("Game stats retrieved", "players", stats.PlayerCount)

	return &stats, nil
}

func (r *Repository) DeleteGame(ctx context.Context, gameID int) error {
	logger := slog.With("component", "game_repository", "operation", "delete_game", "game_id", gameID)
	logger.Info("Deleting game and all related data")

	query := `DELETE FROM games WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, gameID)
	if err != nil {
		logger.Error("Failed to delete game", "error", err)
		return fmt.Errorf("failed to delete game: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Failed to get rows affected", "error", err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		logger.Warn("Game not found for deletion")
		return fmt.Errorf("game not found")
	}

	logger.Info("Game deleted successfully")
	return nil
}

func (r *Repository) UpdateGameCounts(ctx context.Context, gameID int, planetCount int, tx *database.Tx) error {
	exec := r.getExecutor(tx)

	logger := r.logger.With("component", "game_repository", "operation", "update_counts", "game_id", gameID)
	logger.Debug("Updating game counts")

	query := `
		UPDATE games
		SET planet_count = $2, updated_at = NOW()
		WHERE id = $1`

	result, err := exec.ExecContext(ctx, query, gameID, planetCount)
	if err != nil {
		logger.Error("Failed to update game counts", "error", err)
		return fmt.Errorf("failed to update game counts: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Failed to get rows affected", "error", err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		logger.Warn("Game not found for count update")
		return fmt.Errorf("game not found")
	}

	logger.Debug("Game counts updated", "planets", planetCount)
	return nil
}
