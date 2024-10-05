package postgres

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Bass-Peerapon/eventsource-demo/ordering/domain/core"
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
)

type aggregateRepository struct {
	db *sqlx.DB
}

// SaveAggregate implements core.AggregateRepository.
func (s *aggregateRepository) SaveAggregate(aggregate core.Aggregate) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	query := `
        INSERT INTO es_aggregate (id, version, aggregate_type)
        VALUES ($1, $2, $3)
				ON CONFLICT (id) WHERE es_aggregate.version = $2
				DO UPDATE SET
				version = $2
    `
	result, err := tx.Exec(query, aggregate.GetID(), aggregate.GetVersion(), aggregate.GetAggregateType())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrAggregateOutdated
	}

	return tx.Commit()
}

// LoadSnapshot implements core.SnapshotStore.
func (s *aggregateRepository) LoadSnapshot(aggregateID uuid.UUID, version *int) (*core.AggregateSnapshot, error) {
	conds := []string{}
	args := []interface{}{}

	conds = append(conds, "es_aggregate_snapshot.aggregate_id = ?")
	args = append(args, aggregateID)

	if version != nil {
		conds = append(conds, "es_aggregate_snapshot.version <= ?")
		args = append(args, *version)
	}

	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	var snapshot core.AggregateSnapshot
	query := fmt.Sprintf(`
	SELECT
		es_aggregate_snapshot.aggregate_id,
		es_aggregate_snapshot.version,
		es_aggregate_snapshot.event_data
	FROM
		es_aggregate_snapshot
	JOIN
		es_aggregate 
	ON 
		es_aggregate.id = es_aggregate_snapshot.aggregate_id
	%s
	ORDER BY
		es_aggregate_snapshot.version DESC
	LIMIT 1
	`,
		where,
	)
	query = sqlx.Rebind(sqlx.DOLLAR, query)
	if err := s.db.Get(&snapshot, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &snapshot, nil
}

// SaveSnapshot implements core.SnapshotStore.
func (s *aggregateRepository) SaveSnapshot(snapshot *core.AggregateSnapshot) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}

	defer tx.Rollback()
	query := `
		INSERT INTO es_aggregate_snapshot (aggregate_id, version, event_data)
		VALUES ($1, $2, $3)
		`
	eventData, _ := json.Marshal(snapshot.EventData)

	if _, err := tx.Exec(query, snapshot.AggregateID, snapshot.Version, eventData); err != nil {
		return err
	}
	return tx.Commit()
}

func NewAggregateRepository(db *sqlx.DB) core.AggregateRepository {
	return &aggregateRepository{
		db: db,
	}
}
