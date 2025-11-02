package player

import (
	"context"
	"planets-server/internal/shared/config"
	"planets-server/internal/shared/errors"
	"strings"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
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
	cfg := config.GlobalConfig
	isAdminEmail := cfg != nil && email == cfg.Admin.Email

	player, err := s.repo.FindPlayerByEmail(ctx, email)
	if err != nil && errors.GetType(err) != errors.ErrorTypeNotFound {
		return nil, errors.WrapInternal("failed to check for existing player by email", err)
	}

	if player != nil {
		if isAdminEmail && player.Role != PlayerRoleAdmin {
			if err := s.repo.UpdatePlayerRole(ctx, player.ID, PlayerRoleAdmin); err != nil {
				return nil, errors.WrapInternal("failed to upgrade player to admin role", err)
			}
			player.Role = PlayerRoleAdmin
		}
		return player, nil
	}

	username := s.generateUsernameFromEmail(email)

	if isAdminEmail && cfg != nil {
		username = cfg.Admin.Username
		displayName = cfg.Admin.DisplayName
	}

	player, err = s.repo.CreatePlayer(ctx, username, email, displayName, avatarURL)
	if err != nil {
		return nil, errors.WrapInternal("failed to create player", err)
	}

	return player, nil
}

func (s *Service) generateUsernameFromEmail(email string) string {
	if idx := strings.Index(email, "@"); idx > 0 {
		return email[:idx]
	}
	return "player"
}
