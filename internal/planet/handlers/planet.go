package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"planets-server/internal/planet"
	"planets-server/internal/shared/errors"
	"planets-server/internal/shared/response"
)

type PlanetHandler struct {
	service *planet.Service
}

func NewPlanetHandler(service *planet.Service) *PlanetHandler {
	return &PlanetHandler{service: service}
}

func (h *PlanetHandler) GetBySystemID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := slog.With("handler", "get_planets_by_system")

	if r.Method != http.MethodGet {
		response.Error(w, r, logger, errors.MethodNotAllowed(r.Method))
		return
	}

	systemIDStr := r.PathValue("id")
	if systemIDStr == "" {
		response.Error(w, r, logger, errors.Validation("system ID is required"))
		return
	}

	systemID, err := strconv.Atoi(systemIDStr)
	if err != nil {
		response.Error(w, r, logger, errors.WrapValidation("invalid system ID format", err))
		return
	}

	planets, err := h.service.GetBySystemID(ctx, systemID)
	if err != nil {
		response.Error(w, r, logger, err)
		return
	}

	if planets == nil {
		planets = []planet.Planet{}
	}

	response.Success(w, http.StatusOK, planets)
}
