package api

import (
	"net/http"

	"github.com/Bass-Peerapon/eventsource-demo/ordering/application"
	"github.com/Bass-Peerapon/eventsource-demo/ordering/domain/order"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
)

type CommandHandler interface {
	CreateOrderHadler(c echo.Context) error
	UpdatedOrderHandler(c echo.Context) error
	UpdateOrderItemAmountHandler(c echo.Context) error
}

type commandHandler struct {
	commandOrderUsecase application.CommandOrderUsecase
}

type orderRequest struct {
	Name       string `json:"name"`
	OrderItems []struct {
		ID     uuid.UUID `json:"id"`
		Name   string    `json:"name"`
		Amount int       `json:"amount"`
	} `json:"order_items"`
}

type updateOrderItemAmountRequest struct {
	OrderItemID string `json:"order_item_id"`
	Amount      int    `json:"amount"`
}

// CreateOrderHadler implements CommandHandler.
func (h *commandHandler) CreateOrderHadler(c echo.Context) error {
	orderRequest := orderRequest{}

	if err := c.Bind(&orderRequest); err != nil {
		return err
	}

	orderItems := make([]order.OrderItem, 0, len(orderRequest.OrderItems))
	for _, orderItem := range orderRequest.OrderItems {
		orderItems = append(orderItems, order.OrderItem{
			ID:     orderItem.ID,
			Name:   orderItem.Name,
			Amount: orderItem.Amount,
		})
	}

	if err := h.commandOrderUsecase.CreateOrder(orderRequest.Name, orderItems); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, orderRequest)
}

// UpdateOrderItemAmountHandler implements CommandHandler.
func (h *commandHandler) UpdateOrderItemAmountHandler(c echo.Context) error {
	id := uuid.FromStringOrNil(c.Param("id"))

	updateOrderItemAmountRequest := updateOrderItemAmountRequest{}

	if err := c.Bind(&updateOrderItemAmountRequest); err != nil {
		return err
	}

	if err := h.commandOrderUsecase.UpdateOrderItemAmount(id, uuid.FromStringOrNil(updateOrderItemAmountRequest.OrderItemID), updateOrderItemAmountRequest.Amount); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, updateOrderItemAmountRequest)
}

// UpdatedOrderHandler implements CommandHandler.
func (h *commandHandler) UpdatedOrderHandler(c echo.Context) error {
	id := uuid.FromStringOrNil(c.Param("id"))
	orderRequest := orderRequest{}
	if err := c.Bind(&orderRequest); err != nil {
		return err
	}
	orderItems := make([]order.OrderItem, 0, len(orderRequest.OrderItems))
	for _, orderItem := range orderRequest.OrderItems {
		orderItems = append(orderItems, order.OrderItem{
			ID:     orderItem.ID,
			Name:   orderItem.Name,
			Amount: orderItem.Amount,
		})
	}

	if err := h.commandOrderUsecase.UpdatedOrder(id, orderRequest.Name, orderItems); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, orderRequest)
}

func NewCommandHandler(commandOrderUsecase application.CommandOrderUsecase) CommandHandler {
	return &commandHandler{
		commandOrderUsecase: commandOrderUsecase,
	}
}
