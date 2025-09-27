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

func (s *Service) GetGalaxiesByUniverseID(universeID int) ([]Galaxy, error) {
	return s.repo.GetGalaxiesByUniverseID(universeID)
}

func (s *Service) CreateGalaxy(universeID, galaxyX, galaxyY int, name string) error {
	return s.repo.CreateGalaxy(universeID, galaxyX, galaxyY, name)
}

// GenerateGalaxies creates galaxies in a universe according to the provided configuration
func (s *Service) GenerateGalaxies(universeID int) (int, error) {
	logger := s.logger.With("component", "galaxy_service", "operation", "generate_galaxies", "universe_id", universeID)
	logger.Debug("Generating galaxies")

	cfg := config.GlobalConfig
	// For now, we create a single galaxy at (0,0)
	err := s.repo.CreateGalaxy(universeID, 0, 0, cfg.Universe.DefaultGalaxyName)

	if err != nil {
		logger.Error("Failed to create galaxy", "error", err)
		return 0, fmt.Errorf("failed to create galaxy: %w", err)
	}

	logger.Info("Galaxies generated")
	return 1, nil
}
