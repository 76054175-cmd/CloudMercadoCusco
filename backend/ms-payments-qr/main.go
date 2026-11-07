// Servicio ms-payments-qr (SIMULADO en la v1): Pagos con billetera, reembolsos y firma de QR.
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
	httpx.Health(mux, "ms-payments-qr")
	httpx.Serve("ms-payments-qr", httpx.Env("ADDR", ":3002"), mux)
}
