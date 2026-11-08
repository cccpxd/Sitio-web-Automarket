package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"automarket/internal/domain"
	"automarket/internal/httpapi"
	"automarket/internal/service"
	"automarket/internal/views"
	"github.com/gin-gonic/gin"
)

type PublishingHandler struct {
	publishing httpapi.PublishingService
	catalog    httpapi.CatalogService
	renderer   httpapi.Renderer
	logger     *slog.Logger
}

func NewPublishingHandler(publishing httpapi.PublishingService, catalog httpapi.CatalogService, renderer httpapi.Renderer, logger *slog.Logger) *PublishingHandler {
	return &PublishingHandler{publishing: publishing, catalog: catalog, renderer: renderer, logger: logger}
}

func (h *PublishingHandler) PublishForm(c *gin.Context) {
	page := views.PublishPage{
		BasePage:  views.NewBasePage(httpapi.CSRFToken(c), currentUserPtr(c), "Publicar vehículo", "publish"),
		Form:      views.CarForm{ClientToken: newClientToken()},
		Errors:    map[string]string{},
		MaxPhotos: domain.MaxCarPhotos,
		MaxBytes:  domain.MaxPhotoBytes,
	}
	renderHTML(c, h.renderer, h.logger, http.StatusOK, "publish.html", page)
}

func (h *PublishingHandler) Create(c *gin.Context) {
	user, ok := httpapi.CurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login?next=/seller/publicar")
		return
	}
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil && !errors.Is(err, http.ErrNotMultipart) {
		h.renderPublishError(c, http.StatusRequestEntityTooLarge,
			"El formulario es demasiado grande. Revisa el tamaño de las fotografías.")
		return
	}

	form := readCarForm(c)
	photos, err := readUploads(c)
	if err != nil {
		h.renderPublishForm(c, http.StatusRequestEntityTooLarge, form, err)
		return
	}

	car, err := h.publishing.Create(c.Request.Context(), user.ID, service.CreateCarInput{
		Brand:       form.Brand,
		Model:       form.Model,
		Year:        parseYear(form.Year),
		Description: form.Description,
		City:        form.City,
		Price:       form.Price,
		ClientToken: form.ClientToken,
		Photos:      photos,
	})
	if err != nil {
		status := publishStatus(err)
		if status >= http.StatusInternalServerError {
			h.logger.ErrorContext(c.Request.Context(), "create_car_failed", "error", err)
		}
		h.renderPublishForm(c, status, form, err)
		return
	}
	c.Redirect(http.StatusFound, "/cars/"+car.ID)
}

func (h *PublishingHandler) EditForm(c *gin.Context) {
	user, ok := httpapi.CurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	car, err := h.publishing.Get(c.Request.Context(), user.ID, c.Param("id"))
	if err != nil {
		h.renderPublishError(c, publishStatus(err), publishErrorMessage(err))
		return
	}
	if !car.CanEdit(h.now()) {
		h.renderPublishError(c, http.StatusConflict,
			"El plazo de edición de diez minutos ya expiró. Puedes retirar la publicación, pero no editarla.")
		return
	}

	photos := make([]views.PhotoView, 0, len(car.Photos))
	for index, photo := range car.Photos {
		photos = append(photos, views.PhotoView{ID: photo.ID, Alt: car.Title() + " fotografía " + strconv.Itoa(index+1)})
	}
	page := views.PublishPage{
		BasePage: views.NewBasePage(httpapi.CSRFToken(c), &user, "Editar publicación", "sales"),
		Form: views.CarForm{
			Brand:       car.Brand,
			Model:       car.Model,
			City:        car.City,
			Year:        strconv.Itoa(car.Year),
			Price:       domain.FormatCOPAmount(car.PriceCents),
			Description: car.Description,
		},
		Errors:    map[string]string{},
		Photos:    photos,
		IsEdit:    true,
		CarID:     car.ID,
		MaxPhotos: domain.MaxCarPhotos,
		MaxBytes:  domain.MaxPhotoBytes,
	}
	renderHTML(c, h.renderer, h.logger, http.StatusOK, "publish.html", page)
}

func (h *PublishingHandler) Update(c *gin.Context) {
	user, ok := httpapi.CurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil && !errors.Is(err, http.ErrNotMultipart) {
		h.renderPublishError(c, http.StatusRequestEntityTooLarge,
			"El formulario es demasiado grande. Revisa el tamaño de las fotografías.")
		return
	}
	form := readCarForm(c)
	photos, err := readUploads(c)
	if err != nil {
		page := h.editPage(c, user, c.Param("id"), form, photos)
		page.ErrorMessage = uploadErrorMessage(err)
		renderHTML(c, h.renderer, h.logger, publishStatus(err), "publish.html", page)
		return
	}

	_, err = h.publishing.Update(c.Request.Context(), user.ID, c.Param("id"), service.UpdateCarInput{
		Brand:       form.Brand,
		Model:       form.Model,
		Year:        parseYear(form.Year),
		Description: form.Description,
		City:        form.City,
		Price:       form.Price,
		Photos:      photos,
	})
	if err != nil {
		status := publishStatus(err)
		if status >= http.StatusInternalServerError {
			h.logger.ErrorContext(c.Request.Context(), "update_car_failed", "error", err)
		}
		if errors.Is(err, domain.ErrEditWindowClosed) || errors.Is(err, domain.ErrForbidden) || errors.Is(err, domain.ErrConflict) {
			h.renderPublishError(c, status, publishErrorMessage(err))
			return
		}
		page := h.editPage(c, user, c.Param("id"), form, photos)
		page.ErrorMessage = publishErrorMessage(err)
		applyFieldErrors(&page, err)
		renderHTML(c, h.renderer, h.logger, status, "publish.html", page)
		return
	}
	c.Redirect(http.StatusFound, "/seller/ventas?editada=1")
}

func (h *PublishingHandler) Withdraw(c *gin.Context) {
	user, ok := httpapi.CurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	if err := h.publishing.Withdraw(c.Request.Context(), user.ID, c.Param("id")); err != nil {
		h.renderPublishError(c, publishStatus(err), publishErrorMessage(err))
		return
	}
	c.Redirect(http.StatusFound, "/seller/ventas?retirada=1")
}

func (h *PublishingHandler) Sales(c *gin.Context) {
	user, ok := httpapi.CurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	result := domain.SearchResult{Page: 1, PageSize: domain.DefaultPageSize}
	if h.catalog != nil {
		catalogResult, err := h.catalog.Search(c.Request.Context(), domain.SearchQuery{
			Q:        c.Query("q"),
			City:     c.Query("city"),
			Year:     c.Query("year"),
			MinPrice: c.Query("min"),
			MaxPrice: c.Query("max"),
			Page:     c.Query("page"),
		})
		if err != nil {
			h.logger.ErrorContext(c.Request.Context(), "sales_catalog_failed", "error", err)
		} else {
			result = catalogResult
		}
	}

	page := views.SalesPage{
		BasePage: views.NewBasePage(httpapi.CSRFToken(c), &user, "Ventas", "sales"),
		Result:   result,
		Query:    c.Query("q"),
		City:     c.Query("city"),
		Year:     c.Query("year"),
		MinPrice: c.Query("min"),
		MaxPrice: c.Query("max"),
	}
	if filters, filterErr := domain.NormalizeSearchFilters(domain.SearchQuery{
		Q: c.Query("q"), City: c.Query("city"), Year: c.Query("year"), MinPrice: c.Query("min"), MaxPrice: c.Query("max"), Page: c.Query("page"),
	}); filterErr == nil {
		page.Filters = filters
		page.PrevURL = salesPaginationURL(filters, result.Page-1)
		page.NextURL = salesPaginationURL(filters, result.Page+1)
	} else {
		page.ErrorMessage = "Revisa los filtros: el año o el rango de precio no son válidos."
	}
	for _, car := range result.Cars {
		page.Cards = append(page.Cards, views.NewCarCard(car, h.now()))
	}
	mine, err := h.publishing.ListMine(c.Request.Context(), user.ID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "list_my_cars_failed", "error", err)
	}
	for _, car := range mine {
		page.Mine = append(page.Mine, views.NewMyListing(car, h.now()))
	}
	if c.Query("retirada") == "1" {
		page.Flash = "La publicación fue retirada y ya no aparece en el catálogo."
	}
	if c.Query("editada") == "1" {
		page.Flash = "Los cambios se guardaron correctamente."
	}
	renderHTML(c, h.renderer, h.logger, http.StatusOK, "sales.html", page)
}

func salesPaginationURL(filters domain.SearchFilters, page int) string {
	if page < 1 {
		page = 1
	}
	values := []string{}
	if filters.Query != "" {
		values = append(values, "q="+url.QueryEscape(filters.Query))
	}
	if filters.Year > 0 {
		values = append(values, "year="+strconv.Itoa(filters.Year))
	}
	if filters.City != "" {
		values = append(values, "city="+url.QueryEscape(filters.City))
	}
	if filters.MinPriceCents > 0 {
		values = append(values, "min="+url.QueryEscape(domain.FormatCOPAmount(filters.MinPriceCents)))
	}
	if filters.MaxPriceCents > 0 {
		values = append(values, "max="+url.QueryEscape(domain.FormatCOPAmount(filters.MaxPriceCents)))
	}
	if page > 1 {
		values = append(values, "page="+strconv.Itoa(page))
	}
	if len(values) == 0 {
		return "/seller/ventas"
	}
	return "/seller/ventas?" + strings.Join(values, "&")
}

func (h *PublishingHandler) editPage(c *gin.Context, user domain.User, carID string, form views.CarForm, photos []service.RawPhoto) views.PublishPage {
	page := views.PublishPage{
		BasePage:  views.NewBasePage(httpapi.CSRFToken(c), &user, "Editar publicación", "sales"),
		Form:      form,
		Errors:    map[string]string{},
		IsEdit:    true,
		CarID:     carID,
		MaxPhotos: domain.MaxCarPhotos,
		MaxBytes:  domain.MaxPhotoBytes,
	}
	car, err := h.publishing.Get(c.Request.Context(), user.ID, carID)
	if err == nil {
		for index, photo := range car.Photos {
			page.Photos = append(page.Photos, views.PhotoView{ID: photo.ID, Alt: car.Title() + " fotografía " + strconv.Itoa(index+1)})
		}
	}
	return page
}

func (h *PublishingHandler) renderPublishForm(c *gin.Context, status int, form views.CarForm, err error) {
	page := views.PublishPage{
		BasePage:  views.NewBasePage(httpapi.CSRFToken(c), currentUserPtr(c), "Publicar vehículo", "publish"),
		Form:      form,
		Errors:    map[string]string{},
		MaxPhotos: domain.MaxCarPhotos,
		MaxBytes:  domain.MaxPhotoBytes,
	}
	page.ErrorMessage = publishErrorMessage(err)
	applyFieldErrors(&page, err)
	renderHTML(c, h.renderer, h.logger, status, "publish.html", page)
}

func (h *PublishingHandler) renderPublishError(c *gin.Context, status int, message string) {
	page := views.ErrorPage{
		BasePage: views.NewBasePage(httpapi.CSRFToken(c), currentUserPtr(c), "No se pudo completar", ""),
		Status:   status,
		Message:  message,
	}
	renderHTML(c, h.renderer, h.logger, status, "error.html", page)
}

func readCarForm(c *gin.Context) views.CarForm {
	return views.CarForm{
		Brand:       strings.TrimSpace(c.PostForm("brand")),
		Model:       strings.TrimSpace(c.PostForm("model")),
		City:        strings.TrimSpace(c.PostForm("city")),
		Year:        strings.TrimSpace(c.PostForm("year")),
		Price:       strings.TrimSpace(c.PostForm("price")),
		Description: strings.TrimSpace(c.PostForm("description")),
		ClientToken: strings.TrimSpace(c.PostForm("client_token")),
	}
}

func readUploads(c *gin.Context) ([]service.RawPhoto, error) {
	form, err := c.MultipartForm()
	if err != nil {
		if errors.Is(err, http.ErrNotMultipart) {
			return nil, nil
		}
		return nil, domain.ErrPhotoTooLarge
	}
	files := form.File["photos"]
	if len(files) == 0 {
		return nil, nil
	}
	if len(files) > domain.MaxCarPhotos {
		return nil, domain.ErrPhotosTooMany
	}

	uploads := make([]service.RawPhoto, 0, len(files))
	for _, header := range files {
		if header.Size > domain.MaxPhotoBytes {
			return nil, domain.ErrPhotoTooLarge
		}
		file, err := header.Open()
		if err != nil {
			return nil, domain.ErrPhotoDecodeFailed
		}
		data, readErr := io.ReadAll(io.LimitReader(file, domain.MaxPhotoBytes+1))
		_ = file.Close()
		if readErr != nil {
			return nil, domain.ErrPhotoDecodeFailed
		}
		if len(data) > domain.MaxPhotoBytes {
			return nil, domain.ErrPhotoTooLarge
		}
		uploads = append(uploads, service.RawPhoto{Name: header.Filename, Data: data})
	}
	return uploads, nil
}

func parseYear(raw string) int {
	year, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0
	}
	return year
}

func applyFieldErrors(page *views.PublishPage, err error) {
	switch {
	case errors.Is(err, domain.ErrBrandRequired), errors.Is(err, domain.ErrBrandTooLong):
		page.Errors["brand"] = "Indica una marca válida."
	case errors.Is(err, domain.ErrModelRequired), errors.Is(err, domain.ErrModelTooLong):
		page.Errors["model"] = "Indica un modelo válido."
	case errors.Is(err, domain.ErrCityRequired):
		page.Errors["city"] = "Indica una ciudad."
	case errors.Is(err, domain.ErrCityTooLong):
		page.Errors["city"] = "La ciudad no puede superar 80 caracteres."
	case errors.Is(err, domain.ErrYearInvalid):
		page.Errors["year"] = "El año debe estar entre 1900 y el próximo año."
	case errors.Is(err, domain.ErrPriceRequired), errors.Is(err, domain.ErrPriceInvalid):
		page.Errors["price"] = "Indica un precio entero en pesos colombianos, mayor que cero."
	case errors.Is(err, domain.ErrDescriptionLong):
		page.Errors["description"] = "La descripción es demasiado larga."
	case errors.Is(err, domain.ErrPhotoRequired):
		page.Errors["photos"] = "Adjunta al menos una fotografía."
	case errors.Is(err, domain.ErrPhotosTooMany):
		page.Errors["photos"] = "Puedes adjuntar hasta 6 fotografías."
	case errors.Is(err, domain.ErrPhotoTooLarge):
		page.Errors["photos"] = "Cada fotografía debe pesar como máximo 5 MB."
	case errors.Is(err, domain.ErrPhotoTypeInvalid), errors.Is(err, domain.ErrPhotoDecodeFailed):
		page.Errors["photos"] = "Usa fotografías JPEG o PNG válidas."
	case errors.Is(err, domain.ErrPhotoDimensions):
		page.Errors["photos"] = "Las dimensiones de la fotografía no son válidas."
	}
}

func publishErrorMessage(err error) string {
	switch {
	case errors.Is(err, domain.ErrForbidden):
		return "No tienes permisos sobre esta publicación."
	case errors.Is(err, domain.ErrEditWindowClosed):
		return "El plazo de edición de diez minutos ya expiró."
	case errors.Is(err, domain.ErrConflict):
		return "La publicación ya no está activa y no se puede editar."
	case errors.Is(err, domain.ErrNotFound):
		return "La publicación no existe."
	case errors.Is(err, domain.ErrBrandRequired), errors.Is(err, domain.ErrModelRequired):
		return "Completa la marca y el modelo."
	case errors.Is(err, domain.ErrCityRequired):
		return "Completa la ciudad."
	case errors.Is(err, domain.ErrCityTooLong):
		return "La ciudad no puede superar 80 caracteres."
	case errors.Is(err, domain.ErrYearInvalid):
		return "El año debe estar entre 1900 y el próximo año."
	case errors.Is(err, domain.ErrPriceRequired), errors.Is(err, domain.ErrPriceInvalid):
		return "El precio debe ser un valor entero en COP mayor que cero."
	case errors.Is(err, domain.ErrPhotoRequired):
		return "Adjunta al menos una fotografía."
	case errors.Is(err, domain.ErrPhotosTooMany):
		return "Puedes adjuntar hasta 6 fotografías."
	case errors.Is(err, domain.ErrPhotoTooLarge):
		return "Cada fotografía debe pesar como máximo 5 MB."
	case errors.Is(err, domain.ErrPhotoTypeInvalid), errors.Is(err, domain.ErrPhotoDecodeFailed):
		return "El archivo no es una imagen JPEG o PNG válida."
	case errors.Is(err, domain.ErrPhotoDimensions):
		return "Las dimensiones de la fotografía no son válidas."
	case errors.Is(err, domain.ErrDescriptionLong):
		return "La descripción es demasiado larga."
	default:
		return "No se pudo completar la operación."
	}
}

func uploadErrorMessage(err error) string { return publishErrorMessage(err) }

func publishStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrEditWindowClosed), errors.Is(err, domain.ErrConflict):
		return http.StatusConflict
	case errors.Is(err, domain.ErrPhotoTooLarge):
		return http.StatusRequestEntityTooLarge
	case errors.Is(err, domain.ErrBrandRequired), errors.Is(err, domain.ErrBrandTooLong),
		errors.Is(err, domain.ErrModelRequired), errors.Is(err, domain.ErrModelTooLong),
		errors.Is(err, domain.ErrCityRequired), errors.Is(err, domain.ErrCityTooLong),
		errors.Is(err, domain.ErrYearInvalid), errors.Is(err, domain.ErrPriceRequired),
		errors.Is(err, domain.ErrPriceInvalid), errors.Is(err, domain.ErrDescriptionLong),
		errors.Is(err, domain.ErrPhotoRequired), errors.Is(err, domain.ErrPhotosTooMany),
		errors.Is(err, domain.ErrPhotoTypeInvalid), errors.Is(err, domain.ErrPhotoDecodeFailed),
		errors.Is(err, domain.ErrPhotoDimensions):
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

func (h *PublishingHandler) now() time.Time { return time.Now() }

func newClientToken() string {
	buffer := make([]byte, 24)
	if _, err := rand.Read(buffer); err != nil {
		return ""
	}
	return hex.EncodeToString(buffer)
}
