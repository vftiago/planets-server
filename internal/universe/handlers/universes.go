package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"planets-server/internal/universe"
)

type UniverseHandler struct {
	service *universe.Service
	logger  *slog.Logger
}

func NewUniverseHandler(service *universe.Service, logger *slog.Logger) *UniverseHandler {
	return &UniverseHandler{
		service: service,
		logger:  logger,
	}
}

// CreateUniverse handles POST /api/universes - Admin only
func (h *UniverseHandler) CreateUniverse(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With("handler", "create_universe")
	logger.Info("Creating new universe")

	var config universe.UniverseConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		logger.Error("Failed to decode request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate config
	if config.GalaxyCount <= 0 {
		http.Error(w, "Galaxy count must be positive", http.StatusBadRequest)
		return
	}
	if config.SectorsPerGalaxy <= 0 {
		http.Error(w, "Sectors per Galaxy must be positive", http.StatusBadRequest)
		return
	}
	if config.SystemsPerSector <= 0 {
		http.Error(w, "Systems per sector must be positive", http.StatusBadRequest)
		return
	}

	universe, err := h.service.CreateUniverse(config)
	if err != nil {
		logger.Error("Failed to create universe", "error", err)
		http.Error(w, "Failed to create universe", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(universe)
}

// GetUniverses handles GET /api/universes
func (h *UniverseHandler) GetUniverses(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With("handler", "get_universes")
	logger.Debug("Getting all universes")

	universes, err := h.service.ListUniverses()
	if err != nil {
		logger.Error("Failed to get universes", "error", err)
		http.Error(w, "Failed to get universes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(universes)
}

// GetUniverse handles GET /api/universes/{id}
func (h *UniverseHandler) GetUniverse(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With("handler", "get_universe")

	// Extract ID from URL path
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "Universe ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid universe ID", http.StatusBadRequest)
		return
	}

	logger = logger.With("universe_id", id)
	logger.Debug("Getting universe by ID")

	universe, err := h.service.GetUniverse(id)
	if err != nil {
		logger.Error("Failed to get universe", "error", err)
		http.Error(w, "Universe not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(universe)
}

// DeleteUniverse handles DELETE /api/universes/{id} - Admin only
func (h *UniverseHandler) DeleteUniverse(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With("handler", "delete_universe")

	// Extract ID from URL path
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "Universe ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid universe ID", http.StatusBadRequest)
		return
	}

	logger = logger.With("universe_id", id)
	logger.Info("Deleting universe")

	err = h.service.DeleteUniverse(id)
	if err != nil {
		logger.Error("Failed to delete universe", "error", err)
		http.Error(w, "Failed to delete universe", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
