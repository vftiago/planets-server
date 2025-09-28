CREATE TABLE IF NOT EXISTS games (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    
    -- Universe properties
    universe_name VARCHAR(100) NOT NULL DEFAULT 'Game Universe',
    universe_description TEXT,
    planet_count INTEGER NOT NULL DEFAULT 0,
    
    -- Game properties
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
    parent_id INTEGER REFERENCES spatial_entities(id) ON DELETE CASCADE,
    
    -- Entity type and position
    entity_type VARCHAR(20) NOT NULL, -- 'galaxy', 'sector', 'system', 'region', etc.
    level INTEGER NOT NULL,           -- 1=galaxy, 2=sector, 3=system, etc.
    x_coord INTEGER NOT NULL,
    y_coord INTEGER NOT NULL,
    
    -- Metadata
    name VARCHAR(100) NOT NULL,
    description TEXT,
    child_count INTEGER NOT NULL DEFAULT 0,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    -- Constraints
    UNIQUE(parent_id, x_coord, y_coord),
    UNIQUE(level, x_coord, y_coord),
    CHECK (entity_type IN ('galaxy', 'sector', 'region', 'system')),
    CHECK (level > 0),
    CHECK (parent_id IS NOT NULL)
);

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
    is_homeworld BOOLEAN NOT NULL DEFAULT false,
    is_neutral BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(system_id, planet_index)
);
