package core

import (
	"encoding/json"

	"github.com/gofrs/uuid"
)

type Aggregate interface {
	GetID() uuid.UUID
	GetVersion() int
	GetAggregateType() string
	Apply(event Event)
}

type AggregateSnapshot struct {
	AggregateID uuid.UUID   `json:"aggregate_id" db:"aggregate_id"`
	Version     int         `json:"version" db:"version"`
	EventData   interface{} `json:"event_data" db:"event_data"`
}

func (s AggregateSnapshot) UnSerialize(dest Aggregate) error {
	return json.Unmarshal(s.EventData.([]byte), dest)
}
