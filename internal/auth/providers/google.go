package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"golang.org/x/oauth2"
)

type GoogleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

type GoogleProvider struct {
	config *oauth2.Config
}

// NewGoogleProvider creates a new Google OAuth provider
func NewGoogleProvider(config *oauth2.Config) *GoogleProvider {
	return &GoogleProvider{config: config}
}

// GetUserInfo fetches user information from Google API
func (p *GoogleProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*GoogleUserInfo, error) {
	client := p.config.Client(ctx, token)

	logger := slog.With("provider", "google", "operation", "get_user_info")
	logger.Debug("Requesting user info from Google API")

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		logger.Error("Failed to request user info from Google", "error", err)
		return nil, fmt.Errorf("failed to request user info from Google: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Error("Failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		logger.Error("Google API returned error status",
			"status_code", resp.StatusCode,
			"status", resp.Status)
		return nil, fmt.Errorf("google API returned status %d", resp.StatusCode)
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		logger.Error("Failed to decode Google user info", "error", err)
		return nil, fmt.Errorf("failed to decode Google user info: %w", err)
	}

	// Validate required fields
	if userInfo.ID == "" {
		logger.Error("Google user info missing user ID")
		return nil, fmt.Errorf("google user info missing user ID")
	}
	if userInfo.Email == "" {
		logger.Error("Google user info missing email")
		return nil, fmt.Errorf("google user info missing email")
	}

	logger.Debug("Successfully retrieved Google user info",
		"user_id", userInfo.ID,
		"has_email", userInfo.Email != "",
		"has_name", userInfo.Name != "",
		"has_picture", userInfo.Picture != "")

	return &userInfo, nil
}

// ExchangeCode exchanges an authorization code for tokens
func (p *GoogleProvider) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	logger := slog.With("provider", "google", "operation", "exchange_code")
	logger.Debug("Exchanging authorization code for Google access token")

	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		logger.Error("Failed to exchange Google authorization code", "error", err)
		return nil, fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	logger.Debug("Successfully exchanged code for token")
	return token, nil
}

// GetAuthURL generates the OAuth authorization URL
func (p *GoogleProvider) GetAuthURL(state string) string {
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}
