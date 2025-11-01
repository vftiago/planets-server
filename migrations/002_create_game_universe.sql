CREATE TABLE IF NOT EXISTS games (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    
    universe_name VARCHAR(100) NOT NULL DEFAULT 'Game Universe',
    universe_description TEXT,
    planet_count INTEGER NOT NULL DEFAULT 0,
    
    status VARCHAR(20) NOT NULL DEFAULT 'creating',
    current_turn INTEGER NOT NULL DEFAULT 0,
    max_players INTEGER NOT NULL DEFAULT 10,
    turn_interval_hours INTEGER NOT NULL DEFAULT 1,
    next_turn_at TIMESTAMP,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS spatial_entities (
    id SERIAL PRIMARY KEY,
    game_id INTEGER NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    parent_id INTEGER REFERENCES spatial_entities(id) ON DELETE CASCADE,

    entity_type VARCHAR(20) NOT NULL,
    level INTEGER NOT NULL,
    x_coord INTEGER NOT NULL,
    y_coord INTEGER NOT NULL,

    name VARCHAR(100) NOT NULL,
    description TEXT,
    child_count INTEGER NOT NULL DEFAULT 0,

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    CHECK (entity_type IN ('galaxy', 'sector', 'system')),
    CHECK (level > 0),
    CHECK ((level = 1 AND parent_id IS NULL) OR (level > 1 AND parent_id IS NOT NULL))
);

-- Unique constraint for level 1 entities (galaxies): coordinates unique within game
CREATE UNIQUE INDEX idx_spatial_entities_level1_coords ON spatial_entities (game_id, x_coord, y_coord) WHERE level = 1;

-- Unique constraint for level 2+ entities: coordinates unique within parent
CREATE UNIQUE INDEX idx_spatial_entities_parent_coords ON spatial_entities (parent_id, x_coord, y_coord) WHERE parent_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS planets (
    id SERIAL PRIMARY KEY,
    system_id INTEGER REFERENCES spatial_entities(id) ON DELETE CASCADE,
    planet_index INTEGER NOT NULL,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'terrestrial',
    size INTEGER NOT NULL DEFAULT 100,
    population BIGINT NOT NULL DEFAULT 0,
    max_population BIGINT NOT NULL DEFAULT 1000000,
    owner_id INTEGER REFERENCES players(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(system_id, planet_index)
);
