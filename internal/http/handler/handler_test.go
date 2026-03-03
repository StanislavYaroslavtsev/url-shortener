package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/StanislavYaroslavtsev/url-shortener/internal/domain"
	"github.com/StanislavYaroslavtsev/url-shortener/internal/http/handler/mocks"
	"github.com/StanislavYaroslavtsev/url-shortener/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTestHandler(t *testing.T) (*Handler, *mocks.MockLinkServiceInterface) {
	svc := mocks.NewMockLinkServiceInterface(t)
	h := NewHandler(svc, "http://localhost:3000")
	return h, svc
}

func newRequest(t *testing.T, method, url string, body any) *http.Request {
	t.Helper()
	data, err := json.Marshal(body)
	require.NoError(t, err)
	return httptest.NewRequest(method, url, bytes.NewReader(data))
}

func TestHandler_ShortenURL_ReturnsShortURL(t *testing.T) {
	h, svc := newTestHandler(t)

	link, err := domain.NewLink("https://google.com", "abc123", nil)
	require.NoError(t, err)

	svc.EXPECT().Create(mock.Anything, "https://google.com", mock.Anything).Return(link, nil)

	req := newRequest(t, http.MethodPost, "/shorten", map[string]string{
		"url": "https://google.com",
	})
	rec := httptest.NewRecorder()

	h.ShortenURL(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:3000/abc123", resp["short_url"])
}

func TestHandler_ShortenURL_InvalidJSON_ReturnsBadRequest(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewBufferString("not json"))
	rec := httptest.NewRecorder()

	h.ShortenURL(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_ShortenURL_EmptyURL_ReturnsBadRequest(t *testing.T) {
	h, _ := newTestHandler(t)

	req := newRequest(t, http.MethodPost, "/shorten", map[string]string{
		"url": "",
	})
	rec := httptest.NewRecorder()

	h.ShortenURL(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_ShortenURL_ServiceError_ReturnsInternalServerError(t *testing.T) {
	h, svc := newTestHandler(t)

	svc.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).Return(nil, assert.AnError)

	req := newRequest(t, http.MethodPost, "/shorten", map[string]string{
		"url": "https://google.com",
	})
	rec := httptest.NewRecorder()

	h.ShortenURL(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestHandler_RedirectURL_Redirects(t *testing.T) {
	h, svc := newTestHandler(t)

	link, err := domain.NewLink("https://google.com", "abc123", nil)
	require.NoError(t, err)

	svc.EXPECT().Get(mock.Anything, "abc123").Return(link, nil)

	router := chi.NewRouter()
	router.Get("/{id}", h.RedirectURL)

	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "https://google.com", rec.Header().Get("Location"))
}

func TestHandler_RedirectURL_NotFound_Returns404(t *testing.T) {
	h, svc := newTestHandler(t)

	svc.EXPECT().Get(mock.Anything, "abc123").Return(nil, repository.ErrLinkNotFound)

	router := chi.NewRouter()
	router.Get("/{id}", h.RedirectURL)

	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandler_RedirectURL_ExpiredLink_Returns410(t *testing.T) {
	h, svc := newTestHandler(t)

	svc.EXPECT().Get(mock.Anything, "abc123").Return(nil, domain.ErrLinkExpired)

	router := chi.NewRouter()
	router.Get("/{id}", h.RedirectURL)

	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusGone, rec.Code)
}

func TestHandler_Health_AllDepsOk_Returns200(t *testing.T) {
	h, svc := newTestHandler(t)

	svc.EXPECT().Ping(mock.Anything).Return(map[string]error{
		"database": nil,
		"cache":    nil,
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h.Health(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "ok", resp["status"])
}

func TestHandler_Health_DBUnavailable_Returns503(t *testing.T) {
	h, svc := newTestHandler(t)

	svc.EXPECT().Ping(mock.Anything).Return(map[string]error{
		"database": assert.AnError,
		"cache":    nil,
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h.Health(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var resp map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "unavailable", resp["status"])
}

func TestHandler_Health_CacheUnavailable_Returns503(t *testing.T) {
	h, svc := newTestHandler(t)

	svc.EXPECT().Ping(mock.Anything).Return(map[string]error{
		"database": nil,
		"cache":    assert.AnError,
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h.Health(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var resp map[string]any
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "unavailable", resp["status"])
}
