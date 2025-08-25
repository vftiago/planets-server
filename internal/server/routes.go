package server

import (
	"log/slog"
	"net/http"

	"planets-server/internal/auth"
	"planets-server/internal/database"
	"planets-server/internal/handlers"
	"planets-server/internal/middleware"
	"planets-server/internal/models"
)

type Routes struct {
	db           *database.DB
	playerRepo   *models.PlayerRepository
	oauthService *auth.OAuthService
}

func NewRoutes(db *database.DB, playerRepo *models.PlayerRepository, oauthService *auth.OAuthService) *Routes {
	return &Routes{
		db:           db,
		playerRepo:   playerRepo,
		oauthService: oauthService,
	}
}

func (r *Routes) Setup() *http.ServeMux {
	logger := slog.With("component", "routes", "operation", "setup")
	logger.Debug("Setting up application routes")

	mux := http.NewServeMux()

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(r.db)
	gameStatusHandler := handlers.NewGameStatusHandler(r.playerRepo)
	playersHandler := handlers.NewPlayersHandler(r.playerRepo)
	meHandler := handlers.NewMeHandler()
	logoutHandler := handlers.NewLogoutHandler()

	// Public API endpoints
	mux.Handle("/api/health", healthHandler)
	mux.Handle("/api/game/status", gameStatusHandler)
	mux.Handle("/api/players", playersHandler)
	
	// Protected API endpoints
	mux.Handle("/api/me", middleware.JWTMiddleware(meHandler))

	// OAuth endpoints
	mux.HandleFunc("/auth/google", r.oauthService.HandleGoogleAuth)
	mux.HandleFunc("/auth/google/callback", r.oauthService.HandleGoogleCallback)
	mux.HandleFunc("/auth/github", r.oauthService.HandleGitHubAuth)
	mux.HandleFunc("/auth/github/callback", r.oauthService.HandleGitHubCallback)
	mux.Handle("/auth/logout", logoutHandler)

	return mux
}
