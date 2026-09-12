package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"devops-api/internal/config"
)

var (
	startTime      = time.Now()
	totalRequests  uint64
	activeRequests int64
)

// ResponseWriter wrapper to capture HTTP status code for logging and metrics
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware logs each HTTP request in structured JSON format
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		atomic.AddUint64(&totalRequests, 1)
		atomic.AddInt64(&activeRequests, 1)
		defer atomic.AddInt64(&activeRequests, -1)

		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		// Log structured JSON to stdout (Cloud Logging captures stdout automatically)
		logEntry := map[string]any{
			"timestamp":     time.Now().UTC().Format(time.RFC3339),
			"severity":      "INFO",
			"method":        r.Method,
			"path":          r.URL.Path,
			"status":        wrapped.statusCode,
			"duration_ms":   duration.Milliseconds(),
			"remote_ip":     getClientIP(r),
			"user_agent":    r.UserAgent(),
		}
		json.NewEncoder(os.Stdout).Encode(logEntry)
	})
}

// HealthzHandler responds with 200 OK for Kubernetes/Cloud Run liveness probes
func HealthzHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "alive",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

// ReadyzHandler responds with 200 OK when ready to accept traffic
func ReadyzHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ready",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

// InfoHandler returns runtime and infrastructure details
func InfoHandler(cfg *config.Config) http.HandlerFunc {
	hostname, _ := os.Hostname()

	return func(w http.ResponseWriter, r *http.Request) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		info := map[string]any{
			"app": map[string]string{
				"name":        "devops-prototype-api",
				"version":     cfg.Version,
				"environment": cfg.Environment,
				"uptime":      time.Since(startTime).String(),
			},
			"system": map[string]any{
				"hostname":        hostname, // Container ID in Cloud Run / Docker
				"os":              runtime.GOOS,
				"arch":            runtime.GOARCH,
				"go_version":      runtime.Version(),
				"num_cpu":         runtime.NumCPU(),
				"num_goroutine":   runtime.NumGoroutine(),
				"memory_alloc_mb": float64(m.Alloc) / 1024 / 1024,
			},
		}

		writeJSON(w, http.StatusOK, info)
	}
}

// NetworkHandler demonstrates how cloud proxies, load balancers, and Ingress route requests
func NetworkHandler(w http.ResponseWriter, r *http.Request) {
	networkData := map[string]any{
		"client_ip":        getClientIP(r),
		"remote_addr":      r.RemoteAddr,
		"host":             r.Host,
		"proto":            r.Proto,
		"headers": map[string]string{
			"X-Forwarded-For":   r.Header.Get("X-Forwarded-For"),
			"X-Forwarded-Proto": r.Header.Get("X-Forwarded-Proto"),
			"X-Cloud-Trace-Context": r.Header.Get("X-Cloud-Trace-Context"), // GCP Trace Header
			"User-Agent":        r.Header.Get("User-Agent"),
		},
	}

	writeJSON(w, http.StatusOK, networkData)
}

// MetricsHandler provides basic telemetry counters
func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics := map[string]any{
		"total_requests":   atomic.LoadUint64(&totalRequests),
		"active_requests":  atomic.LoadInt64(&activeRequests),
		"uptime_seconds":   time.Since(startTime).Seconds(),
		"memory_alloc_kb":  m.Alloc / 1024,
		"memory_sys_kb":    m.Sys / 1024,
		"num_gc_runs":      m.NumGC,
	}

	writeJSON(w, http.StatusOK, metrics)
}

// RootHandler gives an overview of available routes
func RootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message": "DevOps Cloud Prototype API running successfully!",
		"docs": map[string]string{
			"GET /":               "API Root & docs overview",
			"GET /healthz":        "Liveness check for container orchestration",
			"GET /readyz":         "Readiness check before routing traffic",
			"GET /api/v1/info":    "Runtime, container hostname, and system stats",
			"GET /api/v1/network": "Client IP, headers, and reverse proxy details",
			"GET /metrics":        "Application and memory metrics",
		},
	})
}

func getClientIP(r *http.Request) string {
	// Behind Cloud Run, Cloud Load Balancing, or Nginx
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return xrip
	}
	return r.RemoteAddr
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
