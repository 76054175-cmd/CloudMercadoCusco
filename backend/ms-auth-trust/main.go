// Servicio ms-auth-trust (SIMULADO en la v1): Antifraude: dispositivo, OTP SMS y trust score.
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
	httpx.Health(mux, "ms-auth-trust")
	httpx.Serve("ms-auth-trust", httpx.Env("ADDR", ":3001"), mux)
}
