package sector

import (
	"fmt"
	"log/slog"
	"math"
)

type Service struct {
	repo   *Repository
	logger *slog.Logger
}

func NewService(repo *Repository, logger *slog.Logger) *Service {
	logger.Debug("Initializing sector service")

	return &Service{
		repo:   repo,
		logger: logger,
	}
}

func (s *Service) GetSectorsByGalaxyID(galaxyID int) ([]Sector, error) {
	return s.repo.GetSectorsByGalaxyID(galaxyID)
}

func (s *Service) CreateSector(galaxyID, sectorX, sectorY int, name string) (*Sector, error) {
	return s.repo.CreateSector(galaxyID, sectorX, sectorY, name)
}

// GenerateSectors creates sectors in a galaxy according to the provided configuration
func (s *Service) GenerateSectors(galaxyID int, sectorCount int) (int, error) {
	logger := s.logger.With("component", "sector_service", "operation", "generate_sectors", "galaxy_id", galaxyID, "sector_count", sectorCount)
	logger.Debug("Generating sectors")

	sectorSize := int(math.Sqrt(float64(sectorCount)))
	if sectorSize*sectorSize < sectorCount {
		sectorSize++
	}

	sectorNames := s.generateSectorNames()
	nameIndex := 0
	totalSectors := 0

	for x := 0; x < sectorSize; x++ {
		for y := 0; y < sectorSize; y++ {
			if totalSectors >= sectorCount {
				break
			}

			sectorName := sectorNames[nameIndex%len(sectorNames)]
			nameIndex++

			_, err := s.repo.CreateSector(galaxyID, x, y, sectorName)

			if err != nil {
				logger.Error("Failed to create sector", "error", err, "coordinates", fmt.Sprintf("(%d,%d)", x, y))
				return 0, fmt.Errorf("failed to create sector at (%d,%d): %w", x, y, err)
			}

			totalSectors++
		}
		if totalSectors >= sectorCount {
			break
		}
	}

	logger.Info("Sectors generated", "count", totalSectors)
	return totalSectors, nil
}

// generateSectorNames returns a list of sector names
func (s *Service) generateSectorNames() []string {
	return []string{
		"Alpha", "Beta", "Gamma", "Delta", "Epsilon", "Zeta", "Eta", "Theta",
		"Iota", "Kappa", "Lambda", "Mu", "Nu", "Xi", "Omicron", "Pi",
		"Rho", "Sigma", "Tau", "Upsilon", "Phi", "Chi", "Psi", "Omega",
		"Prime", "Core", "Frontier", "Outer", "Inner", "Central", "Remote",
		"Azure", "Crimson", "Golden", "Silver", "Emerald", "Violet", "Amber",
	}
}
