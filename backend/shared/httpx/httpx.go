// Package httpx contiene utilidades HTTP comunes a los microservicios de QhatuDigital.
package httpx

import (
	"crypto/subtle"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"
)

// Env devuelve la variable de entorno key o fallback si está vacía.
func Env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// JSON escribe v como JSON con el status indicado.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("json encode", "error", err)
	}
}

// Health registra GET /health en mux.
func Health(mux *http.ServeMux, service string) {
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		JSON(w, http.StatusOK, map[string]any{"service": service, "status": "ok", "simulated": true})
	})
}

// RequireInternalSecret exige el header X-Internal-Secret igual a INTERNAL_SECRET.
// /health queda libre para los healthchecks de Docker.
func RequireInternalSecret(next http.Handler) http.Handler {
	secret := os.Getenv("INTERNAL_SECRET")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}
		got := r.Header.Get("X-Internal-Secret")
		if secret == "" || subtle.ConstantTimeCompare([]byte(got), []byte(secret)) != 1 {
			JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Serve arranca un servidor HTTP con timeouts razonables.
func Serve(service, addr string, h http.Handler) {
	srv := &http.Server{
		Addr:              addr,
		Handler:           RequireInternalSecret(h),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
	}
	slog.Info("servicio iniciado", "service", service, "addr", addr)
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("servidor detenido", "service", service, "error", err)
		os.Exit(1)
	}
}
