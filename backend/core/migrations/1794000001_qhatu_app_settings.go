package migrations

import (
	"ganigeorgiev.com/pocketbase/pocketbase/core"
	m "ganigeorgiev.com/pocketbase/pocketbase/migrations"
)

// DefaultSettings son los valores iniciales de app_settings (arquitectura.md §4.6).
// Después de la primera migración se ajustan manualmente desde el panel /_/.
var DefaultSettings = []struct {
	Key         string
	Value       any
	Description string
}{
	{"trust_initial", 100, "Trust score de una cuenta nueva"},
	{"trust_max", 150, "Trust score máximo"},
	{"trust_completed_delta", 2, "Cambio por orden completada"},
	{"trust_no_show_pay_at_market_delta", -30, "Cambio por no recoger una orden 'Pago en mercado'"},
	{"trust_no_show_paid_delta", -10, "Cambio por no recoger una orden ya pagada"},
	{"trust_late_cancel_delta", -5, "Cambio por cancelación tardía"},
	{"trust_late_cancel_after_min", 10, "Minutos desde la creación a partir de los cuales una cancelación es tardía"},
	{"trust_block_below", 40, "Bajo este score la cuenta se bloquea"},
	{"trust_recovery_delta_per_30d", 5, "Recuperación cada 30 días sin incidentes (hasta trust_initial)"},
	{"pay_at_market_web_enabled", true, "Mostrar 'Pago en mercado' en la versión web"},
	{"pay_at_market_min_score", 80, "Trust score mínimo para 'Pago en mercado'"},
	{"pay_at_market_max_open_orders", 1, "Órdenes 'Pago en mercado' abiertas permitidas a la vez"},
	{"pay_at_market_max_amount", 150, "Monto máximo (S/) por orden 'Pago en mercado'"},
	{"max_users_per_device", 3, "Cuentas máximas asociadas a un mismo dispositivo"},
	{"blacklist_no_shows", 2, "No-shows que envían el dispositivo a la lista negra"},
	{"blacklist_window_days", 30, "Ventana (días) para contar no-shows"},
	{"otp_length", 6, "Dígitos del código OTP"},
	{"otp_ttl_sec", 300, "Vigencia del OTP en segundos"},
	{"otp_max_attempts", 3, "Intentos máximos del OTP"},
	{"otp_resend_sec", 60, "Segundos antes de poder reenviar el OTP"},
	{"vendor_accept_timeout_min", 20, "Minutos para que el vendedor acepte antes del rechazo automático"},
	{"pickup_window_hours", 3, "Horas para recoger desde ready_at_main_gate"},
	{"min_order_amount", 10, "Monto mínimo (S/) por orden de mercado, sin envío"},
	{"rappi_flat_fee_sim", 8, "Tarifa simulada (S/) de Rappi"},
	{"rating_window_days", 7, "Días para calificar después de completada la orden"},
	{"rating_prior_votes", 5, "Votos 'virtuales' del promedio bayesiano"},
	{"rating_prior_mean", 2.0, "Media 'virtual' del promedio bayesiano"},
}

func init() {
	m.Register(func(app core.App) error {
		col, err := app.FindCollectionByNameOrId("app_settings")
		if err != nil {
			return err
		}
		for _, s := range DefaultSettings {
			if _, err := app.FindFirstRecordByData(col, "key", s.Key); err == nil {
				continue // ya existe (p. ej. ajustado a mano): no se pisa
			}
			r := core.NewRecord(col)
			r.Set("key", s.Key)
			r.Set("value", s.Value)
			r.Set("description", s.Description)
			if err := app.Save(r); err != nil {
				return err
			}
		}
		return nil
	}, func(app core.App) error {
		_, err := app.DB().NewQuery("DELETE FROM app_settings").Execute()
		return err
	})
}
