package server

import (
	"log/slog"
	"net/http"

	"planets-server/internal/auth"
	authHandlers "planets-server/internal/auth/handlers"
	"planets-server/internal/handlers"
	"planets-server/internal/middleware"
	"planets-server/internal/models"
	"planets-server/internal/shared/database"
)

type Routes struct {
	db         *database.DB
	playerRepo *models.PlayerRepository
	oauthConfig *auth.OAuthConfig
}

func NewRoutes(db *database.DB, playerRepo *models.PlayerRepository, oauthConfig *auth.OAuthConfig) *Routes {
	return &Routes{
		db:         db,
		playerRepo: playerRepo,
		oauthConfig: oauthConfig,
	}
}

func (r *Routes) Setup() *http.ServeMux {
	logger := slog.With("component", "routes", "operation", "setup")
	logger.Debug("Setting up application routes")

	mux := http.NewServeMux()

	// Initialize API handlers
	healthHandler := handlers.NewHealthHandler(r.db)
	gameStatusHandler := handlers.NewGameStatusHandler(r.playerRepo)
	playersHandler := handlers.NewPlayersHandler(r.playerRepo)
	meHandler := handlers.NewMeHandler()
	logoutHandler := handlers.NewLogoutHandler()

	// Initialize OAuth handlers
	googleAuthHandler := authHandlers.NewGoogleAuthHandler(
		r.oauthConfig.GoogleProvider, 
		r.playerRepo, 
		r.oauthConfig.GoogleConfigured,
	)
	githubAuthHandler := authHandlers.NewGitHubAuthHandler(
		r.oauthConfig.GitHubProvider, 
		r.playerRepo, 
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
