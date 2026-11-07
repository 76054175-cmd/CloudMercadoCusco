// Package migrations contiene el esquema de QhatuDigital para PocketBase v0.41.1.
//
// Las migraciones se aplican automáticamente al iniciar `serve`.
// No edites una migración ya publicada: crea una nueva con un timestamp mayor.
package migrations

import (
	"ganigeorgiev.com/pocketbase/pocketbase/core"
	"ganigeorgiev.com/pocketbase/pocketbase/tools/types"
)

// Valores compartidos por varias colecciones (mantener en sync con apps/web/src/lib/constants.ts).
var (
	Sections   = []string{"verduras", "carnes_aves", "abarrotes", "quesos_lacteos", "frutas", "otros"}
	UnitTypes  = []string{"kg", "medio_kg", "cuarto_kg", "atado", "manojo", "unidad", "porcion"}
	Roles      = []string{"customer", "vendor", "hub_operator", "admin"}
	imageMimes = []string{"image/jpeg", "image/png", "image/webp", "image/svg+xml"}
)

// Fragmentos de reglas de API.
const (
	ruleAdmin  = `@request.auth.role = "admin"`
	ruleAuthed = `@request.auth.id != ""`
)

func rule(s string) *string { return types.Pointer(s) }

func ptr(v float64) *float64 { return &v }

func addTimestamps(c *core.Collection) {
	c.Fields.Add(&core.AutodateField{Name: "created", OnCreate: true})
	c.Fields.Add(&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true})
}

func text(name string, required bool) *core.TextField {
	return &core.TextField{Name: name, Required: required}
}

func hiddenText(name string) *core.TextField {
	return &core.TextField{Name: name, Hidden: true}
}

func sel(name string, required bool, values ...string) *core.SelectField {
	return &core.SelectField{Name: name, Required: required, MaxSelect: 1, Values: values}
}

// money es un monto en soles >= 0 (no requerido para permitir 0).
func money(name string) *core.NumberField {
	return &core.NumberField{Name: name, Min: ptr(0)}
}

func integer(name string, min *float64) *core.NumberField {
	return &core.NumberField{Name: name, Min: min, OnlyInt: true}
}

func rel(name string, target *core.Collection, required, cascade bool) *core.RelationField {
	return &core.RelationField{
		Name:          name,
		CollectionId:  target.Id,
		MaxSelect:     1,
		Required:      required,
		CascadeDelete: cascade,
	}
}

func image(name string) *core.FileField {
	return &core.FileField{
		Name:      name,
		MaxSelect: 1,
		MaxSize:   5 << 20, // 5 MB
		MimeTypes: imageMimes,
		Thumbs:    []string{"400x400", "120x120"},
	}
}

func boolean(name string) *core.BoolField { return &core.BoolField{Name: name} }

func date(name string) *core.DateField { return &core.DateField{Name: name} }

func jsonField(name string) *core.JSONField { return &core.JSONField{Name: name, MaxSize: 1 << 20} }
