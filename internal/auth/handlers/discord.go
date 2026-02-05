package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"planets-server/internal/auth"
	"planets-server/internal/auth/providers"
	"planets-server/internal/player"
	"planets-server/internal/shared/config"
	"planets-server/internal/shared/cookies"
	"planets-server/internal/shared/errors"
	"planets-server/internal/shared/response"
)

type DiscordAuthHandler struct {
	provider      *providers.DiscordProvider
	playerService *player.Service
	authService   *auth.Service
	isConfigured  bool
}

func NewDiscordAuthHandler(provider *providers.DiscordProvider, playerService *player.Service, authService *auth.Service, isConfigured bool) *DiscordAuthHandler {
	return &DiscordAuthHandler{
		provider:      provider,
		playerService: playerService,
		authService:   authService,
		isConfigured:  isConfigured,
	}
}

func (h *DiscordAuthHandler) HandleAuth(w http.ResponseWriter, r *http.Request) {
	logger := slog.With("handler", "discord_oauth_init")

	if !h.isConfigured {
		response.Error(w, r, logger, errors.External("Discord OAuth is not properly configured"))
		return
	}

	state, err := auth.GenerateOAuthState("discord", r.UserAgent())
	if err != nil {
		response.Error(w, r, logger, errors.WrapInternal("failed to initialize OAuth flow", err))
		return
	}

	url := h.provider.GetAuthURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *DiscordAuthHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")

	logger := slog.With(
		"handler", "discord_oauth_callback",
		"user_agent", r.UserAgent(),
		"ip", r.RemoteAddr,
		"has_code", code != "",
		"has_state", state != "",
	)

	if errorParam != "" {
		logger.Warn("Discord OAuth authorization denied",
			"oauth_error", errorParam,
			"error_description", r.URL.Query().Get("error_description"))
		redirectWithError(w, r, "oauth_denied")
		return
	}

	if code == "" {
		logger.Error("Discord OAuth callback missing authorization code")
		redirectWithError(w, r, "oauth_error")
		return
	}

	if err := auth.ValidateOAuthState(state, "discord", r.UserAgent()); err != nil {
		logger.Error("OAuth state validation failed",
			"error", err,
			"provider", "discord",
			"state", state)
		redirectWithError(w, r, "oauth_error")
		return
	}

	logger.Info("OAuth state validation successful - proceeding with Discord OAuth callback")

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	token, err := h.provider.ExchangeCode(ctx, code)
	if err != nil {
		logger.Error("Failed to exchange Discord authorization code",
			"error", err,
			"provider", "discord")
		redirectWithError(w, r, "oauth_error")
		return
	}

	logger.Debug("Fetching user information from Discord API")
	userInfo, err := h.provider.GetUserInfo(ctx, token)
	if err != nil {
		logger.Error("Failed to get user info from Discord",
			"error", err,
			"provider", "discord")
		redirectWithError(w, r, "oauth_error")
		return
	}

	userLogger := logger.With(
		"user_email", userInfo.Email,
		"discord_user_id", userInfo.ID,
		"user_name", userInfo.DisplayName())

	if userInfo.Email == "" || !userInfo.Verified {
		userLogger.Error("Discord user missing verified email")
		redirectWithError(w, r, "oauth_error")
		return
	}

	userLogger.Info("Creating or finding player account for Discord user")

	existingPlayerID, err := h.authService.FindPlayerByAuthProvider(ctx, "discord", userInfo.ID)
	if err != nil && errors.GetType(err) != errors.ErrorTypeNotFound {
		userLogger.Error("Database error checking auth provider", "error", err)
		redirectWithError(w, r, "database_error")
		return
	}

	var player *player.Player
	if existingPlayerID > 0 {
		userLogger.Debug("Found existing player via OAuth provider")
		player, err = h.playerService.GetPlayerByID(ctx, existingPlayerID)
		if err != nil {
			userLogger.Error("Failed to get existing player", "error", err)
			redirectWithError(w, r, "database_error")
			return
		}
	} else {
		userLogger.Debug("No existing OAuth link found, finding or creating player by email")
		avatarURL := userInfo.AvatarURL()
		player, err = h.playerService.FindOrCreatePlayerByOAuth(
			ctx,
			"discord",
			userInfo.ID,
			userInfo.Email,
			userInfo.DisplayName(),
			&avatarURL,
		)
		if err != nil {
			userLogger.Error("Failed to create player", "error", err)
			redirectWithError(w, r, "database_error")
			return
		}

		userLogger.Debug("Linking OAuth provider to player account")
		err = h.authService.CreateAuthProvider(ctx, player.ID, "discord", userInfo.ID, userInfo.Email)
		if err != nil {
			userLogger.Error("Failed to create auth provider link", "error", err)
			redirectWithError(w, r, "database_error")
			return
		}
	}

	playerLogger := userLogger.With("player_id", player.ID)

	playerLogger.Debug("Generating JWT token for player")
	jwtToken, err := auth.GenerateJWT(player.ID, player.Username, player.Email, player.Role.String())
	if err != nil {
		playerLogger.Error("Failed to generate JWT token", "error", err)
		redirectWithError(w, r, "auth_error")
		return
	}

	cookies.SetAuthCookie(w, jwtToken)

	playerLogger.Info("Discord OAuth authentication successful",
		"provider", "discord",
		"player_username", player.Username,
		"player_role", player.Role)

	cfg := config.GlobalConfig
	successURL := fmt.Sprintf("%s/auth/callback?success=true", cfg.Frontend.URL)
	http.Redirect(w, r, successURL, http.StatusTemporaryRedirect)
}
