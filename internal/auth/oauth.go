package auth

import (
	"log/slog"
	"planets-server/internal/utils"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

var GitHubOAuthConfig *oauth2.Config
var GoogleOAuthConfig *oauth2.Config

func InitOAuth() {
	logger := slog.With("component", "oauth", "operation", "init")
	logger.Debug("Initializing OAuth configurations")

	baseURL := utils.GetEnv("BASE_URL", "http://localhost:8080")

	// Initialize GitHub OAuth
	GitHubOAuthConfig = &oauth2.Config{
		ClientID:     utils.GetEnv("GITHUB_CLIENT_ID", ""),
		ClientSecret: utils.GetEnv("GITHUB_CLIENT_SECRET", ""),
		RedirectURL:  baseURL + "/auth/github/callback",
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}

	// Initialize Google OAuth
	GoogleOAuthConfig = &oauth2.Config{
		ClientID:     utils.GetEnv("GOOGLE_CLIENT_ID", ""),
		ClientSecret: utils.GetEnv("GOOGLE_CLIENT_SECRET", ""),
		RedirectURL:  baseURL + "/auth/google/callback",
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint:     google.Endpoint,
	}

	// Log configuration status
	githubConfigured := GitHubOAuthConfig.ClientID != "" && GitHubOAuthConfig.ClientSecret != ""
	googleConfigured := GoogleOAuthConfig.ClientID != "" && GoogleOAuthConfig.ClientSecret != ""

	logger.Info("OAuth configuration completed",
		"base_url", baseURL,
		"github_configured", githubConfigured,
		"google_configured", googleConfigured,
		"github_redirect", GitHubOAuthConfig.RedirectURL,
		"google_redirect", GoogleOAuthConfig.RedirectURL,
	)

	if !githubConfigured {
		logger.Warn("GitHub OAuth not configured - missing client credentials")
	}
	if !googleConfigured {
		logger.Warn("Google OAuth not configured - missing client credentials")
	}
}
