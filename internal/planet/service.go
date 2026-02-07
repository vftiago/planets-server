package planet

import (
	"context"
	"fmt"
	"math/rand"
	"planets-server/internal/shared/database"
	"planets-server/internal/shared/errors"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
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

func (s *Service) GeneratePlanets(ctx context.Context, systemIDs []int, minPlanets, maxPlanets int, tx *database.Tx) (int, error) {
	if len(systemIDs) == 0 {
		return 0, nil
	}

	planetNames := s.generatePlanetNames()
	var batchRequests []BatchInsertRequest

	// Prepare all planets for all systems upfront
	for _, systemID := range systemIDs {
		// Check for context cancellation
		if err := ctx.Err(); err != nil {
			return 0, errors.WrapInternal("planet generation cancelled", err)
		}

		planetCount := minPlanets + rand.Intn(maxPlanets-minPlanets+1)

		for i := 0; i < planetCount; i++ {
			planetName := fmt.Sprintf("Planet %s", planetNames[i%len(planetNames)])

			batchRequests = append(batchRequests, BatchInsertRequest{
				SystemID:      systemID,
				PlanetIndex:   i,
				Name:          planetName,
				Type:          s.generateRandomPlanetType(),
				Size:          50 + rand.Intn(151),
				MaxPopulation: int64(100000 + rand.Intn(900000)),
			})
		}
	}

	// Perform batch insert for all planets
	if len(batchRequests) == 0 {
		return 0, nil
	}

	count, err := s.repo.CreatePlanetsBatch(ctx, batchRequests, tx)
	if err != nil {
		return 0, errors.WrapInternal("failed to batch create planets", err)
	}

	return count, nil
}
