package providers

import (
	"context"

	"golang.org/x/oauth2"
)

// OAuthUser is the normalized user info returned by all OAuth providers.
type OAuthUser struct {
	ID            string
	Email         string
	EmailVerified bool
	Name          string
	AvatarURL     string
}

// OAuthProvider is the interface that all OAuth providers implement.
type OAuthProvider interface {
	Name() string
	GetAuthURL(state string) string
	ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error)
	GetUserInfo(ctx context.Context, token *oauth2.Token) (*OAuthUser, error)
}
