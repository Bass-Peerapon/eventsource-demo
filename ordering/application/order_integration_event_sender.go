package application

import (
	"encoding/json"
	"reflect"

	"github.com/Bass-Peerapon/eventsource-demo/ordering/domain/core"
	"github.com/Bass-Peerapon/eventsource-demo/ordering/domain/order"
	"github.com/Bass-Peerapon/eventsource-demo/ordering/infrastructure/messaging"
)

type OrderIntegrationEventSender struct {
	eventRepo     core.EventRepository
	aggregateRepo core.AggregateRepository
	messageBroker messaging.MessageBroker
}

// GetAggregateType implements core.AsyncEventHandler.
func (o OrderIntegrationEventSender) GetAggregateType() string {
	orderAggregate := order.OrderAggregate{}
	return orderAggregate.GetAggregateType()
}

// GetSubscriptionName implements core.AsyncEventHandler.
func (o OrderIntegrationEventSender) GetSubscriptionName() string {
	return reflect.TypeOf(o).Name()
}

// HandleEvent implements core.AsyncEventHandler.
func (o OrderIntegrationEventSender) HandleEvent(event core.Event) error {
	orderAggregate := order.OrderAggregate{}
	snapshot, err := o.aggregateRepo.LoadSnapshot(event.AggregateID, &event.Version)
	if err != nil {
		return err
	}
	if snapshot != nil {
		snapshot.UnSerialize(&orderAggregate)
	}
	loadedEvents, err := o.eventRepo.LoadEvents(event.AggregateID, &orderAggregate.Version, &event.Version)
	if err != nil {
		return err
	}

	for _, event := range loadedEvents {
		orderAggregate.Apply(event)
	}

	bu, _ := json.Marshal(orderAggregate)

	if err := o.messageBroker.Publish(messaging.TOPIC_ORDER_EVENT, event.EventType, bu); err != nil {
		return err
	}
	return nil
}

func NewOrderIntegrationEventSender(eventStore core.EventRepository, snapshotStore core.AggregateRepository, messageBroker messaging.MessageBroker) core.AsyncEventHandler {
	return OrderIntegrationEventSender{
		eventRepo:     eventStore,
		aggregateRepo: snapshotStore,
		messageBroker: messageBroker,
	}
}
