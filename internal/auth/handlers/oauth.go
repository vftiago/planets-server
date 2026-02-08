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
	"planets-server/internal/shared/cookies"
	"planets-server/internal/shared/errors"
	"planets-server/internal/shared/response"
)

type OAuthHandler struct {
	provider      providers.OAuthProvider
	playerService *player.Service
	authService   *auth.Service
	isConfigured  bool
}

func NewOAuthHandler(provider providers.OAuthProvider, playerService *player.Service, authService *auth.Service, isConfigured bool) *OAuthHandler {
	return &OAuthHandler{
		provider:      provider,
		playerService: playerService,
		authService:   authService,
		isConfigured:  isConfigured,
	}
}

func (h *OAuthHandler) HandleAuth(w http.ResponseWriter, r *http.Request) {
	name := h.provider.Name()
	logger := slog.With("handler", name+"_oauth_init")

	if !h.isConfigured {
		response.Error(w, r, logger, errors.External(fmt.Sprintf("%s OAuth is not properly configured", name)))
		return
	}

	redirectURI := resolveRedirectURI(r.URL.Query().Get("redirect_uri"))

	state, err := auth.GenerateOAuthState(name, r.UserAgent(), redirectURI)
	if err != nil {
		response.Error(w, r, logger, errors.WrapInternal("failed to initialize OAuth flow", err))
		return
	}

	authURL := h.provider.GetAuthURL(state)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (h *OAuthHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	name := h.provider.Name()
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")

	logger := slog.With(
		"handler", name+"_oauth_callback",
		"user_agent", r.UserAgent(),
		"ip", r.RemoteAddr,
		"has_code", code != "",
		"has_state", state != "",
	)

	// Try to recover redirect URI from state even in early-exit cases.
	// Falls back to FRONTEND_CLIENT_URL if state is missing or invalid.
	redirectURI := ""
	if state != "" {
		if entry, err := auth.ValidateOAuthState(state, name, r.UserAgent()); err == nil {
			redirectURI = entry.RedirectURI
		}
	}

	if errorParam != "" {
		logger.Warn("OAuth authorization denied",
			"provider", name,
			"oauth_error", errorParam,
			"error_description", r.URL.Query().Get("error_description"))
		redirectWithError(w, r, redirectURI, "oauth_denied")
		return
	}

	if code == "" {
		logger.Error("OAuth callback missing authorization code", "provider", name)
		redirectWithError(w, r, redirectURI, "oauth_error")
		return
	}
	logger.Info("OAuth state validation successful - proceeding with OAuth callback", "provider", name)

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	token, err := h.provider.ExchangeCode(ctx, code)
	if err != nil {
		logger.Error("Failed to exchange authorization code",
			"error", err,
			"provider", name)
		redirectWithError(w, r, redirectURI, "oauth_error")
		return
	}

	logger.Debug("Fetching user information from provider API", "provider", name)
	userInfo, err := h.provider.GetUserInfo(ctx, token)
	if err != nil {
		logger.Error("Failed to get user info",
			"error", err,
			"provider", name)
		redirectWithError(w, r, redirectURI, "oauth_error")
		return
	}

	userLogger := logger.With(
		"user_email", userInfo.Email,
		"provider_user_id", userInfo.ID,
		"user_name", userInfo.Name)

	if userInfo.Email == "" || !userInfo.EmailVerified {
		userLogger.Error("User missing verified email", "provider", name)
		redirectWithError(w, r, redirectURI, "oauth_error")
		return
	}

	userLogger.Info("Creating or finding player account", "provider", name)

	existingPlayerID, err := h.authService.FindPlayerByAuthProvider(ctx, name, userInfo.ID)
	if err != nil && errors.GetType(err) != errors.ErrorTypeNotFound {
		userLogger.Error("Database error checking auth provider", "error", err)
		redirectWithError(w, r, redirectURI, "database_error")
		return
	}

	var p *player.Player
	if existingPlayerID > 0 {
		userLogger.Debug("Found existing player via OAuth provider")
		p, err = h.playerService.GetPlayerByID(ctx, existingPlayerID)
		if err != nil {
			userLogger.Error("Failed to get existing player", "error", err)
			redirectWithError(w, r, redirectURI, "database_error")
			return
		}
	} else {
		userLogger.Debug("No existing OAuth link found, finding or creating player by email")
		p, err = h.playerService.FindOrCreatePlayerByOAuth(
			ctx,
			name,
			userInfo.ID,
			userInfo.Email,
			userInfo.Name,
			&userInfo.AvatarURL,
		)
		if err != nil {
			userLogger.Error("Failed to create player", "error", err)
			redirectWithError(w, r, redirectURI, "database_error")
			return
		}

		userLogger.Debug("Linking OAuth provider to player account")
		err = h.authService.CreateAuthProvider(ctx, p.ID, name, userInfo.ID, userInfo.Email)
		if err != nil {
			userLogger.Error("Failed to create auth provider link", "error", err)
			redirectWithError(w, r, redirectURI, "database_error")
			return
		}
	}

	playerLogger := userLogger.With("player_id", p.ID)

	playerLogger.Debug("Generating JWT token for player")
	jwtToken, err := auth.GenerateJWT(p.ID, p.Username, p.Email, p.Role.String())
	if err != nil {
		playerLogger.Error("Failed to generate JWT token", "error", err)
		redirectWithError(w, r, redirectURI, "auth_error")
		return
	}

	cookies.SetAuthCookie(w, jwtToken)

	playerLogger.Info("OAuth authentication successful",
		"provider", name,
		"player_username", p.Username,
		"player_role", p.Role)

	successURL := fmt.Sprintf("%s/auth/callback?success=true", redirectURI)
	http.Redirect(w, r, successURL, http.StatusTemporaryRedirect)
}
