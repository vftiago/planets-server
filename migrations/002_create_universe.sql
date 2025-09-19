CREATE TABLE IF NOT EXISTS galaxies (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    sector_count INTEGER NOT NULL DEFAULT 16,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
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
    system_id INTEGER REFERENCES systems(id) ON DELETE CASCADE,
    planet_index INTEGER NOT NULL,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'terrestrial',
    size INTEGER NOT NULL DEFAULT 100,
    population BIGINT NOT NULL DEFAULT 0,
    max_population BIGINT NOT NULL DEFAULT 1000000,
    owner_id INTEGER REFERENCES players(id) ON DELETE SET NULL,
    is_homeworld BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(system_id, planet_index)
);

CREATE INDEX IF NOT EXISTS idx_sectors_galaxy_id ON sectors(galaxy_id);
CREATE INDEX IF NOT EXISTS idx_sectors_coordinates ON sectors(galaxy_id, sector_x, sector_y);
CREATE INDEX IF NOT EXISTS idx_systems_sector_id ON systems(sector_id);
CREATE INDEX IF NOT EXISTS idx_systems_coordinates ON systems(sector_id, system_x, system_y);
CREATE INDEX IF NOT EXISTS idx_planets_system_id ON planets(system_id);
CREATE INDEX IF NOT EXISTS idx_planets_owner_id ON planets(owner_id);
CREATE INDEX IF NOT EXISTS idx_planets_homeworld ON planets(is_homeworld) WHERE is_homeworld = true;

CREATE OR REPLACE FUNCTION update_system_planet_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE systems SET planet_count = planet_count + 1 WHERE id = NEW.system_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE systems SET planet_count = planet_count - 1 WHERE id = OLD.system_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_sector_system_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE sectors SET system_count = system_count + 1 WHERE id = NEW.sector_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE sectors SET system_count = system_count - 1 WHERE id = OLD.sector_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_system_planet_count
    AFTER INSERT OR DELETE ON planets
    FOR EACH ROW EXECUTE FUNCTION update_system_planet_count();

CREATE TRIGGER trigger_update_sector_system_count
    AFTER INSERT OR DELETE ON systems
    FOR EACH ROW EXECUTE FUNCTION trigger_update_sector_system_count();

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_galaxies_updated_at BEFORE UPDATE ON galaxies
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sectors_updated_at BEFORE UPDATE ON sectors
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_systems_updated_at BEFORE UPDATE ON systems
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_planets_updated_at BEFORE UPDATE ON planets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
