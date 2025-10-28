package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"planets-server/internal/player"
)

type PlayersHandler struct {
	service *player.Service
}

func NewPlayersHandler(service *player.Service) *PlayersHandler {
	return &PlayersHandler{service: service}
}

func (h *PlayersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := slog.With("handler", "players", "remote_addr", r.RemoteAddr)
	logger.Debug("Players list requested")

	w.Header().Set("Content-Type", "application/json")

	players, err := h.service.GetAllPlayers(ctx)
	if err != nil {
		logger.Error("Failed to fetch players", "error", err)
		http.Error(w, "Failed to fetch players", http.StatusInternalServerError)
		return
	}

	if players == nil {
		players = []player.Player{}
	}

	if err := json.NewEncoder(w).Encode(players); err != nil {
		logger.Error("Failed to encode players response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	logger.Debug("Players list completed", "player_count", len(players))
}
