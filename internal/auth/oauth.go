package auth

import (
	"fmt"
	"planets-server/internal/utils"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

var GitHubOAuthConfig *oauth2.Config
var GoogleOAuthConfig *oauth2.Config

func InitOAuth() {
	baseURL := utils.GetEnv("BASE_URL", "http://localhost:8080")

	GitHubOAuthConfig = &oauth2.Config{
		ClientID:     utils.GetEnv("GITHUB_CLIENT_ID", ""),
		ClientSecret: utils.GetEnv("GITHUB_CLIENT_SECRET", ""),
		RedirectURL:  fmt.Sprintf("%s/auth/github/callback", baseURL),
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}

	GoogleOAuthConfig = &oauth2.Config{
		ClientID:     utils.GetEnv("GOOGLE_CLIENT_ID", ""),
		ClientSecret: utils.GetEnv("GOOGLE_CLIENT_SECRET", ""),
		RedirectURL:  fmt.Sprintf("%s/auth/google/callback", baseURL),
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint:     google.Endpoint,
	}
}
