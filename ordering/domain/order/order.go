package order

import (
	"reflect"

	"github.com/Bass-Peerapon/eventsource-demo/ordering/domain/core"
	"github.com/gofrs/uuid"
)

type OrderAggregate struct {
	ID          uuid.UUID    `json:"id"`
	Name        string       `json:"name"`
	OrderItems  []OrderItem  `json:"order_items"`
	IsSubmitted bool         `json:"is_submitted"`
	Version     int          `json:"version"`
	Events      []core.Event `json:"-"`
}

type OrderItem struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Amount int       `json:"amount"`
}

func (o *OrderAggregate) GetID() uuid.UUID {
	return o.ID
}

func (o *OrderAggregate) GetVersion() int {
	return o.Version
}

func (o *OrderAggregate) GetAggregateType() string {
	return reflect.TypeOf(o).Elem().Name()
}

func (o *OrderAggregate) Apply(event core.Event) {
	switch event.EventType {
	case reflect.TypeOf(OrderCreatedEvent{}).Name():
		createdEvent := event.EventData.(OrderCreatedEvent)
		o.ID = event.AggregateID
		o.Name = createdEvent.Name
		o.OrderItems = createdEvent.OrderItems
	case reflect.TypeOf(OrderUpdatedEvent{}).Name():
		updatedEvent := event.EventData.(OrderUpdatedEvent)
		o.Name = updatedEvent.Name
		o.OrderItems = updatedEvent.OrderItems
	case reflect.TypeOf(OrderItemAmountUpdatedEvent{}).Name():
		updatedEvent := event.EventData.(OrderItemAmountUpdatedEvent)
		for i, orderItem := range o.OrderItems {
			if orderItem.ID == updatedEvent.ID {
				o.OrderItems[i].Amount = updatedEvent.Amount
				break
			}
		}
	}
	o.Version++
}

func (o *OrderAggregate) appendEvent(event ...core.Event) {
	for _, e := range event {
		e.Version = o.Version + len(o.Events) + 1
		o.Events = append(o.Events, e)
	}
}

func CreateOrderWithItems(name string, orderItems []OrderItem) *OrderAggregate {
	order := OrderAggregate{}

	id, _ := uuid.NewV4()

	eventData := OrderCreatedEvent{
		Name:       name,
		OrderItems: orderItems,
	}

	createdOrderEvent := core.NewEvent(id, eventData.GetEventType(), eventData)
	order.appendEvent(createdOrderEvent)
	order.Apply(createdOrderEvent)

	return &order
}

func (o *OrderAggregate) UpdatedOrderWithItems(name string, orderItems []OrderItem) error {
	if o.IsSubmitted {
		return ErrOrderIsSubmitted
	}
	eventData := OrderUpdatedEvent{
		Name:       name,
		OrderItems: orderItems,
	}
	updatedOrderEvent := core.NewEvent(o.GetID(), eventData.GetEventType(), eventData)
	o.appendEvent(updatedOrderEvent)
	o.Apply(updatedOrderEvent)

	return nil
}

func (o *OrderAggregate) UpdateOrderItemAmount(id uuid.UUID, amount int) error {
	if o.IsSubmitted {
		return ErrOrderIsSubmitted
	}

	if amount < 0 {
		return ErrItemAmountLessThanZero
	}

	var exist bool
	for _, item := range o.OrderItems {
		if item.ID == id {
			exist = true
			break
		}
	}

	if !exist {
		return ErrItemNotFound
	}

	eventData := OrderItemAmountUpdatedEvent{
		ID:     id,
		Amount: amount,
	}
	updatedOrderEvent := core.NewEvent(o.GetID(), eventData.GetEventType(), eventData)
	o.appendEvent(updatedOrderEvent)
	o.Apply(updatedOrderEvent)
	return nil
}
