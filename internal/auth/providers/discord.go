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

type discordAPIResponse struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	GlobalName    string `json:"global_name"`
	Email         string `json:"email"`
	Verified      bool   `json:"verified"`
	Avatar        string `json:"avatar"`
	Discriminator string `json:"discriminator"`
}

func (u *discordAPIResponse) avatarURL() string {
	if u.Avatar == "" {
		return ""
	}
	return fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", u.ID, u.Avatar)
}

func (u *discordAPIResponse) displayName() string {
	if u.GlobalName != "" {
		return u.GlobalName
	}
	return u.Username
}

type DiscordProvider struct {
	config *oauth2.Config
}

func NewDiscordProvider(config *oauth2.Config) *DiscordProvider {
	return &DiscordProvider{config: config}
}

func (p *DiscordProvider) Name() string { return "discord" }

func (p *DiscordProvider) GetAuthURL(state string) string {
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

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

func (p *DiscordProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*OAuthUser, error) {
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
		return nil, fmt.Errorf("discord API returned status %d", resp.StatusCode)
	}

	var raw discordAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		logger.Error("Failed to decode Discord user info", "error", err)
		return nil, fmt.Errorf("failed to decode Discord user info: %w", err)
	}

	if raw.ID == "" {
		logger.Error("Discord user info missing user ID")
		return nil, fmt.Errorf("discord user info missing user ID")
	}

	logger.Debug("Successfully retrieved Discord user info",
		"user_id", raw.ID,
		"has_email", raw.Email != "",
		"email_verified", raw.Verified,
		"has_global_name", raw.GlobalName != "",
		"has_avatar", raw.Avatar != "")

	return &OAuthUser{
		ID:            raw.ID,
		Email:         raw.Email,
		EmailVerified: raw.Verified,
		Name:          raw.displayName(),
		AvatarURL:     raw.avatarURL(),
	}, nil
}
