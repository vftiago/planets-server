# Spatial Entities vs. Coordinate-Based Paths: Technical Analysis

## Current Implementation

The codebase uses a `spatial_entities` table with an `entity_type` discriminator to represent galaxies, sectors, and systems. Each entity has:

- Parent-child relationships via `parent_id`
- Grid coordinates (`x_coord`, `y_coord`) within its parent
- Metadata (name, description)
- Child count for navigation

## Proposed Alternative

Replace spatial entities with coordinate-based paths on planets (e.g., `1.0.0.4` = galaxy 1, sector 0, system 0, planet 4). The hierarchy becomes implicit rather than explicit.

---

## Analysis: Frontend Navigation

### With Spatial Entities

```sql
-- List all galaxies in a game
SELECT * FROM spatial_entities WHERE game_id = ? AND entity_type = 'galaxy';

-- List sectors in a galaxy
SELECT * FROM spatial_entities WHERE parent_id = ? AND entity_type = 'sector';

-- Get breadcrumb path for a planet
WITH RECURSIVE path AS (
  SELECT * FROM spatial_entities WHERE id = ?
  UNION ALL
  SELECT se.* FROM spatial_entities se JOIN path p ON se.id = p.parent_id
)
SELECT * FROM path ORDER BY level;
```

**Pros:**

- Each level has its own identity (name: "Andromeda", "Sector 7G", "Sol System")
- Natural REST endpoints: `/galaxies`, `/galaxies/:id/sectors`, etc.
- Breadcrumb data comes directly from entity names
- Can add galaxy/sector/system-specific metadata (description, discovery date, etc.)

**Cons:**

- More rows in the database
- JOINs required to traverse hierarchy

### With Coordinate Paths

```sql
-- "List all galaxies" = list unique first coordinate segments
SELECT DISTINCT split_part(coords, '.', 1) as galaxy_id FROM planets WHERE game_id = ?;

-- "List sectors in galaxy 1" = list unique second segments where first = 1
SELECT DISTINCT split_part(coords, '.', 2) as sector_id
FROM planets WHERE game_id = ? AND coords LIKE '1.%';
```

**Pros:**

- Simpler schema (no spatial_entities table)
- Planet location is self-contained

**Cons:**

- **No metadata for intermediate levels** - "Galaxy 1" has no name, description, or properties
- Awkward aggregation queries with string parsing
- REST endpoints become unnatural: What does `GET /galaxies/1` return if galaxies don't exist as entities?
- Breadcrumbs would show coordinates, not names (unless you store names elsewhere, which defeats the purpose)

### Verdict: Navigation

**Spatial Entities wins.** Without them, you lose the ability to name and describe galaxies/sectors/systems. Your frontend would show "Galaxy 1 > Sector 0 > System 3" instead of "Andromeda > Core Worlds > Sol System". This significantly impacts game immersion.

You _could_ work around this with a separate lightweight `spatial_names` lookup table, but at that point you're just recreating spatial_entities with extra steps.

---

## Analysis: Property Inheritance

### With Spatial Entities

```sql
-- Add inheritable properties to spatial_entities
ALTER TABLE spatial_entities ADD COLUMN age_billion_years DECIMAL;
ALTER TABLE spatial_entities ADD COLUMN radiation_level VARCHAR(20);
ALTER TABLE spatial_entities ADD COLUMN mineral_modifier DECIMAL DEFAULT 1.0;

-- Query a planet with inherited properties
SELECT
  p.*,
  system.radiation_level,
  sector.mineral_modifier,
  galaxy.age_billion_years
FROM planets p
JOIN spatial_entities system ON p.system_id = system.id
JOIN spatial_entities sector ON system.parent_id = sector.id
JOIN spatial_entities galaxy ON sector.parent_id = galaxy.id
WHERE p.id = ?;
```

**Pros:**

- Clear, normalized storage for each level's properties
- Properties can be updated at source (change galaxy age, all descendants reflect it)
- Different levels can have different properties
- Easy to add new inheritable properties

**Cons:**

- JOINs required to resolve full inheritance chain
- Need to decide which properties live at which level

### With Coordinate Paths

**Option A: Denormalize onto planets**

```sql
ALTER TABLE planets ADD COLUMN galaxy_age_billion_years DECIMAL;
ALTER TABLE planets ADD COLUMN sector_mineral_modifier DECIMAL;
ALTER TABLE planets ADD COLUMN system_radiation_level VARCHAR(20);
```

**Pros:**

- Zero JOINs - all data on the planet row
- Fast reads

**Cons:**

- **Massive data duplication** - if a galaxy has 10,000 planets, galaxy_age is stored 10,000 times
- **Expensive updates** - changing a galaxy's age requires updating all its planets
- **No single source of truth** - if updates fail partially, data becomes inconsistent
- Schema changes for new properties affect the planets table

**Option B: Separate lookup tables**

```sql
CREATE TABLE galaxy_properties (galaxy_idx INT, game_id INT, age_billion_years DECIMAL, ...);
CREATE TABLE sector_properties (galaxy_idx INT, sector_idx INT, game_id INT, mineral_modifier DECIMAL, ...);
CREATE TABLE system_properties (galaxy_idx INT, sector_idx INT, system_idx INT, game_id INT, radiation_level VARCHAR, ...);
```

**Pros:**

- Normalized storage
- Single source of truth

**Cons:**

- **This is just spatial_entities with extra tables** - you've recreated the hierarchy
- Coordinate parsing still needed to JOIN
- Three tables instead of one
- No names/descriptions unless you add them (then it's definitely just spatial_entities)

**Option C: JSON blob inheritance context**

```sql
ALTER TABLE planets ADD COLUMN inherited_context JSONB;
-- {"galaxy": {"age": 4.5}, "sector": {"mineral_mod": 1.2}, "system": {"radiation": "high"}}
```

**Pros:**

- Flexible schema
- All data on planet row

**Cons:**

- Same duplication/update problems as Option A
- Harder to query/index
- No type safety

### Verdict: Property Inheritance

**Spatial Entities wins decisively.** Every alternative either:

1. Duplicates data massively (denormalization)
2. Recreates spatial_entities in a more awkward form (lookup tables)
3. Loses the ability to have a single source of truth for level properties

The inheritance use case fundamentally requires entities at each level to exist and hold properties.

---

## Additional Considerations

### Ships and Structures

You mentioned ships might inherit properties from their location. Consider:

**With Spatial Entities:**

```sql
-- Ship can be at any level
ALTER TABLE ships ADD COLUMN location_entity_id INT REFERENCES spatial_entities(id);
ALTER TABLE ships ADD COLUMN location_planet_id INT REFERENCES planets(id);
-- One of these is set depending on where the ship is
```

A ship in transit between systems can reference the sector. A ship orbiting a planet references the planet. Clean and flexible.

**With Coordinate Paths:**

```sql
-- Ship needs its own coordinate
ALTER TABLE ships ADD COLUMN coords VARCHAR(50); -- "1.0.0" for system, "1.0.0.4" for planet orbit
```

How do you distinguish "ship at system 1.0.0" from "ship at planet 1.0.0.0"? You'd need conventions or additional flags. Querying "all ships in sector 1.2" requires LIKE queries on strings.

### Future Entity Types

What if you want to add:

- **Nebulae** that span multiple systems and affect sensor range
- **Asteroid belts** between planets in a system
- **Space stations** at specific coordinates
- **Wormholes** connecting distant locations

**With Spatial Entities:** Add new `entity_type` values. The infrastructure exists.

**With Coordinate Paths:** Each new type needs its own coordinate convention or table. Asteroid belts don't fit the hierarchical coordinate model.

### Query Performance

**Spatial Entities:** JOINs are required but predictable. Indexes on `parent_id` and `entity_type` make traversal fast. For hot paths, you could add a materialized `path` column.

**Coordinate Paths:** String parsing (`split_part`) is slower than integer comparisons. Aggregation queries scan more rows. However, single-planet lookups are faster (no JOINs).

**Optimization for either approach:** If inheritance lookups become a bottleneck, denormalize the most-accessed inherited properties onto planets as a cache, with the spatial_entities remaining the source of truth.

---

## Hybrid Approach (If You Want Simpler Queries)

Keep spatial_entities but add a denormalized path to planets:

```sql
ALTER TABLE planets ADD COLUMN coord_path VARCHAR(50); -- "1.0.0.4"
-- OR
ALTER TABLE planets ADD COLUMN galaxy_id INT;
ALTER TABLE planets ADD COLUMN sector_id INT;
-- (in addition to existing system_id)
```

This gives you:

- Fast planet lookups without JOINs for basic queries
- Spatial entities still exist for names, descriptions, and inherited properties
- Best of both worlds at the cost of some denormalization

---

## Recommendation

**Keep Spatial Entities.** Both of your concerns (frontend navigation and property inheritance) are better served by explicit entities:

| Requirement                    | Spatial Entities     | Coordinate Paths                      |
| ------------------------------ | -------------------- | ------------------------------------- |
| Named galaxies/sectors/systems | ✅ Native            | ❌ Requires workaround                |
| Breadcrumb navigation          | ✅ Entity names      | ⚠️ Coordinate numbers only            |
| Property inheritance           | ✅ Clean, normalized | ❌ Duplication or recreating entities |
| Ships at any level             | ✅ FK to entity      | ⚠️ Coordinate conventions             |
| Future entity types            | ✅ Add entity_type   | ❌ New conventions per type           |
| Query simplicity               | ⚠️ JOINs needed      | ✅ For single planets                 |

The coordinate-path approach optimizes for one thing (simple single-planet queries) while making everything else harder. The spatial entities approach is more flexible and better supports the game features you described.

### If Simplification Is Still Desired

Consider these lighter alternatives:

1. **Keep spatial_entities but lazy-load properties** - Don't JOIN for inheritance on every request. Cache inherited properties or compute them only when needed.

2. **Add a denormalized `full_path` to planets** - For display/filtering purposes, store "Andromeda > Core Worlds > Sol" on the planet. Spatial entities remain the source of truth.

3. **Flatten the hierarchy** - If sectors don't add meaningful gameplay value, remove that level. Galaxy → System → Planet is simpler than Galaxy → Sector → System → Planet.

The current implementation is sound. I'd recommend keeping it unless the JOINs become a measured performance problem (which is unlikely at game scale).

❯ What if I let go of inherited properties as well?

● Then spatial_entities become purely navigational scaffolding. At that point, the question is: do you need them at all?

Without names or inherited properties, spatial_entities only provide:

1. A place to query "list all galaxies" / "list sectors in galaxy 3"
2. Grid coordinates (x, y) for visual layout
3. Foreign key integrity

You could achieve #1 and #2 with just planets and coord_path:

-- The only spatial table you need
planets (
id,
game_id,
coord_path VARCHAR(20), -- "3.5.1.0" (galaxy.sector.system.planet)
grid_x INT, -- for 2D map layout within system
grid_y INT,
type, size, population, ...
)

-- List galaxies
SELECT DISTINCT split_part(coord_path, '.', 1) as galaxy FROM planets WHERE game_id = ?;

-- List sectors in galaxy 3
SELECT DISTINCT split_part(coord_path, '.', 2) as sector FROM planets WHERE coord_path LIKE '3.%';

-- List systems in sector 3.5
SELECT DISTINCT split_part(coord_path, '.', 3) as system FROM planets WHERE coord_path LIKE '3.5.%';

-- Ship route display
SELECT coord_path FROM planets WHERE id IN (origin_id, destination_id);
-- Returns: "3.5.1.0", "3.5.2.8"

What you lose:

- No place to store galaxy/sector/system metadata (but you're saying you don't need it)
- No FK integrity on coordinates (planet "99.99.99.0" could exist with no validation)
- Aggregation requires string parsing (minor performance cost)

What you gain:

- Single table for all spatial data
- Zero JOINs, ever
- Coordinates are the native identifier, not derived
- Simpler mental model
- Delete spatial_entities, spatial service, spatial repository entirely

The critical question: Are you certain you won't want level-specific properties later? Examples:

- Sector control (which faction owns it)
- System hazards (radiation, debris fields)
- Discovery state (has player visited this galaxy/sector?)
- Generation parameters (this galaxy is "old", this sector is "mineral-rich")

If any of those sound plausible, keep a lightweight spatial_entities. If not, you can genuinely just use planets with coord_path and delete a chunk of code.
