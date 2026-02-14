package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"planets-server/internal/game"
	appconfig "planets-server/internal/shared/config"
	"planets-server/internal/shared/errors"
	"planets-server/internal/shared/response"
)

type GameHandler struct {
	service *game.Service
}

func NewGameHandler(service *game.Service) *GameHandler {
	return &GameHandler{service: service}
}

func (h *GameHandler) CreateGame(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := slog.With("handler", "create_game")

	if r.Method != http.MethodPost {
		response.Error(w, r, logger, errors.MethodNotAllowed(r.Method))
		return
	}

	defaults := appconfig.GlobalConfig.Game

	gameConfig := game.GameConfig{
		MaxPlayers:          defaults.MaxPlayers,
		TurnIntervalHours:   defaults.TurnIntervalHours,
		GalaxyCount:         defaults.GalaxyCount,
		SectorsPerGalaxy:    defaults.SectorsPerGalaxy,
		SystemsPerSector:    defaults.SystemsPerSector,
		MinPlanetsPerSystem: defaults.MinPlanetsPerSystem,
		MaxPlanetsPerSystem: defaults.MaxPlanetsPerSystem,
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB
	if err := json.NewDecoder(r.Body).Decode(&gameConfig); err != nil {
		response.Error(w, r, logger, errors.WrapValidation("invalid JSON in request body", err))
		return
	}

	createdGame, err := h.service.CreateGame(ctx, gameConfig)
	if err != nil {
		response.Error(w, r, logger, err)
		return
	}

	response.Success(w, http.StatusCreated, createdGame)
}

func (h *GameHandler) GetGames(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := slog.With("handler", "get_games")

	if r.Method != http.MethodGet {
		response.Error(w, r, logger, errors.MethodNotAllowed(r.Method))
		return
	}

	games, err := h.service.GetAllGames(ctx)
	if err != nil {
		response.Error(w, r, logger, err)
		return
	}

	if games == nil {
		games = []game.Game{}
	}

	response.Success(w, http.StatusOK, games)
}

func (h *GameHandler) GetGameStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := slog.With("handler", "get_game_stats")

	if r.Method != http.MethodGet {
		response.Error(w, r, logger, errors.MethodNotAllowed(r.Method))
		return
	}

	gameIDStr := r.PathValue("id")
	if gameIDStr == "" {
		response.Error(w, r, logger, errors.Validation("game ID is required"))
		return
	}

	gameID, err := strconv.Atoi(gameIDStr)
	if err != nil {
		response.Error(w, r, logger, errors.WrapValidation("invalid game ID format", err))
		return
	}

	stats, err := h.service.GetGameStats(ctx, gameID)
	if err != nil {
		response.Error(w, r, logger, err)
		return
	}

	response.Success(w, http.StatusOK, stats)
}
