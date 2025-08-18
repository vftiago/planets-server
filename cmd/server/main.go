package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

type GameStatusResponse struct {
	Game          string `json:"game"`
	Turn          int    `json:"turn"`
	OnlinePlayers int    `json:"online_players"`
}

func main() {
	mux := http.NewServeMux()
	
	mux.HandleFunc("/api/health", healthHandler)
	mux.HandleFunc("/api/game/status", gameStatusHandler)
	
	fmt.Println("Planets! server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}

func gameStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := GameStatusResponse{
		Game:          "Planets!",
		Turn:          1,
		OnlinePlayers: 0,
	}
	
	json.NewEncoder(w).Encode(response)
}
