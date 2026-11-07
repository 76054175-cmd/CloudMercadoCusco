package migrations

import (
	"strings"

	"ganigeorgiev.com/pocketbase/pocketbase/core"
	m "ganigeorgiev.com/pocketbase/pocketbase/migrations"
)

// Campos de users que el propio usuario no puede asignarse ni modificar.
var protectedUserFields = []string{
	"role", "market_id", "sms_verified", "device_hw_id", "trust_score", "no_show_count", "is_blocked",
}

func notSet(fields ...string) string {
	parts := make([]string, len(fields))
	for i, f := range fields {
		parts[i] = "@request.body." + f + ":isset = false"
	}
	return strings.Join(parts, " && ")
}

func init() {
	m.Register(func(app core.App) error {
		// ---------------------------------------------------------- markets
		markets := core.NewBaseCollection("markets")
		markets.Fields.Add(
			text("name", true),
			&core.TextField{Name: "slug", Required: true, Pattern: `^[a-z0-9-]+$`},
			&core.TextField{Name: "code_prefix", Required: true, Pattern: `^[A-Z]{3}$`},
			image("bubble_icon"),
			boolean("active"),
			text("hub_location_desc", false),
			&core.TextField{Name: "open_time", Pattern: `^([01]\d|2[0-3]):[0-5]\d$`},
			&core.TextField{Name: "close_time", Pattern: `^([01]\d|2[0-3]):[0-5]\d$`},
			boolean("has_own_delivery"),
			money("delivery_fee"),
			boolean("rappi_enabled"),
			hiddenText("payout_account_desc"),
		)
		addTimestamps(markets)
		markets.AddIndex("idx_markets_slug", true, "slug", "")
		markets.AddIndex("idx_markets_code_prefix", true, "code_prefix", "")
		markets.ListRule = rule(`active = true || ` + ruleAdmin)
		markets.ViewRule = markets.ListRule
		markets.CreateRule = rule(ruleAdmin)
		markets.UpdateRule = rule(ruleAdmin)
		markets.DeleteRule = rule(ruleAdmin)
		if err := app.Save(markets); err != nil {
			return err
		}

		// ------------------------------------------------------------ users
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}
		users.Fields.Add(
			sel("role", false, Roles...),
			rel("market_id", markets, false, false),
			text("phone", false),
			boolean("sms_verified"),
			hiddenText("device_hw_id"),
			&core.NumberField{Name: "trust_score", Min: ptr(0), Max: ptr(150)},
			integer("no_show_count", ptr(0)),
			boolean("is_blocked"),
		)
		users.AddIndex("idx_users_role_market", false, "role, market_id", "")
		guard := notSet(protectedUserFields...)
		users.ListRule = rule(`id = @request.auth.id || ` + ruleAdmin)
		users.ViewRule = users.ListRule
		// El registro de clientes solo ocurre vía Google OAuth2 (ver hooks/users.go),
		// pero OAuth2 pasa por esta regla, por eso debe permitir el alta sin campos protegidos.
		users.CreateRule = rule(ruleAdmin + ` || (` + guard + `)`)
		users.UpdateRule = rule(ruleAdmin + ` || (id = @request.auth.id && ` + guard + `)`)
		users.DeleteRule = rule(ruleAdmin)
		users.OAuth2.MappedFields.Name = "name"
		users.OAuth2.MappedFields.AvatarURL = "avatar"
		if err := app.Save(users); err != nil {
			return err
		}

		// ----------------------------------------------------------- stalls
		stalls := core.NewBaseCollection("stalls")
		stalls.Fields.Add(
			rel("market_id", markets, true, false),
			rel("owner_user_id", users, false, false),
			text("name", true),
			text("stall_number", true),
			sel("section", true, Sections...),
			boolean("active"),
			date("enabled_until"),
			jsonField("metrics_summary"),
		)
		addTimestamps(stalls)
		stalls.AddIndex("idx_stalls_market_number", true, "market_id, stall_number", "")
		stalls.AddIndex("idx_stalls_owner", false, "owner_user_id", "")
		stalls.ListRule = rule(`active = true || owner_user_id = @request.auth.id || ` + ruleAdmin)
		stalls.ViewRule = stalls.ListRule
		stalls.CreateRule = rule(ruleAdmin)
		stalls.UpdateRule = rule(ruleAdmin + ` || (owner_user_id = @request.auth.id && ` +
			notSet("active", "enabled_until", "metrics_summary", "owner_user_id", "market_id") + `)`)
		stalls.DeleteRule = rule(ruleAdmin)
		if err := app.Save(stalls); err != nil {
			return err
		}

		// ---------------------------------------------------- catalog_items
		catalog := core.NewBaseCollection("catalog_items")
		catalog.Fields.Add(
			text("name", true),
			sel("section", true, Sections...),
			image("default_image"),
		)
		addTimestamps(catalog)
		catalog.AddIndex("idx_catalog_items_name", true, "name COLLATE NOCASE", "")
		catalog.ListRule = rule("")
		catalog.ViewRule = rule("")
		catalog.CreateRule = rule(ruleAdmin)
		catalog.UpdateRule = rule(ruleAdmin)
		catalog.DeleteRule = rule(ruleAdmin)
		if err := app.Save(catalog); err != nil {
			return err
		}

		// --------------------------------------------------------- products
		products := core.NewBaseCollection("products")
		products.Fields.Add(
			rel("stall_id", stalls, true, true),
			rel("catalog_item_id", catalog, false, false),
			text("name", true),
			image("image"),
			&core.NumberField{Name: "price", Required: true, Min: ptr(0.01)},
			boolean("in_stock"),
			sel("unit_type", true, UnitTypes...),
		)
		addTimestamps(products)
		products.AddIndex("idx_products_stall", false, "stall_id", "")
		products.AddIndex("idx_products_catalog_item", false, "catalog_item_id", "")
		ownsActiveStall := `(@request.auth.role = "vendor" && stall_id.owner_user_id = @request.auth.id && stall_id.active = true)`
		products.ListRule = rule(`stall_id.active = true || stall_id.owner_user_id = @request.auth.id || ` + ruleAdmin)
		products.ViewRule = products.ListRule
		products.CreateRule = rule(ruleAdmin + ` || ` + ownsActiveStall)
		products.UpdateRule = rule(ruleAdmin + ` || (` + ownsActiveStall +
			` && (@request.body.stall_id:isset = false || @request.body.stall_id = stall_id))`)
		products.DeleteRule = rule(ruleAdmin + ` || ` + ownsActiveStall)
		if err := app.Save(products); err != nil {
			return err
		}

		// ----------------------------------------------------------- orders
		orders := core.NewBaseCollection("orders")
		orders.Fields.Add(
			text("checkout_group_id", true),
			rel("customer_id", users, true, false),
			rel("market_id", markets, true, false),
			&core.TextField{Name: "business_date", Pattern: `^\d{4}-\d{2}-\d{2}$`},
			text("customer_short_id", false),
			sel("delivery_type", true, "pickup_main_gate", "market_delivery", "rappi_delivery"),
			text("delivery_address", false),
			text("delivery_reference", false),
			&core.NumberField{Name: "delivery_lat"},
			&core.NumberField{Name: "delivery_lng"},
			text("contact_phone", false),
			sel("payment_method", true, "digital_wallet", "pay_at_market"),
			sel("payment_status", false, "pending", "paid", "verified_at_gate", "partially_refunded", "refunded", "failed_fraud"),
			hiddenText("qr_token_hash"),
			sel("status", false, "created", "preparing_in_stalls", "ready_at_main_gate", "dispatched", "completed", "cancelled", "no_show"),
			money("items_subtotal"),
			money("delivery_fee"),
			money("total_amount"),
			money("refund_amount"),
			date("pickup_deadline"),
			hiddenText("device_id_snapshot"),
		)
		addTimestamps(orders)
		orders.AddIndex("idx_orders_customer", false, "customer_id", "")
		orders.AddIndex("idx_orders_group", false, "checkout_group_id", "")
		orders.AddIndex("idx_orders_market_status", false, "market_id, status", "")
		orders.AddIndex("idx_orders_short_id", true, "market_id, business_date, customer_short_id", "customer_short_id != ''")
		orders.ListRule = rule(`customer_id = @request.auth.id || (@request.auth.role = "hub_operator" && market_id = @request.auth.market_id) || ` + ruleAdmin)
		orders.ViewRule = orders.ListRule
		// create/update/delete: solo vía rutas /api/qhatu/* del core (nil = solo superusers).
		if err := app.Save(orders); err != nil {
			return err
		}

		// ------------------------------------------------------- sub_orders
		subOrders := core.NewBaseCollection("sub_orders")
		subOrders.Fields.Add(
			rel("order_id", orders, true, true),
			rel("stall_id", stalls, true, false),
			text("customer_short_id", false),
			jsonField("items"),
			money("subtotal"),
			money("rejected_amount"),
			sel("status", false, "pending_vendor", "preparing", "delivered_to_main_gate", "rejected"),
		)
		addTimestamps(subOrders)
		subOrders.AddIndex("idx_sub_orders_order", false, "order_id", "")
		subOrders.AddIndex("idx_sub_orders_stall_status", false, "stall_id, status", "")
		subOrders.ListRule = rule(`order_id.customer_id = @request.auth.id || stall_id.owner_user_id = @request.auth.id || ` +
			`(@request.auth.role = "hub_operator" && order_id.market_id = @request.auth.market_id) || ` + ruleAdmin)
		subOrders.ViewRule = subOrders.ListRule
		if err := app.Save(subOrders); err != nil {
			return err
		}

		// --------------------------------------------------- vendor_ratings
		ratings := core.NewBaseCollection("vendor_ratings")
		score := func(name string) *core.NumberField {
			return &core.NumberField{Name: name, Required: true, Min: ptr(1), Max: ptr(3), OnlyInt: true}
		}
		ratings.Fields.Add(
			rel("sub_order_id", subOrders, true, false),
			rel("stall_id", stalls, true, false),
			rel("customer_id", users, true, false),
			score("q_quality"),
			score("q_exact_weight"),
			score("q_hygiene"),
			score("q_speed"),
		)
		addTimestamps(ratings)
		ratings.AddIndex("idx_vendor_ratings_sub_order", true, "sub_order_id", "")
		ratings.AddIndex("idx_vendor_ratings_stall", false, "stall_id", "")
		ratings.ListRule = rule(`customer_id = @request.auth.id || ` + ruleAdmin)
		ratings.ViewRule = ratings.ListRule
		ratings.CreateRule = rule(ruleAuthed + ` && customer_id = @request.auth.id` +
			` && sub_order_id.order_id.customer_id = @request.auth.id` +
			` && sub_order_id.order_id.status = "completed"` +
			` && sub_order_id.status = "delivered_to_main_gate"` +
			` && stall_id = sub_order_id.stall_id`)
		if err := app.Save(ratings); err != nil {
			return err
		}

		// ---------------------------------------------------------- refunds
		refunds := core.NewBaseCollection("refunds")
		refunds.Fields.Add(
			rel("order_id", orders, true, false),
			rel("sub_order_id", subOrders, false, false),
			money("amount"),
			sel("reason", true, "item_rejected", "stall_rejected", "customer_cancelled", "vendor_timeout"),
			sel("status", true, "pending", "processed", "not_applicable"),
			text("provider_ref", false),
		)
		addTimestamps(refunds)
		refunds.AddIndex("idx_refunds_order", false, "order_id", "")
		refunds.ListRule = rule(`order_id.customer_id = @request.auth.id || ` + ruleAdmin)
		refunds.ViewRule = refunds.ListRule
		if err := app.Save(refunds); err != nil {
			return err
		}

		// -------------------------------------------------- device_registry
		devices := core.NewBaseCollection("device_registry")
		devices.Fields.Add(
			text("device_hw_id", true),
			&core.RelationField{Name: "user_ids", CollectionId: users.Id, MaxSelect: 50},
			boolean("blacklisted"),
			sel("blacklist_reason", false, "no_show_repeated", "manual"),
			date("blacklisted_at"),
		)
		addTimestamps(devices)
		devices.AddIndex("idx_device_registry_hw", true, "device_hw_id", "")
		devices.ListRule = rule(ruleAdmin)
		devices.ViewRule = rule(ruleAdmin)
		devices.UpdateRule = rule(ruleAdmin)
		if err := app.Save(devices); err != nil {
			return err
		}

		// ----------------------------------------------------- trust_events
		trustEvents := core.NewBaseCollection("trust_events")
		trustEvents.Fields.Add(
			rel("user_id", users, true, true),
			rel("order_id", orders, false, false),
			sel("event", true, "order_completed", "no_show", "late_cancel", "manual_adjust", "recovery"),
			&core.NumberField{Name: "delta"},
			&core.NumberField{Name: "score_after"},
		)
		addTimestamps(trustEvents)
		trustEvents.AddIndex("idx_trust_events_user", false, "user_id", "")
		trustEvents.ListRule = rule(ruleAdmin)
		trustEvents.ViewRule = rule(ruleAdmin)
		if err := app.Save(trustEvents); err != nil {
			return err
		}

		// --------------------------------------------------- daily_counters
		counters := core.NewBaseCollection("daily_counters")
		counters.Fields.Add(
			rel("market_id", markets, true, true),
			&core.TextField{Name: "date", Required: true, Pattern: `^\d{4}-\d{2}-\d{2}$`},
			integer("last_seq", ptr(0)),
		)
		addTimestamps(counters)
		counters.AddIndex("idx_daily_counters_market_date", true, "market_id, date", "")
		counters.ListRule = rule(ruleAdmin)
		counters.ViewRule = rule(ruleAdmin)
		if err := app.Save(counters); err != nil {
			return err
		}

		// ----------------------------------------------- vendor_settlements
		settlements := core.NewBaseCollection("vendor_settlements")
		settlements.Fields.Add(
			rel("market_id", markets, true, false),
			rel("stall_id", stalls, true, false),
			&core.TextField{Name: "period_date", Required: true, Pattern: `^\d{4}-\d{2}-\d{2}$`},
			integer("orders_count", ptr(0)),
			money("gross_amount"),
			money("refunded_amount"),
			money("net_amount"),
			sel("status", true, "pending", "paid"),
			date("paid_at"),
			text("paid_ref", false),
		)
		addTimestamps(settlements)
		settlements.AddIndex("idx_vendor_settlements_stall_date", true, "stall_id, period_date", "")
		settlements.ListRule = rule(`stall_id.owner_user_id = @request.auth.id || ` + ruleAdmin)
		settlements.ViewRule = settlements.ListRule
		settlements.UpdateRule = rule(ruleAdmin + ` && ` +
			notSet("market_id", "stall_id", "period_date", "orders_count", "gross_amount", "refunded_amount", "net_amount"))
		if err := app.Save(settlements); err != nil {
			return err
		}

		// ----------------------------------------------------- app_settings
		settings := core.NewBaseCollection("app_settings")
		settings.Fields.Add(
			&core.TextField{Name: "key", Required: true, Pattern: `^[a-z0-9_]+$`},
			jsonField("value"),
			text("description", false),
		)
		addTimestamps(settings)
		settings.AddIndex("idx_app_settings_key", true, "key", "")
		// Lectura pública: no contiene secretos y el frontend necesita algunos valores.
		settings.ListRule = rule("")
		settings.ViewRule = rule("")
		settings.UpdateRule = rule(ruleAdmin + ` && @request.body.key:isset = false`)
		return app.Save(settings)
	}, func(app core.App) error {
		for _, name := range []string{
			"app_settings", "vendor_settlements", "daily_counters", "trust_events", "device_registry",
			"refunds", "vendor_ratings", "sub_orders", "orders", "products", "catalog_items", "stalls",
		} {
			c, err := app.FindCollectionByNameOrId(name)
			if err != nil {
				continue
			}
			if err := app.Delete(c); err != nil {
				return err
			}
		}

		users, err := app.FindCollectionByNameOrId("users")
		if err == nil {
			for _, f := range protectedUserFields {
				users.Fields.RemoveByName(f)
			}
			users.Fields.RemoveByName("phone")
			users.RemoveIndex("idx_users_role_market")
			if err := app.Save(users); err != nil {
				return err
			}
		}

		if c, err := app.FindCollectionByNameOrId("markets"); err == nil {
			return app.Delete(c)
		}
		return nil
	})
}
