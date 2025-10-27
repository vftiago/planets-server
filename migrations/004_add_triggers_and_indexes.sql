-- Indexes
CREATE INDEX IF NOT EXISTS idx_games_status ON games(status);
CREATE INDEX IF NOT EXISTS idx_games_next_turn ON games(next_turn_at) WHERE next_turn_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_spatial_entities_game_id ON spatial_entities(game_id);
CREATE INDEX IF NOT EXISTS idx_spatial_entities_parent_id ON spatial_entities(parent_id);
CREATE INDEX IF NOT EXISTS idx_spatial_entities_type_level ON spatial_entities(entity_type, level);
CREATE INDEX IF NOT EXISTS idx_spatial_entities_coords ON spatial_entities(parent_id, x_coord, y_coord);
CREATE INDEX IF NOT EXISTS idx_spatial_entities_game_level ON spatial_entities(game_id, level);
CREATE INDEX IF NOT EXISTS idx_planets_system_id ON planets(system_id);
CREATE INDEX IF NOT EXISTS idx_planets_owner_id ON planets(owner_id);
CREATE INDEX IF NOT EXISTS idx_game_players_game_id ON game_players(game_id);
CREATE INDEX IF NOT EXISTS idx_game_players_player_id ON game_players(player_id);
CREATE INDEX IF NOT EXISTS idx_player_stats_game_player ON player_stats(game_id, player_id);

-- Triggers
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
