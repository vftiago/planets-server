package handlers

import (
	"fmt"
	"net/http"
	"net/url"

	"planets-server/internal/shared/config"
)

// redirectWithError redirects to the given base URL (or FRONTEND_URL fallback) with an error code
func redirectWithError(w http.ResponseWriter, r *http.Request, baseURL, errorCode string) {
	if baseURL == "" {
		baseURL = config.GlobalConfig.Frontend.ClientURL
	}
	errorURL := fmt.Sprintf("%s/?error=%s", baseURL, errorCode)

	http.Redirect(w, r, errorURL, http.StatusTemporaryRedirect)
}

// resolveRedirectURI checks the redirect_uri against the allowlist of known origins.
// Returns the validated URI or the default FRONTEND_URL if invalid/missing.
func resolveRedirectURI(rawURI string) string {
	cfg := config.GlobalConfig
	if rawURI == "" {
		return cfg.Frontend.ClientURL
	}

	parsed, err := url.Parse(rawURI)
	if err != nil || parsed.Host == "" {
		return cfg.Frontend.ClientURL
	}

	origin := parsed.Scheme + "://" + parsed.Host

	var allowed []string
	if cfg.Frontend.ClientURL != "" {
		allowed = append(allowed, cfg.Frontend.ClientURL)
	}
	if cfg.Frontend.AdminURL != "" {
		allowed = append(allowed, cfg.Frontend.AdminURL)
	}

	for _, a := range allowed {
		ap, err := url.Parse(a)
		if err != nil {
			continue
		}
		if origin == ap.Scheme+"://"+ap.Host {
			return origin
		}
	}

	return cfg.Frontend.ClientURL
}
