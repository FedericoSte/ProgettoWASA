# --- Stage 1: Compilazione del Frontend (Vue.js) ---
FROM node:20 AS web-builder
WORKDIR /app/webui

# Abilita Corepack per usare la versione corretta di Yarn definita nel package.json (Yarn v4.5.0)
RUN corepack enable

COPY webui/package.json webui/yarn.lock ./
COPY webui/.yarn ./.yarn
COPY webui/.yarnrc.yml ./

RUN yarn install --immutable --immutable-cache
COPY webui/ .
RUN yarn run build-embed

# --- Stage 2: Compilazione del Backend (Go) ---
FROM golang:1.21 AS go-builder
WORKDIR /src
COPY go.mod go.sum ./
COPY vendor/ ./vendor/
COPY . .
COPY --from=web-builder /app/webui/dist ./webui/dist

RUN go build -mod=vendor -tags webui -o webapi ./cmd/webapi/

# --- Stage 3: Immagine di Runtime Finale ---
FROM debian:bookworm-slim
WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/*

COPY --from=go-builder /src/webapi /app/webapi
RUN mkdir -p /app/data

EXPOSE 3000

CMD ["./webapi", "-db", "/app/data/database.db"]