# PROYECTO DE ARQUITECTURA DE SOFTWARE

## 1. NOMBRE DE PROYECTO
**QhatuDigital Cusco** — Plataforma Web y Móvil Multi-Mercado con Consolidación de Pedidos y Despacho Inteligente.

---

## 2. INTEGRANTES DEL GRUPO
* **Rye Gabriel Gregory Paucar Quejia**
* **Luis Alberto Tecsi Huallpa**
* **Nelson Luis Pauccara Huillca**
* **Salome Celeste Ccahua Huamani**
* **Johan Robinho Quispe Juarez**

---

## 3. INTRODUCCIÓN (¿De qué trata y para qué sirve?)
* **¿De qué trata?** 
  Es una plataforma digital (construida con **SvelteKit v5+**, **Tailwind CSS**, **Skeleton UI** en modo estático, y un backend en **Go** basado en **PocketBase v0.41.1**; preparada para empaquetarse como app móvil con **Capacitor** en una versión futura) que reúne en un solo lugar a los principales mercados de abastos de la ciudad del Cusco (como Vinocanchón, Wanchaq, San Pedro, Ttio y Rosaspata). A través de una interfaz inicial minimalista tipo burbuja, el ciudadano puede explorar libremente los productos de múltiples puestos, vendedores y mercados, comparar precios del mismo producto entre puestos, armar un único carrito de compras unificado (independiente del mercado) y pagar mediante billetera digital o contraentrega validada.

* **¿Para qué sirve? (Problema que resuelve):** 
  Resuelve la pérdida de tiempo, la congestión y el esfuerzo físico de caminar puesto por puesto dentro de mercados de gran extensión. Cuando el usuario confirma su compra (autenticándose con Google OAuth2), el sistema crea una orden por cada mercado, la divide automáticamente por puesto y avisa a cada vendedor involucrado. Los vendedores aceptan o rechazan cada ítem (los rechazados se reembolsan completos, respetando siempre el precio pactado), alistan sus productos y los llevan a un **Punto de Acopio Central (Puerta Principal del mercado)**, donde el cliente —o un servicio de delivery como Rappi o flota propia del mercado— recoge todo el pedido consolidado en un solo paso mostrando un comprobante digital con **Código QR de token único**. Además, incorpora un sistema minimalista de calificación por cuestionario cerrado (calidad, peso exacto, higiene y rapidez) libre de spam.

---

## 4. MICROSERVICIOS A UTILIZAR
El sistema implementa una arquitectura basada en un núcleo BaaS/Core y microservicios satélite especializados que se comunican mediante eventos y APIs REST/Webhooks:

1. **Servicio Core de Catálogo, Usuarios y Sub-Órdenes (`pb-core-service` - PocketBase v0.41.1):**
   * Gestiona el catálogo público de mercados, puestos y productos, la autenticación con Google OAuth2, la división de una orden global en sub-órdenes por puesto en tiempo real (SSE) y las cuentas CMS de los vendedores.
2. **Microservicio de Identidad Móvil y Antifraude (`ms-auth-trust`):**
   * Evalúa la confiabilidad del comprador (`Trust Score`) para habilitar la modalidad de *"Pago en mercado"*. Verifica el identificador del dispositivo (en la versión web, un identificador generado en el navegador; en la futura app móvil, `Android_ID` / `iOS UUID` vía `@capacitor/device`) para evitar falsas compras y gestiona la validación por **SMS (OTP)** obligatoria en la primera compra.
3. **Microservicio de Pagos y Firma de Comprobantes QR (`ms-payments-qr`):**
   * Procesa la validación de pagos por billeteras digitales y genera el comprobante digital cifrado con el **Código QR de Token Único**, el cual se valida en la puerta principal del mercado al entregar los productos al cliente o al repartidor.
4. **Microservicio de Despacho y Logística (`ms-logistics-dispatch`):**
   * Coordina el estado de recepción de bolsas en el punto de acopio (Puerta Principal) e integra los adaptadores de envío tanto para el servicio de delivery propio de cada mercado como para plataformas externas (**Rappi**).
5. **Microservicio de Reputación y Métricas de Vendedores (`ms-ratings-aggregator`):**
   * Procesa las respuestas del cuestionario minimalista cerrado (sin texto libre) tras cada entrega exitosa y recalcula en tiempo real los indicadores de calidad, peso exacto y rapidez de cada puesto.

---

## 5. ALCANCE DE LA VERSIÓN 1
* Solo **aplicación web** (SvelteKit en modo SPA). La app móvil con Capacitor queda para la versión 2.
* Las integraciones externas (**pagos Yape/Plin, SMS y Rappi**) son **simuladas** con la misma interfaz que tendrá la integración real.
* Los vendedores pagan el alquiler de su puesto en un sistema externo; en esta plataforma el **administrador** los habilita o deshabilita.
* Los pagos de los clientes ingresan primero a la **cuenta central de cada mercado**, que luego liquida a cada vendedor según el reporte de ventas por puesto.
* Los parámetros de negocio (trust score, tiempos, montos) son configurables desde el panel de PocketBase para su ajuste fino manual.

---

## 6. CÓMO EJECUTAR EL PROYECTO
Solo se necesita **Git**, **Docker** (con Docker Compose) y **make**; no hace falta instalar Go ni Node.

```bash
git clone https://github.com/76054175-cmd/CloudMercadoCusco.git
cd CloudMercadoCusco
make dev
```

| URL | Qué es |
|---|---|
| http://localhost:5173 | Aplicación web (desarrollo, recarga en caliente) |
| http://localhost:8090/_/ | Panel de administración de PocketBase |

La primera ejecución carga automáticamente los mercados, puestos, productos y usuarios de prueba desde `data/seed/`. Credenciales de desarrollo en `.env` (se crea desde `.env.example`). `make help` lista todos los comandos.

**Producción:** `make build` genera una imagen Docker de ~43 MB con un único binario Go que sirve la web y la API.

## 7. ESTRUCTURA DEL REPOSITORIO
| Carpeta | Contenido |
|---|---|
| `apps/web/` | Frontend SvelteKit 3 (Svelte 5 Runes) + Tailwind CSS 4 + Skeleton UI 5, compilado como sitio estático |
| `backend/core/` | `pb-core-service`: PocketBase v0.41.1 como librería Go (esquema, reglas, hooks y web embebida) |
| `backend/ms-*/` | Microservicios Go (simulados en la v1) |
| `data/seed/` | Catálogo base, datos de prueba e imágenes, en JSON |

## 8. CÓMO COLABORAR
Todos los integrantes pueden subir **imágenes y datos** (mercados, puestos, productos) editando `data/seed/`, y código, mediante ramas y Pull Requests. La guía paso a paso está en [CONTRIBUTING.md](CONTRIBUTING.md) y el formato de los datos en [data/seed/README.md](data/seed/README.md).

## 9. ESTADO
- [x] **Fase 1:** cimientos, esquema PocketBase, datos de prueba y selector de mercados tipo burbuja.
- [x] **Fase 2:** catálogo con filtros y buscador, comparación entre puestos y carrito multi-mercado.
- [x] `make seed-export`: lo cargado en el panel `/_/` se pasa a `data/seed/` para subirlo al repo.
- [ ] Fase 3: login con Google, antifraude y checkout con QR.
- [ ] Fase 4: módulo de vendedor, acopio en puerta y administración.
- [ ] Fase 5: calificaciones.
