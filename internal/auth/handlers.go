package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"planets-server/internal/models"
	"planets-server/internal/utils"

	"golang.org/x/oauth2"
)

type OAuthService struct {
	playerRepo *models.PlayerRepository
}

func NewOAuthService(playerRepo *models.PlayerRepository) *OAuthService {
	return &OAuthService{playerRepo: playerRepo}
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func sendErrorResponse(w http.ResponseWriter, statusCode int, errorType, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	errorResp := ErrorResponse{
		Error:   errorType,
		Message: message,
		Code:    statusCode,
	}
	
	if err := json.NewEncoder(w).Encode(errorResp); err != nil {
		slog.Error("Failed to encode error response", 
			"error", err,
			"status_code", statusCode,
			"error_type", errorType)
	}
}

func redirectWithError(w http.ResponseWriter, r *http.Request, errorType, message string) {
	frontendURL := utils.GetEnv("FRONTEND_URL", "http://localhost:3000")
	errorURL := fmt.Sprintf("%s/auth/error?error=%s&message=%s", 
		frontendURL, errorType, message)
	
	slog.Debug("Redirecting to frontend with error",
		"frontend_url", frontendURL,
		"error_type", errorType,
		"message", message)
	
	http.Redirect(w, r, errorURL, http.StatusTemporaryRedirect)
}

// Google OAuth initiation
func (s *OAuthService) HandleGoogleAuth(w http.ResponseWriter, r *http.Request) {
	logger := slog.With(
		"handler", "google_oauth_init",
		"user_agent", r.UserAgent(),
		"ip", r.RemoteAddr,
	)
	
	// Validate that OAuth is properly configured
	if GoogleOAuthConfig.ClientID == "" || GoogleOAuthConfig.ClientSecret == "" {
		logger.Error("Google OAuth not configured - missing client credentials")
		sendErrorResponse(w, http.StatusServiceUnavailable, 
			"oauth_not_configured", "Google OAuth is not properly configured")
		return
	}

	// Generate and store secure state token
	state, err := GenerateOAuthState("google", r.UserAgent())
	if err != nil {
		logger.Error("Failed to generate state token", "error", err)
		sendErrorResponse(w, http.StatusInternalServerError,
			"internal_error", "Failed to initialize OAuth flow")
		return
	}
	
	url := GoogleOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	
	logger.Info("Initiating Google OAuth flow", 
		"redirect_url", url,
		"state_length", len(state))
	
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Google OAuth callback
func (s *OAuthService) HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")
	
	logger := slog.With(
		"handler", "google_oauth_callback",
		"user_agent", r.UserAgent(),
		"ip", r.RemoteAddr,
		"has_code", code != "",
		"has_state", state != "",
	)
	
	// Check if user denied authorization
	if errorParam != "" {
		logger.Warn("Google OAuth authorization denied",
			"oauth_error", errorParam,
			"error_description", r.URL.Query().Get("error_description"))
		redirectWithError(w, r, "oauth_denied", "Authorization was denied")
		return
	}
	
	// Validate authorization code
	if code == "" {
		logger.Error("Google OAuth callback missing authorization code")
		redirectWithError(w, r, "oauth_error", "Missing authorization code")
		return
	}
	
	// Validate state token against stored value
	if err := ValidateOAuthState(state, "google", r.UserAgent()); err != nil {
		logger.Error("OAuth state validation failed", 
			"error", err,
			"provider", "google",
			"state", state)
		redirectWithError(w, r, "oauth_error", "Invalid request state - possible CSRF attack")
		return
	}
	
	logger.Info("OAuth state validation successful - proceeding with Google OAuth callback")

	// Exchange code for token with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()
	
	logger.Debug("Exchanging authorization code for Google access token")
	token, err := GoogleOAuthConfig.Exchange(ctx, code)
	if err != nil {
		logger.Error("Failed to exchange Google authorization code",
			"error", err,
			"provider", "google")
		redirectWithError(w, r, "oauth_error", "Failed to exchange authorization code")
		return
	}

	// Get user info from Google
	logger.Debug("Fetching user information from Google API")
	userInfo, err := s.getGoogleUserInfo(ctx, token)
	if err != nil {
		logger.Error("Failed to get user info from Google",
			"error", err,
			"provider", "google")
		redirectWithError(w, r, "oauth_error", "Failed to retrieve user information")
		return
	}
	
	// Add user context to logger
	userLogger := logger.With(
		"user_email", userInfo.Email,
		"google_user_id", userInfo.ID,
		"user_name", userInfo.Name)
	
	// Validate required user info
	if userInfo.Email == "" {
		userLogger.Error("Google user info missing required email field")
		redirectWithError(w, r, "oauth_error", "Email address is required")
		return
	}

	// Create/find player
	userLogger.Info("Creating or finding player account for Google user")
	player, err := s.playerRepo.FindOrCreatePlayerByOAuth(
		"google",
		userInfo.ID,
		userInfo.Email,
		userInfo.Name,
		&userInfo.Picture,
	)
	if err != nil {
		userLogger.Error("Failed to create or find player account",
			"error", err,
			"provider", "google")
		redirectWithError(w, r, "database_error", "Failed to create user account")
		return
	}

	// Add player context to logger
	playerLogger := userLogger.With("player_id", player.ID)

	// Generate JWT
	playerLogger.Debug("Generating JWT token for player")
	jwtToken, err := GenerateJWT(player)
	if err != nil {
		playerLogger.Error("Failed to generate JWT token",
			"error", err)
		redirectWithError(w, r, "auth_error", "Failed to create authentication token")
		return
	}

	// Set HttpOnly cookie
	isProduction := utils.GetEnv("ENVIRONMENT", "development") == "production"
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    jwtToken,
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   86400, // 24 hours
	})

	// Check if this is a newly created player (rough approximation)
	isNewPlayer := time.Since(player.CreatedAt) < time.Minute
	
	playerLogger.Info("Google OAuth authentication successful",
		"provider", "google",
		"new_player", isNewPlayer,
		"player_username", player.Username)
	
	frontendURL := utils.GetEnv("FRONTEND_URL", "http://localhost:3000")
	successURL := fmt.Sprintf("%s/auth/callback?success=true", frontendURL)
	http.Redirect(w, r, successURL, http.StatusTemporaryRedirect)
}

// GitHub OAuth initiation
func (s *OAuthService) HandleGitHubAuth(w http.ResponseWriter, r *http.Request) {
	logger := slog.With(
		"handler", "github_oauth_init",
		"user_agent", r.UserAgent(),
		"ip", r.RemoteAddr,
	)
	
	// Validate that OAuth is properly configured
	if GitHubOAuthConfig.ClientID == "" || GitHubOAuthConfig.ClientSecret == "" {
		logger.Error("GitHub OAuth not configured - missing client credentials")
		sendErrorResponse(w, http.StatusServiceUnavailable, 
			"oauth_not_configured", "GitHub OAuth is not properly configured")
		return
	}

	// Generate and store secure state token
	state, err := GenerateOAuthState("github", r.UserAgent())
	if err != nil {
		logger.Error("Failed to generate state token", "error", err)
		sendErrorResponse(w, http.StatusInternalServerError,
			"internal_error", "Failed to initialize OAuth flow")
		return
	}
	
	url := GitHubOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	
	logger.Info("Initiating GitHub OAuth flow", 
		"redirect_url", url,
		"state_length", len(state))
	
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GitHub OAuth callback
func (s *OAuthService) HandleGitHubCallback(w http.ResponseWriter, r *http.Request) {
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
	if err := ValidateOAuthState(state, "github", r.UserAgent()); err != nil {
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
	
	logger.Debug("Exchanging authorization code for GitHub access token")
	token, err := GitHubOAuthConfig.Exchange(ctx, code)
	if err != nil {
		logger.Error("Failed to exchange GitHub authorization code",
			"error", err,
			"provider", "github")
		redirectWithError(w, r, "oauth_error", "Failed to exchange authorization code")
		return
	}

	// Get user info from GitHub
	logger.Debug("Fetching user information from GitHub API")
	userInfo, err := s.getGitHubUserInfo(ctx, token)
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
	player, err := s.playerRepo.FindOrCreatePlayerByOAuth(
		"github",
		fmt.Sprintf("%d", userInfo.ID),
		userInfo.Email,
		userInfo.Name,
		&userInfo.AvatarURL,
	)
	if err != nil {
		userLogger.Error("Failed to create or find player account",
			"error", err,
			"provider", "github")
		redirectWithError(w, r, "database_error", "Failed to create user account")
		return
	}

	// Add player context to logger
	playerLogger := userLogger.With("player_id", player.ID)

	// Generate JWT
	playerLogger.Debug("Generating JWT token for player")
	jwtToken, err := GenerateJWT(player)
	if err != nil {
		playerLogger.Error("Failed to generate JWT token",
			"error", err)
		redirectWithError(w, r, "auth_error", "Failed to create authentication token")
		return
	}

	// Set HttpOnly cookie
	isProduction := utils.GetEnv("ENVIRONMENT", "development") == "production"
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    jwtToken,
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   86400, // 24 hours
	})

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

// Google user info structure
type GoogleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

// GitHub user info structure
type GitHubUserInfo struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

// Get user info from Google with proper error handling and context
func (s *OAuthService) getGoogleUserInfo(ctx context.Context, token *oauth2.Token) (*GoogleUserInfo, error) {
	client := GoogleOAuthConfig.Client(ctx, token)
	
	logger := slog.With("api_call", "google_userinfo")
	logger.Debug("Requesting user info from Google API")
	
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		logger.Error("Failed to request user info from Google", "error", err)
		return nil, fmt.Errorf("failed to request user info from Google: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("Google API returned error status", 
			"status_code", resp.StatusCode,
			"status", resp.Status)
		return nil, fmt.Errorf("Google API returned status %d", resp.StatusCode)
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		logger.Error("Failed to decode Google user info", "error", err)
		return nil, fmt.Errorf("failed to decode Google user info: %w", err)
	}

	// Validate required fields
	if userInfo.ID == "" {
		logger.Error("Google user info missing user ID")
		return nil, fmt.Errorf("Google user info missing user ID")
	}
	if userInfo.Email == "" {
		logger.Error("Google user info missing email")
		return nil, fmt.Errorf("Google user info missing email")
	}

	logger.Debug("Successfully retrieved Google user info",
		"user_id", userInfo.ID,
		"has_email", userInfo.Email != "",
		"has_name", userInfo.Name != "",
		"has_picture", userInfo.Picture != "")

	return &userInfo, nil
}

// Get user info from GitHub with proper error handling and context
func (s *OAuthService) getGitHubUserInfo(ctx context.Context, token *oauth2.Token) (*GitHubUserInfo, error) {
	client := GitHubOAuthConfig.Client(ctx, token)
	
	logger := slog.With("api_call", "github_userinfo")
	logger.Debug("Requesting user info from GitHub API")
	
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		logger.Error("Failed to request user info from GitHub", "error", err)
		return nil, fmt.Errorf("failed to request user info from GitHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("GitHub API returned error status",
			"status_code", resp.StatusCode,
			"status", resp.Status)
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var userInfo GitHubUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		logger.Error("Failed to decode GitHub user info", "error", err)
		return nil, fmt.Errorf("failed to decode GitHub user info: %w", err)
	}

	// Validate required fields
	if userInfo.ID == 0 {
		logger.Error("GitHub user info missing user ID")
		return nil, fmt.Errorf("GitHub user info missing user ID")
	}

	// GitHub might not return email in the user endpoint, try to get it
	if userInfo.Email == "" {
		logger.Debug("GitHub user info missing email, attempting to fetch from emails endpoint")
		if err := s.fetchGitHubUserEmail(ctx, client, &userInfo); err != nil {
			logger.Warn("Failed to fetch GitHub user email", "error", err)
			// Don't fail here - we'll validate email requirement in the caller
		}
	}

	logger.Debug("Successfully retrieved GitHub user info",
		"user_id", userInfo.ID,
		"has_email", userInfo.Email != "",
		"has_name", userInfo.Name != "",
		"has_avatar", userInfo.AvatarURL != "")

	return &userInfo, nil
}

// fetchGitHubUserEmail attempts to get the user's primary email from GitHub
func (s *OAuthService) fetchGitHubUserEmail(ctx context.Context, client *http.Client, userInfo *GitHubUserInfo) error {
	logger := slog.With("api_call", "github_emails", "github_user_id", userInfo.ID)
	
	logger.Debug("Requesting email information from GitHub API")
	emailResp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		logger.Error("Failed to request emails from GitHub", "error", err)
		return fmt.Errorf("failed to request emails from GitHub: %w", err)
	}
	defer emailResp.Body.Close()

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
			userInfo.Email = email.Email
			logger.Debug("Found primary verified email", "email", email.Email)
			return nil
		}
	}

	// Fallback to any verified email
	for _, email := range emails {
		if email.Verified {
			userInfo.Email = email.Email
			logger.Debug("Found verified email (not primary)", "email", email.Email)
			return nil
		}
	}

	logger.Warn("No verified email found for GitHub user",
		"total_emails", len(emails))
	return fmt.Errorf("no verified email found")
}
