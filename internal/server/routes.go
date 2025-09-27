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
	"planets-server/internal/universe"
	universeHandlers "planets-server/internal/universe/handlers"
)

type Routes struct {
	db              *database.DB
	playerService   *player.Service
	authService     *auth.Service
	gameService     *game.Service
	universeService *universe.Service
	oauthConfig     *auth.OAuthConfig
	logger          *slog.Logger
}

func NewRoutes(db *database.DB, playerService *player.Service, authService *auth.Service, gameService *game.Service, universeService *universe.Service, oauthConfig *auth.OAuthConfig, logger *slog.Logger) *Routes {
	return &Routes{
		db:              db,
		playerService:   playerService,
		authService:     authService,
		gameService:     gameService,
		universeService: universeService,
		oauthConfig:     oauthConfig,
		logger:          logger,
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
	universeHandler := universeHandlers.NewUniverseHandler(r.universeService, r.logger)

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
	mux.HandleFunc("/api/games/{id}/stats", gameHandler.GetGameStats)

	// Universe API (public read-only)
	mux.HandleFunc("/api/universes", universeHandler.GetUniverses)
	mux.HandleFunc("/api/universes/{id}", universeHandler.GetUniverse)

	// Protected endpoints (authenticated users)
	mux.Handle("/api/players/me", middleware.JWTMiddleware(meHandler))

	// Admin-only endpoints (authenticated + admin role)
	mux.Handle("/api/games/create", middleware.RequireAdmin(http.HandlerFunc(gameHandler.CreateGame)))
	mux.Handle("/api/universes/create", middleware.RequireAdmin(http.HandlerFunc(universeHandler.CreateUniverse)))
	mux.HandleFunc("/api/universes/{id}/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			middleware.RequireAdmin(http.HandlerFunc(universeHandler.DeleteUniverse)).ServeHTTP(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// OAuth endpoints
	mux.HandleFunc("/auth/google", googleAuthHandler.HandleAuth)
	mux.HandleFunc("/auth/google/callback", googleAuthHandler.HandleCallback)
	mux.HandleFunc("/auth/github", githubAuthHandler.HandleAuth)
	mux.HandleFunc("/auth/github/callback", githubAuthHandler.HandleCallback)
	mux.Handle("/auth/logout", logoutHandler)

	logger.Info("Routes configured successfully",
		"public_endpoints", []string{"/api/server/health", "/api/game/status", "/api/players", "/api/games", "/api/games/stats", "/api/universes/*"},
		"protected_endpoints", []string{"/api/players/me"},
		"admin_endpoints", []string{"/api/games/create", "/api/universes/create", "/api/universes/{id}/delete"},
		"auth_endpoints", []string{"/auth/google", "/auth/github", "/auth/logout"},
	)

	return mux
}
