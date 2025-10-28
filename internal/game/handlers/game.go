package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"planets-server/internal/game"
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
		response.Error(w, r, logger, errors.Validation("method not allowed"))
		return
	}

	var request struct {
		Game     game.GameConfig     `json:"game"`
		Universe game.UniverseConfig `json:"universe"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Error(w, r, logger, errors.WrapValidation("invalid JSON in request body", err))
		return
	}

	// Validate required game fields
	if request.Game.Name == "" {
		response.Error(w, r, logger, errors.Validation("game name is required"))
		return
	}

	// Set defaults for optional game fields
	if request.Game.MaxPlayers == 0 {
		request.Game.MaxPlayers = 10
	}
	if request.Game.TurnIntervalHours == 0 {
		request.Game.TurnIntervalHours = 1
	}
	if request.Game.UniverseName == "" {
		request.Game.UniverseName = "Game Universe"
	}

	// Set defaults for universe configuration
	if request.Universe.GalaxyCount == 0 {
		request.Universe.GalaxyCount = 1
	}
	if request.Universe.SectorsPerGalaxy == 0 {
		request.Universe.SectorsPerGalaxy = 10
	}
	if request.Universe.SystemsPerSector == 0 {
		request.Universe.SystemsPerSector = 10
	}
	if request.Universe.MinPlanetsPerSystem == 0 {
		request.Universe.MinPlanetsPerSystem = 1
	}
	if request.Universe.MaxPlanetsPerSystem == 0 {
		request.Universe.MaxPlanetsPerSystem = 8
	}

	createdGame, err := h.service.CreateGame(ctx, request.Game, request.Universe)
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
		response.Error(w, r, logger, errors.Validation("method not allowed"))
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
		response.Error(w, r, logger, errors.Validation("method not allowed"))
		return
	}

	gameIDStr := r.URL.Query().Get("id")
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
