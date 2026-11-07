// Package seed importa los datos de prueba/catálogo desde data/seed (JSON + imágenes).
//
// Reglas:
//   - Idempotente: cada registro se identifica por una clave natural
//     (slug, email, nombre, número de puesto...).
//   - Por defecto SOLO CREA lo que falta: no pisa lo editado a mano en /_/.
//     Con Options.Update se sobrescriben los campos presentes en el JSON.
//   - Usuarios y superusuario solo se importan con Options.Dev (APP_ENV=dev).
//   - Options.DryRun valida todo dentro de una transacción y la revierte.
package seed

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"ganigeorgiev.com/pocketbase/pocketbase/core"
	"ganigeorgiev.com/pocketbase/pocketbase/tools/filesystem"
)

type Options struct {
	Dir      string // carpeta data/seed
	Update   bool   // sobrescribir registros existentes
	DryRun   bool   // validar y revertir
	Dev      bool   // importar usuarios de prueba
	Password string // contraseña de usuarios de prueba (SEED_DEV_PASSWORD)
}

type Report struct {
	Created, Updated, Skipped int
}

func (r Report) String() string {
	return fmt.Sprintf("creados=%d actualizados=%d sin cambios=%d", r.Created, r.Updated, r.Skipped)
}

var errDryRun = errors.New("dry-run")

// --- formato de los archivos JSON (ver data/seed/README.md) ---

type marketJSON struct {
	Slug            string  `json:"slug"`
	Name            string  `json:"name"`
	CodePrefix      string  `json:"code_prefix"`
	Active          bool    `json:"active"`
	HubLocationDesc string  `json:"hub_location_desc"`
	OpenTime        string  `json:"open_time"`
	CloseTime       string  `json:"close_time"`
	HasOwnDelivery  bool    `json:"has_own_delivery"`
	DeliveryFee     float64 `json:"delivery_fee"`
	RappiEnabled    bool    `json:"rappi_enabled"`
	BubbleIcon      string  `json:"bubble_icon"`
}

type userJSON struct {
	Email       string `json:"email"`
	Name        string `json:"name"`
	Role        string `json:"role"`
	Market      string `json:"market"`
	Phone       string `json:"phone"`
	SMSVerified bool   `json:"sms_verified"`
}

type catalogJSON struct {
	Name         string `json:"name"`
	Section      string `json:"section"`
	DefaultImage string `json:"default_image"`
}

type stallJSON struct {
	Market       string `json:"market"`
	StallNumber  string `json:"stall_number"`
	Name         string `json:"name"`
	Section      string `json:"section"`
	OwnerEmail   string `json:"owner_email"`
	Active       bool   `json:"active"`
	EnabledUntil string `json:"enabled_until"`
}

type productJSON struct {
	Market      string  `json:"market"`
	StallNumber string  `json:"stall_number"`
	Name        string  `json:"name"`
	CatalogItem string  `json:"catalog_item"`
	UnitType    string  `json:"unit_type"`
	Price       float64 `json:"price"`
	InStock     bool    `json:"in_stock"`
	Image       string  `json:"image"`
}

// Run importa todo en orden de dependencias.
func Run(app core.App, opt Options) (Report, error) {
	var rep Report
	err := app.RunInTransaction(func(tx core.App) error {
		s := &seeder{app: tx, opt: opt, rep: &rep}
		steps := []func() error{s.markets, s.superuser, s.users, s.catalog, s.stalls, s.products}
		for _, step := range steps {
			if err := step(); err != nil {
				return err
			}
		}
		if opt.DryRun {
			return errDryRun
		}
		return nil
	})
	if errors.Is(err, errDryRun) {
		err = nil
	}
	return rep, err
}

type seeder struct {
	app core.App
	opt Options
	rep *Report
}

func (s *seeder) load(name string, v any) (bool, error) {
	b, err := os.ReadFile(filepath.Join(s.opt.Dir, name))
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	dec := json.NewDecoder(strings.NewReader(string(b)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return false, fmt.Errorf("%s: %w", name, err)
	}
	return true, nil
}

// upsert busca por filtro; crea si no existe o actualiza si Options.Update.
// apply rellena los campos; isNew permite asignar campos solo al crear.
func (s *seeder) upsert(collection, filter string, params map[string]any, label string,
	apply func(r *core.Record, isNew bool) error,
) (*core.Record, error) {
	rec, err := s.app.FindFirstRecordByFilter(collection, filter, params)
	isNew := err != nil
	if isNew {
		col, cerr := s.app.FindCollectionByNameOrId(collection)
		if cerr != nil {
			return nil, cerr
		}
		rec = core.NewRecord(col)
	} else if !s.opt.Update {
		s.rep.Skipped++
		return rec, nil
	}
	if err := apply(rec, isNew); err != nil {
		return nil, fmt.Errorf("%s %q: %w", collection, label, err)
	}
	if err := s.app.Save(rec); err != nil {
		return nil, fmt.Errorf("%s %q: %w", collection, label, err)
	}
	if isNew {
		s.rep.Created++
	} else {
		s.rep.Updated++
	}
	return rec, nil
}

// setImage adjunta data/seed/images/<rel> al campo si está vacío (o si Update).
func (s *seeder) setImage(r *core.Record, field, rel string) error {
	if rel == "" || (r.GetString(field) != "" && !s.opt.Update) {
		return nil
	}
	p := filepath.Join(s.opt.Dir, "images", filepath.FromSlash(rel))
	f, err := filesystem.NewFileFromPath(p)
	if err != nil {
		return fmt.Errorf("imagen %q: %w", rel, err)
	}
	r.Set(field, f)
	return nil
}

func (s *seeder) marketBySlug(slug string) (*core.Record, error) {
	r, err := s.app.FindFirstRecordByData("markets", "slug", slug)
	if err != nil {
		return nil, fmt.Errorf("mercado %q no existe (revisa markets.json)", slug)
	}
	return r, nil
}

func (s *seeder) markets() error {
	var items []marketJSON
	if ok, err := s.load("markets.json", &items); !ok || err != nil {
		return err
	}
	for _, it := range items {
		_, err := s.upsert("markets", "slug = {:slug}", map[string]any{"slug": it.Slug}, it.Slug,
			func(r *core.Record, _ bool) error {
				r.Set("slug", it.Slug)
				r.Set("name", it.Name)
				r.Set("code_prefix", it.CodePrefix)
				r.Set("active", it.Active)
				r.Set("hub_location_desc", it.HubLocationDesc)
				r.Set("open_time", it.OpenTime)
				r.Set("close_time", it.CloseTime)
				r.Set("has_own_delivery", it.HasOwnDelivery)
				r.Set("delivery_fee", it.DeliveryFee)
				r.Set("rappi_enabled", it.RappiEnabled)
				return s.setImage(r, "bubble_icon", it.BubbleIcon)
			})
		if err != nil {
			return err
		}
	}
	return nil
}

// superuser crea el superusuario de desarrollo desde PB_SUPERUSER_EMAIL/PASSWORD.
func (s *seeder) superuser() error {
	email, pass := os.Getenv("PB_SUPERUSER_EMAIL"), os.Getenv("PB_SUPERUSER_PASSWORD")
	if !s.opt.Dev || email == "" || pass == "" {
		return nil
	}
	if _, err := s.app.FindAuthRecordByEmail(core.CollectionNameSuperusers, email); err == nil {
		return nil
	}
	col, err := s.app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
	if err != nil {
		return err
	}
	r := core.NewRecord(col)
	r.SetEmail(email)
	r.SetPassword(pass)
	if err := s.app.Save(r); err != nil {
		return fmt.Errorf("superusuario: %w", err)
	}
	s.rep.Created++
	return nil
}

func (s *seeder) users() error {
	if !s.opt.Dev {
		return nil
	}
	if s.opt.Password == "" {
		slog.Warn("SEED_DEV_PASSWORD vacío: se omiten los usuarios de prueba")
		return nil
	}
	var items []userJSON
	if ok, err := s.load("users.dev.json", &items); !ok || err != nil {
		return err
	}
	for _, it := range items {
		_, err := s.upsert("users", "email = {:email}", map[string]any{"email": it.Email}, it.Email,
			func(r *core.Record, isNew bool) error {
				r.SetEmail(it.Email)
				r.SetVerified(true)
				if isNew {
					r.SetPassword(s.opt.Password)
				}
				r.Set("name", it.Name)
				r.Set("role", it.Role)
				r.Set("phone", it.Phone)
				r.Set("sms_verified", it.SMSVerified)
				r.Set("market_id", "")
				if it.Market != "" {
					m, err := s.marketBySlug(it.Market)
					if err != nil {
						return err
					}
					r.Set("market_id", m.Id)
				}
				return nil
			})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *seeder) catalog() error {
	var items []catalogJSON
	if ok, err := s.load("catalog_items.json", &items); !ok || err != nil {
		return err
	}
	for _, it := range items {
		_, err := s.upsert("catalog_items", "name = {:name}", map[string]any{"name": it.Name}, it.Name,
			func(r *core.Record, _ bool) error {
				r.Set("name", it.Name)
				r.Set("section", it.Section)
				return s.setImage(r, "default_image", it.DefaultImage)
			})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *seeder) stalls() error {
	var items []stallJSON
	if ok, err := s.load("stalls.json", &items); !ok || err != nil {
		return err
	}
	for _, it := range items {
		m, err := s.marketBySlug(it.Market)
		if err != nil {
			return fmt.Errorf("stalls %q: %w", it.Name, err)
		}
		_, err = s.upsert("stalls", "market_id = {:m} && stall_number = {:n}",
			map[string]any{"m": m.Id, "n": it.StallNumber}, it.Name,
			func(r *core.Record, isNew bool) error {
				r.Set("market_id", m.Id)
				r.Set("stall_number", it.StallNumber)
				r.Set("name", it.Name)
				r.Set("section", it.Section)
				r.Set("active", it.Active)
				r.Set("enabled_until", it.EnabledUntil)
				if isNew {
					r.Set("metrics_summary", map[string]any{
						"quality": 2.0, "exact_weight": 2.0, "hygiene": 2.0, "speed": 2.0,
						"overall": 2.0, "total_votes": 0,
					})
				}
				if it.OwnerEmail != "" {
					// En producción el dueño se asigna en /_/; si el usuario no existe se deja vacío.
					if u, err := s.app.FindAuthRecordByEmail("users", it.OwnerEmail); err == nil {
						r.Set("owner_user_id", u.Id)
					}
				}
				return nil
			})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *seeder) products() error {
	var items []productJSON
	if ok, err := s.load("products.json", &items); !ok || err != nil {
		return err
	}
	for _, it := range items {
		m, err := s.marketBySlug(it.Market)
		if err != nil {
			return fmt.Errorf("products %q: %w", it.Name, err)
		}
		stall, err := s.app.FindFirstRecordByFilter("stalls", "market_id = {:m} && stall_number = {:n}",
			map[string]any{"m": m.Id, "n": it.StallNumber})
		if err != nil {
			return fmt.Errorf("products %q: puesto %q no existe en %q", it.Name, it.StallNumber, it.Market)
		}
		label := it.Name + " (" + it.UnitType + ")"
		_, err = s.upsert("products", "stall_id = {:s} && name = {:name} && unit_type = {:u}",
			map[string]any{"s": stall.Id, "name": it.Name, "u": it.UnitType}, label,
			func(r *core.Record, _ bool) error {
				r.Set("stall_id", stall.Id)
				r.Set("name", it.Name)
				r.Set("unit_type", it.UnitType)
				r.Set("price", it.Price)
				r.Set("in_stock", it.InStock)
				r.Set("catalog_item_id", "")
				if it.CatalogItem != "" {
					ci, err := s.app.FindFirstRecordByData("catalog_items", "name", it.CatalogItem)
					if err != nil {
						return fmt.Errorf("catalog_item %q no existe (revisa catalog_items.json)", it.CatalogItem)
					}
					r.Set("catalog_item_id", ci.Id)
				}
				return s.setImage(r, "image", it.Image)
			})
		if err != nil {
			return err
		}
	}
	return nil
}
