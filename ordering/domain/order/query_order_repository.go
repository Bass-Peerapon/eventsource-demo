package order

import "github.com/gofrs/uuid"

type Order struct {
	ID          uuid.UUID
	Name        string
	OrderItems  []OrderItem
	IsSubmitted bool
}

type QueryOrderRepository interface {
	GetOrders() ([]Order, error)
}
