# --- Stage 1: Compilazione del Frontend (Vue.js) ---
FROM node:20 AS web-builder
WORKDIR /app/webui
COPY webui/package.json webui/yarn.lock ./
RUN yarn install --immutable --immutable-cache
COPY webui/ .
RUN yarn run build-embed

# --- Stage 2: Compilazione del Backend (Go) ---
FROM golang:1.21 AS go-builder
WORKDIR /src
COPY go.mod go.sum ./
COPY vendor/ ./vendor/
COPY . .
# Copia la build del frontend prodotta nel primo stage nella posizione attesa da Go
COPY --from=web-builder /app/webui/dist ./webui/dist

# Compilazione del binario con il tag webui abilitato
RUN go build -mod=vendor -tags webui -o webapi ./cmd/webapi/

# --- Stage 3: Immagine di Runtime Finale (Leggera) ---
FROM debian:bookworm-slim
WORKDIR /app

# Installa le dipendenze di sistema minime (CA certificates e SQLite)
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/*

# Copia l'eseguibile compilato dallo stage Go
COPY --from=go-builder /src/webapi /app/webapi

# Crea una cartella per i dati persistenti (es. database)
RUN mkdir -p /app/data

EXPOSE 3000

# Avvia l'applicazione configurando il database nel volume persistente
CMD ["./webapi", "-db", "/app/data/database.db"]