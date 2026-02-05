package handlers

import (
	"fmt"
	"net/http"

	"planets-server/internal/shared/config"
)

// redirectWithError redirects to frontend with an error code
func redirectWithError(w http.ResponseWriter, r *http.Request, errorCode string) {
	cfg := config.GlobalConfig
	errorURL := fmt.Sprintf("%s/auth/error?error=%s", cfg.Frontend.URL, errorCode)

	http.Redirect(w, r, errorURL, http.StatusTemporaryRedirect)
}
