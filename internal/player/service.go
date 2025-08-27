package player

import (
	"fmt"
	"log/slog"
	"strings"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	logger := slog.With("component", "player_service", "operation", "init")
	logger.Debug("Initializing player service")
	return &Service{repo: repo}
}

func (s *Service) GetPlayerCount() (int, error) {
	return s.repo.GetPlayerCount()
}

func (s *Service) GetAllPlayers() ([]Player, error) {
	return s.repo.GetAllPlayers()
}

func (s *Service) GetPlayerByID(id int) (*Player, error) {
	return s.repo.GetPlayerByID(id)
}

func (s *Service) CreatePlayer(username, email, displayName string, avatarURL *string) (*Player, error) {
	return s.repo.CreatePlayer(username, email, displayName, avatarURL)
}

func (s *Service) FindOrCreatePlayerByOAuth(provider, providerUserID, email, displayName string, avatarURL *string) (*Player, error) {
	logger := slog.With(
		"component", "player_service",
		"operation", "find_or_create_oauth",
		"provider", provider,
		"email", email,
	)
	logger.Debug("Finding or creating player by OAuth")

	// Check if a player with this email already exists (account linking)
	logger.Debug("Checking for existing player by email for account linking")
	player, err := s.repo.FindPlayerByEmail(email)
	if err != nil {
		logger.Error("Database error checking for player by email", "error", err)
		return nil, fmt.Errorf("database error: %w", err)
	}

	if player != nil {
		logger.Info("Found existing player by email", "player_id", player.ID)
		return player, nil
	}

	// No existing player found, create new one
	logger.Info("Creating new player with OAuth provider")
	username := s.generateUsernameFromEmail(email)
	logger.Debug("Generated username from email", "username", username, "email", email)

	player, err = s.repo.CreatePlayer(username, email, displayName, avatarURL)
	if err != nil {
		logger.Error("Failed to create player", "error", err)
		return nil, fmt.Errorf("failed to create player: %w", err)
	}

	logger.Info("Successfully created new player with OAuth",
		"player_id", player.ID,
		"username", player.Username,
		"provider", provider)

	return player, nil
}

func (s *Service) generateUsernameFromEmail(email string) string {
	if idx := strings.Index(email, "@"); idx > 0 {
		return email[:idx]
	}
	return "player"
}