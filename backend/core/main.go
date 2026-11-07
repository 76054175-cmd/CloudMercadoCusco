// pb-core-service: PocketBase v0.41.1 extendido en Go + web estática embebida.
package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	_ "time/tzdata" // America/Lima disponible aunque la imagen no tenga tzdata

	"ganigeorgiev.com/pocketbase/pocketbase"
	"ganigeorgiev.com/pocketbase/pocketbase/apis"
	"ganigeorgiev.com/pocketbase/pocketbase/core"
	"ganigeorgiev.com/pocketbase/pocketbase/plugins/migratecmd"
	"ganigeorgiev.com/pocketbase/pocketbase/tools/hook"
	"github.com/spf13/cobra"

	"github.com/76054175-cmd/CloudMercadoCusco/backend/core/hooks"
	_ "github.com/76054175-cmd/CloudMercadoCusco/backend/core/migrations"
	"github.com/76054175-cmd/CloudMercadoCusco/backend/core/seed"
	"github.com/76054175-cmd/CloudMercadoCusco/backend/core/web"
)

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	app := pocketbase.New()
	isDev := env("APP_ENV", "dev") == "dev"

	// `migrate create` genera migraciones en Go dentro de ./migrations.
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		TemplateLang: migratecmd.TemplateLangGo,
		Automigrate:  false, // el esquema se versiona a mano en Go
		Dir:          "migrations",
	})

	hooks.Register(app)
	registerSeed(app, isDev)
	registerStatic(app)

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

func seedOptions(isDev bool) seed.Options {
	return seed.Options{
		Dir:      env("QHATU_SEED_DIR", "../data/seed"),
		Dev:      isDev,
		Password: os.Getenv("SEED_DEV_PASSWORD"),
	}
}

// registerSeed agrega el comando `seed` y el auto-seed al iniciar (SEED_ON_START=true).
func registerSeed(app *pocketbase.PocketBase, isDev bool) {
	var update, dryRun bool
	cmd := &cobra.Command{
		Use:   "seed",
		Short:        "Importa data/seed (mercados, puestos, productos e imágenes)",
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := app.RunAllMigrations(); err != nil {
				return err
			}
			opt := seedOptions(isDev)
			opt.Update, opt.DryRun = update, dryRun
			rep, err := seed.Run(app, opt)
			if err != nil {
				return err
			}
			prefix := "seed"
			if dryRun {
				prefix = "seed (dry-run, sin guardar)"
			}
			log.Printf("%s OK: %s", prefix, rep)
			return nil
		},
	}
	cmd.Flags().BoolVar(&update, "update", false, "sobrescribe registros existentes con los valores del JSON")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "solo valida los archivos, no guarda nada")
	app.RootCmd.AddCommand(cmd)

	if env("SEED_ON_START", "false") != "true" {
		return
	}
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		rep, err := seed.Run(se.App, seedOptions(isDev))
		if err != nil {
			slog.Error("auto-seed falló", "error", err)
		} else {
			slog.Info("auto-seed", "resultado", rep.String())
		}
		return se.Next()
	})
}

// registerStatic sirve la SPA embebida con fallback a index.html.
func registerStatic(app *pocketbase.PocketBase) {
	if !web.HasBuild() {
		return
	}
	app.OnServe().Bind(&hook.Handler[*core.ServeEvent]{
		Func: func(se *core.ServeEvent) error {
			if !se.Router.HasRoute(http.MethodGet, "/{path...}") {
				se.Router.GET("/{path...}", apis.Static(web.FS(), true))
			}
			return se.Next()
		},
		Priority: 999,
	})
}
