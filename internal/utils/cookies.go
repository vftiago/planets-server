package utils

import (
	"net/http"
	"net/url"
	"strings"
)

// SetAuthCookie sets an authentication cookie with proper security settings
func SetAuthCookie(w http.ResponseWriter, token string) {
	isProduction := GetEnv("ENVIRONMENT", "development") == "production"
	
	// Extract domain from frontend URL for proper cookie domain
	frontendURL := GetEnv("FRONTEND_URL", "http://localhost:3000")
	var domain string
	if parsedURL, err := url.Parse(frontendURL); err == nil && parsedURL.Host != "" {
		host := strings.Split(parsedURL.Host, ":")[0]
		if host != "localhost" && host != "127.0.0.1" {
			domain = host
		}
	}

	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		Domain:   domain,
		MaxAge:   86400, // 24 hours
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: http.SameSiteLaxMode, // Better for OAuth flows than Strict
	}

	http.SetCookie(w, cookie)
}

// ClearAuthCookie clears the authentication cookie
func ClearAuthCookie(w http.ResponseWriter) {
	isProduction := GetEnv("ENVIRONMENT", "development") == "production"
	
	frontendURL := GetEnv("FRONTEND_URL", "http://localhost:3000")
	var domain string
	if parsedURL, err := url.Parse(frontendURL); err == nil && parsedURL.Host != "" {
		host := strings.Split(parsedURL.Host, ":")[0]
		if host != "localhost" && host != "127.0.0.1" {
			domain = host
		}
	}

	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		Domain:   domain,
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1, // Expire immediately
	}

	http.SetCookie(w, cookie)
}
