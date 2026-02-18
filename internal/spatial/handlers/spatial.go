package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"planets-server/internal/shared/errors"
	"planets-server/internal/shared/response"
	"planets-server/internal/spatial"
)

type SpatialHandler struct {
	service *spatial.Service
}

func NewSpatialHandler(service *spatial.Service) *SpatialHandler {
	return &SpatialHandler{service: service}
}

func (h *SpatialHandler) GetChildren(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := slog.With("handler", "get_children")

	if r.Method != http.MethodGet {
		response.Error(w, r, logger, errors.MethodNotAllowed(r.Method))
		return
	}

	entityIDStr := r.PathValue("id")
	if entityIDStr == "" {
		response.Error(w, r, logger, errors.Validation("entity ID is required"))
		return
	}

	entityID, err := strconv.Atoi(entityIDStr)
	if err != nil {
		response.Error(w, r, logger, errors.WrapValidation("invalid entity ID format", err))
		return
	}

	children, err := h.service.GetChildren(ctx, entityID)
	if err != nil {
		response.Error(w, r, logger, err)
		return
	}

	if children == nil {
		children = []spatial.SpatialEntity{}
	}

	response.Success(w, http.StatusOK, children)
}

func (h *SpatialHandler) GetAncestors(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := slog.With("handler", "get_ancestors")

	if r.Method != http.MethodGet {
		response.Error(w, r, logger, errors.MethodNotAllowed(r.Method))
		return
	}

	entityIDStr := r.PathValue("id")
	if entityIDStr == "" {
		response.Error(w, r, logger, errors.Validation("entity ID is required"))
		return
	}

	entityID, err := strconv.Atoi(entityIDStr)
	if err != nil {
		response.Error(w, r, logger, errors.WrapValidation("invalid entity ID format", err))
		return
	}

	ancestors, err := h.service.GetAncestors(ctx, entityID)
	if err != nil {
		response.Error(w, r, logger, err)
		return
	}

	if ancestors == nil {
		ancestors = []spatial.SpatialEntity{}
	}

	response.Success(w, http.StatusOK, ancestors)
}
