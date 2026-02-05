package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"golang.org/x/oauth2"
)

var DiscordEndpoint = oauth2.Endpoint{
	AuthURL:  "https://discord.com/api/oauth2/authorize",
	TokenURL: "https://discord.com/api/oauth2/token",
}

type DiscordUserInfo struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	GlobalName    string `json:"global_name"`
	Email         string `json:"email"`
	Verified      bool   `json:"verified"`
	Avatar        string `json:"avatar"`
	Discriminator string `json:"discriminator"`
}

// AvatarURL returns the full CDN URL for the user's avatar, or empty string if none.
func (u *DiscordUserInfo) AvatarURL() string {
	if u.Avatar == "" {
		return ""
	}
	return fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", u.ID, u.Avatar)
}

// DisplayName returns the global display name, falling back to username.
func (u *DiscordUserInfo) DisplayName() string {
	if u.GlobalName != "" {
		return u.GlobalName
	}
	return u.Username
}

type DiscordProvider struct {
	config *oauth2.Config
}

// NewDiscordProvider creates a new Discord OAuth provider
func NewDiscordProvider(config *oauth2.Config) *DiscordProvider {
	return &DiscordProvider{config: config}
}

// GetUserInfo fetches user information from Discord API
func (p *DiscordProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*DiscordUserInfo, error) {
	client := p.config.Client(ctx, token)

	logger := slog.With("provider", "discord", "operation", "get_user_info")
	logger.Debug("Requesting user info from Discord API")

	resp, err := client.Get("https://discord.com/api/users/@me")
	if err != nil {
		logger.Error("Failed to request user info from Discord", "error", err)
		return nil, fmt.Errorf("failed to request user info from Discord: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Error("Failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		logger.Error("Discord API returned error status",
			"status_code", resp.StatusCode,
			"status", resp.Status)
		return nil, fmt.Errorf("Discord API returned status %d", resp.StatusCode)
	}

	var userInfo DiscordUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		logger.Error("Failed to decode Discord user info", "error", err)
		return nil, fmt.Errorf("failed to decode Discord user info: %w", err)
	}

	if userInfo.ID == "" {
		logger.Error("Discord user info missing user ID")
		return nil, fmt.Errorf("Discord user info missing user ID")
	}

	logger.Debug("Successfully retrieved Discord user info",
		"user_id", userInfo.ID,
		"has_email", userInfo.Email != "",
		"email_verified", userInfo.Verified,
		"has_global_name", userInfo.GlobalName != "",
		"has_avatar", userInfo.Avatar != "")

	return &userInfo, nil
}

// ExchangeCode exchanges an authorization code for tokens
func (p *DiscordProvider) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	logger := slog.With("provider", "discord", "operation", "exchange_code")
	logger.Debug("Exchanging authorization code for Discord access token")

	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		logger.Error("Failed to exchange Discord authorization code", "error", err)
		return nil, fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	logger.Debug("Successfully exchanged code for token")
	return token, nil
}

// GetAuthURL generates the OAuth authorization URL
func (p *DiscordProvider) GetAuthURL(state string) string {
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}
