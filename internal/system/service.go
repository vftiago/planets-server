package system

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
	logger.Debug("Initializing system service")

	return &Service{
		repo:   repo,
		logger: logger,
	}
}

func (s *Service) GetSystemsBySectorID(sectorID int) ([]System, error) {
	return s.repo.GetSystemsBySectorID(sectorID)
}

func (s *Service) CreateSystem(sectorID, systemX, systemY int, name string) (*System, error) {
	return s.repo.CreateSystem(sectorID, systemX, systemY, name)
}

// GenerateSystems creates systems in a sector according to the provided configuration
func (s *Service) GenerateSystems(sectorID int, systemsPerSector int) (int, error) {
	logger := s.logger.With("operation", "generate_systems", "sector_id", sectorID, "systems_per_sector", systemsPerSector)
	logger.Debug("Generating systems")

	systemsPerSide := int(math.Sqrt(float64(systemsPerSector)))
	if systemsPerSide*systemsPerSide != systemsPerSector {
		systemsPerSide = int(math.Ceil(math.Sqrt(float64(systemsPerSector))))
	}

	systemNames := s.generateSystemNames()
	nameIndex := 0
	totalSystems := 0

	for x := 0; x < systemsPerSide; x++ {
		for y := 0; y < systemsPerSide; y++ {
			if totalSystems >= systemsPerSector {
				break
			}

			systemName := systemNames[nameIndex%len(systemNames)]
			nameIndex++

			_, err := s.repo.CreateSystem(sectorID, x, y, systemName)
			if err != nil {
				logger.Error("Failed to create system", "error", err, "coordinates", fmt.Sprintf("(%d,%d)", x, y))
				return 0, fmt.Errorf("failed to create system at (%d,%d): %w", x, y, err)
			}
			totalSystems++
		}
		if totalSystems >= systemsPerSector {
			break
		}
	}

	logger.Info("Systems generated", "count", totalSystems)
	return totalSystems, nil
}

// generateSystemNames returns a list of system names
func (s *Service) generateSystemNames() []string {
	return []string{
		"Altair", "Vega", "Sirius", "Arcturus", "Capella", "Rigel", "Procyon",
		"Betelgeuse", "Aldebaran", "Spica", "Antares", "Pollux", "Fomalhaut",
		"Deneb", "Regulus", "Adhara", "Castor", "Gacrux", "Bellatrix", "Elnath",
		"Miaplacidus", "Alnilam", "Alnair", "Alioth", "Dubhe", "Mirfak", "Wezen",
		"Sargas", "Kaus", "Avior", "Menkalinan", "Atria", "Alhena", "Peacock",
		"Alsephina", "Mirzam", "Polaris", "Alphard", "Hamal", "Algieba", "Diphda",
		"Mizar", "Nunki", "Menkent", "Mirach", "Alpheratz", "Rasalhague", "Kochab",
		"Saiph", "Zubenelgenubi", "Enif", "Schedar", "Markab", "Unukalhai", "Tau",
	}
}
