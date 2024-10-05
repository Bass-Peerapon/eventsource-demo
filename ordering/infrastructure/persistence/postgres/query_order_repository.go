package postgres

import (
	"github.com/Bass-Peerapon/eventsource-demo/ordering/domain/order"
	"github.com/jmoiron/sqlx"
)

type queryOrderRepository struct {
	db *sqlx.DB
}

// GetOrders implements order.QueryOrderRepository.
func (q *queryOrderRepository) GetOrders() ([]order.Order, error) {
	var orders []order.Order
	if err := q.db.Select(&orders, "SELECT * FROM orders"); err != nil {
		return nil, err
	}
	return orders, nil
}

func NewQueryOrderRepository(db *sqlx.DB) order.QueryOrderRepository {
	return &queryOrderRepository{
		db: db,
	}
}
