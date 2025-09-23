package server

import (
	"log/slog"
	"net/http"

	"planets-server/internal/auth"
	authHandlers "planets-server/internal/auth/handlers"
	"planets-server/internal/game"
	gameHandlers "planets-server/internal/game/handlers"
	"planets-server/internal/middleware"
	"planets-server/internal/player"
	playerHandler "planets-server/internal/player/handlers"
	serverHandlers "planets-server/internal/server/handlers"
	"planets-server/internal/shared/database"
)

type Routes struct {
	db            *database.DB
	playerService *player.Service
	authService   *auth.Service
	gameService   *game.Service
	oauthConfig   *auth.OAuthConfig
}

func NewRoutes(db *database.DB, playerService *player.Service, authService *auth.Service, gameService *game.Service, oauthConfig *auth.OAuthConfig) *Routes {
	return &Routes{
		db:            db,
		playerService: playerService,
		authService:   authService,
		gameService:   gameService,
		oauthConfig:   oauthConfig,
	}
}

func (r *Routes) Setup() *http.ServeMux {
	logger := slog.With("component", "routes", "operation", "setup")
	logger.Debug("Setting up application routes")

	mux := http.NewServeMux()

	healthHandler := serverHandlers.NewHealthHandler(r.db)
	gameStatusHandler := gameHandlers.NewGameStatusHandler(r.playerService)
	playersHandler := playerHandler.NewPlayersHandler(r.playerService)
	meHandler := playerHandler.NewMeHandler()
	logoutHandler := authHandlers.NewLogoutHandler()

	gameHandler := gameHandlers.NewGameHandler(r.gameService)

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

	// Public endpoints
	mux.Handle("/api/server/health", healthHandler)
	mux.Handle("/api/game/status", gameStatusHandler)
	mux.Handle("/api/players", playersHandler)
	mux.HandleFunc("/api/games", gameHandler.GetGames)
	mux.HandleFunc("/api/games/stats", gameHandler.GetGameStats)

	// Protected endpoints (authenticated users)
	mux.Handle("/api/players/me", middleware.JWTMiddleware(meHandler))

	// Admin-only endpoints (authenticated + admin role)
	mux.Handle("/api/games/create", middleware.RequireAdmin(http.HandlerFunc(gameHandler.CreateGame)))

	// OAuth endpoints
	mux.HandleFunc("/auth/google", googleAuthHandler.HandleAuth)
	mux.HandleFunc("/auth/google/callback", googleAuthHandler.HandleCallback)
	mux.HandleFunc("/auth/github", githubAuthHandler.HandleAuth)
	mux.HandleFunc("/auth/github/callback", githubAuthHandler.HandleCallback)
	mux.Handle("/auth/logout", logoutHandler)

	logger.Info("Routes configured successfully",
		"public_endpoints", []string{"/api/server/health", "/api/game/status", "/api/players", "/api/games", "/api/games/stats"},
		"protected_endpoints", []string{"/api/players/me"},
		"admin_endpoints", []string{"/api/games/create"},
		"auth_endpoints", []string{"/auth/google", "/auth/github", "/auth/logout"},
	)

	return mux
}
