package handlers

import (
	"log/slog"
	"net/http"

	"planets-server/internal/player"
	"planets-server/internal/shared/response"
)

type PlayersHandler struct {
	service *player.Service
}

func NewPlayersHandler(service *player.Service) *PlayersHandler {
	return &PlayersHandler{service: service}
}

func (h *PlayersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := slog.With("handler", "players")

	players, err := h.service.GetAllPlayers(ctx)
	if err != nil {
		response.Error(w, r, logger, err)
		return
	}

	if players == nil {
		players = []player.Player{}
	}

	response.Success(w, http.StatusOK, players)
}
