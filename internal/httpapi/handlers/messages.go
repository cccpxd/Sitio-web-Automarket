package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"automarket/internal/domain"
	"automarket/internal/httpapi"
	"automarket/internal/views"
	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	messaging httpapi.MessagingService
	catalog   httpapi.CatalogService
	renderer  httpapi.Renderer
	logger    *slog.Logger
}

func NewMessageHandler(messaging httpapi.MessagingService, catalog httpapi.CatalogService, renderer httpapi.Renderer, logger *slog.Logger) *MessageHandler {
	return &MessageHandler{messaging: messaging, catalog: catalog, renderer: renderer, logger: logger}
}

func (h *MessageHandler) ContactForm(c *gin.Context) {
	if h.catalog == nil {
		renderErrorPage(c, h.renderer, h.logger, http.StatusNotFound, "La publicación no existe o ya no está disponible.")
		return
	}
	car, err := h.catalog.GetActive(c.Request.Context(), c.Param("id"))
	if err != nil {
		renderErrorPage(c, h.renderer, h.logger, http.StatusNotFound, "La publicación no existe o ya no está disponible.")
		return
	}

	user := currentUserPtr(c)
	if user != nil && car.SellerID == user.ID {
		renderErrorPage(c, h.renderer, h.logger, http.StatusForbidden, "No puedes contactar tu propia publicación.")
		return
	}

	page := views.ContactPage{
		BasePage: views.NewBasePage(httpapi.CSRFToken(c), user, "Contactar al vendedor", "catalog"),
		Car:      views.NewCarView(car),
		Message:  h.messaging.InitialMessage(car),
	}
	renderHTML(c, h.renderer, h.logger, http.StatusOK, "contact.html", page)
}

func (h *MessageHandler) ContactSubmit(c *gin.Context) {
	user, ok := httpapi.CurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login?next=/cars/"+c.Param("id")+"/contacto")
		return
	}
	car, err := h.catalog.GetActive(c.Request.Context(), c.Param("id"))
	if err != nil {
		renderErrorPage(c, h.renderer, h.logger, http.StatusNotFound, "La publicación no existe o ya no está disponible.")
		return
	}
	message := c.PostForm("message")
	conversation, err := h.messaging.Start(c.Request.Context(), car.ID, user.ID, message)
	if err != nil {
		page := views.ContactPage{
			BasePage: views.NewBasePage(httpapi.CSRFToken(c), &user, "Contactar al vendedor", "catalog"),
			Car:      views.NewCarView(car),
			Message:  message,
		}
		page.ErrorMessage = messageErrorMessage(err)
		renderHTML(c, h.renderer, h.logger, messageErrorStatus(err), "contact.html", page)
		return
	}
	c.Redirect(http.StatusFound, "/mensajes/"+conversation.ID)
}

func (h *MessageHandler) Inbox(c *gin.Context) {
	user, ok := httpapi.CurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login?next=/mensajes")
		return
	}
	conversations, err := h.messaging.List(c.Request.Context(), user.ID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "list_messages_failed", "error", err)
	}
	page := views.MessagesPage{
		BasePage: views.NewBasePage(httpapi.CSRFToken(c), &user, "Mensajes", "messages"),
	}
	for _, conversation := range conversations {
		page.Conversations = append(page.Conversations, views.NewConversationView(conversation, user.ID))
	}
	renderHTML(c, h.renderer, h.logger, http.StatusOK, "messages.html", page)
}

func (h *MessageHandler) Thread(c *gin.Context) {
	user, ok := httpapi.CurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login?next=/mensajes")
		return
	}
	conversation, err := h.messaging.Get(c.Request.Context(), c.Param("id"), user)
	if err != nil {
		renderErrorPage(c, h.renderer, h.logger, messageErrorStatus(err), messageErrorMessage(err))
		return
	}
	page := views.ConversationPage{
		BasePage:     views.NewBasePage(httpapi.CSRFToken(c), &user, "Conversación", "messages"),
		Conversation: views.NewConversationView(conversation, user.ID),
		ReadOnly:     user.Role == domain.RoleAdmin,
	}
	renderHTML(c, h.renderer, h.logger, http.StatusOK, "conversation.html", page)
}

func (h *MessageHandler) Reply(c *gin.Context) {
	user, ok := httpapi.CurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	if _, err := h.messaging.Reply(c.Request.Context(), c.Param("id"), user, c.PostForm("message")); err != nil {
		conversation, getErr := h.messaging.Get(c.Request.Context(), c.Param("id"), user)
		if getErr != nil {
			renderErrorPage(c, h.renderer, h.logger, messageErrorStatus(getErr), messageErrorMessage(getErr))
			return
		}
		page := views.ConversationPage{
			BasePage:     views.NewBasePage(httpapi.CSRFToken(c), &user, "Conversación", "messages"),
			Conversation: views.NewConversationView(conversation, user.ID),
			Error:        messageErrorMessage(err),
			ReadOnly:     user.Role == domain.RoleAdmin,
		}
		renderHTML(c, h.renderer, h.logger, messageErrorStatus(err), "conversation.html", page)
		return
	}
	c.Redirect(http.StatusFound, "/mensajes/"+c.Param("id"))
}

func messageErrorMessage(err error) string {
	switch {
	case errors.Is(err, domain.ErrForbidden):
		return "No tienes acceso a esta conversación."
	case errors.Is(err, domain.ErrNotFound):
		return "La conversación no existe."
	case errors.Is(err, domain.ErrSelfMessage):
		return "No puedes enviarte mensajes sobre tu propia publicación."
	case errors.Is(err, domain.ErrMessageEmpty):
		return "Escribe un mensaje antes de enviarlo."
	case errors.Is(err, domain.ErrMessageTooLong):
		return "El mensaje es demasiado largo."
	case errors.Is(err, domain.ErrConversationRead):
		return "El administrador puede leer la conversación, pero no responder."
	default:
		return "No se pudo completar la operación con el mensaje."
	}
}

func messageErrorStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrMessageEmpty), errors.Is(err, domain.ErrMessageTooLong),
		errors.Is(err, domain.ErrSelfMessage), errors.Is(err, domain.ErrConversationRead):
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

func renderErrorPage(c *gin.Context, renderer httpapi.Renderer, logger *slog.Logger, status int, message string) {
	page := views.ErrorPage{
		BasePage: views.NewBasePage(httpapi.CSRFToken(c), currentUserPtr(c), "No disponible", ""),
		Status:   status,
		Message:  message,
	}
	renderHTML(c, renderer, logger, status, "error.html", page)
}
