package postgres

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/Bass-Peerapon/eventsource-demo/ordering/domain/core"
	"github.com/Bass-Peerapon/eventsource-demo/ordering/domain/order"
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
)

type eventRepository struct {
	db *sqlx.DB
}

// LoadEvents implements core.EventStore.
func (e *eventRepository) LoadEvents(aggregateID uuid.UUID, fromVersion *int, toVersion *int) ([]core.Event, error) {
	conds := []string{}
	args := []interface{}{}

	conds = append(conds, "aggregate_id = ?")
	args = append(args, aggregateID)

	if fromVersion != nil {
		conds = append(conds, "version >= ?")
		args = append(args, *fromVersion)
	}
	if toVersion != nil {
		conds = append(conds, "version <= ?")
		args = append(args, *toVersion)
	}

	where := ""
	if len(conds) > 0 {
		where = fmt.Sprintf("WHERE %s", strings.Join(conds, " AND "))
	}

	query := fmt.Sprintf(`
SELECT
    id,
    transaction_id,
    aggregate_id,
    event_type,
    event_data,
    version,
    created_at
FROM
    es_event
%s
ORDER BY
    version
	`,
		where,
	)

	query = sqlx.Rebind(sqlx.DOLLAR, query)
	var events []event
	if err := e.db.Select(&events, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	loadedEvents := make([]core.Event, 0, len(events))
	for _, event := range events {
		switch event.EventType {
		case reflect.TypeOf(order.OrderCreatedEvent{}).Name():
			eventData := order.OrderCreatedEvent{}
			json.Unmarshal(event.EventData, &eventData)
			loadedEvents = append(loadedEvents, core.Event{
				ID:            event.ID,
				TransactionID: event.TransactionID,
				AggregateID:   event.AggregateID,
				EventType:     event.EventType,
				EventData:     eventData,
				Version:       event.Version,
				CreatedAt:     event.CreatedAt,
			})

		case reflect.TypeOf(order.OrderUpdatedEvent{}).Name():
			eventData := order.OrderUpdatedEvent{}
			json.Unmarshal(event.EventData, &eventData)
			loadedEvents = append(loadedEvents, core.Event{
				ID:            event.ID,
				TransactionID: event.TransactionID,
				AggregateID:   event.AggregateID,
				EventType:     event.EventType,
				EventData:     eventData,
				Version:       event.Version,
				CreatedAt:     event.CreatedAt,
			})
		case reflect.TypeOf(order.OrderItemAmountUpdatedEvent{}).Name():
			eventData := order.OrderItemAmountUpdatedEvent{}
			json.Unmarshal(event.EventData, &eventData)
			loadedEvents = append(loadedEvents, core.Event{
				ID:            event.ID,
				TransactionID: event.TransactionID,
				AggregateID:   event.AggregateID,
				EventType:     event.EventType,
				EventData:     eventData,
				Version:       event.Version,
				CreatedAt:     event.CreatedAt,
			})
		}
	}

	return loadedEvents, nil
}

// SaveEvent implements core.EventStore.
func (e *eventRepository) SaveEvents(events []core.Event) error {
	tx, err := e.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, event := range events {
		query := `
			INSERT INTO es_event (transaction_id, aggregate_id, version, event_type, event_data, created_at)
			VALUES (pg_current_xact_id() ,$1, $2, $3, $4, $5)
		`
		eventData, _ := json.Marshal(event.EventData)
		if _, err := tx.Exec(query, event.AggregateID, event.Version, event.EventType, eventData, event.CreatedAt); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func NewEventRepository(db *sqlx.DB) core.EventRepository {
	return &eventRepository{
		db: db,
	}
}

type event struct {
	ID            int64           `json:"id" db:"id"`
	TransactionID int64           `json:"transaction_id" db:"transaction_id"`
	AggregateID   uuid.UUID       `json:"aggregate_id" db:"aggregate_id"`
	EventType     string          `json:"event_type" db:"event_type"`
	EventData     json.RawMessage `json:"event_data" db:"event_data"`
	Version       int             `json:"version" db:"version"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
}
