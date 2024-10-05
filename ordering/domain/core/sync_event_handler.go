package core

type SyncEventHandler interface {
	HandleEvent(aggregaate Aggregate) error
	GetAggregateType() string
}
