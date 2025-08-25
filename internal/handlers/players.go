package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"planets-server/internal/models"
)

type PlayersHandler struct {
	playerRepo *models.PlayerRepository
}

func NewPlayersHandler(playerRepo *models.PlayerRepository) *PlayersHandler {
	return &PlayersHandler{playerRepo: playerRepo}
}

func (h *PlayersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := slog.With("handler", "players", "remote_addr", r.RemoteAddr)
	logger.Debug("Players list requested")
	
	w.Header().Set("Content-Type", "application/json")
	
	players, err := h.playerRepo.GetAllPlayers()
	if err != nil {
		logger.Error("Failed to fetch players", "error", err)
		http.Error(w, "Failed to fetch players", http.StatusInternalServerError)
		return
	}

	if players == nil {
		players = []models.Player{}
	}
	
	if err := json.NewEncoder(w).Encode(players); err != nil {
		logger.Error("Failed to encode players response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	
	logger.Debug("Players list completed", "player_count", len(players))
}
