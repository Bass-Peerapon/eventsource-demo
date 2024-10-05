package core

import (
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
)

type EventRepository interface {
	SaveEvents(events []Event) error
	LoadEvents(aggregateID uuid.UUID, fromVersion *int, toVersion *int) ([]Event, error)
}

type AggregateRepository interface {
	SaveAggregate(aggregate Aggregate) error
	SaveSnapshot(snapshot *AggregateSnapshot) error
	LoadSnapshot(aggregateID uuid.UUID, version *int) (*AggregateSnapshot, error)
}

type EventSubscriptionRepository interface {
	CreateSubscription(subscriptionName string) error
	ReadCheckpointAndLockSubscription(subscriptionName string) (*sqlx.Tx, *EventSubscriptionCheckpoint, error)
	ReadEventsAfterCheckpoint(tx *sqlx.Tx, aggregateType string, lastTransactionID int64, lastEventID int64) ([]Event, error)
	UpdateEventSubscription(tx *sqlx.Tx, subscriptionName string, lastTransactionID int64, lastEventID int64) (bool, error)
}
