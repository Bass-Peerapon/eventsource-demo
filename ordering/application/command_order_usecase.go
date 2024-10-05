package application

import (
	"errors"

	"github.com/Bass-Peerapon/eventsource-demo/ordering/domain/core"
	"github.com/Bass-Peerapon/eventsource-demo/ordering/domain/order"
	"github.com/Bass-Peerapon/eventsource-demo/ordering/infrastructure/persistence/postgres"
	"github.com/gofrs/uuid"
)

type CommandOrderUsecase interface {
	CreateOrder(name string, orderItems []order.OrderItem) error
	UpdatedOrder(id uuid.UUID, name string, orderItems []order.OrderItem) error
	UpdateOrderItemAmount(id uuid.UUID, orderItemID uuid.UUID, amount int) error
}

type commandOrderUsecase struct {
	eventRepo       core.EventRepository
	aggregateRepo   core.AggregateRepository
	orderProjection OrderProjection
}

// CreateOrder implements OrderUsecase.
func (o *commandOrderUsecase) CreateOrder(name string, orderItems []order.OrderItem) error {
	order := order.CreateOrderWithItems(name, orderItems)
	if err := o.aggregateRepo.SaveAggregate(order); err != nil {
		return err
	}
	if err := o.eventRepo.SaveEvents(order.Events); err != nil {
		return err
	}
	if err := o.orderProjection.HandleEvent(order); err != nil {
		return err
	}
	return nil
}

// UpdateOrderItemAmount implements OrderUsecase.
func (o *commandOrderUsecase) UpdateOrderItemAmount(id uuid.UUID, orderItemID uuid.UUID, amount int) error {
	orderAggregate := order.OrderAggregate{}
	snapshot, err := o.aggregateRepo.LoadSnapshot(id, nil)
	if err != nil {
		return err
	}
	if snapshot != nil {
		snapshot.UnSerialize(&orderAggregate)
	}
	loadedEvents, err := o.eventRepo.LoadEvents(id, &orderAggregate.Version, nil)
	if err != nil {
		return err
	}

	for _, event := range loadedEvents {
		orderAggregate.Apply(event)
	}

	if err := orderAggregate.UpdateOrderItemAmount(orderItemID, amount); err != nil {
		return err
	}

	if err := o.aggregateRepo.SaveAggregate(&orderAggregate); err != nil {
		if errors.Is(err, postgres.ErrAggregateOutdated) {
			return o.UpdateOrderItemAmount(id, orderItemID, amount)
		}
		return err
	}
	if err := o.eventRepo.SaveEvents(orderAggregate.Events); err != nil {
		return err
	}

	for _, event := range orderAggregate.Events {
		if event.Version%10 == 0 {
			snapshot := core.AggregateSnapshot{
				AggregateID: id,
				Version:     event.Version,
				EventData:   event.EventData,
			}
			if err := o.aggregateRepo.SaveSnapshot(&snapshot); err != nil {
				return err
			}
		}
	}

	if err := o.orderProjection.HandleEvent(&orderAggregate); err != nil {
		return err
	}

	return nil
}

// UpdatedOrder implements OrderUsecase.
func (o *commandOrderUsecase) UpdatedOrder(id uuid.UUID, name string, orderItems []order.OrderItem) error {
	orderAggregate := order.OrderAggregate{}
	snapshot, err := o.aggregateRepo.LoadSnapshot(id, nil)
	if err != nil {
		return err
	}
	if snapshot != nil {
		snapshot.UnSerialize(&orderAggregate)
	}

	loadedEvents, err := o.eventRepo.LoadEvents(id, &orderAggregate.Version, nil)
	if err != nil {
		return err
	}

	for _, event := range loadedEvents {
		orderAggregate.Apply(event)
	}

	items := make([]order.OrderItem, 0, len(orderItems))
	for _, v := range orderItems {
		items = append(items, order.OrderItem{
			ID:     v.ID,
			Name:   v.Name,
			Amount: v.Amount,
		})
	}

	if err := orderAggregate.UpdatedOrderWithItems(name, items); err != nil {
		return err
	}
	if err := o.aggregateRepo.SaveAggregate(&orderAggregate); err != nil {
		if errors.Is(err, postgres.ErrAggregateOutdated) {
			return o.UpdatedOrder(id, name, orderItems)
		}
		return err
	}

	if err := o.eventRepo.SaveEvents(orderAggregate.Events); err != nil {
		return err
	}

	for _, event := range orderAggregate.Events {
		if event.Version%10 == 0 {
			snapshot := core.AggregateSnapshot{
				AggregateID: id,
				Version:     event.Version,
				EventData:   event.EventData,
			}
			if err := o.aggregateRepo.SaveSnapshot(&snapshot); err != nil {
				return err
			}
		}
	}

	if err := o.orderProjection.HandleEvent(&orderAggregate); err != nil {
		return err
	}
	return nil
}

func NewCommandOrderUsecase(eventStore core.EventRepository, aggregateRepo core.AggregateRepository, orderProjection OrderProjection) CommandOrderUsecase {
	return &commandOrderUsecase{
		eventRepo:       eventStore,
		aggregateRepo:   aggregateRepo,
		orderProjection: orderProjection,
	}
}
