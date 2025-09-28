package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"planets-server/internal/game"
)

type GameHandler struct {
	service *game.Service
}

func NewGameHandler(service *game.Service) *GameHandler {
	return &GameHandler{service: service}
}

func (h *GameHandler) CreateGame(w http.ResponseWriter, r *http.Request) {
	logger := slog.With("handler", "create_game", "remote_addr", r.RemoteAddr)
	logger.Debug("Game creation requested")

	if r.Method != http.MethodPost {
		logger.Warn("Invalid method for game creation", "method", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Game     game.GameConfig     `json:"game"`
		Universe game.UniverseConfig `json:"universe"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logger.Error("Failed to decode request", "error", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required game fields
	if request.Game.Name == "" {
		logger.Error("Game name is required")
		http.Error(w, "Game name is required", http.StatusBadRequest)
		return
	}

	// Set defaults for optional game fields
	if request.Game.MaxPlayers == 0 {
		request.Game.MaxPlayers = 10
	}
	if request.Game.TurnIntervalHours == 0 {
		request.Game.TurnIntervalHours = 1
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

	logger.Info("Creating game with universe",
		"name", request.Game.Name,
		"max_players", request.Game.MaxPlayers,
		"galaxies", request.Universe.GalaxyCount,
		"sectors_per_galaxy", request.Universe.SectorsPerGalaxy,
		"systems_per_sector", request.Universe.SystemsPerSector)

	createdGame, err := h.service.CreateGame(request.Game, request.Universe)
	if err != nil {
		logger.Error("Failed to create game", "error", err)
		http.Error(w, "Failed to create game", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(createdGame); err != nil {
		logger.Error("Failed to encode game response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	logger.Info("Game created successfully",
		"game_id", createdGame.ID,
		"name", createdGame.Name,
		"status", createdGame.Status,
		"galaxies", createdGame.GalaxyCount,
		"sectors", createdGame.SectorCount,
		"systems", createdGame.SystemCount,
		"planets", createdGame.PlanetCount)
}

func (h *GameHandler) GetGames(w http.ResponseWriter, r *http.Request) {
	logger := slog.With("handler", "get_games", "remote_addr", r.RemoteAddr)
	logger.Debug("Games list requested")

	if r.Method != http.MethodGet {
		logger.Warn("Invalid method for get games", "method", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	games, err := h.service.GetAllGames()
	if err != nil {
		logger.Error("Failed to get games", "error", err)
		http.Error(w, "Failed to get games", http.StatusInternalServerError)
		return
	}

	if games == nil {
		games = []game.Game{}
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(games); err != nil {
		logger.Error("Failed to encode games response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	logger.Debug("Games list completed", "game_count", len(games))
}

func (h *GameHandler) GetGameStats(w http.ResponseWriter, r *http.Request) {
	logger := slog.With("handler", "get_game_stats", "remote_addr", r.RemoteAddr)
	logger.Debug("Game stats requested")

	if r.Method != http.MethodGet {
		logger.Warn("Invalid method for get game stats", "method", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	gameIDStr := r.URL.Query().Get("id")
	if gameIDStr == "" {
		logger.Error("Missing game ID parameter")
		http.Error(w, "Game ID is required", http.StatusBadRequest)
		return
	}

	gameID, err := strconv.Atoi(gameIDStr)
	if err != nil {
		logger.Error("Invalid game ID", "game_id", gameIDStr, "error", err)
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}

	stats, err := h.service.GetGameStats(gameID)
	if err != nil {
		logger.Error("Failed to get game stats", "error", err, "game_id", gameID)
		http.Error(w, "Failed to get game stats", http.StatusInternalServerError)
		return
	}

	if stats == nil {
		logger.Debug("Game not found", "game_id", gameID)
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(stats); err != nil {
		logger.Error("Failed to encode game stats response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	logger.Debug("Game stats completed", "game_id", gameID)
}
