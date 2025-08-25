package handlers

import (
	"log/slog"
	"net/http"

	"planets-server/internal/utils"
)

type LogoutHandler struct{}

func NewLogoutHandler() *LogoutHandler {
	return &LogoutHandler{}
}

func (h *LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := slog.With("handler", "logout", "remote_addr", r.RemoteAddr)
	logger.Debug("Logout requested")
	
	// Clear the auth cookie using the utility function
	utils.ClearAuthCookie(w)
	
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Logged out")); err != nil {
		logger.Error("Failed to write logout response", "error", err)
		return
	}
	
	logger.Info("User logged out successfully")
}
