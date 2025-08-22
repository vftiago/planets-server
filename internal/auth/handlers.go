package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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

// Google OAuth initiation
func (s *OAuthService) HandleGoogleAuth(w http.ResponseWriter, r *http.Request) {
	url := GoogleOAuthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Google OAuth callback
func (s *OAuthService) HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No authorization code", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	token, err := GoogleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	// Get user info from Google
	userInfo, err := s.getGoogleUserInfo(token)
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	fmt.Printf("DEBUG Google userInfo: ID='%s', Email='%s', Name='%s'\n", userInfo.ID, userInfo.Email, userInfo.Name)

	// Create/find player
	player, err := s.playerRepo.FindOrCreatePlayerByOAuth(
		"google",
		userInfo.ID,
		userInfo.Email,
		userInfo.Name,
		&userInfo.Picture,
	)
	if err != nil {
		http.Error(w, "Failed to create/find player", http.StatusInternalServerError)
		return
	}

	// Generate JWT
	jwtToken, err := GenerateJWT(player)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Redirect to frontend with token
	frontendURL := utils.GetEnv("FRONTEND_URL", "http://localhost:3000")
	redirectURL := fmt.Sprintf("%s/auth/callback?token=%s", frontendURL, jwtToken)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// GitHub OAuth initiation
func (s *OAuthService) HandleGitHubAuth(w http.ResponseWriter, r *http.Request) {
	url := GitHubOAuthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GitHub OAuth callback
func (s *OAuthService) HandleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No authorization code", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	token, err := GitHubOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	// Get user info from GitHub
	userInfo, err := s.getGitHubUserInfo(token)
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	// Create/find player
	player, err := s.playerRepo.FindOrCreatePlayerByOAuth(
		"github",
		fmt.Sprintf("%d", userInfo.ID),
		userInfo.Email,
		userInfo.Name,
		&userInfo.AvatarURL,
	)
	if err != nil {
		http.Error(w, "Failed to create/find player", http.StatusInternalServerError)
		return
	}

	// Generate JWT
	jwtToken, err := GenerateJWT(player)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Redirect to frontend with token
	frontendURL := utils.GetEnv("FRONTEND_URL", "http://localhost:3000")
	redirectURL := fmt.Sprintf("%s/auth/callback?token=%s", frontendURL, jwtToken)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
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

// Get user info from Google
func (s *OAuthService) getGoogleUserInfo(token *oauth2.Token) (*GoogleUserInfo, error) {
	client := GoogleOAuthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// Get user info from GitHub
func (s *OAuthService) getGitHubUserInfo(token *oauth2.Token) (*GitHubUserInfo, error) {
	client := GitHubOAuthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo GitHubUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	// GitHub might not return email in the user endpoint
	if userInfo.Email == "" {
		emailResp, err := client.Get("https://api.github.com/user/emails")
		if err == nil {
			defer emailResp.Body.Close()
			var emails []struct {
				Email   string `json:"email"`
				Primary bool   `json:"primary"`
			}
			if json.NewDecoder(emailResp.Body).Decode(&emails) == nil {
				for _, email := range emails {
					if email.Primary {
						userInfo.Email = email.Email
						break
					}
				}
			}
		}
	}

	return &userInfo, nil
}
