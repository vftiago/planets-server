package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"planets-server/internal/auth"
	"planets-server/internal/auth/providers"
	"planets-server/internal/player"
	"planets-server/internal/utils"
)

type GitHubAuthHandler struct {
	provider        *providers.GitHubProvider
	playerService   *player.Service
	authService     *auth.Service
	isConfigured    bool
}

// NewGitHubAuthHandler creates a new GitHub OAuth handler
func NewGitHubAuthHandler(provider *providers.GitHubProvider, playerService *player.Service, authService *auth.Service, isConfigured bool) *GitHubAuthHandler {
	return &GitHubAuthHandler{
		provider:        provider,
		playerService:   playerService,
		authService:     authService,
		isConfigured:    isConfigured,
	}
}

// HandleAuth initiates GitHub OAuth flow
func (h *GitHubAuthHandler) HandleAuth(w http.ResponseWriter, r *http.Request) {
	logger := slog.With(
		"handler", "github_oauth_init",
		"user_agent", r.UserAgent(),
		"ip", r.RemoteAddr,
	)
	
	// Validate that OAuth is properly configured
	if !h.isConfigured {
		logger.Error("GitHub OAuth not configured - missing client credentials")
		sendErrorResponse(w, http.StatusServiceUnavailable, 
			"oauth_not_configured", "GitHub OAuth is not properly configured")
		return
	}

	// Generate and store secure state token
	state, err := auth.GenerateOAuthState("github", r.UserAgent())
	if err != nil {
		logger.Error("Failed to generate state token", "error", err)
		sendErrorResponse(w, http.StatusInternalServerError,
			"internal_error", "Failed to initialize OAuth flow")
		return
	}
	
	url := h.provider.GetAuthURL(state)
	
	logger.Info("Initiating GitHub OAuth flow", 
		"redirect_url", url,
		"state_length", len(state))
	
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// HandleCallback processes GitHub OAuth callback
func (h *GitHubAuthHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")
	
	logger := slog.With(
		"handler", "github_oauth_callback",
		"user_agent", r.UserAgent(),
		"ip", r.RemoteAddr,
		"has_code", code != "",
		"has_state", state != "",
	)
	
	// Check if user denied authorization
	if errorParam != "" {
		logger.Warn("GitHub OAuth authorization denied",
			"oauth_error", errorParam,
			"error_description", r.URL.Query().Get("error_description"))
		redirectWithError(w, r, "oauth_denied", "Authorization was denied")
		return
	}
	
	// Validate authorization code
	if code == "" {
		logger.Error("GitHub OAuth callback missing authorization code")
		redirectWithError(w, r, "oauth_error", "Missing authorization code")
		return
	}
	
	// Validate state token against stored value
	if err := auth.ValidateOAuthState(state, "github", r.UserAgent()); err != nil {
		logger.Error("OAuth state validation failed", 
			"error", err,
			"provider", "github",
			"state", state)
		redirectWithError(w, r, "oauth_error", "Invalid request state - possible CSRF attack")
		return
	}

	logger.Info("OAuth state validation successful - proceeding with GitHub OAuth callback")

	// Exchange code for token with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	token, err := h.provider.ExchangeCode(ctx, code)
	if err != nil {
		logger.Error("Failed to exchange GitHub authorization code",
			"error", err,
			"provider", "github")
		redirectWithError(w, r, "oauth_error", "Failed to exchange authorization code")
		return
	}

	// Get user info from GitHub
	logger.Debug("Fetching user information from GitHub API")
	userInfo, err := h.provider.GetUserInfo(ctx, token)
	if err != nil {
		logger.Error("Failed to get user info from GitHub",
			"error", err,
			"provider", "github")
		redirectWithError(w, r, "oauth_error", "Failed to retrieve user information")
		return
	}
	
	// Add user context to logger
	userLogger := logger.With(
		"user_email", userInfo.Email,
		"github_user_id", userInfo.ID,
		"user_name", userInfo.Name)
	
	// Validate required user info
	if userInfo.Email == "" {
		userLogger.Error("GitHub user info missing required email field")
		redirectWithError(w, r, "oauth_error", "Email address is required for registration")
		return
	}

	// Create/find player
	userLogger.Info("Creating or finding player account for GitHub user")
	
	githubUserID := strconv.Itoa(userInfo.ID)
	
	// First check if auth provider exists
	existingPlayerID, err := h.authService.FindPlayerByAuthProvider("github", githubUserID)
	if err != nil {
		userLogger.Error("Failed to check auth provider", "error", err)
		redirectWithError(w, r, "database_error", "Failed to authenticate user")
		return
	}
	
	var player *player.Player
	if existingPlayerID > 0 {
		// Player exists via OAuth
		player, err = h.playerService.GetPlayerByID(existingPlayerID)
		if err != nil {
			userLogger.Error("Failed to get existing player", "error", err)
			redirectWithError(w, r, "database_error", "Failed to get user account")
			return
		}
	} else {
		// Find or create player
		player, err = h.playerService.FindOrCreatePlayerByOAuth(
			"github",
			githubUserID,
			userInfo.Email,
			userInfo.Name,
			&userInfo.AvatarURL,
		)
		if err != nil {
			userLogger.Error("Failed to create player", "error", err)
			redirectWithError(w, r, "database_error", "Failed to create user account")
			return
		}
		
		// Link auth provider
		err = h.authService.CreateAuthProvider(player.ID, "github", githubUserID, userInfo.Email)
		if err != nil {
			userLogger.Error("Failed to create auth provider", "error", err)
			redirectWithError(w, r, "database_error", "Failed to link account")
			return
		}
	}

	// Add player context to logger
	playerLogger := userLogger.With("player_id", player.ID)

	// Generate JWT
	playerLogger.Debug("Generating JWT token for player")
	jwtToken, err := h.authService.GenerateJWT(player.ID, player.Username, player.Email)
	if err != nil {
		playerLogger.Error("Failed to generate JWT token", "error", err)
		redirectWithError(w, r, "auth_error", "Failed to create authentication token")
		return
	}

	// Set HttpOnly cookie
	utils.SetAuthCookie(w, jwtToken)

	// Check if this is a newly created player
	isNewPlayer := time.Since(player.CreatedAt) < time.Minute
	
	playerLogger.Info("GitHub OAuth authentication successful",
		"provider", "github",
		"new_player", isNewPlayer,
		"player_username", player.Username)
	
	frontendURL := utils.GetEnv("FRONTEND_URL", "http://localhost:3000")
	successURL := fmt.Sprintf("%s/auth/callback?success=true", frontendURL)
	http.Redirect(w, r, successURL, http.StatusTemporaryRedirect)
}
