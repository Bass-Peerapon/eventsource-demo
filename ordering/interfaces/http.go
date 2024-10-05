package interfaces

import (
	"github.com/Bass-Peerapon/eventsource-demo/ordering/interfaces/api"
	"github.com/labstack/echo/v4"
)

type Route struct {
	e *echo.Echo
}

func NewRoute(e *echo.Echo) *Route {
	return &Route{
		e: e,
	}
}

func (r *Route) RegisterCommandOrderHandler(h api.CommandHandler) {
	r.e.POST("/orders", h.CreateOrderHadler)
	r.e.PUT("/orders/:id", h.UpdatedOrderHandler)
	r.e.PUT("/orders/:id/items/:item_id", h.UpdateOrderItemAmountHandler)
}

func (r *Route) RegisterQueryOrderHandler(h api.QueryHandler) {
	r.e.GET("/orders", h.GetOrdersHandler)
}
