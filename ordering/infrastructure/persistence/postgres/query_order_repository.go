package postgres

import (
	"encoding/json"

	"github.com/Bass-Peerapon/eventsource-demo/ordering/domain/order"
	"github.com/jmoiron/sqlx"
)

type queryOrderRepository struct {
	db *sqlx.DB
}

// SaveOrder implements order.QueryOrderRepository.
func (q *queryOrderRepository) SaveOrder(OrderAggregate *order.OrderAggregate) error {
	tx, err := q.db.Beginx()
	if err != nil {
		return err
	}

	defer tx.Rollback()
	orderItems, _ := json.Marshal(OrderAggregate.OrderItems)
	query := `
	INSERT INTO orders (id, version, name, order_items, is_submitted)
	    VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (id)
	    DO UPDATE SET
	        version = $2, name = $3, order_items = $4, is_submitted = $5, updated_at = NOW()
			`
	_, err = tx.Exec(query,
		OrderAggregate.ID, OrderAggregate.Version, OrderAggregate.Name, orderItems, OrderAggregate.IsSubmitted,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
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
