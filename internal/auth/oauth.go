package auth

import (
	"log/slog"
	"planets-server/internal/auth/providers"
	"planets-server/internal/utils"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

type OAuthConfig struct {
	GitHubConfig    *oauth2.Config
	GoogleConfig    *oauth2.Config
	GitHubProvider  *providers.GitHubProvider
	GoogleProvider  *providers.GoogleProvider
	GitHubConfigured bool
	GoogleConfigured bool
}

// InitOAuth initializes OAuth configuration and returns the config
func InitOAuth() *OAuthConfig {
	logger := slog.With("component", "oauth", "operation", "init")
	logger.Debug("Initializing OAuth configurations")

	baseURL := utils.GetEnv("BASE_URL", "http://localhost:8080")

	// Initialize GitHub OAuth
	githubConfig := &oauth2.Config{
		ClientID:     utils.GetEnv("GITHUB_CLIENT_ID", ""),
		ClientSecret: utils.GetEnv("GITHUB_CLIENT_SECRET", ""),
		RedirectURL:  baseURL + "/auth/github/callback",
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}

	// Initialize Google OAuth
	googleConfig := &oauth2.Config{
		ClientID:     utils.GetEnv("GOOGLE_CLIENT_ID", ""),
		ClientSecret: utils.GetEnv("GOOGLE_CLIENT_SECRET", ""),
		RedirectURL:  baseURL + "/auth/google/callback",
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint:     google.Endpoint,
	}

	// Check configuration status
	githubConfigured := githubConfig.ClientID != "" && githubConfig.ClientSecret != ""
	googleConfigured := googleConfig.ClientID != "" && googleConfig.ClientSecret != ""

	// Create providers
	githubProvider := providers.NewGitHubProvider(githubConfig)
	googleProvider := providers.NewGoogleProvider(googleConfig)

	logger.Info("OAuth configuration completed",
		"base_url", baseURL,
		"github_configured", githubConfigured,
		"google_configured", googleConfigured,
		"github_redirect", githubConfig.RedirectURL,
		"google_redirect", googleConfig.RedirectURL,
	)

	if !githubConfigured {
		logger.Warn("GitHub OAuth not configured - missing client credentials")
	}
	if !googleConfigured {
		logger.Warn("Google OAuth not configured - missing client credentials")
	}

	return &OAuthConfig{
		GitHubConfig:     githubConfig,
		GoogleConfig:     googleConfig,
		GitHubProvider:   githubProvider,
		GoogleProvider:   googleProvider,
		GitHubConfigured: githubConfigured,
		GoogleConfigured: googleConfigured,
	}
}
