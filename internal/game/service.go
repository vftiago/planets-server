package game

import (
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"time"

	"planets-server/internal/galaxy"
	"planets-server/internal/planet"
	"planets-server/internal/sector"
	"planets-server/internal/system"
)

type Service struct {
	gameRepo    *Repository
	galaxyRepo  *galaxy.Repository
	sectorRepo  *sector.Repository
	systemRepo  *system.Repository
	planetRepo  *planet.Repository
}

func NewService(
	gameRepo *Repository,
	galaxyRepo *galaxy.Repository,
	sectorRepo *sector.Repository,
	systemRepo *system.Repository,
	planetRepo *planet.Repository,
) *Service {
	logger := slog.With("component", "game_service", "operation", "init")
	logger.Debug("Initializing game service")
	return &Service{
		gameRepo:   gameRepo,
		galaxyRepo: galaxyRepo,
		sectorRepo: sectorRepo,
		systemRepo: systemRepo,
		planetRepo: planetRepo,
	}
}

func (s *Service) CreateGame(config GameConfig) (*Game, error) {
	logger := slog.With("component", "game_service", "operation", "create_game", "name", config.Name)
	logger.Info("Creating new game with universe generation")

	game, err := s.gameRepo.CreateGame(config)
	if err != nil {
		logger.Error("Failed to create game", "error", err)
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	logger.Info("Game created, starting universe generation", "game_id", game.ID)

	if err := s.generateUniverse(game.ID, config.UniverseConfig); err != nil {
		logger.Error("Failed to generate universe", "error", err, "game_id", game.ID)
		
		if deleteErr := s.gameRepo.DeleteGame(game.ID); deleteErr != nil {
			logger.Error("Failed to cleanup game after universe generation failure", "error", deleteErr)
		}
		
		return nil, fmt.Errorf("failed to generate universe: %w", err)
	}

	if err := s.gameRepo.ActivateGame(game.ID); err != nil {
		logger.Error("Failed to activate game", "error", err)
		return nil, fmt.Errorf("failed to activate game: %w", err)
	}

	game.Status = GameStatusActive
	game.CurrentTurn = 1
	nextTurn := time.Now().Add(1 * time.Hour).Truncate(time.Hour)
	game.NextTurnAt = &nextTurn

	logger.Info("Game created and activated successfully", 
		"game_id", game.ID, 
		"name", game.Name,
		"next_turn_at", nextTurn)

	return game, nil
}

func (s *Service) generateUniverse(gameID int, config UniverseConfig) error {
	logger := slog.With("component", "game_service", "operation", "generate_universe", "game_id", gameID)
	logger.Info("Starting universe generation",
		"sector_count", config.SectorCount,
		"systems_per_sector", config.SystemsPerSector,
		"planet_range", fmt.Sprintf("%d-%d", config.MinPlanetsPerSystem, config.MaxPlanetsPerSystem))

	startTime := time.Now()

	galaxy, err := s.galaxyRepo.CreateGalaxy(gameID, config.GalaxyName, 
		fmt.Sprintf("Generated galaxy with %d sectors", config.SectorCount), 
		config.SectorCount)
	if err != nil {
		logger.Error("Failed to create galaxy", "error", err)
		return fmt.Errorf("failed to create galaxy: %w", err)
	}

	logger.Info("Galaxy created", "galaxy_id", galaxy.ID, "name", galaxy.Name)

	sectorsPerSide := int(math.Sqrt(float64(config.SectorCount)))
	if sectorsPerSide*sectorsPerSide != config.SectorCount {
		sectorsPerSide = int(math.Ceil(math.Sqrt(float64(config.SectorCount))))
	}

	totalSectors := 0
	totalSystems := 0
	totalPlanets := 0

	for x := 0; x < sectorsPerSide; x++ {
		for y := 0; y < sectorsPerSide; y++ {
			if totalSectors >= config.SectorCount {
				break
			}

			sectorName := s.generateSectorName(x, y)
			sector, err := s.sectorRepo.CreateSector(galaxy.ID, x, y, sectorName)
			if err != nil {
				logger.Error("Failed to create sector", "error", err, "x", x, "y", y)
				return fmt.Errorf("failed to create sector at (%d,%d): %w", x, y, err)
			}

			systemsInSector, planetsInSector, err := s.generateSectorContent(sector, config)
			if err != nil {
				logger.Error("Failed to generate sector content", "error", err, "sector_id", sector.ID)
				return fmt.Errorf("failed to generate content for sector %d: %w", sector.ID, err)
			}

			totalSectors++
			totalSystems += systemsInSector
			totalPlanets += planetsInSector

			if totalSectors%10 == 0 {
				logger.Debug("Generation progress", 
					"sectors_completed", totalSectors,
					"total_sectors", config.SectorCount,
					"systems_generated", totalSystems,
					"planets_generated", totalPlanets)
			}
		}
		if totalSectors >= config.SectorCount {
			break
		}
	}

	generationTime := time.Since(startTime)
	logger.Info("Universe generation completed",
		"game_id", gameID,
		"galaxy_id", galaxy.ID,
		"sectors", totalSectors,
		"systems", totalSystems,
		"planets", totalPlanets,
		"generation_time", generationTime)

	return nil
}

func (s *Service) generateSectorContent(sec *sector.Sector, config UniverseConfig) (int, int, error) {
	logger := slog.With("component", "game_service", "operation", "generate_sector", "sector_id", sec.ID)

	systemsPerSide := int(math.Sqrt(float64(config.SystemsPerSector)))
	if systemsPerSide*systemsPerSide != config.SystemsPerSector {
		systemsPerSide = int(math.Ceil(math.Sqrt(float64(config.SystemsPerSector))))
	}

	systemCount := 0
	totalPlanets := 0

	for x := 0; x < systemsPerSide; x++ {
		for y := 0; y < systemsPerSide; y++ {
			if systemCount >= config.SystemsPerSector {
				break
			}

			systemName := s.generateSystemName(sec, x, y)
			sys, err := s.systemRepo.CreateSystem(sec.ID, x, y, systemName)
			if err != nil {
				logger.Error("Failed to create system", "error", err, "x", x, "y", y)
				return 0, 0, fmt.Errorf("failed to create system at (%d,%d): %w", x, y, err)
			}

			planetsInSystem, err := s.generateSystemContent(sys, config)
			if err != nil {
				logger.Error("Failed to generate system content", "error", err, "system_id", sys.ID)
				return 0, 0, fmt.Errorf("failed to generate content for system %d: %w", sys.ID, err)
			}

			systemCount++
			totalPlanets += planetsInSystem
		}
		if systemCount >= config.SystemsPerSector {
			break
		}
	}

	logger.Debug("Sector content generated", 
		"sector_id", sec.ID,
		"systems", systemCount,
		"planets", totalPlanets)

	return systemCount, totalPlanets, nil
}

func (s *Service) generateSystemContent(sys *system.System, config UniverseConfig) (int, error) {
	planetCount := s.generatePlanetCount(config.MinPlanetsPerSystem, config.MaxPlanetsPerSystem)
	
	var planets []planet.Planet
	for i := 0; i < planetCount; i++ {
		planetType := s.generatePlanetType(i, planetCount)
		planetName := s.generatePlanetName(sys, i)
		size := s.generatePlanetSize(planetType)
		maxPop := s.generateMaxPopulation(planetType, size)

		pl := planet.Planet{
			SystemID:      sys.ID,
			PlanetIndex:   i,
			Name:          planetName,
			Type:          planetType,
			Size:          size,
			Population:    0,
			MaxPopulation: maxPop,
			OwnerID:       nil,
			IsHomeworld:   false,
		}
		planets = append(planets, pl)
	}

	if err := s.planetRepo.CreatePlanetsBatch(planets); err != nil {
		return 0, fmt.Errorf("failed to create planets: %w", err)
	}

	return planetCount, nil
}

func (s *Service) generatePlanetCount(min, max int) int {
	weights := make([]float64, max-min+1)
	center := float64(min+max) / 2.0
	
	for i := range weights {
		planetNum := float64(min + i)
		distance := math.Abs(planetNum - center)
		weights[i] = math.Exp(-distance * distance / 8.0)
	}
	
	totalWeight := 0.0
	for _, w := range weights {
		totalWeight += w
	}
	
	for i := range weights {
		weights[i] /= totalWeight
	}
	
	r := rand.Float64()
	cumulative := 0.0
	for i, weight := range weights {
		cumulative += weight
		if r <= cumulative {
			return min + i
		}
	}
	
	return max
}

func (s *Service) generatePlanetType(planetIndex, totalPlanets int) planet.PlanetType {
	types := []planet.PlanetType{
		planet.PlanetTypeBarren,
		planet.PlanetTypeTerrestrial,
		planet.PlanetTypeGasGiant,
		planet.PlanetTypeIce,
		planet.PlanetTypeVolcanic,
	}

	if planetIndex == 0 {
		if rand.Float64() < 0.4 {
			return planet.PlanetTypeVolcanic
		}
	}

	if planetIndex < totalPlanets/3 {
		weights := []float64{0.3, 0.4, 0.1, 0.1, 0.3}
		return s.weightedChoice(types, weights)
	} else if planetIndex < 2*totalPlanets/3 {
		weights := []float64{0.2, 0.5, 0.2, 0.2, 0.1}
		return s.weightedChoice(types, weights)
	} else {
		weights := []float64{0.1, 0.1, 0.4, 0.4, 0.1}
		return s.weightedChoice(types, weights)
	}
}

func (s *Service) weightedChoice(options []planet.PlanetType, weights []float64) planet.PlanetType {
	totalWeight := 0.0
	for _, w := range weights {
		totalWeight += w
	}
	
	r := rand.Float64() * totalWeight
	cumulative := 0.0
	
	for i, weight := range weights {
		cumulative += weight
		if r <= cumulative {
			return options[i]
		}
	}
	
	return options[len(options)-1]
}

func (s *Service) generatePlanetSize(planetType planet.PlanetType) int {
	switch planetType {
	case planet.PlanetTypeGasGiant:
		return 300 + rand.Intn(200)
	case planet.PlanetTypeTerrestrial:
		return 80 + rand.Intn(40)
	case planet.PlanetTypeIce:
		return 60 + rand.Intn(30)
	case planet.PlanetTypeVolcanic:
		return 70 + rand.Intn(35)
	case planet.PlanetTypeBarren:
		return 50 + rand.Intn(50)
	default:
		return 100
	}
}

func (s *Service) generateMaxPopulation(planetType planet.PlanetType, size int) int64 {
	basePopulation := int64(size * 10000)
	
	switch planetType {
	case planet.PlanetTypeTerrestrial:
		return basePopulation * 2
	case planet.PlanetTypeIce:
		return basePopulation / 2
	case planet.PlanetTypeVolcanic:
		return basePopulation / 3
	case planet.PlanetTypeBarren:
		return basePopulation / 4
	case planet.PlanetTypeGasGiant:
		return 0
	default:
		return basePopulation
	}
}

func (s *Service) generateSectorName(x, y int) string {
	prefixes := []string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon", "Zeta", "Eta", "Theta"}
	suffixes := []string{"Quadrant", "Sector", "Region", "Zone", "Territory"}
	
	prefix := prefixes[(x+y)%len(prefixes)]
	suffix := suffixes[(x*y)%len(suffixes)]
	
	return fmt.Sprintf("%s %s %d-%d", prefix, suffix, x, y)
}

func (s *Service) generateSystemName(sec *sector.Sector, x, y int) string {
	prefixes := []string{"Kepler", "Vega", "Altair", "Rigel", "Sirius", "Procyon", "Canopus", "Aldebaran"}
	suffixes := []string{"Prime", "Major", "Minor", "Central", "Outer", "Inner", "Deep", "Far"}
	
	seed := sec.ID + x*100 + y
	prefix := prefixes[seed%len(prefixes)]
	suffix := suffixes[(seed/10)%len(suffixes)]
	
	return fmt.Sprintf("%s %s", prefix, suffix)
}

func (s *Service) generatePlanetName(sys *system.System, planetIndex int) string {
	romanNumerals := []string{"I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX", "X", "XI", "XII"}
	
	if planetIndex < len(romanNumerals) {
		return fmt.Sprintf("%s %s", sys.Name, romanNumerals[planetIndex])
	}
	
	return fmt.Sprintf("%s %d", sys.Name, planetIndex+1)
}

func (s *Service) GetGameByID(gameID int) (*Game, error) {
	return s.gameRepo.GetGameByID(gameID)
}

func (s *Service) GetAllGames() ([]Game, error) {
	return s.gameRepo.GetAllGames()
}

func (s *Service) GetGameStats(gameID int) (*GameStats, error) {
	return s.gameRepo.GetGameStats(gameID)
}

func (s *Service) DeleteGame(gameID int) error {
	logger := slog.With("component", "game_service", "operation", "delete_game", "game_id", gameID)
	logger.Info("Deleting game and all related data")

	return s.gameRepo.DeleteGame(gameID)
}
