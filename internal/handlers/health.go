package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"planets-server/internal/database"
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
	logger := slog.With("handler", "health", "remote_addr", r.RemoteAddr)
	logger.Debug("Health check requested")
	
	w.Header().Set("Content-Type", "application/json")
	
	dbStatus := "disconnected"
	if err := h.db.Ping(); err == nil {
		dbStatus = "connected"
	} else {
		logger.Warn("Database ping failed", "error", err)
	}
	
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Database:  dbStatus,
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode health response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	
	logger.Debug("Health check completed", "db_status", dbStatus)
}
