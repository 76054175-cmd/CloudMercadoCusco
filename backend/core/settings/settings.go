// Package settings lee los parámetros ajustables de la colección app_settings.
package settings

import (
	"encoding/json"

	"ganigeorgiev.com/pocketbase/pocketbase/core"
)

func raw(app core.App, key string) (json.RawMessage, bool) {
	rec, err := app.FindFirstRecordByData("app_settings", "key", key)
	if err != nil {
		return nil, false
	}
	b, err := json.Marshal(rec.Get("value"))
	if err != nil {
		return nil, false
	}
	return b, true
}

// Float devuelve el valor numérico de key, o fallback si no existe o no es número.
func Float(app core.App, key string, fallback float64) float64 {
	b, ok := raw(app, key)
	if !ok {
		return fallback
	}
	var v float64
	if err := json.Unmarshal(b, &v); err != nil {
		return fallback
	}
	return v
}

// Bool devuelve el valor booleano de key, o fallback.
func Bool(app core.App, key string, fallback bool) bool {
	b, ok := raw(app, key)
	if !ok {
		return fallback
	}
	var v bool
	if err := json.Unmarshal(b, &v); err != nil {
		return fallback
	}
	return v
}
