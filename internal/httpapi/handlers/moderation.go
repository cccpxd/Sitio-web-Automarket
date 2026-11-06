package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"automarket/internal/domain"
	"automarket/internal/httpapi"
	"automarket/internal/views"
	"github.com/gin-gonic/gin"
)

type ModerationHandler struct {
	moderation httpapi.ModerationService
	catalog    httpapi.CatalogService
	admin      httpapi.AdminService
	renderer   httpapi.Renderer
	logger     *slog.Logger
}

func NewModerationHandler(moderation httpapi.ModerationService, catalog httpapi.CatalogService, admin httpapi.AdminService, renderer httpapi.Renderer, logger *slog.Logger) *ModerationHandler {
	return &ModerationHandler{moderation: moderation, catalog: catalog, admin: admin, renderer: renderer, logger: logger}
}

// Panel renders the administrator dashboard with all listings and the audit log.
func (h *ModerationHandler) Panel(c *gin.Context) {
	user, ok := httpapi.CurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login?next=/admin/moderacion")
		return
	}
	page := views.ModerationPage{
		BasePage: views.NewBasePage(httpapi.CSRFToken(c), &user, "Moderación", "admin"),
	}
	if h.catalog != nil {
		cars, err := h.catalog.ListAll(c.Request.Context())
		if err != nil {
			h.logger.ErrorContext(c.Request.Context(), "list_all_cars_failed", "error", err)
		}
		for _, car := range cars {
			page.Cars = append(page.Cars, views.NewCarCard(car, time.Time{}))
		}
	}
	if h.moderation != nil {
		entries, err := h.moderation.List(c.Request.Context())
		if err != nil {
			h.logger.ErrorContext(c.Request.Context(), "list_moderation_failed", "error", err)
		}
		page.Entries = entries
	}
	if c.Query("retirada") == "1" {
		page.Flash = "La publicación fue retirada y la acción quedó registrada."
	}
	renderHTML(c, h.renderer, h.logger, http.StatusOK, "moderation.html", page)
}

// Confirm shows the mandatory-reason form.
func (h *ModerationHandler) Confirm(c *gin.Context) {
	user, ok := httpapi.CurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	if h.catalog == nil {
		renderErrorPage(c, h.renderer, h.logger, http.StatusNotFound, "La publicación no existe.")
		return
	}
	car, err := h.catalog.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		renderErrorPage(c, h.renderer, h.logger, moderationStatus(err), moderationMessage(err))
		return
	}
	page := views.ModerationConfirmPage{
		BasePage: views.NewBasePage(httpapi.CSRFToken(c), &user, "Retirar publicación", "admin"),
		Car:      views.NewCarView(car),
	}
	renderHTML(c, h.renderer, h.logger, http.StatusOK, "moderation_confirm.html", page)
}

// Remove applies the administrative removal with a mandatory reason.
func (h *ModerationHandler) Remove(c *gin.Context) {
	user, ok := httpapi.CurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	reason := c.PostForm("reason")
	if err := h.moderation.Remove(c.Request.Context(), user.ID, c.Param("id"), reason); err != nil {
		if errors.Is(err, domain.ErrReasonRequired) || errors.Is(err, domain.ErrReasonTooLong) {
			if h.catalog != nil {
				if car, getErr := h.catalog.Get(c.Request.Context(), c.Param("id")); getErr == nil {
					page := views.ModerationConfirmPage{
						BasePage: views.NewBasePage(httpapi.CSRFToken(c), &user, "Retirar publicación", "admin"),
						Car:      views.NewCarView(car),
						Reason:   reason,
					}
					page.ErrorMessage = moderationMessage(err)
					renderHTML(c, h.renderer, h.logger, http.StatusUnprocessableEntity, "moderation_confirm.html", page)
					return
				}
			}
		}
		if moderationStatus(err) >= http.StatusInternalServerError {
			h.logger.ErrorContext(c.Request.Context(), "moderate_car_failed", "error", err)
		}
		renderErrorPage(c, h.renderer, h.logger, moderationStatus(err), moderationMessage(err))
		return
	}
	c.Redirect(http.StatusFound, "/admin/moderacion?retirada=1")
}

func moderationMessage(err error) string {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return "La publicación no existe."
	case errors.Is(err, domain.ErrReasonRequired):
		return "Indica la razón de la retirada."
	case errors.Is(err, domain.ErrReasonTooLong):
		return "La razón es demasiado larga."
	case errors.Is(err, domain.ErrConflict):
		return "La publicación no se puede retirar en su estado actual."
	case errors.Is(err, domain.ErrForbidden):
		return "No tienes permisos para retirar publicaciones."
	default:
		return "No se pudo completar la retirada."
	}
}

func moderationStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrReasonRequired), errors.Is(err, domain.ErrReasonTooLong):
		return http.StatusUnprocessableEntity
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}
