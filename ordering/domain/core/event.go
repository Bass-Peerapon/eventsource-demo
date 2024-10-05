package core

import (
	"time"

	"github.com/gofrs/uuid"
)

type Event struct {
	ID            int64       `json:"id"`
	TransactionID int64       `json:"transaction_id"`
	AggregateID   uuid.UUID   `json:"aggregate_id"`
	EventType     string      `json:"event_type"`
	EventData     interface{} `json:"event_data"`
	Version       int         `json:"version"`
	CreatedAt     time.Time   `json:"created_at"`
}

func NewEvent(aggregateID uuid.UUID, eventType string, eventData interface{}) Event {
	return Event{
		AggregateID: aggregateID,
		EventType:   eventType,
		EventData:   eventData,
		CreatedAt:   time.Now(),
	}
}
