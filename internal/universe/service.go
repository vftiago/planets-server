package universe

import (
	"fmt"
	"log/slog"

	"planets-server/internal/galaxy"
	"planets-server/internal/planet"
	"planets-server/internal/sector"
	"planets-server/internal/system"
)

type Service struct {
	repo          *Repository
	galaxyService *galaxy.Service
	sectorService *sector.Service
	systemService *system.Service
	planetService *planet.Service
	logger        *slog.Logger
}

func NewService(repo *Repository, galaxyService *galaxy.Service, sectorService *sector.Service, systemService *system.Service, planetService *planet.Service, logger *slog.Logger) *Service {
	return &Service{
		repo:          repo,
		galaxyService: galaxyService,
		sectorService: sectorService,
		systemService: systemService,
		planetService: planetService,
		logger:        logger,
	}
}

// CreateUniverse creates a new universe with the given configuration
func (s *Service) CreateUniverse(config UniverseConfig) (*Universe, error) {
	s.logger.Info("Creating new universe")

	universe := &Universe{
		// Counts will be updated after generation
		GalaxyCount: 0,
		SectorCount: 0,
		SystemCount: 0,
		PlanetCount: 0,
	}

	err := s.repo.CreateUniverse(universe)
	if err != nil {
		return nil, fmt.Errorf("failed to create universe: %w", err)
	}

	// Generate the universe content
	err = s.generateUniverse(universe.ID, config)
	if err != nil {
		s.logger.Error("Failed to generate universe", "universe_id", universe.ID, "error", err)
		// Clean up the universe if generation failed
		s.repo.DeleteUniverse(universe.ID)
		return nil, fmt.Errorf("failed to generate universe: %w", err)
	}

	// Reload the universe with updated counts
	return s.repo.GetUniverse(universe.ID)
}

// GetUniverse retrieves a universe by ID
func (s *Service) GetUniverse(id int) (*Universe, error) {
	return s.repo.GetUniverse(id)
}

// ListUniverses retrieves all universes
func (s *Service) ListUniverses() ([]*Universe, error) {
	return s.repo.ListUniverses()
}

// DeleteUniverse deletes a universe and all its data
func (s *Service) DeleteUniverse(id int) error {
	s.logger.Info("Deleting universe", "universe_id", id)
	return s.repo.DeleteUniverse(id)
}

// generateUniverse orchestrates the generation of universe content using repositories
func (s *Service) generateUniverse(universeID int, config UniverseConfig) error {
	s.logger.Info("Starting universe generation", "universe_id", universeID)

	var totalGalaxies, totalSectors, totalSystems, totalPlanets int

	// Generate galaxy using galaxy service
	galaxyCount, err := s.galaxyService.GenerateGalaxies(universeID)
	if err != nil {
		return fmt.Errorf("failed to generate galaxies: %w", err)
	}
	totalGalaxies += galaxyCount

	// Get all galaxies to generate their content
	galaxies, err := s.galaxyService.GetGalaxiesByUniverseID(universeID)
	if err != nil {
		return fmt.Errorf("failed to get galaxies: %w", err)
	}

	for _, galaxy := range galaxies {
		// Generate sectors for this galaxy
		sectorCount, err := s.sectorService.GenerateSectors(galaxy.ID, config.SectorsPerGalaxy)
		if err != nil {
			return fmt.Errorf("failed to generate sectors for galaxy %d: %w", galaxy.ID, err)
		}
		totalSectors += sectorCount

		// Get all sectors in this galaxy to generate their content
		sectors, err := s.sectorService.GetSectorsByGalaxyID(galaxy.ID)
		if err != nil {
			return fmt.Errorf("failed to get sectors for galaxy %d: %w", galaxy.ID, err)
		}

		for _, sector := range sectors {
			// Generate systems for this sector
			systemCount, err := s.systemService.GenerateSystems(sector.ID, config.SystemsPerSector)
			if err != nil {
				return fmt.Errorf("failed to generate systems for sector %d: %w", sector.ID, err)
			}
			totalSystems += systemCount

			// Get all systems in this sector to generate their content
			systems, err := s.systemService.GetSystemsBySectorID(sector.ID)
			if err != nil {
				return fmt.Errorf("failed to get systems for sector %d: %w", sector.ID, err)
			}

			for _, system := range systems {
				// Generate planets for this system
				planetCount, err := s.planetService.GeneratePlanets(system.ID, config.MinPlanetsPerSystem, config.MaxPlanetsPerSystem)
				if err != nil {
					return fmt.Errorf("failed to generate planets for system %d: %w", system.ID, err)
				}
				totalPlanets += planetCount
			}
		}
	}

	// Update the universe with final counts
	err = s.repo.UpdateUniverseCounts(universeID, totalGalaxies, totalSectors, totalSystems, totalPlanets)
	if err != nil {
		return fmt.Errorf("failed to update universe counts: %w", err)
	}

	s.logger.Info("Universe generation completed",
		"universe_id", universeID,
		"galaxies", totalGalaxies,
		"sectors", totalSectors,
		"systems", totalSystems,
		"planets", totalPlanets)

	return nil
}
