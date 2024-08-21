package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemStorage_UpdateGauge(t *testing.T) {
	store := NewMemStorage()

	store.UpdateGauge("Alloc", 12345.67)

	assert.Equal(t, 12345.67, store.Gauges["Alloc"])
}

func TestMemStorage_UpdateCounter(t *testing.T) {
	store := NewMemStorage()

	store.UpdateCounter("PollCount", 1)
	store.UpdateCounter("PollCount", 2)

	assert.Equal(t, int64(3), store.Counters["PollCount"])
}

func TestHTTPHandler_UpdateGauge(t *testing.T) {
	store := NewMemStorage()
	handler := http.HandlerFunc(store.HTTPHandler)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/Alloc/12345.67", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, 12345.67, store.Gauges["Alloc"])
}

func TestHTTPHandler_UpdateCounter(t *testing.T) {
	store := NewMemStorage()
	handler := http.HandlerFunc(store.HTTPHandler)

	req := httptest.NewRequest(http.MethodPost, "/update/counter/PollCount/1", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	req = httptest.NewRequest(http.MethodPost, "/update/counter/PollCount/2", nil)
	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, int64(3), store.Counters["PollCount"])
}

func TestHTTPHandler_InvalidMetricType(t *testing.T) {
	store := NewMemStorage()
	handler := http.HandlerFunc(store.HTTPHandler)

	req := httptest.NewRequest(http.MethodPost, "/update/invalid/Alloc/12345.67", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHTTPHandler_MissingMetricName(t *testing.T) {
	store := NewMemStorage()
	handler := http.HandlerFunc(store.HTTPHandler)

	req := httptest.NewRequest(http.MethodPost, "/update/gauge//12345.67", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}
