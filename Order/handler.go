package order

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

func (h *Handler) PostHandler(c echo.Context) error {
	var order Order
	err := c.Bind(&order) 
	if err != nil {
		res := make(map[string]any)
		res["error"] = "Bad Request"
		log.Errorf("Bad Request")
		return c.JSON(http.StatusBadRequest, res)
	}
	log.Debug("The Request is ok")

	err = h.Repo.SaveOrder(c.Request().Context(), &order)
	var order_event = OrderEvent{
		Order: order,
		EventID: uuid.New().String(),
	}
	if err != nil {
		log.Error("Could not save order to order datavase")
		return err
	}
	if err := h.Producer.PublishOrderCreated(c.Request().Context(), &order_event); err != nil {
		log.Errorf("Failed to publish order to Kafka: %v", err)
		return c.JSON(500, map[string]string{"message": "failed to publish event"})
	}
	log.Infof("Order %d published to Kafka", order.ID)
	log.Infof("Post request proccessed successfully")
	return c.JSON(http.StatusCreated, order)
}