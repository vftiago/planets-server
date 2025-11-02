package player

import (
	"context"
	"fmt"
	"log/slog"
	"planets-server/internal/shared/config"
	"planets-server/internal/shared/errors"
	"strings"
)

type Service struct {
	repo   *Repository
	logger *slog.Logger
}

func NewService(repo *Repository, logger *slog.Logger) *Service {
	logger.Debug("Initializing player service")

	return &Service{
		repo:   repo,
		logger: logger,
	}
}

func (s *Service) GetPlayerCount(ctx context.Context) (int, error) {
	return s.repo.GetPlayerCount(ctx)
}

func (s *Service) GetAllPlayers(ctx context.Context) ([]Player, error) {
	return s.repo.GetAllPlayers(ctx)
}

func (s *Service) GetPlayerByID(ctx context.Context, id int) (*Player, error) {
	return s.repo.GetPlayerByID(ctx, id)
}

func (s *Service) CreatePlayer(ctx context.Context, username, email, displayName string, avatarURL *string) (*Player, error) {
	return s.repo.CreatePlayer(ctx, username, email, displayName, avatarURL)
}

func (s *Service) FindOrCreatePlayerByOAuth(ctx context.Context, provider, providerUserID, email, displayName string, avatarURL *string) (*Player, error) {
	logger := s.logger.With(
		"component", "player_service",
		"operation", "find_or_create_oauth",
		"provider", provider,
		"email", email,
	)
	logger.Debug("Finding or creating player by OAuth")

	cfg := config.GlobalConfig
	isAdminEmail := cfg != nil && email == cfg.Admin.Email

	player, err := s.repo.FindPlayerByEmail(ctx, email)
	if err != nil && errors.GetType(err) != errors.ErrorTypeNotFound {
		logger.Error("Database error checking for player by email", "error", err)
		return nil, fmt.Errorf("database error: %w", err)
	}

	if player != nil {
		logger.Info("Found existing player by email", "player_id", player.ID, "role", player.Role)
		if isAdminEmail && player.Role != PlayerRoleAdmin {
			logger.Info("Upgrading existing user to admin role", "player_id", player.ID)
			if err := s.repo.UpdatePlayerRole(ctx, player.ID, PlayerRoleAdmin); err != nil {
				logger.Error("Failed to upgrade user to admin", "error", err)
				return nil, fmt.Errorf("failed to upgrade to admin: %w", err)
			}
			player.Role = PlayerRoleAdmin
		}
		return player, nil
	}

	logger.Info("No existing player found, creating new player with OAuth provider")
	username := s.generateUsernameFromEmail(email)

	if isAdminEmail && cfg != nil {
		username = cfg.Admin.Username
		displayName = cfg.Admin.DisplayName
		logger.Info("Creating new admin user via OAuth")
	}

	player, err = s.repo.CreatePlayer(ctx, username, email, displayName, avatarURL)
	if err != nil {
		logger.Error("Failed to create player", "error", err)
		return nil, fmt.Errorf("failed to create player: %w", err)
	}

	logger.Info("Successfully created new player with OAuth",
		"player_id", player.ID,
		"username", player.Username,
		"role", player.Role,
		"provider", provider)

	return player, nil
}

func (s *Service) generateUsernameFromEmail(email string) string {
	if idx := strings.Index(email, "@"); idx > 0 {
		return email[:idx]
	}
	return "player"
}
