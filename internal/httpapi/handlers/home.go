package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"automarket/internal/domain"
	"automarket/internal/httpapi"
	"automarket/internal/views"
	"github.com/gin-gonic/gin"
)

type PageHandler struct {
	renderer httpapi.Renderer
	logger   *slog.Logger
	check    func(ctx context.Context) error
	catalog  httpapi.CatalogService
}

func NewPageHandler(renderer httpapi.Renderer, logger *slog.Logger, check func(ctx context.Context) error, catalog ...httpapi.CatalogService) *PageHandler {
	h := &PageHandler{renderer: renderer, logger: logger, check: check}
	if len(catalog) > 0 {
		h.catalog = catalog[0]
	}
	return h
}

func (h *PageHandler) SetCatalogService(catalog httpapi.CatalogService) { h.catalog = catalog }

// Home renders the public catalog. The landing page is the catalog so that
// visitors always see real listings or an honest empty state.
func (h *PageHandler) Home(c *gin.Context) { h.Catalog(c) }

func (h *PageHandler) Catalog(c *gin.Context) {
	raw := domain.SearchQuery{
		Q:        c.Query("q"),
		City:     c.Query("city"),
		Year:     c.Query("year"),
		MinPrice: c.Query("min"),
		MaxPrice: c.Query("max"),
		Page:     c.Query("page"),
	}

	result := domain.SearchResult{Page: 1, PageSize: domain.DefaultPageSize}
	filters, filterErr := domain.NormalizeSearchFilters(raw)
	if filterErr == nil && h.catalog != nil {
		searchResult, err := h.catalog.Search(c.Request.Context(), raw)
		if err != nil {
			h.logger.ErrorContext(c.Request.Context(), "catalog_search_failed", "error", err)
		} else {
			result = searchResult
		}
	} else if filterErr != nil {
		filters = domain.SearchFilters{Page: 1, PageSize: domain.DefaultPageSize}
	}

	page := views.NewCatalogPage(httpapi.CSRFToken(c), currentUserPtr(c), result, filters)
	if filterErr != nil {
		page.ErrorMessage = "Revisa los filtros: el año o el rango de precio no son válidos."
	}
	renderHTML(c, h.renderer, h.logger, statusForFilterError(filterErr), "catalog.html", page)
}

func (h *PageHandler) Detail(c *gin.Context) {
	if h.catalog == nil {
		h.renderError(c, http.StatusNotFound, "La publicación no existe o ya no está disponible.")
		return
	}
	car, err := h.catalog.GetActive(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			h.renderError(c, http.StatusNotFound, "La publicación no existe o ya no está disponible.")
			return
		}
		h.logger.ErrorContext(c.Request.Context(), "car_detail_failed", "error", err)
		h.renderError(c, http.StatusInternalServerError, "No se pudo cargar la publicación.")
		return
	}

	user := currentUserPtr(c)
	page := views.CarDetailPage{
		BasePage:   views.NewBasePage(httpapi.CSRFToken(c), user, car.Title(), "catalog"),
		Car:        views.NewCarView(car),
		CanContact: true,
	}
	for index, photo := range car.Photos {
		page.Photos = append(page.Photos, views.PhotoView{
			ID:    photo.ID,
			Alt:   car.Title() + " fotografía " + itoa(index+1),
			Index: index,
		})
	}
	if user != nil && user.ID == car.SellerID {
		page.IsOwner = true
		page.CanContact = false
	}
	renderHTML(c, h.renderer, h.logger, http.StatusOK, "detail.html", page)
}

func (h *PageHandler) renderError(c *gin.Context, status int, message string) {
	page := views.ErrorPage{
		BasePage: views.NewBasePage(httpapi.CSRFToken(c), currentUserPtr(c), "No disponible", ""),
		Status:   status,
		Message:  message,
	}
	renderHTML(c, h.renderer, h.logger, status, "error.html", page)
}

func statusForFilterError(err error) int {
	if err != nil {
		return http.StatusBadRequest
	}
	return http.StatusOK
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

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	negative := value < 0
	if negative {
		value = -value
	}
	var buffer [20]byte
	index := len(buffer)
	for value > 0 {
		index--
		buffer[index] = byte('0' + value%10)
		value /= 10
	}
	if negative {
		index--
		buffer[index] = '-'
	}
	return string(buffer[index:])
}
