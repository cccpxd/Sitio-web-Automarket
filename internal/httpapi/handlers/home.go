package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"automarket/internal/httpapi"
	"automarket/internal/views"
	"github.com/gin-gonic/gin"
)

type PageHandler struct {
	renderer httpapi.Renderer
	logger   *slog.Logger
	check    func(ctx context.Context) error
	cars     httpapi.CarService
}

func NewPageHandler(renderer httpapi.Renderer, logger *slog.Logger, check func(ctx context.Context) error, cars ...httpapi.CarService) *PageHandler {
	h := &PageHandler{renderer: renderer, logger: logger, check: check}
	if len(cars) > 0 {
		h.cars = cars[0]
	}
	return h
}

func (h *PageHandler) SetCarService(cars httpapi.CarService) {
	h.cars = cars
}

func (h *PageHandler) Home(c *gin.Context) {
	page := views.HomePage{Title: "Automarket"}
	if h.cars != nil {
		cars, err := h.cars.ListCars(c.Request.Context())
		if err == nil {
			page = views.NewHomePage(cars)
		}
	}
	renderHTML(c, h.renderer, h.logger, http.StatusOK, "index.html", page)
}

type healthResponse struct {
	Status string `json:"status"`
}

func (h *PageHandler) Health(c *gin.Context) {
	if h.check != nil {
		if err := h.check(c.Request.Context()); err != nil {
			h.logger.ErrorContext(c.Request.Context(), "health_check_failed", "error", err)
			c.JSON(http.StatusServiceUnavailable, healthResponse{Status: "unavailable"})
			return
		}
	}
	c.JSON(http.StatusOK, healthResponse{Status: "ok"})
}
