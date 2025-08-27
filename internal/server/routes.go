package server

import (
	"log/slog"
	"net/http"

	"planets-server/internal/auth"
	authHandlers "planets-server/internal/auth/handlers"
	"planets-server/internal/handlers"
	"planets-server/internal/middleware"
	"planets-server/internal/player"
	"planets-server/internal/shared/database"
)

type Routes struct {
	db            *database.DB
	playerService *player.Service
	authService   *auth.Service
	oauthConfig   *auth.OAuthConfig
}

func NewRoutes(db *database.DB, playerService *player.Service, authService *auth.Service, oauthConfig *auth.OAuthConfig) *Routes {
	return &Routes{
		db:            db,
		playerService: playerService,
		authService:   authService,
		oauthConfig:   oauthConfig,
	}
}

func (r *Routes) Setup() *http.ServeMux {
	logger := slog.With("component", "routes", "operation", "setup")
	logger.Debug("Setting up application routes")

	mux := http.NewServeMux()

	// Initialize API handlers
	healthHandler := handlers.NewHealthHandler(r.db)
	gameStatusHandler := handlers.NewGameStatusHandler(r.playerService)
	playersHandler := player.NewPlayersHandler(r.playerService)
	meHandler := player.NewMeHandler()
	logoutHandler := handlers.NewLogoutHandler()

	// Initialize OAuth handlers
	googleAuthHandler := authHandlers.NewGoogleAuthHandler(
		r.oauthConfig.GoogleProvider, 
		r.playerService,
		r.authService,
		r.oauthConfig.GoogleConfigured,
	)
	githubAuthHandler := authHandlers.NewGitHubAuthHandler(
		r.oauthConfig.GitHubProvider, 
		r.playerService,
		r.authService,
		r.oauthConfig.GitHubConfigured,
	)

	// Public API endpoints
	mux.Handle("/api/health", healthHandler)
	mux.Handle("/api/game/status", gameStatusHandler)
	mux.Handle("/api/players", playersHandler)
	
	// Protected API endpoints
	mux.Handle("/api/me", middleware.JWTMiddleware(meHandler))

	// OAuth endpoints
	mux.HandleFunc("/auth/google", googleAuthHandler.HandleAuth)
	mux.HandleFunc("/auth/google/callback", googleAuthHandler.HandleCallback)
	mux.HandleFunc("/auth/github", githubAuthHandler.HandleAuth)
	mux.HandleFunc("/auth/github/callback", githubAuthHandler.HandleCallback)
	mux.Handle("/auth/logout", logoutHandler)

	return mux
}
