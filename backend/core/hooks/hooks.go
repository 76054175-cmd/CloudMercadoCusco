// Package hooks registra las reglas de negocio del core de QhatuDigital.
package hooks

import (
	"log/slog"
	"net/http"
	"os"

	"ganigeorgiev.com/pocketbase/pocketbase/core"

	"github.com/76054175-cmd/CloudMercadoCusco/backend/core/settings"
)

// Register engancha todos los hooks y rutas custom en app.
func Register(app core.App) {
	registerUsers(app)
	registerRoutes(app)
	registerGoogleOAuthFromEnv(app)
}

func registerUsers(app core.App) {
	// Valores por defecto (PocketBase no tiene defaults a nivel de campo).
	app.OnRecordCreate("users").BindFunc(func(e *core.RecordEvent) error {
		if e.Record.GetString("role") == "" {
			e.Record.Set("role", "customer")
		}
		if e.Record.GetFloat("trust_score") == 0 {
			e.Record.Set("trust_score", settings.Float(e.App, "trust_initial", 100))
		}
		return e.Next()
	})

	// Clientes: solo Google OAuth2. Personal (vendor/hub_operator/admin): lo crea un admin.
	app.OnRecordCreateRequest("users").BindFunc(func(e *core.RecordRequestEvent) error {
		if e.HasSuperuserAuth() {
			return e.Next()
		}
		if e.Auth != nil && e.Auth.GetString("role") == "admin" {
			return e.Next()
		}
		info, err := e.RequestInfo()
		if err != nil {
			return err
		}
		if info.Context == core.RequestInfoContextOAuth2 {
			return e.Next()
		}
		return e.ForbiddenError("El registro de clientes se realiza solo con Google.", nil)
	})
}

func registerRoutes(app core.App) {
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.GET("/api/qhatu/health", func(e *core.RequestEvent) error {
			return e.JSON(http.StatusOK, map[string]any{"status": "ok", "service": "pb-core-service"})
		})
		return se.Next()
	})
}

// registerGoogleOAuthFromEnv configura el proveedor Google si GOOGLE_CLIENT_ID y
// GOOGLE_CLIENT_SECRET están definidos. Si no, se configura a mano en /_/.
func registerGoogleOAuthFromEnv(app core.App) {
	id, secret := os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_CLIENT_SECRET")
	if id == "" || secret == "" {
		return
	}
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		col, err := se.App.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}
		col.OAuth2.Enabled = true
		found := false
		for i := range col.OAuth2.Providers {
			if col.OAuth2.Providers[i].Name == "google" {
				col.OAuth2.Providers[i].ClientId = id
				col.OAuth2.Providers[i].ClientSecret = secret
				found = true
			}
		}
		if !found {
			col.OAuth2.Providers = append(col.OAuth2.Providers, core.OAuth2ProviderConfig{
				Name: "google", ClientId: id, ClientSecret: secret,
			})
		}
		if err := se.App.Save(col); err != nil {
			slog.Error("no se pudo configurar Google OAuth2", "error", err)
		}
		return se.Next()
	})
}
