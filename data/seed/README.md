# Datos de catálogo y prueba (`data/seed`)

Aquí viven los **datos y las imágenes** que se cargan automáticamente en PocketBase.
Cualquier integrante puede agregar mercados, puestos, productos e imágenes editando estos archivos.

| Archivo | Qué contiene | Clave única (no la repitas) |
|---|---|---|
| `markets.json` | Mercados | `slug` |
| `catalog_items.json` | Producto genérico ("Papa amarilla") para comparar precios entre puestos | `name` |
| `stalls.json` | Puestos de cada mercado | `market` + `stall_number` |
| `products.json` | Lo que vende cada puesto, con **su** precio | `market` + `stall_number` + `name` + `unit_type` |
| `users.dev.json` | Usuarios de **prueba** (solo desarrollo, nunca datos reales) | `email` |
| `images/` | Fotos: `images/markets/`, `images/catalog/`, `images/products/` | ruta del archivo |

## Valores permitidos

- `section`: `verduras`, `carnes_aves`, `abarrotes`, `quesos_lacteos`, `frutas`, `otros`
- `unit_type`: `kg`, `medio_kg`, `cuarto_kg`, `atado`, `manojo`, `unidad`, `porcion`
- `role` (usuarios de prueba): `customer`, `vendor`, `hub_operator`, `admin`
- `market`: el `slug` de un mercado de `markets.json` (ej. `vinocanchon`)
- `catalog_item`: el `name` exacto de `catalog_items.json`, o `""` si no aplica
- Horas en formato `HH:mm` (24 h). Precios en soles con punto decimal (`3.50`).

## Ejemplo: agregar un producto con foto

1. Copia la foto a `data/seed/images/products/papa-huayro-juana.webp`.
2. Agrega en `products.json`:
   ```json
   {
     "market": "vinocanchon",
     "stall_number": "Pabellón A - Puesto 12",
     "name": "Papa huayro",
     "catalog_item": "",
     "unit_type": "kg",
     "price": 4.2,
     "in_stock": true,
     "image": "products/papa-huayro-juana.webp"
   }
   ```
3. Valida: `make seed-check` (debe terminar en `seed (dry-run, sin guardar) OK`).
4. Carga en tu entorno: `make seed`.

## Reglas para imágenes

- Formato **WebP o JPG** (PNG/SVG solo para íconos). Máximo **1 MB** (el sistema rechaza > 5 MB).
- Cuadradas, idealmente **800×800 px**.
- Nombres en minúsculas, sin espacios ni tildes: `queso-paria.webp`.
- Solo fotos **propias o con permiso** de uso. Nada descargado de Google sin licencia.
- Si un producto no tiene foto, se usa la del `catalog_item` y, si tampoco hay, un ícono de la sección.

## Cómo se cargan

- Al iniciar en desarrollo (`make dev`) el core importa todo **solo creando lo que falta**:
  lo que edites a mano en el panel `/_/` no se pisa.
- `make seed-update` sí sobrescribe con los valores de los JSON.
- `users.dev.json` **nunca** se importa en producción.
- Los JSON no aceptan campos desconocidos: un error de tipeo hace fallar `make seed-check`, a propósito.
