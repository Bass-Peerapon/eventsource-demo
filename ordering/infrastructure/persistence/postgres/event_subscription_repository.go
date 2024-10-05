package postgres

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/Bass-Peerapon/eventsource-demo/ordering/domain/core"
	"github.com/Bass-Peerapon/eventsource-demo/ordering/domain/order"
	"github.com/jmoiron/sqlx"
)

type eventSubscriptionRepository struct {
	db *sqlx.DB
}

// NewEventSubscriptionRepository ฟังก์ชันสำหรับสร้าง EventSubscriptionRepository ใหม่
func NewEventSubscriptionRepository(db *sqlx.DB) core.EventSubscriptionRepository {
	return &eventSubscriptionRepository{
		db: db,
	}
}

// CreateSubscriptionIfAbsent สร้าง subscription หากยังไม่มี
func (r *eventSubscriptionRepository) CreateSubscription(subscriptionName string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}

	defer tx.Rollback()
	query := `
INSERT INTO es_event_subscription (subscription_name, last_transaction_id, last_event_id)
    VALUES ($1, '0', 0)
ON CONFLICT
    DO NOTHING
	`
	_, err = tx.Exec(query, subscriptionName)
	if err != nil {
		return fmt.Errorf("failed to create subscription if absent: %w", err)
	}
	return tx.Commit()
}

// ReadCheckpointAndLockSubscription อ่าน checkpoint และทำการล็อก subscription
func (r *eventSubscriptionRepository) ReadCheckpointAndLockSubscription(subscriptionName string) (*sqlx.Tx, *core.EventSubscriptionCheckpoint, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	query := `
SELECT
    last_transaction_id,
    last_event_id
FROM
    es_event_subscription
WHERE
    subscription_name = $1
FOR UPDATE
    SKIP LOCKED
	`
	rows, err := tx.Query(query, subscriptionName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read checkpoint: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		var lastTransactionIDStr string
		var lastEventID int64
		if err := rows.Scan(&lastTransactionIDStr, &lastEventID); err != nil {
			return nil, nil, fmt.Errorf("failed to scan checkpoint: %w", err)
		}

		lastTransactionID, err := strconv.ParseInt(lastTransactionIDStr, 10, 64)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse last transaction ID: %w", err)
		}

		checkpoint := core.EventSubscriptionCheckpoint{
			LasttransactionID: lastTransactionID,
			LastEventID:       lastEventID,
		}
		return tx, &checkpoint, nil
	}

	return tx, nil, nil
}

// ReadEventsAfterCheckpoint implements core.EventStore.
func (r *eventSubscriptionRepository) ReadEventsAfterCheckpoint(tx *sqlx.Tx, aggregateType string, lastTransactionID int64, lastEventID int64) ([]core.Event, error) {
	query := `
SELECT
		es_event.id,
		es_event.transaction_id,
		es_event.aggregate_id,
    es_event.event_type,
		es_event.event_data,
		es_event.version,
    es_event.created_at
FROM
    es_event
JOIN 
		es_aggregate ON es_aggregate.id = es_event.aggregate_id
WHERE
		aggregate_type = $1
AND
		(es_event.transaction_id , es_event.id) > ($2::xid8, $3)
AND
		es_event.transaction_id < pg_snapshot_xmin(pg_current_snapshot())
ORDER BY
    es_event.transaction_id, es_event.id
	`
	var events []event
	if err := tx.Select(&events, query, aggregateType, lastTransactionID, lastEventID); err != nil {
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

// UpdateEventSubscription อัปเดต event subscription ด้วยข้อมูลล่าสุดที่ประมวลผล
func (r *eventSubscriptionRepository) UpdateEventSubscription(tx *sqlx.Tx, subscriptionName string, lastTransactionID int64, lastEventID int64) (bool, error) {
	query := `
UPDATE
    es_event_subscription
SET
    last_transaction_id = $1,
    last_event_id = $2
WHERE
    subscription_name = $3
	`
	result, err := tx.Exec(query, lastTransactionID, lastEventID, subscriptionName)
	if err != nil {
		return false, fmt.Errorf("failed to update event subscription: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected > 0, nil
}
