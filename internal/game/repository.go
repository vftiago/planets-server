package game

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"
)

type Repository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewRepository(db *sql.DB, logger *slog.Logger) *Repository {
	logger.Debug("Initializing game repository")

	return &Repository{
		db:     db,
		logger: logger,
	}
}

func (r *Repository) CreateGame(config GameConfig) (*Game, error) {
	logger := r.logger.With(
		"component", "game_repository",
		"operation", "create_game",
		"name", config.Name,
		"max_players", config.MaxPlayers,
		"universe_id", config.UniverseID,
	)
	logger.Info("Creating new game")

	query := `
		INSERT INTO games (name, description, universe_id, status, current_turn, max_players, turn_interval_hours)
		VALUES ($1, $2, $3, 'creating', 0, $4, $5)
		RETURNING id, universe_id, name, description, status, current_turn, max_players, turn_interval_hours, next_turn_at, created_at, updated_at
	`

	var game Game
	err := r.db.QueryRow(query, config.Name, config.Description, config.UniverseID, config.MaxPlayers, config.TurnIntervalHours).Scan(
		&game.ID,
		&game.UniverseID,
		&game.Name,
		&game.Description,
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

func (r *Repository) GetGameByID(gameID int) (*Game, error) {
	logger := slog.With("component", "game_repository", "operation", "get_game", "game_id", gameID)
	logger.Debug("Getting game by ID")

	query := `
		SELECT id, universe_id, name, description, status, current_turn, max_players, turn_interval_hours, next_turn_at, created_at, updated_at
		FROM games
		WHERE id = $1
	`

	var game Game
	err := r.db.QueryRow(query, gameID).Scan(
		&game.ID,
		&game.UniverseID,
		&game.Name,
		&game.Description,
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

func (r *Repository) GetAllGames() ([]Game, error) {
	logger := slog.With("component", "game_repository", "operation", "get_all_games")
	logger.Debug("Getting all games")

	query := `
		SELECT id, universe_id, name, description, status, current_turn, max_players, turn_interval_hours, next_turn_at, created_at, updated_at
		FROM games
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
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
			&game.UniverseID,
			&game.Name,
			&game.Description,
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

func (r *Repository) UpdateGameStatus(gameID int, status GameStatus) error {
	logger := slog.With("component", "game_repository", "operation", "update_status", "game_id", gameID, "status", status)
	logger.Debug("Updating game status")

	query := `UPDATE games SET status = $1 WHERE id = $2`
	result, err := r.db.Exec(query, status, gameID)
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

func (r *Repository) ActivateGame(gameID int) error {
	logger := slog.With("component", "game_repository", "operation", "activate_game", "game_id", gameID)
	logger.Info("Activating game")

	nextTurnAt := time.Now().Add(1 * time.Hour).Truncate(time.Hour)

	query := `
		UPDATE games 
		SET status = 'active', current_turn = 1, next_turn_at = $1
		WHERE id = $2 AND status = 'creating'
	`

	result, err := r.db.Exec(query, nextTurnAt, gameID)
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

func (r *Repository) GetGameStats(gameID int) (*GameStats, error) {
	logger := slog.With("component", "game_repository", "operation", "get_game_stats", "game_id", gameID)
	logger.Debug("Getting game statistics")

	query := `
		SELECT 
			g.id,
			g.universe_id,
			g.name,
			g.status,
			g.current_turn,
			COALESCE(player_count.count, 0) as player_count,
			g.max_players,
			g.next_turn_at
		FROM games g
		WHERE g.id = $1
	`

	var stats GameStats
	err := r.db.QueryRow(query, gameID).Scan(
		&stats.ID,
		&stats.UniverseID,
		&stats.Name,
		&stats.Status,
		&stats.CurrentTurn,
		&stats.PlayerCount,
		&stats.MaxPlayers,
		&stats.NextTurnAt,
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

func (r *Repository) DeleteGame(gameID int) error {
	logger := slog.With("component", "game_repository", "operation", "delete_game", "game_id", gameID)
	logger.Info("Deleting game and all related data")

	query := `DELETE FROM games WHERE id = $1`
	result, err := r.db.Exec(query, gameID)
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
