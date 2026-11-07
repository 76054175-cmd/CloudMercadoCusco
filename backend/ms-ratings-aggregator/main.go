// Servicio ms-ratings-aggregator (SIMULADO en la v1): Agregación de calificaciones.
//
// Fase 1: solo expone /health. La lógica se implementa en la fase indicada
// en arquitectura.md.
package main

import (
	"net/http"

	"github.com/76054175-cmd/CloudMercadoCusco/backend/shared/httpx"
)

func main() {
	mux := http.NewServeMux()
	httpx.Health(mux, "ms-ratings-aggregator")
	httpx.Serve("ms-ratings-aggregator", httpx.Env("ADDR", ":3004"), mux)
}
