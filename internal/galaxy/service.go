package galaxy

import (
	"fmt"
	"log/slog"
	"planets-server/internal/shared/config"
)

type Service struct {
	repo   *Repository
	logger *slog.Logger
}

func NewService(repo *Repository, logger *slog.Logger) *Service {
	logger.Debug("Initializing galaxy service")

	return &Service{
		repo:   repo,
		logger: logger,
	}
}

func (s *Service) GetGalaxiesByGameID(gameID int) ([]Galaxy, error) {
	return s.repo.GetGalaxiesByGameID(gameID)
}

func (s *Service) CreateGalaxy(gameID, galaxyX, galaxyY int, name string) error {
	return s.repo.CreateGalaxy(gameID, galaxyX, galaxyY, name)
}

// GenerateGalaxies creates galaxies in a game according to the provided configuration
func (s *Service) GenerateGalaxies(gameID int) (int, error) {
	logger := s.logger.With("component", "galaxy_service", "operation", "generate_galaxies", "game_id", gameID)
	logger.Debug("Generating galaxies")

	cfg := config.GlobalConfig
	// For now, we create a single galaxy at (0,0)
	err := s.repo.CreateGalaxy(gameID, 0, 0, cfg.Universe.DefaultGalaxyName)

	if err != nil {
		logger.Error("Failed to create galaxy", "error", err)
		return 0, fmt.Errorf("failed to create galaxy: %w", err)
	}

	logger.Info("Galaxies generated")
	return 1, nil
}
