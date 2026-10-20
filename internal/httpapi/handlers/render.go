package handlers

import (
	"log/slog"

	"automarket/internal/httpapi"
	"github.com/gin-gonic/gin"
)

func renderHTML(c *gin.Context, renderer httpapi.Renderer, logger *slog.Logger, status int, templateName string, data any) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Status(status)
	if err := renderer.Render(c.Writer, templateName, data); err != nil {
		logger.ErrorContext(c.Request.Context(), "render_template_failed",
			"template", templateName,
			"error", err,
		)
	}
}
