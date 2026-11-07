# Imagen de producción: un solo binario Go (PocketBase v0.41.1 + API + web estática embebida).

# 1) Web estática (SvelteKit + adapter-static)
FROM node:24-bookworm-slim AS web
ENV COREPACK_ENABLE_DOWNLOAD_PROMPT=0
RUN corepack enable
WORKDIR /src/apps/web
COPY apps/web/package.json apps/web/pnpm-lock.yaml apps/web/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile
COPY apps/web/ ./
RUN pnpm build

# 2) Binario Go con la web embebida
FROM golang:1.27-alpine AS core
WORKDIR /src/backend
COPY backend/ ./
COPY --from=web /src/apps/web/build/ ./core/web/dist/
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/qhatu ./core && mkdir -p /out/pb_data

# 3) Imagen final mínima
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=core /out/qhatu /app/qhatu
COPY --from=core --chown=nonroot:nonroot /out/pb_data /app/pb_data
COPY data/seed/ /app/data/seed/
ENV APP_ENV=production \
    QHATU_SEED_DIR=/app/data/seed \
    TZ=America/Lima
EXPOSE 8090
VOLUME ["/app/pb_data"]
ENTRYPOINT ["/app/qhatu"]
CMD ["serve", "--http=0.0.0.0:8090", "--dir=/app/pb_data"]
