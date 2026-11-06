package handlers

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"automarket/internal/domain"
	"automarket/internal/httpapi"
	"github.com/gin-gonic/gin"
)

type PhotoHandler struct {
	photos httpapi.PhotoService
	logger *slog.Logger
}

func NewPhotoHandler(photos httpapi.PhotoService, logger *slog.Logger) *PhotoHandler {
	return &PhotoHandler{photos: photos, logger: logger}
}

// Serve streams one photograph with a content hash validator.
func (h *PhotoHandler) Serve(c *gin.Context) {
	if h.photos == nil {
		c.Status(http.StatusNotFound)
		return
	}
	var viewer *domain.User
	if user, ok := httpapi.CurrentUser(c); ok {
		viewer = &user
	}
	photo, err := h.photos.GetPublic(c.Request.Context(), c.Param("id"), viewer)
	switch {
	case errors.Is(err, domain.ErrNotFound):
		c.Status(http.StatusNotFound)
		return
	case errors.Is(err, domain.ErrForbidden):
		c.Status(http.StatusForbidden)
		return
	case err != nil:
		h.logger.ErrorContext(c.Request.Context(), "photo_serve_failed", "error", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Header("Content-Type", photo.ContentType)
	c.Header("Content-Length", strconv.FormatInt(int64(len(photo.Data)), 10))
	etag := `"` + photo.SHA256 + `"`
	c.Header("ETag", etag)
	if photo.CarStatus.IsActive() {
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		c.Header("Cache-Control", "private, max-age=0, must-revalidate")
	}
	if etagMatches(c.GetHeader("If-None-Match"), etag) {
		c.Status(http.StatusNotModified)
		return
	}
	http.ServeContent(c.Writer, c.Request, photo.ID, photo.CreatedAt, bytes.NewReader(photo.Data))
}

func etagMatches(header, current string) bool {
	for _, candidate := range strings.Split(header, ",") {
		candidate = strings.TrimSpace(candidate)
		if candidate == "*" || candidate == current {
			return true
		}
		if strings.HasPrefix(candidate, "W/") && strings.TrimPrefix(candidate, "W/") == current {
			return true
		}
	}
	return false
}
