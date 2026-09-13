package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"devops-api/internal/config"
)

func TestHealthzHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	HealthzHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", res.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["status"] != "alive" {
		t.Errorf("expected status 'alive', got '%s'", body["status"])
	}
}

func TestReadyzHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()

	ReadyzHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", res.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["status"] != "ready" {
		t.Errorf("expected status 'ready', got '%s'", body["status"])
	}
}

func TestInfoHandler(t *testing.T) {
	cfg := &config.Config{
		Port:        "8080",
		Environment: "test",
		Version:     "0.1.0",
	}

	handler := InfoHandler(cfg)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/info", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", res.StatusCode)
	}
}

func TestPrometheusMetricsHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/metrics/prometheus", nil)
	w := httptest.NewRecorder()

	PrometheusMetricsHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", res.StatusCode)
	}

	contentType := res.Header.Get("Content-Type")
	if contentType == "" {
		t.Errorf("expected non-empty Content-Type")
	}
}

func TestChaosToggleAndReadyz(t *testing.T) {
	// Toggle to unready
	reqToggle := httptest.NewRequest(http.MethodPost, "/api/v1/chaos/toggle-ready", nil)
	wToggle := httptest.NewRecorder()
	ChaosToggleReadyHandler(wToggle, reqToggle)

	// Readyz should now return 503
	reqReady := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	wReady := httptest.NewRecorder()
	ReadyzHandler(wReady, reqReady)

	if wReady.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 Service Unavailable after chaos toggle, got %d", wReady.Code)
	}

	// Toggle back to ready
	wToggle2 := httptest.NewRecorder()
	ChaosToggleReadyHandler(wToggle2, reqToggle)

	// Readyz should return 200 OK
	wReady2 := httptest.NewRecorder()
	ReadyzHandler(wReady2, reqReady)

	if wReady2.Code != http.StatusOK {
		t.Errorf("expected 200 OK after restoring readiness, got %d", wReady2.Code)
	}
}
