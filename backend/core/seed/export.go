package seed

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"unicode"

	"ganigeorgiev.com/pocketbase/pocketbase/core"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// maxImageBytes es el límite que valida el CI para data/seed/images.
const maxImageBytes = 1 << 20

// ExportOptions configura Export.
type ExportOptions struct {
	Dir string // carpeta data/seed de destino
}

// ExportReport resume lo exportado.
type ExportReport struct {
	Markets, Catalog, Stalls, Products, Images int
	Warnings                                   []string
}

func (r ExportReport) String() string {
	return fmt.Sprintf("mercados=%d catálogo=%d puestos=%d productos=%d imágenes=%d avisos=%d",
		r.Markets, r.Catalog, r.Stalls, r.Products, r.Images, len(r.Warnings))
}

// Export escribe en opt.Dir lo que hay en la base (markets, catalog_items, stalls,
// products y sus imágenes) con el mismo formato que lee Run. Es lo inverso de Run:
// permite cargar datos desde el panel /_/ y subirlos al repo.
//
// No exporta usuarios: owner_email solo se conserva si el usuario es de prueba
// (aparece en users.dev.json), para no publicar correos reales en el repo.
func Export(app core.App, opt ExportOptions) (ExportReport, error) {
	e := &exporter{app: app, dir: opt.Dir, devEmails: map[string]bool{}}
	var dev []userJSON
	if b, err := os.ReadFile(filepath.Join(opt.Dir, "users.dev.json")); err == nil {
		_ = json.Unmarshal(b, &dev)
	}
	for _, u := range dev {
		e.devEmails[strings.ToLower(u.Email)] = true
	}

	fsys, err := app.NewFilesystem()
	if err != nil {
		return e.rep, err
	}
	defer fsys.Close()
	e.readFile = func(r *core.Record, name string) ([]byte, error) {
		rd, err := fsys.GetReader(r.BaseFilesPath() + "/" + name)
		if err != nil {
			return nil, err
		}
		defer rd.Close()
		return io.ReadAll(rd)
	}

	for _, step := range []func() error{e.markets, e.catalog, e.stalls, e.products} {
		if err := step(); err != nil {
			return e.rep, err
		}
	}
	for _, w := range e.rep.Warnings {
		slog.Warn("seed export", "aviso", w)
	}
	return e.rep, nil
}

type exporter struct {
	app       core.App
	dir       string
	devEmails map[string]bool
	readFile  func(r *core.Record, name string) ([]byte, error)
	rep       ExportReport

	marketSlug map[string]string // id -> slug
	stallKey   map[string][2]string
	catalogBy  map[string]string // id -> name
}

func (e *exporter) warn(format string, a ...any) {
	e.rep.Warnings = append(e.rep.Warnings, fmt.Sprintf(format, a...))
}

func (e *exporter) all(collection string) ([]*core.Record, error) {
	return e.app.FindAllRecords(collection)
}

func (e *exporter) markets() error {
	recs, err := e.all("markets")
	if err != nil {
		return err
	}
	e.marketSlug = map[string]string{}
	out := make([]marketJSON, 0, len(recs))
	for _, r := range recs {
		slug := r.GetString("slug")
		e.marketSlug[r.Id] = slug
		img, err := e.image(r, "bubble_icon", "markets", slug)
		if err != nil {
			return err
		}
		out = append(out, marketJSON{
			Slug: slug, Name: r.GetString("name"), CodePrefix: r.GetString("code_prefix"),
			Active: r.GetBool("active"), HubLocationDesc: r.GetString("hub_location_desc"),
			OpenTime: r.GetString("open_time"), CloseTime: r.GetString("close_time"),
			HasOwnDelivery: r.GetBool("has_own_delivery"), DeliveryFee: r.GetFloat("delivery_fee"),
			RappiEnabled: r.GetBool("rappi_enabled"), BubbleIcon: img,
		})
	}
	slices.SortFunc(out, func(a, b marketJSON) int { return strings.Compare(a.Slug, b.Slug) })
	e.rep.Markets = len(out)
	return writeChecked(e, "markets.json", out, func(m marketJSON) string { return m.Slug })
}

func (e *exporter) catalog() error {
	recs, err := e.all("catalog_items")
	if err != nil {
		return err
	}
	e.catalogBy = map[string]string{}
	out := make([]catalogJSON, 0, len(recs))
	for _, r := range recs {
		name := r.GetString("name")
		e.catalogBy[r.Id] = name
		img, err := e.image(r, "default_image", "catalog", name)
		if err != nil {
			return err
		}
		out = append(out, catalogJSON{Name: name, Section: r.GetString("section"), DefaultImage: img})
	}
	slices.SortFunc(out, func(a, b catalogJSON) int {
		return cmpChain(strings.Compare(a.Section, b.Section), strings.Compare(a.Name, b.Name))
	})
	e.rep.Catalog = len(out)
	return writeChecked(e, "catalog_items.json", out, func(c catalogJSON) string { return c.Name })
}

func (e *exporter) stalls() error {
	recs, err := e.all("stalls")
	if err != nil {
		return err
	}
	e.stallKey = map[string][2]string{}
	out := make([]stallJSON, 0, len(recs))
	for _, r := range recs {
		market := e.marketSlug[r.GetString("market_id")]
		number := r.GetString("stall_number")
		e.stallKey[r.Id] = [2]string{market, number}

		owner := ""
		if id := r.GetString("owner_user_id"); id != "" {
			if u, err := e.app.FindRecordById("users", id); err == nil {
				if e.devEmails[strings.ToLower(u.Email())] {
					owner = u.Email()
				} else {
					e.warn("puesto %q: el dueño no es un usuario de prueba, se exporta sin owner_email", r.GetString("name"))
				}
			}
		}
		until := ""
		if d := r.GetDateTime("enabled_until"); !d.IsZero() {
			until = d.Time().Format("2006-01-02")
		}
		out = append(out, stallJSON{
			Market: market, StallNumber: number, Name: r.GetString("name"),
			Section: r.GetString("section"), OwnerEmail: owner,
			Active: r.GetBool("active"), EnabledUntil: until,
		})
	}
	slices.SortFunc(out, func(a, b stallJSON) int {
		return cmpChain(strings.Compare(a.Market, b.Market), strings.Compare(a.StallNumber, b.StallNumber))
	})
	e.rep.Stalls = len(out)
	return writeChecked(e, "stalls.json", out, func(s stallJSON) string { return s.Market + " / " + s.StallNumber })
}

func (e *exporter) products() error {
	recs, err := e.all("products")
	if err != nil {
		return err
	}
	out := make([]productJSON, 0, len(recs))
	for _, r := range recs {
		key := e.stallKey[r.GetString("stall_id")]
		name, unit := r.GetString("name"), r.GetString("unit_type")
		img, err := e.image(r, "image", "products", key[0]+" "+key[1]+" "+name+" "+unit)
		if err != nil {
			return err
		}
		out = append(out, productJSON{
			Market: key[0], StallNumber: key[1], Name: name,
			CatalogItem: e.catalogBy[r.GetString("catalog_item_id")],
			UnitType:    unit, Price: r.GetFloat("price"), InStock: r.GetBool("in_stock"), Image: img,
		})
	}
	slices.SortFunc(out, func(a, b productJSON) int {
		return cmpChain(strings.Compare(a.Market, b.Market), strings.Compare(a.StallNumber, b.StallNumber),
			strings.Compare(a.Name, b.Name), strings.Compare(a.UnitType, b.UnitType))
	})
	e.rep.Products = len(out)
	return writeChecked(e, "products.json", out, func(p productJSON) string {
		return p.Market + " / " + p.StallNumber + " / " + p.Name + " (" + p.UnitType + ")"
	})
}

// image copia el archivo del campo a images/<folder>/<slug>.<ext> y devuelve la ruta relativa.
// Si el destino ya existe con el mismo contenido no lo reescribe.
func (e *exporter) image(r *core.Record, field, folder, label string) (string, error) {
	name := r.GetString(field)
	if name == "" {
		return "", nil
	}
	data, err := e.readFile(r, name)
	if err != nil {
		e.warn("%s %q: no se pudo leer la imagen %q: %v", r.Collection().Name, label, name, err)
		return "", nil
	}
	if len(data) > maxImageBytes {
		e.warn("%s %q: la imagen pesa %d KB (máx. 1024 KB): redúcela antes del PR", r.Collection().Name, label, len(data)/1024)
	}
	rel := path.Join(folder, slugify(label)+strings.ToLower(path.Ext(name)))
	dst := filepath.Join(e.dir, "images", filepath.FromSlash(rel))
	if old, err := os.ReadFile(dst); err == nil && sha256.Sum256(old) == sha256.Sum256(data) {
		return rel, nil
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return "", err
	}
	e.rep.Images++
	return rel, nil
}

// writeChecked escribe el archivo y avisa qué registros del JSON anterior ya no están en la
// base (p. ej. si se exportó sin haber hecho `git pull` + `make seed`).
func writeChecked[T any](e *exporter, name string, items []T, key func(T) string) error {
	var prev []T
	if b, err := os.ReadFile(filepath.Join(e.dir, name)); err == nil {
		_ = json.Unmarshal(b, &prev)
	}
	now := make(map[string]bool, len(items))
	for _, it := range items {
		now[key(it)] = true
	}
	for _, it := range prev {
		if k := key(it); !now[k] {
			e.warn("%s: se quitó %q (estaba en el JSON pero no en tu base; si no lo borraste tú, haz git pull + make seed y exporta de nuevo)", name, k)
		}
	}
	return writeRows(e.dir, name, items)
}

// writeRows guarda un arreglo JSON con un registro por línea (diffs limpios en git).
func writeRows[T any](dir, name string, items []T) error {
	var buf bytes.Buffer
	buf.WriteString("[\n")
	for i, it := range items {
		var line bytes.Buffer
		enc := json.NewEncoder(&line)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(it); err != nil {
			return err
		}
		buf.WriteString("  " + spaceJSON(strings.TrimSpace(line.String())))
		if i < len(items)-1 {
			buf.WriteByte(',')
		}
		buf.WriteByte('\n')
	}
	buf.WriteString("]\n")
	return os.WriteFile(filepath.Join(dir, name), buf.Bytes(), 0o644)
}

// spaceJSON da formato `{ "a": 1, "b": "x" }` a un objeto JSON compacto plano.
func spaceJSON(s string) string {
	var out strings.Builder
	inStr, esc := false, false
	for _, c := range s {
		switch {
		case esc:
			esc = false
		case c == '\\' && inStr:
			esc = true
		case c == '"':
			inStr = !inStr
		case !inStr && c == ':':
			out.WriteString(": ")
			continue
		case !inStr && c == ',':
			out.WriteString(", ")
			continue
		case !inStr && c == '{':
			out.WriteString("{ ")
			continue
		case !inStr && c == '}':
			out.WriteString(" }")
			continue
		}
		out.WriteRune(c)
	}
	return out.String()
}

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

// slugify: "Pabellón A - Puesto 12 Papa" -> "pabellon-a-puesto-12-papa".
func slugify(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	s, _, _ = transform.String(t, strings.ToLower(s))
	return strings.Trim(nonSlug.ReplaceAllString(s, "-"), "-")
}

func cmpChain(cs ...int) int {
	for _, c := range cs {
		if c != 0 {
			return c
		}
	}
	return 0
}
