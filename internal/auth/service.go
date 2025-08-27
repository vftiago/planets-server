package auth

import (
	"log/slog"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	logger := slog.With("component", "auth_service", "operation", "init")
	logger.Debug("Initializing auth service")
	return &Service{repo: repo}
}

func (s *Service) CreateAuthProvider(playerID int, provider, providerUserID, providerEmail string) error {
	return s.repo.CreateAuthProvider(playerID, provider, providerUserID, providerEmail)
}

func (s *Service) FindPlayerByAuthProvider(provider, providerUserID string) (int, error) {
	return s.repo.FindPlayerByAuthProvider(provider, providerUserID)
}
