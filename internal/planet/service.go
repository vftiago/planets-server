package planet

import (
	"fmt"
	"log/slog"
	"math/rand"
)

type Service struct {
	repo   *Repository
	logger *slog.Logger
}

func NewService(repo *Repository, logger *slog.Logger) *Service {
	logger.Debug("Initializing planet service")

	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// GeneratePlanets creates planets in a system according to the provided configuration
func (s *Service) GeneratePlanets(systemID int, minPlanets, maxPlanets int) (int, error) {
	logger := s.logger.With("component", "planet_service", "operation", "generate_planets", "system_id", systemID, "min_planets", minPlanets, "max_planets", maxPlanets)
	logger.Debug("Generating planets")

	planetCount := minPlanets + rand.Intn(maxPlanets-minPlanets+1)
	planetNames := s.generatePlanetNames()

	for i := 0; i < planetCount; i++ {
		planetName := fmt.Sprintf("Planet %s", planetNames[i%len(planetNames)])

		_, err := s.repo.CreatePlanet(systemID, i, planetName, s.generateRandomPlanetType(), 50+rand.Intn(151), int64(100000+rand.Intn(900000)))
		if err != nil {
			logger.Error("Failed to create planet", "error", err, "planet_name", planetName)
			return 0, fmt.Errorf("failed to create planet: %w", err)
		}
	}

	logger.Info("Planets generated", "count", planetCount)
	return planetCount, nil
}

// generatePlanetNames returns a list of planet suffixes
func (s *Service) generatePlanetNames() []string {
	return []string{
		"I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX", "X",
		"Prime", "Alpha", "Beta", "Gamma", "Major", "Minor", "Core", "Outer",
	}
}

// generateRandomPlanetType returns a random planet type
func (s *Service) generateRandomPlanetType() PlanetType {
	types := []PlanetType{
		PlanetTypeBarren,
		PlanetTypeTerrestrial,
		PlanetTypeGasGiant,
		PlanetTypeIce,
		PlanetTypeVolcanic,
	}

	// Weight terrestrial planets more heavily
	weights := []int{15, 40, 20, 15, 10} // Terrestrial is 40% chance
	totalWeight := 0
	for _, w := range weights {
		totalWeight += w
	}

	roll := rand.Intn(totalWeight)
	currentWeight := 0
	for i, weight := range weights {
		currentWeight += weight
		if roll < currentWeight {
			return types[i]
		}
	}

	return PlanetTypeTerrestrial // fallback
}
