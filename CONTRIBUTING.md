# Cómo colaborar en QhatuDigital

## 1. Preparar tu entorno (una sola vez)

Solo necesitas **Git**, **Docker** y **make**. No hace falta instalar Go ni Node.

```bash
git clone https://github.com/76054175-cmd/CloudMercadoCusco.git
cd CloudMercadoCusco
make dev
```

La primera vez tarda varios minutos (descarga imágenes y compila). Luego:

| URL | Qué es |
|---|---|
| http://localhost:5173 | Web (se recarga sola al guardar) |
| http://localhost:8090/_/ | Panel de PocketBase (usuario/clave en `.env`: `PB_SUPERUSER_*`) |

Usuarios de prueba: ver `data/seed/users.dev.json` (contraseña `SEED_DEV_PASSWORD` de `.env`).

## 2. Flujo de trabajo con ramas

`main` siempre debe funcionar. Nadie sube directo a `main`: todo entra por **Pull Request**.

```bash
git checkout main && git pull
git checkout -b datos/puestos-san-pedro      # o feat/..., fix/..., docs/...
# ... cambios ...
make seed-check                               # si tocaste data/seed
make check                                    # si tocaste código
git add data/seed/...                         # agrega archivos concretos, no "git add ."
git commit -m "Agregar puestos de San Pedro"
git push -u origin datos/puestos-san-pedro
```

Luego abre el Pull Request en GitHub. GitHub ejecuta las validaciones automáticamente y Luis lo revisa y lo une a `main`.

Prefijos de rama: `datos/` (imágenes y JSON), `feat/` (funcionalidad), `fix/` (errores), `docs/` (documentación).

## 3. Subir datos e imágenes

Todo está explicado en [`data/seed/README.md`](data/seed/README.md). Resumen:

1. Imágenes en `data/seed/images/<markets|catalog|products>/`.
2. Datos en los JSON de `data/seed/`.
3. `make seed-check` para validar y `make seed` para verlo en tu web local.

> **Datos reales de producción** (fotos definitivas, vendedores reales) se cargan desde el panel
> `/_/` del servidor, no por el repo. El repo es para el catálogo base y los datos de prueba.

## 4. Lo que nunca se sube

- `.env`, tokens, contraseñas reales o datos personales de vendedores/clientes.
- `node_modules/`, `build/`, `pb_data/` (ya están en `.gitignore`).

## 5. Comandos útiles

```bash
make help        # lista todos los comandos
make reset       # borra la base local y la vuelve a crear desde data/seed
make logs        # logs del backend
```
