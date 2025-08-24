package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"planets-server/internal/auth"
	"planets-server/internal/database"
	"planets-server/internal/middleware"
	"planets-server/internal/models"
	"planets-server/internal/utils"

	"github.com/joho/godotenv"
)

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Database  string `json:"database"`
}

type GameStatusResponse struct {
	Game          string `json:"game"`
	Turn          int    `json:"turn"`
	OnlinePlayers int    `json:"online_players"`
}

var db *database.DB
var playerRepo *models.PlayerRepository
var oauthService *auth.OAuthService

func initLogger() {
	var handler slog.Handler
	
	if utils.GetEnv("ENVIRONMENT", "development") == "production" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}
	
	slog.SetDefault(slog.New(handler))
}

func main() {
	// Initialize logger
	initLogger()
	
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		slog.Debug("No .env file found, using system environment variables")
	}

	// Initialize OAuth
	auth.InitOAuth()
	slog.Info("OAuth configuration initialized")
	
	// Connect to database
	var err error
	db, err = database.Connect()
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("Connected to database successfully")

	// Run migrations
	slog.Info("Running database migrations...")
	if err := db.RunMigrations(); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}
	slog.Info("Migrations completed successfully")

	// Initialize services
	playerRepo = models.NewPlayerRepository(db.DB)
	oauthService = auth.NewOAuthService(playerRepo)
	slog.Info("Services initialized")

	// Setup CORS
	corsMiddleware := middleware.SetupCORS()
	frontendURL := utils.GetEnv("FRONTEND_URL", "http://localhost:3000")
	slog.Info("CORS configured", "allowed_origin", frontendURL)

	// Setup routes
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/api/health", healthHandler)
	mux.HandleFunc("/api/game/status", gameStatusHandler)
	mux.HandleFunc("/api/players", playersHandler)
	mux.Handle("/api/me", middleware.JWTMiddleware(http.HandlerFunc(meHandler))) // Protected route

	// OAuth endpoints
	mux.HandleFunc("/auth/google", oauthService.HandleGoogleAuth)
	mux.HandleFunc("/auth/google/callback", oauthService.HandleGoogleCallback)
	mux.HandleFunc("/auth/github", oauthService.HandleGitHubAuth)
	mux.HandleFunc("/auth/github/callback", oauthService.HandleGitHubCallback)
	mux.HandleFunc("/auth/logout", logoutHandler)

	// Wrap mux with CORS middleware
	handler := corsMiddleware.Handler(mux)

	port := ":8080"
	slog.Info("Starting Planets! server", "port", port)
	if err := http.ListenAndServe(port, handler); err != nil {
		slog.Error("Server failed to start", "error", err, "port", port)
		os.Exit(1)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	logger := slog.With("handler", "health", "remote_addr", r.RemoteAddr)
	logger.Debug("Health check requested")
	
	w.Header().Set("Content-Type", "application/json")
	
	dbStatus := "disconnected"
	if err := db.Ping(); err == nil {
		dbStatus = "connected"
	} else {
		logger.Warn("Database ping failed", "error", err)
	}
	
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Database:  dbStatus,
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode health response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	
	logger.Debug("Health check completed", "db_status", dbStatus)
}

func gameStatusHandler(w http.ResponseWriter, r *http.Request) {
	logger := slog.With("handler", "game_status", "remote_addr", r.RemoteAddr)
	logger.Debug("Game status requested")
	
	w.Header().Set("Content-Type", "application/json")

	playerCount, err := playerRepo.GetPlayerCount()
	if err != nil {
		logger.Warn("Failed to get player count", "error", err)
		playerCount = 0
	}

	response := GameStatusResponse{
		Game:          "Planets!",
		Turn:          1,
		OnlinePlayers: playerCount,
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode game status response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	
	logger.Debug("Game status completed", "player_count", playerCount)
}

func playersHandler(w http.ResponseWriter, r *http.Request) {
	logger := slog.With("handler", "players", "remote_addr", r.RemoteAddr)
	logger.Debug("Players list requested")
	
	w.Header().Set("Content-Type", "application/json")
	
	players, err := playerRepo.GetAllPlayers()
	if err != nil {
		logger.Error("Failed to fetch players", "error", err)
		http.Error(w, "Failed to fetch players", http.StatusInternalServerError)
		return
	}

	if players == nil {
		players = []models.Player{}
	}
	
	if err := json.NewEncoder(w).Encode(players); err != nil {
		logger.Error("Failed to encode players response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	
	logger.Debug("Players list completed", "player_count", len(players))
}

func meHandler(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("user").(*auth.Claims)
	logger := slog.With(
		"handler", "me", 
		"remote_addr", r.RemoteAddr,
		"player_id", claims.PlayerID,
		"username", claims.Username,
	)
	logger.Debug("User info requested")
	
	w.Header().Set("Content-Type", "application/json")
	
	response := map[string]interface{}{
		"player_id": claims.PlayerID,
		"username":  claims.Username,
		"email":     claims.Email,
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode user info response", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	
	logger.Debug("User info completed")
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	logger := slog.With("handler", "logout", "remote_addr", r.RemoteAddr)
	logger.Debug("Logout requested")
	
	// Clear the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		HttpOnly: true,
		Secure:   utils.GetEnv("ENVIRONMENT", "development") == "production",
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   -1, // Expire immediately
	})
	
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Logged out")); err != nil {
		logger.Error("Failed to write logout response", "error", err)
		return
	}
	
	logger.Info("User logged out successfully")
}
