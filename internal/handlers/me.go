package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"planets-server/internal/middleware"
)

type MeHandler struct{}

func NewMeHandler() *MeHandler {
	return &MeHandler{}
}

func (h *MeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		logger := slog.With("handler", "me", "remote_addr", r.RemoteAddr)
		logger.Error("No user claims found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	logger := slog.With(
		"handler", "me", 
		"remote_addr", r.RemoteAddr,
		"player_id", claims.PlayerID,
		"username", claims.Username,
	)
	logger.Debug("User info requested")
	
	w.Header().Set("Content-Type", "application/json")
	
	response := map[string]interface{}{
		"player_id": claims.PlayerID,
		"username":  claims.Username,
		"email":     claims.Email,
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode user info response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	
	logger.Debug("User info completed")
}
