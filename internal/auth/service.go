package auth

import (
	"context"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) CreateAuthProvider(ctx context.Context, playerID int, provider, providerUserID, providerEmail string) error {
	return s.repo.CreateAuthProvider(ctx, playerID, provider, providerUserID, providerEmail)
}

func (s *Service) FindPlayerByAuthProvider(ctx context.Context, provider, providerUserID string) (int, error) {
	return s.repo.FindPlayerByAuthProvider(ctx, provider, providerUserID)
}
