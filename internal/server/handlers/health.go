package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"planets-server/internal/shared/database"
	"planets-server/internal/shared/response"
)

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Database  string `json:"database"`
}

type HealthHandler struct {
	db *database.DB
}

func NewHealthHandler(db *database.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := slog.With("handler", "health")

	dbStatus := "disconnected"
	if err := h.db.Ping(); err == nil {
		dbStatus = "connected"
	} else {
		logger.Warn("Database ping failed", "error", err)
	}

	resp := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Database:  dbStatus,
	}

	response.Success(w, http.StatusOK, resp)
}
