-- Indexes
CREATE INDEX IF NOT EXISTS idx_games_status ON games(status);
CREATE INDEX IF NOT EXISTS idx_games_next_turn ON games(next_turn_at) WHERE next_turn_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_galaxies_game_id ON galaxies(game_id);
CREATE INDEX IF NOT EXISTS idx_sectors_galaxy_id ON sectors(galaxy_id);
CREATE INDEX IF NOT EXISTS idx_sectors_coordinates ON sectors(galaxy_id, sector_x, sector_y);
CREATE INDEX IF NOT EXISTS idx_systems_sector_id ON systems(sector_id);
CREATE INDEX IF NOT EXISTS idx_systems_coordinates ON systems(sector_id, system_x, system_y);
CREATE INDEX IF NOT EXISTS idx_planets_game_id ON planets(game_id);
CREATE INDEX IF NOT EXISTS idx_planets_system_id ON planets(system_id);
CREATE INDEX IF NOT EXISTS idx_planets_owner_id ON planets(owner_id);
CREATE INDEX IF NOT EXISTS idx_planets_game_owner ON planets(game_id, owner_id);
CREATE INDEX IF NOT EXISTS idx_planets_homeworld ON planets(is_homeworld) WHERE is_homeworld = true;
CREATE INDEX IF NOT EXISTS idx_planets_homeworld_game ON planets(game_id, is_homeworld) WHERE is_homeworld = true;
CREATE INDEX IF NOT EXISTS idx_game_players_game_id ON game_players(game_id);
CREATE INDEX IF NOT EXISTS idx_game_players_player_id ON game_players(player_id);
CREATE INDEX IF NOT EXISTS idx_game_players_homeworld ON game_players(homeworld_planet_id);
CREATE INDEX IF NOT EXISTS idx_player_stats_game_player ON player_stats(game_id, player_id);

-- Functions (DEFINE THESE FIRST!)
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_counts()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        IF TG_TABLE_NAME = 'galaxies' THEN
            UPDATE games SET galaxy_count = galaxy_count + 1 WHERE id = NEW.game_id;
        ELSIF TG_TABLE_NAME = 'sectors' THEN
            UPDATE galaxies SET sector_count = sector_count + 1 WHERE id = NEW.galaxy_id;
        ELSIF TG_TABLE_NAME = 'systems' THEN
            UPDATE sectors SET system_count = system_count + 1 WHERE id = NEW.sector_id;
        ELSIF TG_TABLE_NAME = 'planets' THEN
            UPDATE systems SET planet_count = planet_count + 1 WHERE id = NEW.system_id;
        END IF;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        IF TG_TABLE_NAME = 'galaxies' THEN
            UPDATE games SET galaxy_count = galaxy_count - 1 WHERE id = OLD.game_id;
        ELSIF TG_TABLE_NAME = 'sectors' THEN
            UPDATE galaxies SET sector_count = sector_count - 1 WHERE id = OLD.galaxy_id;
        ELSIF TG_TABLE_NAME = 'systems' THEN
            UPDATE sectors SET system_count = system_count - 1 WHERE id = OLD.sector_id;
        ELSIF TG_TABLE_NAME = 'planets' THEN
            UPDATE systems SET planet_count = planet_count - 1 WHERE id = OLD.system_id;
        END IF;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_player_stats()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' AND NEW.owner_id IS NOT NULL THEN
        INSERT INTO player_stats (game_id, player_id, total_planets, total_population)
        VALUES (NEW.game_id, NEW.owner_id, 1, NEW.population)
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
                WHERE game_id = OLD.game_id AND player_id = OLD.owner_id;
            END IF;
            IF NEW.owner_id IS NOT NULL THEN
                INSERT INTO player_stats (game_id, player_id, total_planets, total_population)
                VALUES (NEW.game_id, NEW.owner_id, 1, NEW.population)
                ON CONFLICT (game_id, player_id)
                DO UPDATE SET 
                    total_planets = player_stats.total_planets + 1,
                    total_population = player_stats.total_population + NEW.population;
            END IF;
        ELSIF NEW.owner_id IS NOT NULL AND OLD.population != NEW.population THEN
            UPDATE player_stats 
            SET total_population = total_population + (NEW.population - OLD.population)
            WHERE game_id = NEW.game_id AND player_id = NEW.owner_id;
        END IF;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' AND OLD.owner_id IS NOT NULL THEN
        UPDATE player_stats 
        SET total_planets = total_planets - 1,
            total_population = total_population - OLD.population
        WHERE game_id = OLD.game_id AND player_id = OLD.owner_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Triggers (AFTER functions are defined)
CREATE TRIGGER update_games_updated_at BEFORE UPDATE ON games FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_galaxies_updated_at BEFORE UPDATE ON galaxies FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_sectors_updated_at BEFORE UPDATE ON sectors FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_systems_updated_at BEFORE UPDATE ON systems FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_planets_updated_at BEFORE UPDATE ON planets FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_player_stats_updated_at BEFORE UPDATE ON player_stats FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_update_galaxy_count AFTER INSERT OR DELETE ON galaxies FOR EACH ROW EXECUTE FUNCTION update_counts();
CREATE TRIGGER trigger_update_sector_count AFTER INSERT OR DELETE ON sectors FOR EACH ROW EXECUTE FUNCTION update_counts();
CREATE TRIGGER trigger_update_system_count AFTER INSERT OR DELETE ON systems FOR EACH ROW EXECUTE FUNCTION update_counts();
CREATE TRIGGER trigger_update_planet_count AFTER INSERT OR DELETE ON planets FOR EACH ROW EXECUTE FUNCTION update_counts();
CREATE TRIGGER trigger_update_player_stats AFTER INSERT OR UPDATE OR DELETE ON planets FOR EACH ROW EXECUTE FUNCTION update_player_stats();
