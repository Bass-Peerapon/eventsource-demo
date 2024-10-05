package order

import (
	"reflect"

	"github.com/gofrs/uuid"
)

type OrderCreatedEvent struct {
	Name       string      `json:"name"`
	OrderItems []OrderItem `json:"order_items"`
}

func (c OrderCreatedEvent) GetEventType() string {
	return reflect.TypeOf(c).Name()
}

type OrderUpdatedEvent struct {
	Name       string      `json:"name"`
	OrderItems []OrderItem `json:"order_items"`
}

func (u OrderUpdatedEvent) GetEventType() string {
	return reflect.TypeOf(u).Name()
}

type OrderItemAmountUpdatedEvent struct {
	ID     uuid.UUID `json:"id"`
	Amount int       `json:"amount"`
}

func (u OrderItemAmountUpdatedEvent) GetEventType() string {
	return reflect.TypeOf(u).Name()
}
