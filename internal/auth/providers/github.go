package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"golang.org/x/oauth2"
)

type githubAPIResponse struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

type GitHubProvider struct {
	config *oauth2.Config
}

func NewGitHubProvider(config *oauth2.Config) *GitHubProvider {
	return &GitHubProvider{config: config}
}

func (p *GitHubProvider) Name() string { return "github" }

func (p *GitHubProvider) GetAuthURL(state string) string {
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (p *GitHubProvider) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	logger := slog.With("provider", "github", "operation", "exchange_code")
	logger.Debug("Exchanging authorization code for GitHub access token")

	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		logger.Error("Failed to exchange GitHub authorization code", "error", err)
		return nil, fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	logger.Debug("Successfully exchanged code for token")
	return token, nil
}

func (p *GitHubProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*OAuthUser, error) {
	client := p.config.Client(ctx, token)

	logger := slog.With("provider", "github", "operation", "get_user_info")
	logger.Debug("Requesting user info from GitHub API")

	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		logger.Error("Failed to request user info from GitHub", "error", err)
		return nil, fmt.Errorf("failed to request user info from GitHub: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Error("Failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		logger.Error("GitHub API returned error status",
			"status_code", resp.StatusCode,
			"status", resp.Status)
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var raw githubAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		logger.Error("Failed to decode GitHub user info", "error", err)
		return nil, fmt.Errorf("failed to decode GitHub user info: %w", err)
	}

	if raw.ID == 0 {
		logger.Error("GitHub user info missing user ID")
		return nil, fmt.Errorf("GitHub user info missing user ID")
	}

	// GitHub might not return email in the user endpoint, try to get it
	emailVerified := true
	if raw.Email == "" {
		logger.Debug("GitHub user info missing email, attempting to fetch from emails endpoint")
		if err := p.fetchUserEmail(client, &raw); err != nil {
			logger.Warn("Failed to fetch GitHub user email", "error", err)
			emailVerified = false
		}
	}

	logger.Debug("Successfully retrieved GitHub user info",
		"user_id", raw.ID,
		"has_email", raw.Email != "",
		"has_name", raw.Name != "",
		"has_avatar", raw.AvatarURL != "")

	return &OAuthUser{
		ID:            strconv.Itoa(raw.ID),
		Email:         raw.Email,
		EmailVerified: emailVerified,
		Name:          raw.Name,
		AvatarURL:     raw.AvatarURL,
	}, nil
}

// fetchUserEmail attempts to get the user's primary email from GitHub
func (p *GitHubProvider) fetchUserEmail(client *http.Client, raw *githubAPIResponse) error {
	logger := slog.With("provider", "github", "operation", "fetch_email", "github_user_id", raw.ID)

	logger.Debug("Requesting email information from GitHub API")
	emailResp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		logger.Error("Failed to request emails from GitHub", "error", err)
		return fmt.Errorf("failed to request emails from GitHub: %w", err)
	}
	defer func() {
		if err := emailResp.Body.Close(); err != nil {
			logger.Error("Failed to close email response body", "error", err)
		}
	}()

	if emailResp.StatusCode != http.StatusOK {
		logger.Error("GitHub emails API returned error status",
			"status_code", emailResp.StatusCode,
			"status", emailResp.Status)
		return fmt.Errorf("GitHub emails API returned status %d", emailResp.StatusCode)
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.NewDecoder(emailResp.Body).Decode(&emails); err != nil {
		logger.Error("Failed to decode GitHub emails", "error", err)
		return fmt.Errorf("failed to decode GitHub emails: %w", err)
	}

	logger.Debug("Retrieved GitHub emails", "email_count", len(emails))

	// Find primary verified email
	for _, email := range emails {
		if email.Primary && email.Verified {
			raw.Email = email.Email
			logger.Debug("Found primary verified email", "email", email.Email)
			return nil
		}
	}

	// Fallback to any verified email
	for _, email := range emails {
		if email.Verified {
			raw.Email = email.Email
			logger.Debug("Found verified email (not primary)", "email", email.Email)
			return nil
		}
	}

	logger.Warn("No verified email found for GitHub user", "total_emails", len(emails))
	return fmt.Errorf("no verified email found")
}
