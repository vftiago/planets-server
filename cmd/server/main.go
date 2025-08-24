package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	auth.InitOAuth()
	
	var err error
	db, err = database.Connect()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	fmt.Println("Connected to database successfully")

	// Run migrations
	fmt.Println("Running database migrations...")
	if err := db.RunMigrations(); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}
	fmt.Println("Migrations completed successfully")

	// Initialize services
	playerRepo = models.NewPlayerRepository(db.DB)
	oauthService = auth.NewOAuthService(playerRepo)
	fmt.Println("Services initialized")

	// Setup CORS
	corsMiddleware := middleware.SetupCORS()
	fmt.Printf("CORS configured for origin: %s\n", utils.GetEnv("FRONTEND_URL", "http://localhost:3000"))

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

	fmt.Println("Planets! server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	dbStatus := "disconnected"
	if err := db.Ping(); err == nil {
		dbStatus = "connected"
	}
	
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Database:  dbStatus,
	}
	json.NewEncoder(w).Encode(response)
}

func gameStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	playerCount, err := playerRepo.GetPlayerCount()
	if err != nil {
		playerCount = 0
	}

	response := GameStatusResponse{
		Game:          "Planets!",
		Turn:          1,
		OnlinePlayers: playerCount,
	}
	json.NewEncoder(w).Encode(response)
}

func playersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	players, err := playerRepo.GetAllPlayers()
	if err != nil {
		http.Error(w, "Failed to fetch players", http.StatusInternalServerError)
		return
	}

	if players == nil {
		players = []models.Player{}
	}
	
	json.NewEncoder(w).Encode(players)
}

func meHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	claims := r.Context().Value("user").(*auth.Claims)
	
	response := map[string]interface{}{
		"player_id": claims.PlayerID,
		"username":  claims.Username,
		"email":     claims.Email,
	}
	
	json.NewEncoder(w).Encode(response)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
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
	w.Write([]byte("Logged out"))
}
