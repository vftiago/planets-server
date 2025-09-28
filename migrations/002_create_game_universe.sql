CREATE TABLE IF NOT EXISTS games (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    galaxy_count INTEGER NOT NULL DEFAULT 0,
    sector_count INTEGER NOT NULL DEFAULT 0,
    system_count INTEGER NOT NULL DEFAULT 0,
    planet_count INTEGER NOT NULL DEFAULT 0,
        status VARCHAR(20) NOT NULL DEFAULT 'creating',
    current_turn INTEGER NOT NULL DEFAULT 0,
    max_players INTEGER NOT NULL DEFAULT 10,
    turn_interval_hours INTEGER NOT NULL DEFAULT 1,
    next_turn_at TIMESTAMP,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS galaxies (
    id SERIAL PRIMARY KEY,
    game_id INTEGER REFERENCES games(id) ON DELETE CASCADE,
    galaxy_x INTEGER NOT NULL DEFAULT 0,
    galaxy_y INTEGER NOT NULL DEFAULT 0,
    name VARCHAR(100) NOT NULL,
    sector_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(game_id, galaxy_x, galaxy_y)
);

CREATE TABLE IF NOT EXISTS sectors (
    id SERIAL PRIMARY KEY,
    galaxy_id INTEGER REFERENCES galaxies(id) ON DELETE CASCADE,
    sector_x INTEGER NOT NULL,
    sector_y INTEGER NOT NULL,
    name VARCHAR(100) NOT NULL,
    system_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(galaxy_id, sector_x, sector_y)
);

CREATE TABLE IF NOT EXISTS systems (
    id SERIAL PRIMARY KEY,
    sector_id INTEGER REFERENCES sectors(id) ON DELETE CASCADE,
    system_x INTEGER NOT NULL,
    system_y INTEGER NOT NULL,
    name VARCHAR(100) NOT NULL,
    planet_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(sector_id, system_x, system_y)
);

CREATE TABLE IF NOT EXISTS planets (
    id SERIAL PRIMARY KEY,
    game_id INTEGER REFERENCES games(id) ON DELETE CASCADE,
    system_id INTEGER REFERENCES systems(id) ON DELETE CASCADE,
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
