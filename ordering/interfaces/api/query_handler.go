package api

import (
	"net/http"

	"github.com/Bass-Peerapon/eventsource-demo/ordering/application"
	"github.com/labstack/echo/v4"
)

type QueryHandler interface {
	GetOrdersHandler(c echo.Context) error
}

type queryHandler struct {
	queryOrderUsecase application.QueryOrderUsecase
}

// GetOrdersHandler implements QueryHandler.
func (q *queryHandler) GetOrdersHandler(c echo.Context) error {
	orders, err := q.queryOrderUsecase.GetOrders()
	if err != nil {
		return err
	}
	resp := map[string]interface{}{
		"orders": orders,
	}
	return c.JSON(http.StatusOK, resp)
}

func NewQueryHandler(queryOrderUsecase application.QueryOrderUsecase) QueryHandler {
	return &queryHandler{
		queryOrderUsecase: queryOrderUsecase,
	}
}
