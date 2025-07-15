package payment

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handler) GetHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{"status": "ok"})
}