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
}

func NewPageHandler(renderer httpapi.Renderer, logger *slog.Logger, check func(ctx context.Context) error) *PageHandler {
	return &PageHandler{renderer: renderer, logger: logger, check: check}
}

func (h *PageHandler) Home(c *gin.Context) {
	renderHTML(c, h.renderer, h.logger, http.StatusOK, "index.html", views.HomePage{
		Title: "Automarket",
	})
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
