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
	GoogleProvider    providers.OAuthProvider
	GitHubProvider    providers.OAuthProvider
	DiscordProvider   providers.OAuthProvider
	GoogleConfigured  bool
	GitHubConfigured  bool
	DiscordConfigured bool
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

	discordConfig := &oauth2.Config{
		ClientID:     cfg.OAuth.Discord.ClientID,
		ClientSecret: cfg.OAuth.Discord.ClientSecret,
		RedirectURL:  cfg.OAuth.Discord.RedirectURL,
		Scopes:       cfg.OAuth.Discord.Scopes,
		Endpoint:     providers.DiscordEndpoint,
	}

	githubConfigured := cfg.GitHubOAuthConfigured()
	googleConfigured := cfg.GoogleOAuthConfigured()
	discordConfigured := cfg.DiscordOAuthConfigured()

	logger.Info("OAuth configuration completed",
		"server_url", cfg.Server.URL,
		"github_configured", githubConfigured,
		"google_configured", googleConfigured,
		"discord_configured", discordConfigured,
		"github_redirect", githubConfig.RedirectURL,
		"google_redirect", googleConfig.RedirectURL,
		"discord_redirect", discordConfig.RedirectURL,
	)

	if !githubConfigured {
		logger.Warn("GitHub OAuth not configured - missing client credentials")
	}
	if !googleConfigured {
		logger.Warn("Google OAuth not configured - missing client credentials")
	}
	if !discordConfigured {
		logger.Warn("Discord OAuth not configured - missing client credentials")
	}

	return &OAuthConfig{
		GoogleProvider:    providers.NewGoogleProvider(googleConfig),
		GitHubProvider:    providers.NewGitHubProvider(githubConfig),
		DiscordProvider:   providers.NewDiscordProvider(discordConfig),
		GoogleConfigured:  googleConfigured,
		GitHubConfigured:  githubConfigured,
		DiscordConfigured: discordConfigured,
	}
}
