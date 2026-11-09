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
			card := views.NewCarCard(car, time.Time{})
			page.Cars = append(page.Cars, card)
			if car.Status == domain.CarStatusPending {
				page.Pending = append(page.Pending, card)
			}
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
	if c.Query("aprobada") == "1" {
		page.Flash = "La publicación fue aprobada y ya es visible."
	}
	if c.Query("rechazada") == "1" {
		page.Flash = "La publicación fue rechazada y la razón quedó registrada."
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
		BasePage:    views.NewBasePage(httpapi.CSRFToken(c), &user, "Retirar publicación", "admin"),
		Car:         views.NewCarView(car),
		ActionTitle: "Confirmar retirada",
		ActionURL:   "/admin/publicaciones/" + car.ID + "/retirar",
		ReasonLabel: "Razón de la retirada (obligatoria)",
		SubmitLabel: "Confirmar retirada",
		CancelURL:   "/admin/moderacion",
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
						BasePage:    views.NewBasePage(httpapi.CSRFToken(c), &user, "Retirar publicación", "admin"),
						Car:         views.NewCarView(car),
						ActionTitle: "Confirmar retirada",
						ActionURL:   "/admin/publicaciones/" + car.ID + "/retirar",
						ReasonLabel: "Razón de la retirada (obligatoria)",
						SubmitLabel: "Confirmar retirada",
						CancelURL:   "/admin/moderacion",
						Reason:      reason,
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

func (h *ModerationHandler) Approve(c *gin.Context) {
	user, ok := httpapi.CurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	if err := h.moderation.Approve(c.Request.Context(), user.ID, c.Param("id")); err != nil {
		if moderationStatus(err) >= http.StatusInternalServerError {
			h.logger.ErrorContext(c.Request.Context(), "approve_car_failed", "error", err)
		}
		renderErrorPage(c, h.renderer, h.logger, moderationStatus(err), moderationMessage(err))
		return
	}
	c.Redirect(http.StatusFound, "/admin/moderacion?aprobada=1")
}

func (h *ModerationHandler) RejectConfirm(c *gin.Context) {
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
		BasePage:    views.NewBasePage(httpapi.CSRFToken(c), &user, "Rechazar publicación", "admin"),
		Car:         views.NewCarView(car),
		ActionTitle: "Rechazar publicación",
		ActionURL:   "/admin/publicaciones/" + car.ID + "/rechazar",
		ReasonLabel: "Razón del rechazo (obligatoria)",
		SubmitLabel: "Confirmar rechazo",
		CancelURL:   "/admin/moderacion",
	}
	renderHTML(c, h.renderer, h.logger, http.StatusOK, "moderation_confirm.html", page)
}

func (h *ModerationHandler) Reject(c *gin.Context) {
	user, ok := httpapi.CurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	reason := c.PostForm("reason")
	if err := h.moderation.Reject(c.Request.Context(), user.ID, c.Param("id"), reason); err != nil {
		if (errors.Is(err, domain.ErrReasonRequired) || errors.Is(err, domain.ErrReasonTooLong)) && h.catalog != nil {
			if car, getErr := h.catalog.Get(c.Request.Context(), c.Param("id")); getErr == nil {
				page := views.ModerationConfirmPage{
					BasePage: views.NewBasePage(httpapi.CSRFToken(c), &user, "Rechazar publicación", "admin"),
					Car:      views.NewCarView(car), Reason: reason,
					ActionTitle: "Rechazar publicación",
					ActionURL:   "/admin/publicaciones/" + car.ID + "/rechazar",
					ReasonLabel: "Razón del rechazo (obligatoria)",
					SubmitLabel: "Confirmar rechazo",
					CancelURL:   "/admin/moderacion",
				}
				page.ErrorMessage = moderationMessage(err)
				renderHTML(c, h.renderer, h.logger, http.StatusUnprocessableEntity, "moderation_confirm.html", page)
				return
			}
		}
		renderErrorPage(c, h.renderer, h.logger, moderationStatus(err), moderationMessage(err))
		return
	}
	c.Redirect(http.StatusFound, "/admin/moderacion?rechazada=1")
}

func moderationMessage(err error) string {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return "La publicación no existe."
	case errors.Is(err, domain.ErrReasonRequired):
		return "Indica una razón."
	case errors.Is(err, domain.ErrReasonTooLong):
		return "La razón es demasiado larga."
	case errors.Is(err, domain.ErrConflict):
		return "La publicación no se puede moderar en su estado actual."
	case errors.Is(err, domain.ErrForbidden):
		return "No tienes permisos para moderar publicaciones."
	default:
		return "No se pudo completar la acción."
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
