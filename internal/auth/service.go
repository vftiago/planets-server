package auth

import (
	"context"
	"log/slog"
)

type Service struct {
	repo   *Repository
	logger *slog.Logger
}

func NewService(repo *Repository, logger *slog.Logger) *Service {
	logger.Debug("Initializing auth service")

	return &Service{
		repo:   repo,
		logger: logger,
	}
}

func (s *Service) CreateAuthProvider(ctx context.Context, playerID int, provider, providerUserID, providerEmail string) error {
	return s.repo.CreateAuthProvider(ctx, playerID, provider, providerUserID, providerEmail)
}

func (s *Service) FindPlayerByAuthProvider(ctx context.Context, provider, providerUserID string) (int, error) {
	return s.repo.FindPlayerByAuthProvider(ctx, provider, providerUserID)
}
