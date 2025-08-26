package auth

import (
	"log/slog"
	"planets-server/internal/auth/providers"
	"planets-server/internal/shared/config"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

type OAuthConfig struct {
	GitHubConfig     *oauth2.Config
	GoogleConfig     *oauth2.Config
	GitHubProvider   *providers.GitHubProvider
	GoogleProvider   *providers.GoogleProvider
	GitHubConfigured bool
	GoogleConfigured bool
}

func InitOAuth() *OAuthConfig {
	cfg := config.GlobalConfig
	logger := slog.With("component", "oauth", "operation", "init")
	logger.Debug("Initializing OAuth configurations")

	githubConfig := &oauth2.Config{
		ClientID:     cfg.OAuth.GitHub.ClientID,
		ClientSecret: cfg.OAuth.GitHub.ClientSecret,
		RedirectURL:  cfg.OAuth.GitHub.RedirectURL,
		Scopes:       cfg.OAuth.GitHub.Scopes,
		Endpoint:     github.Endpoint,
	}

	googleConfig := &oauth2.Config{
		ClientID:     cfg.OAuth.Google.ClientID,
		ClientSecret: cfg.OAuth.Google.ClientSecret,
		RedirectURL:  cfg.OAuth.Google.RedirectURL,
		Scopes:       cfg.OAuth.Google.Scopes,
		Endpoint:     google.Endpoint,
	}

	githubConfigured := cfg.GitHubOAuthConfigured()
	googleConfigured := cfg.GoogleOAuthConfigured()

	githubProvider := providers.NewGitHubProvider(githubConfig)
	googleProvider := providers.NewGoogleProvider(googleConfig)

	logger.Info("OAuth configuration completed",
		"base_url", cfg.Server.BaseURL,
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
