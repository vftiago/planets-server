DROP TABLE IF EXISTS player_stats CASCADE;
DROP TABLE IF EXISTS game_players CASCADE;
DROP TABLE IF EXISTS planets CASCADE;
DROP TABLE IF EXISTS spatial_entities CASCADE;
DROP TABLE IF EXISTS games CASCADE;
DROP TABLE IF EXISTS player_auth_providers CASCADE;
DROP TABLE IF EXISTS players CASCADE;

DROP TYPE IF EXISTS planet_type CASCADE;
DROP TYPE IF EXISTS entity_type CASCADE;

DROP FUNCTION IF EXISTS update_updated_at_column CASCADE;
DROP FUNCTION IF EXISTS update_spatial_child_counts CASCADE;
DROP FUNCTION IF EXISTS update_planet_counts CASCADE;
DROP FUNCTION IF EXISTS update_player_stats CASCADE;

CREATE TYPE planet_type AS ENUM ('barren', 'terrestrial', 'gas_giant', 'ice', 'volcanic');
CREATE TYPE entity_type AS ENUM ('galaxy', 'sector', 'system');

CREATE TABLE players (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    avatar_url TEXT,
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT check_player_role CHECK (role IN ('user', 'admin'))
);

CREATE TABLE player_auth_providers (
    id SERIAL PRIMARY KEY,
    player_id INTEGER REFERENCES players(id) ON DELETE CASCADE,
    provider VARCHAR(20) NOT NULL,
    provider_user_id VARCHAR(100),
    provider_email VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(provider, provider_user_id),
    UNIQUE(player_id, provider)
);

CREATE TABLE games (
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

CREATE TABLE spatial_entities (
    id SERIAL PRIMARY KEY,
    game_id INTEGER NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    parent_id INTEGER REFERENCES spatial_entities(id) ON DELETE CASCADE,
    entity_type entity_type NOT NULL,
    level INTEGER NOT NULL,
    x_coord INTEGER NOT NULL,
    y_coord INTEGER NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    child_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    CHECK (level > 0),
    CHECK ((level = 1 AND parent_id IS NULL) OR (level > 1 AND parent_id IS NOT NULL))
);

CREATE UNIQUE INDEX idx_spatial_entities_level1_coords ON spatial_entities (game_id, x_coord, y_coord) WHERE level = 1;
CREATE UNIQUE INDEX idx_spatial_entities_parent_coords ON spatial_entities (parent_id, x_coord, y_coord) WHERE parent_id IS NOT NULL;

CREATE TABLE planets (
    id SERIAL PRIMARY KEY,
    system_id INTEGER REFERENCES spatial_entities(id) ON DELETE CASCADE,
    planet_index INTEGER NOT NULL,
    name VARCHAR(100) NOT NULL,
    type planet_type NOT NULL DEFAULT 'terrestrial',
    size INTEGER NOT NULL DEFAULT 100,
    population BIGINT NOT NULL DEFAULT 0,
    max_population BIGINT NOT NULL DEFAULT 1000000,
    owner_id INTEGER REFERENCES players(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(system_id, planet_index)
);

CREATE TABLE game_players (
    id SERIAL PRIMARY KEY,
    game_id INTEGER REFERENCES games(id) ON DELETE CASCADE,
    player_id INTEGER REFERENCES players(id) ON DELETE CASCADE,
    joined_at TIMESTAMP DEFAULT NOW(),
    is_active BOOLEAN NOT NULL DEFAULT true,
    UNIQUE(game_id, player_id)
);

CREATE TABLE player_stats (
    id SERIAL PRIMARY KEY,
    game_id INTEGER REFERENCES games(id) ON DELETE CASCADE,
    player_id INTEGER REFERENCES players(id) ON DELETE CASCADE,
    total_planets INTEGER NOT NULL DEFAULT 0,
    total_population BIGINT NOT NULL DEFAULT 0,
    total_ships INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(game_id, player_id)
);

CREATE INDEX idx_players_email ON players(email);
CREATE INDEX idx_players_role ON players(role);
CREATE INDEX idx_auth_providers_player_id ON player_auth_providers(player_id);
CREATE INDEX idx_auth_providers_provider ON player_auth_providers(provider, provider_user_id);
CREATE INDEX idx_games_status ON games(status);
CREATE INDEX idx_games_next_turn ON games(next_turn_at) WHERE next_turn_at IS NOT NULL;
CREATE INDEX idx_spatial_entities_game_id ON spatial_entities(game_id);
CREATE INDEX idx_spatial_entities_parent_id ON spatial_entities(parent_id);
CREATE INDEX idx_spatial_entities_type_level ON spatial_entities(entity_type, level);
CREATE INDEX idx_spatial_entities_coords ON spatial_entities(parent_id, x_coord, y_coord);
CREATE INDEX idx_spatial_entities_game_level ON spatial_entities(game_id, level);
CREATE INDEX idx_planets_system_id ON planets(system_id);
CREATE INDEX idx_planets_owner_id ON planets(owner_id);
CREATE INDEX idx_game_players_game_id ON game_players(game_id);
CREATE INDEX idx_game_players_player_id ON game_players(player_id);
CREATE INDEX idx_player_stats_game_player ON player_stats(game_id, player_id);
CREATE INDEX idx_games_planet_count ON games(planet_count) WHERE planet_count > 0;
CREATE INDEX idx_spatial_entities_game_type ON spatial_entities(game_id, entity_type);
CREATE INDEX idx_planets_type ON planets(type);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_spatial_child_counts()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE spatial_entities SET child_count = child_count + 1 WHERE id = NEW.parent_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE spatial_entities SET child_count = child_count - 1 WHERE id = OLD.parent_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_planet_counts()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE spatial_entities SET child_count = child_count + 1 WHERE id = NEW.system_id;
        UPDATE games SET planet_count = planet_count + 1 
        WHERE id = (SELECT game_id FROM spatial_entities WHERE id = NEW.system_id);
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE spatial_entities SET child_count = child_count - 1 WHERE id = OLD.system_id;
        UPDATE games SET planet_count = planet_count - 1 
        WHERE id = (SELECT game_id FROM spatial_entities WHERE id = OLD.system_id);
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_player_stats()
RETURNS TRIGGER AS $$
DECLARE
    v_game_id INTEGER;
BEGIN
    SELECT game_id INTO v_game_id FROM spatial_entities WHERE id = COALESCE(NEW.system_id, OLD.system_id);
    
    IF TG_OP = 'INSERT' AND NEW.owner_id IS NOT NULL THEN
        INSERT INTO player_stats (game_id, player_id, total_planets, total_population)
        VALUES (v_game_id, NEW.owner_id, 1, NEW.population)
        ON CONFLICT (game_id, player_id)
        DO UPDATE SET 
            total_planets = player_stats.total_planets + 1,
            total_population = player_stats.total_population + NEW.population;
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        IF OLD.owner_id IS DISTINCT FROM NEW.owner_id THEN
            IF OLD.owner_id IS NOT NULL THEN
                UPDATE player_stats 
                SET total_planets = total_planets - 1,
                    total_population = total_population - OLD.population
                WHERE game_id = v_game_id AND player_id = OLD.owner_id;
            END IF;
            IF NEW.owner_id IS NOT NULL THEN
                INSERT INTO player_stats (game_id, player_id, total_planets, total_population)
                VALUES (v_game_id, NEW.owner_id, 1, NEW.population)
                ON CONFLICT (game_id, player_id)
                DO UPDATE SET 
                    total_planets = player_stats.total_planets + 1,
                    total_population = player_stats.total_population + NEW.population;
            END IF;
        ELSIF NEW.owner_id IS NOT NULL AND OLD.population != NEW.population THEN
            UPDATE player_stats 
            SET total_population = total_population + (NEW.population - OLD.population)
            WHERE game_id = v_game_id AND player_id = NEW.owner_id;
        END IF;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' AND OLD.owner_id IS NOT NULL THEN
        UPDATE player_stats 
        SET total_planets = total_planets - 1,
            total_population = total_population - OLD.population
        WHERE game_id = v_game_id AND player_id = OLD.owner_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_games_updated_at BEFORE UPDATE ON games FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_spatial_entities_updated_at BEFORE UPDATE ON spatial_entities FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_planets_updated_at BEFORE UPDATE ON planets FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_player_stats_updated_at BEFORE UPDATE ON player_stats FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_update_spatial_child_counts AFTER INSERT OR DELETE ON spatial_entities FOR EACH ROW EXECUTE FUNCTION update_spatial_child_counts();
CREATE TRIGGER trigger_update_planet_counts AFTER INSERT OR DELETE ON planets FOR EACH ROW EXECUTE FUNCTION update_planet_counts();
CREATE TRIGGER trigger_update_player_stats AFTER INSERT OR UPDATE OR DELETE ON planets FOR EACH ROW EXECUTE FUNCTION update_player_stats();
