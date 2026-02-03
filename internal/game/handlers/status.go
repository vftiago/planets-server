package handlers

import (
	"log/slog"
	"net/http"

	"planets-server/internal/player"
	"planets-server/internal/shared/errors"
	"planets-server/internal/shared/response"
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
	ctx := r.Context()
	logger := slog.With("handler", "game_status")

	playerCount, err := h.playerService.GetPlayerCount(ctx)
	if err != nil {
		response.Error(w, r, logger, errors.WrapInternal("failed to get player count", err))
		return
	}

	resp := GameStatusResponse{
		Game:          "Planets!",
		Turn:          1,
		OnlinePlayers: playerCount,
	}

	response.Success(w, http.StatusOK, resp)
}
