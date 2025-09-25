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

	var config game.GameConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		logger.Error("Failed to decode game config", "error", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if config.Name == "" {
		config.Name = "New Game"
	}
	if config.MaxPlayers == 0 {
		config.MaxPlayers = 10
	}
	if config.TurnIntervalHours == 0 {
		config.TurnIntervalHours = 1
	}
	if config.UniverseConfig.GalaxyName == "" {
		config.UniverseConfig.GalaxyName = "Andromeda"
	}
	if config.UniverseConfig.SectorCount == 0 {
		config.UniverseConfig.SectorCount = 16
	}
	if config.UniverseConfig.SystemsPerSector == 0 {
		config.UniverseConfig.SystemsPerSector = 16
	}
	if config.UniverseConfig.MinPlanetsPerSystem == 0 {
		config.UniverseConfig.MinPlanetsPerSystem = 3
	}
	if config.UniverseConfig.MaxPlanetsPerSystem == 0 {
		config.UniverseConfig.MaxPlanetsPerSystem = 12
	}

	logger.Info("Creating game", 
		"name", config.Name,
		"max_players", config.MaxPlayers,
		"sector_count", config.UniverseConfig.SectorCount)

	createdGame, err := h.service.CreateGame(config)
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
		"status", createdGame.Status)
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

func (h *GameHandler) GetCurrentGame(w http.ResponseWriter, r *http.Request) {
	logger := slog.With("handler", "get_current_game", "remote_addr", r.RemoteAddr)
	logger.Debug("Current game requested")

	if r.Method != http.MethodGet {
		logger.Warn("Invalid method for get current game", "method", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentGame, err := h.service.GetCurrentGame()
	if err != nil {
		logger.Error("Failed to get current game", "error", err)
		http.Error(w, "Failed to get current game", http.StatusInternalServerError)
		return
	}

	if currentGame == nil {
		logger.Debug("No current game found")
		http.Error(w, "No game found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(currentGame); err != nil {
		logger.Error("Failed to encode current game response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	logger.Debug("Current game response completed", "game_id", currentGame.ID, "status", currentGame.Status)
}

func (h *GameHandler) GetCurrentGameStats(w http.ResponseWriter, r *http.Request) {
	logger := slog.With("handler", "get_current_game_stats", "remote_addr", r.RemoteAddr)
	logger.Debug("Current game stats requested")

	if r.Method != http.MethodGet {
		logger.Warn("Invalid method for get current game stats", "method", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentGame, err := h.service.GetCurrentGame()
	if err != nil {
		logger.Error("Failed to get current game", "error", err)
		http.Error(w, "Failed to get current game", http.StatusInternalServerError)
		return
	}

	if currentGame == nil {
		logger.Debug("No current game found for stats")
		http.Error(w, "No game found", http.StatusNotFound)
		return
	}

	stats, err := h.service.GetGameStats(currentGame.ID)
	if err != nil {
		logger.Error("Failed to get game stats", "error", err, "game_id", currentGame.ID)
		http.Error(w, "Failed to get game stats", http.StatusInternalServerError)
		return
	}

	if stats == nil {
		logger.Debug("Game stats not found", "game_id", currentGame.ID)
		http.Error(w, "Game stats not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(stats); err != nil {
		logger.Error("Failed to encode game stats response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	logger.Debug("Current game stats completed", 
		"game_id", currentGame.ID,
		"planets", stats.PlanetCount,
		"systems", stats.SystemCount)
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

	logger.Debug("Game stats completed", 
		"game_id", gameID,
		"planets", stats.PlanetCount,
		"systems", stats.SystemCount)
}
