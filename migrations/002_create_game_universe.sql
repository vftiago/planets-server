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
    parent_id INTEGER NOT NULL,
    
    entity_type VARCHAR(20) NOT NULL,
    level INTEGER NOT NULL,
    x_coord INTEGER NOT NULL,
    y_coord INTEGER NOT NULL,
    
    name VARCHAR(100) NOT NULL,
    description TEXT,
    child_count INTEGER NOT NULL DEFAULT 0,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(parent_id, x_coord, y_coord),
    CHECK (entity_type IN ('galaxy', 'sector', 'region', 'system')),
    CHECK (level > 0),
    CHECK ((level = 1 AND parent_id = game_id) OR (level > 1))
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
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(system_id, planet_index)
);
