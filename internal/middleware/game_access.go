package middleware

import (
	"log/slog"
	"net/http"
	"strconv"

	"planets-server/internal/shared/database"
	"planets-server/internal/shared/errors"
	"planets-server/internal/shared/response"
)

type GameAccessMiddleware struct {
	db *database.DB
}

func NewGameAccessMiddleware(db *database.DB) *GameAccessMiddleware {
	return &GameAccessMiddleware{db: db}
}

func (m *GameAccessMiddleware) Require(next http.Handler) http.Handler {
	return JWTMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := slog.With(
			"middleware", "game_access",
			"method", r.Method,
			"path", r.URL.Path,
		)

		claims := GetUserFromContext(r)
		if claims == nil {
			response.Error(w, r, logger, errors.Unauthorized("authentication required"))
			return
		}

		// Admins can access all spatial entities
		if claims.Role == "admin" {
			next.ServeHTTP(w, r)
			return
		}

		// Parse spatial entity ID from path
		entityIDStr := r.PathValue("id")
		if entityIDStr == "" {
			response.Error(w, r, logger, errors.Validation("entity ID is required"))
			return
		}

		entityID, err := strconv.Atoi(entityIDStr)
		if err != nil {
			response.Error(w, r, logger, errors.WrapValidation("invalid entity ID format", err))
			return
		}

		// Look up game_id from spatial_entities table
		var gameID int
		err = m.db.QueryRowContext(r.Context(),
			`SELECT game_id FROM spatial_entities WHERE id = $1`, entityID,
		).Scan(&gameID)
		if err != nil {
			response.Error(w, r, logger, errors.NotFoundf("spatial entity not found with id: %d", entityID))
			return
		}

		// Check if player is a member of the game
		var exists bool
		err = m.db.QueryRowContext(r.Context(),
			`SELECT EXISTS(SELECT 1 FROM game_players WHERE game_id = $1 AND player_id = $2)`,
			gameID, claims.PlayerID,
		).Scan(&exists)
		if err != nil {
			response.Error(w, r, logger, errors.WrapInternal("failed to check game membership", err))
			return
		}

		if !exists {
			response.Error(w, r, logger, errors.Forbidden("game access required"))
			return
		}

		next.ServeHTTP(w, r)
	}))
}
