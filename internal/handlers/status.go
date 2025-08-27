package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"planets-server/internal/player"
)

type GameStatusResponse struct {
	Game          string `json:"game"`
	Turn          int    `json:"turn"`
	OnlinePlayers int    `json:"online_players"`
}

type GameStatusHandler struct {
	playerService *player.Service
}

func NewGameStatusHandler(playerService *player.Service) *GameStatusHandler {
	return &GameStatusHandler{playerService: playerService}
}

func (h *GameStatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := slog.With("handler", "game_status", "remote_addr", r.RemoteAddr)
	logger.Debug("Game status requested")

	w.Header().Set("Content-Type", "application/json")

	playerCount, err := h.playerService.GetPlayerCount()
	if err != nil {
		logger.Warn("Failed to get player count", "error", err)
		playerCount = 0
	}

	response := GameStatusResponse{
		Game:          "Planets!",
		Turn:          1,
		OnlinePlayers: playerCount,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode game status response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	logger.Debug("Game status completed", "player_count", playerCount)
}
