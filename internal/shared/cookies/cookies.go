package cookies

import (
	"net/http"
	"net/url"
	"planets-server/internal/shared/config"
	"strings"
)

func SetAuthCookie(w http.ResponseWriter, token string) {
	cfg := config.GlobalConfig

	cookie := createAuthCookie()
	cookie.Value = token
	cookie.MaxAge = int(cfg.Auth.TokenExpiration.Seconds())
	
	http.SetCookie(w, cookie)
}

func ClearAuthCookie(w http.ResponseWriter) {
	cookie := createAuthCookie()
	cookie.Value = ""
	cookie.MaxAge = -1
	
	http.SetCookie(w, cookie)
}

func createAuthCookie() *http.Cookie {
	cfg := config.GlobalConfig

	return &http.Cookie{
		Name:     "auth_token",
		Path:     "/",
		Domain:   extractDomain(cfg.Frontend.URL),
		HttpOnly: true,
		Secure:   cfg.Auth.CookieSecure,
		SameSite: parseSameSite(cfg.Auth.CookieSameSite),
	}
}

func extractDomain(frontendURL string) string {
	parsedURL, err := url.Parse(frontendURL)
	if err != nil || parsedURL.Host == "" {
		return ""
	}
	
	host := strings.Split(parsedURL.Host, ":")[0]
	if host == "localhost" || host == "127.0.0.1" {
		return ""
	}
	
	return host
}

func parseSameSite(sameSiteStr string) http.SameSite {
	switch sameSiteStr {
	case "strict":
		return http.SameSiteStrictMode
	case "lax":
		return http.SameSiteLaxMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}
