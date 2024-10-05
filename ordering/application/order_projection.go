package application

import (
	"github.com/Bass-Peerapon/eventsource-demo/ordering/domain/core"
	"github.com/Bass-Peerapon/eventsource-demo/ordering/domain/order"
)

type OrderProjection interface {
	core.SyncEventHandler
}

type orderProjection struct {
	orderRepository order.QueryOrderRepository
}

// GetAggregateType implements core.SyncEventHandler.
func (o *orderProjection) GetAggregateType() string {
	oderAggregate := order.OrderAggregate{}
	return oderAggregate.GetAggregateType()
}

// HandleEvent implements core.SyncEventHandler.
func (o *orderProjection) HandleEvent(aggregate core.Aggregate) error {
	orderAggregate := aggregate.(*order.OrderAggregate)
	return o.orderRepository.SaveOrder(orderAggregate)
}

func NewOrderProjection(orderRepository order.QueryOrderRepository) core.SyncEventHandler {
	return &orderProjection{
		orderRepository: orderRepository,
	}
}
