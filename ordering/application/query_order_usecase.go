package application

import (
	"github.com/Bass-Peerapon/eventsource-demo/ordering/domain/order"
)

type QueryOrderUsecase interface {
	GetOrders() ([]order.Order, error)
}

type queryOrderUsecase struct {
	orderRepository order.QueryOrderRepository
}

// GetOrders implements QueryOrderUsecase.
func (q *queryOrderUsecase) GetOrders() ([]order.Order, error) {
	return q.orderRepository.GetOrders()
}

func NewQueryOrderUsecase(orderRepository order.QueryOrderRepository) QueryOrderUsecase {
	return &queryOrderUsecase{
		orderRepository: orderRepository,
	}
}
