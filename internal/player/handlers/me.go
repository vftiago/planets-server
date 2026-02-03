package handlers

import (
	"log/slog"
	"net/http"

	"planets-server/internal/middleware"
	"planets-server/internal/shared/errors"
	"planets-server/internal/shared/response"
)

type MeHandler struct{}

func NewMeHandler() *MeHandler {
	return &MeHandler{}
}

func (h *MeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := slog.With("handler", "me")

	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		response.Error(w, r, logger, errors.Unauthorized("no user claims found in context"))
		return
	}

	resp := map[string]interface{}{
		"player_id": claims.PlayerID,
		"username":  claims.Username,
		"email":     claims.Email,
		"role":      claims.Role,
	}

	response.Success(w, http.StatusOK, resp)
}
