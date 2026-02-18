package spatial

import (
	"context"
	"database/sql"
	"planets-server/internal/shared/database"
	"planets-server/internal/shared/errors"

	"github.com/lib/pq"
)

type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) getExecutor(tx *database.Tx) database.Executor {
	if tx != nil {
		return tx
	}
	return r.db
}

// BatchInsertRequest represents a single entity to be inserted in a batch
type BatchInsertRequest struct {
	GameID      int
	ParentID    *int
	EntityType  EntityType
	Level       int
	XCoord      int
	YCoord      int
	Name string
}

// CreateEntitiesBatch creates multiple spatial entities in a single database operation using JSON
// Returns only the IDs of created entities to minimize memory usage
func (r *Repository) CreateEntitiesBatch(ctx context.Context, entities []BatchInsertRequest, tx *database.Tx) ([]int, error) {
	if len(entities) == 0 {
		return []int{}, nil
	}

	exec := r.getExecutor(tx)

	// Build arrays for each column
	gameIDs := make([]int, len(entities))
	parentIDs := make([]*int, len(entities))
	entityTypes := make([]string, len(entities))
	levels := make([]int, len(entities))
	xCoords := make([]int, len(entities))
	yCoords := make([]int, len(entities))
	names := make([]string, len(entities))

	for i, entity := range entities {
		gameIDs[i] = entity.GameID
		parentIDs[i] = entity.ParentID
		entityTypes[i] = string(entity.EntityType)
		levels[i] = entity.Level
		xCoords[i] = entity.XCoord
		yCoords[i] = entity.YCoord
		names[i] = entity.Name
	}

	query := `
		INSERT INTO spatial_entities (game_id, parent_id, entity_type, level, x_coord, y_coord, name, child_count)
		SELECT
			unnest($1::int[]),
			unnest($2::int[]),
			unnest($3::entity_type[]),
			unnest($4::int[]),
			unnest($5::int[]),
			unnest($6::int[]),
			unnest($7::text[]),
			0
		RETURNING id`

	rows, err := exec.QueryContext(ctx, query,
		pq.Array(gameIDs),
		pq.Array(parentIDs),
		pq.Array(entityTypes),
		pq.Array(levels),
		pq.Array(xCoords),
		pq.Array(yCoords),
		pq.Array(names),
	)
	if err != nil {
		return nil, errors.WrapInternal("failed to batch create spatial entities", err)
	}
	defer func() { _ = rows.Close() }()

	var entityIDs []int
	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			return nil, errors.WrapInternal("failed to scan entity ID", err)
		}
		entityIDs = append(entityIDs, id)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.WrapInternal("error iterating entity IDs", err)
	}

	return entityIDs, nil
}

func (r *Repository) scanEntity(scanner interface{ Scan(...any) error }) (SpatialEntity, error) {
	var e SpatialEntity
	err := scanner.Scan(
		&e.ID, &e.GameID, &e.ParentID, &e.EntityType, &e.Level,
		&e.XCoord, &e.YCoord, &e.Name, &e.ChildCount, &e.CreatedAt, &e.UpdatedAt,
	)
	return e, err
}

const entityColumns = `id, game_id, parent_id, entity_type, level, x_coord, y_coord, name, child_count, created_at, updated_at`

func (r *Repository) GetByID(ctx context.Context, entityID int) (*SpatialEntity, error) {
	query := `SELECT ` + entityColumns + ` FROM spatial_entities WHERE id = $1`

	entity, err := r.scanEntity(r.db.QueryRowContext(ctx, query, entityID))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NotFoundf("spatial entity not found with id: %d", entityID)
		}
		return nil, errors.WrapInternal("failed to get spatial entity by id", err)
	}

	return &entity, nil
}

func (r *Repository) GetChildren(ctx context.Context, parentID int) ([]SpatialEntity, error) {
	query := `SELECT ` + entityColumns + ` FROM spatial_entities WHERE parent_id = $1 ORDER BY x_coord, y_coord`

	rows, err := r.db.QueryContext(ctx, query, parentID)
	if err != nil {
		return nil, errors.WrapInternal("failed to query children", err)
	}
	defer func() { _ = rows.Close() }()

	var entities []SpatialEntity
	for rows.Next() {
		entity, err := r.scanEntity(rows)
		if err != nil {
			return nil, errors.WrapInternal("failed to scan spatial entity", err)
		}
		entities = append(entities, entity)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.WrapInternal("error iterating spatial entities", err)
	}

	return entities, nil
}

func (r *Repository) GetAncestors(ctx context.Context, entityID int) ([]SpatialEntity, error) {
	query := `
		WITH RECURSIVE ancestors AS (
			SELECT ` + entityColumns + `
			FROM spatial_entities WHERE id = $1
			UNION ALL
			SELECT se.id, se.game_id, se.parent_id, se.entity_type, se.level,
				se.x_coord, se.y_coord, se.name, se.child_count, se.created_at, se.updated_at
			FROM spatial_entities se
			INNER JOIN ancestors a ON se.id = a.parent_id
		)
		SELECT * FROM ancestors ORDER BY level ASC`

	rows, err := r.db.QueryContext(ctx, query, entityID)
	if err != nil {
		return nil, errors.WrapInternal("failed to query ancestors", err)
	}
	defer func() { _ = rows.Close() }()

	var entities []SpatialEntity
	for rows.Next() {
		entity, err := r.scanEntity(rows)
		if err != nil {
			return nil, errors.WrapInternal("failed to scan ancestor entity", err)
		}
		entities = append(entities, entity)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.WrapInternal("error iterating ancestor entities", err)
	}

	return entities, nil
}
