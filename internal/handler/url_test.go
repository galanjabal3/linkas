package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/galanjabal3/linkas/internal/model"
)

func setupTestRouter(h *URLHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(gin.Recovery())

	api := router.Group("/api")
	{
		api.POST("/shorten", h.Shorten)
		api.GET("/stats/:slug", h.GetStats)
		api.DELETE("/:slug", h.Delete)
		api.GET("/health", h.Health)
	}

	return router
}

func TestHealthEndpoint(t *testing.T) {
	h := &URLHandler{baseURL: "http://localhost:8080"}
	router := setupTestRouter(h)

	req, _ := http.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != "healthy" {
		t.Errorf("expected status 'healthy', got %s", resp["status"])
	}
}

func TestShortenInvalidJSON(t *testing.T) {
	h := &URLHandler{baseURL: "http://localhost:8080"}
	router := setupTestRouter(h)

	req, _ := http.NewRequest("POST", "/api/shorten", bytes.NewBufferString("invalid"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestShortenMissingURL(t *testing.T) {
	h := &URLHandler{baseURL: "http://localhost:8080"}
	router := setupTestRouter(h)

	body, _ := json.Marshal(map[string]string{})
	req, _ := http.NewRequest("POST", "/api/shorten", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestShortenInvalidURLFormat(t *testing.T) {
	h := &URLHandler{baseURL: "http://localhost:8080"}
	router := setupTestRouter(h)

	body, _ := json.Marshal(model.ShortenRequest{URL: "not-a-url"})
	req, _ := http.NewRequest("POST", "/api/shorten", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestShortenEmptyBody(t *testing.T) {
	h := &URLHandler{baseURL: "http://localhost:8080"}
	router := setupTestRouter(h)

	req, _ := http.NewRequest("POST", "/api/shorten", bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestGetStatsInvalidSlug(t *testing.T) {
	h := &URLHandler{baseURL: "http://localhost:8080"}
	router := setupTestRouter(h)

	req, _ := http.NewRequest("GET", "/api/stats/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Without database, this will either 404 or 500
	if w.Code != http.StatusNotFound && w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 404 or 500, got %d", w.Code)
	}
}

func TestDeleteInvalidSlug(t *testing.T) {
	h := &URLHandler{baseURL: "http://localhost:8080"}
	router := setupTestRouter(h)

	req, _ := http.NewRequest("DELETE", "/api/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Without database, this will either 404 or 500
	if w.Code != http.StatusNotFound && w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 404 or 500, got %d", w.Code)
	}
}

func TestNewURLHandler(t *testing.T) {
	h := &URLHandler{baseURL: "http://example.com"}

	if h.baseURL != "http://example.com" {
		t.Errorf("expected baseURL 'http://example.com', got %s", h.baseURL)
	}
}

func TestGetQRCodeInvalidSlug(t *testing.T) {
	h := &URLHandler{baseURL: "http://localhost:8080"}
	router := setupTestRouter(h)

	req, _ := http.NewRequest("GET", "/api/qrcode/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound && w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 404 or 500, got %d", w.Code)
	}
}
