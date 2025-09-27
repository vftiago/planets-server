package planet

import (
	"time"
)

type PlanetType string

const (
	PlanetTypeBarren      PlanetType = "barren"
	PlanetTypeTerrestrial PlanetType = "terrestrial"
	PlanetTypeGasGiant    PlanetType = "gas_giant"
	PlanetTypeIce         PlanetType = "ice"
	PlanetTypeVolcanic    PlanetType = "volcanic"
)

type Planet struct {
	ID            int        `json:"id"`
	SystemID      int        `json:"system_id"`
	PlanetIndex   int        `json:"planet_index"`
	Name          string     `json:"name"`
	Type          PlanetType `json:"type"`
	Size          int        `json:"size"`
	Population    int64      `json:"population"`
	MaxPopulation int64      `json:"max_population"`
	OwnerID       *int       `json:"owner_id"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
