package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/galanjabal3/linkas/internal/model"
	"github.com/galanjabal3/linkas/internal/service"
)

type URLHandler struct {
	service *service.URLService
	baseURL string
}

func NewURLHandler(service *service.URLService, baseURL string) *URLHandler {
	return &URLHandler{
		service: service,
		baseURL: baseURL,
	}
}

func handleServiceError(c *gin.Context, err error, context string) {
	if errors.Is(err, service.ErrURLExpired) {
		c.JSON(http.StatusGone, model.ErrorResponse{
			Error:   "url_expired",
			Message: "This URL has expired",
		})
		return
	}
	if errors.Is(err, service.ErrURLNotFound) {
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Error:   "not_found",
			Message: "Short URL not found",
		})
		return
	}
	c.JSON(http.StatusInternalServerError, model.ErrorResponse{
		Error:   context + "_failed",
		Message: err.Error(),
	})
}

// Shorten godoc
// @Summary      Shorten URL
// @Description  Create a new short URL from a long URL
// @Tags         urls
// @Accept       json
// @Produce      json
// @Param        request body model.ShortenRequest true "URL to shorten"
// @Success      201 {object} model.ShortenResponse
// @Failure      400 {object} model.ErrorResponse
// @Failure      409 {object} model.ErrorResponse
// @Router       /api/shorten [post]
func (h *URLHandler) Shorten(c *gin.Context) {
	var req model.ShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	url, err := h.service.Shorten(req.URL, req.CustomSlug, req.ExpiresIn)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrInvalidURL) || errors.Is(err, service.ErrInvalidSlug) {
			status = http.StatusBadRequest
		} else if errors.Is(err, service.ErrSlugExists) {
			status = http.StatusConflict
		}
		c.JSON(status, model.ErrorResponse{
			Error:   "shorten_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, model.ShortenResponse{
		Slug:        url.Slug,
		ShortURL:    h.baseURL + "/" + url.Slug,
		OriginalURL: url.OriginalURL,
		ExpiresAt:   url.ExpiresAt,
	})
}

// RedirectBySlug handles redirect for path-based slugs (e.g., "video/01")
func (h *URLHandler) RedirectBySlug(c *gin.Context, slug string) {
	originalURL, err := h.service.Redirect(slug)
	if err != nil {
		handleServiceError(c, err, "redirect")
		return
	}

	_ = h.service.RecordClick(
		slug,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
		c.GetHeader("Referer"),
	)

	c.Redirect(http.StatusMovedPermanently, originalURL)
}

// GetStats godoc
// @Summary      Get URL statistics
// @Description  Get click statistics for a short URL
// @Tags         urls
// @Produce      json
// @Param        slug path string true "Short URL slug"
// @Success      200 {object} model.StatsResponse
// @Failure      404 {object} model.ErrorResponse
// @Router       /api/stats/{slug} [get]
func (h *URLHandler) GetStats(c *gin.Context) {
	slug := strings.TrimPrefix(c.Param("slug"), "/")
	stats, err := h.service.GetStats(slug)
	if err != nil {
		handleServiceError(c, err, "stats")
		return
	}

	c.JSON(http.StatusOK, stats)
}

// Delete godoc
// @Summary      Delete short URL
// @Description  Delete a short URL by slug
// @Tags         urls
// @Produce      json
// @Param        slug path string true "Short URL slug"
// @Success      200 {object} map[string]string
// @Failure      404 {object} model.ErrorResponse
// @Router       /api/{slug} [delete]
func (h *URLHandler) Delete(c *gin.Context) {
	slug := strings.TrimPrefix(c.Param("slug"), "/")
	if err := h.service.Delete(slug); err != nil {
		handleServiceError(c, err, "delete")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "URL deleted successfully",
	})
}

// GetQRCode godoc
// @Summary      Get QR Code
// @Description  Generate QR code image for a short URL
// @Tags         urls
// @Produce      png
// @Param        slug path string true "Short URL slug"
// @Success      200 {file} binary
// @Failure      404 {object} model.ErrorResponse
// @Router       /api/qrcode/{slug} [get]
func (h *URLHandler) GetQRCode(c *gin.Context) {
	slug := strings.TrimPrefix(c.Param("slug"), "/")
	qr, err := h.service.GenerateQRCode(slug, h.baseURL)
	if err != nil {
		handleServiceError(c, err, "qrcode")
		return
	}

	c.Data(http.StatusOK, "image/png", qr)
}

// Health godoc
// @Summary      Health check
// @Description  Check if the API is running
// @Tags         system
// @Produce      json
// @Success      200 {object} map[string]string
// @Router       /api/health [get]
func (h *URLHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}
