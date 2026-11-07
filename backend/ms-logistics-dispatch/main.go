// Servicio ms-logistics-dispatch (SIMULADO en la v1): Despacho: flota propia y Rappi.
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
	httpx.Health(mux, "ms-logistics-dispatch")
	httpx.Serve("ms-logistics-dispatch", httpx.Env("ADDR", ":3003"), mux)
}
