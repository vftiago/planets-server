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
	logger        *slog.Logger
}

func NewRoutes(db *database.DB, playerService *player.Service, authService *auth.Service, gameService *game.Service, oauthConfig *auth.OAuthConfig, logger *slog.Logger) *Routes {
	return &Routes{
		db:            db,
		playerService: playerService,
		authService:   authService,
		gameService:   gameService,
		oauthConfig:   oauthConfig,
		logger:        logger,
	}
}

func (r *Routes) Setup() *http.ServeMux {
	logger := slog.With("component", "routes", "operation", "setup")
	logger.Debug("Setting up application routes")

	mux := http.NewServeMux()

	healthHandler := serverHandlers.NewHealthHandler(r.db)
	playersHandler := playerHandler.NewPlayersHandler(r.playerService)
	meHandler := playerHandler.NewMeHandler()
	logoutHandler := authHandlers.NewLogoutHandler()

	gameHandler := gameHandlers.NewGameHandler(r.gameService)

	googleAuthHandler := authHandlers.NewOAuthHandler(
		r.oauthConfig.GoogleProvider,
		r.playerService,
		r.authService,
		r.oauthConfig.GoogleConfigured,
	)
	githubAuthHandler := authHandlers.NewOAuthHandler(
		r.oauthConfig.GitHubProvider,
		r.playerService,
		r.authService,
		r.oauthConfig.GitHubConfigured,
	)
	discordAuthHandler := authHandlers.NewOAuthHandler(
		r.oauthConfig.DiscordProvider,
		r.playerService,
		r.authService,
		r.oauthConfig.DiscordConfigured,
	)

	// Protected endpoints (authenticated users)
	mux.Handle("/api/players", middleware.JWTMiddleware(playersHandler))
	mux.Handle("/api/games", middleware.JWTMiddleware(http.HandlerFunc(gameHandler.GetGames)))
	mux.Handle("/api/games/{id}/stats", middleware.JWTMiddleware(http.HandlerFunc(gameHandler.GetGameStats)))
	mux.Handle("/api/players/me", middleware.JWTMiddleware(meHandler))

	// Admin-only endpoints (authenticated + admin role)
	mux.Handle("/api/server/health", middleware.RequireAdmin(healthHandler))
	mux.Handle("/api/games/create", middleware.RequireAdmin(http.HandlerFunc(gameHandler.CreateGame)))
	mux.Handle("/api/games/{id}/delete", middleware.RequireAdmin(http.HandlerFunc(gameHandler.DeleteGame)))

	// OAuth endpoints
	mux.Handle("/auth/google", http.HandlerFunc(googleAuthHandler.HandleAuth))
	mux.Handle("/auth/google/callback", http.HandlerFunc(googleAuthHandler.HandleCallback))
	mux.Handle("/auth/github", http.HandlerFunc(githubAuthHandler.HandleAuth))
	mux.Handle("/auth/github/callback", http.HandlerFunc(githubAuthHandler.HandleCallback))
	mux.Handle("/auth/discord", http.HandlerFunc(discordAuthHandler.HandleAuth))
	mux.Handle("/auth/discord/callback", http.HandlerFunc(discordAuthHandler.HandleCallback))
	mux.Handle("/auth/logout", logoutHandler)

	logger.Info("Routes configured successfully",
		"protected_endpoints", []string{"/api/players", "/api/games", "/api/games/{id}/stats", "/api/players/me"},
		"admin_endpoints", []string{"/api/server/health", "/api/games/create", "/api/games/{id}/delete"},
		"auth_endpoints", []string{"/auth/google", "/auth/github", "/auth/discord", "/auth/logout"},
	)

	return mux
}
