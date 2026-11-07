# --- compilación (sin CGO: SQLite con modernc.org/sqlite) ---
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# --- imagen final ---
FROM alpine:3.20
RUN adduser -D -u 10001 app && mkdir -p /app/data/pdfs && chown -R app /app
WORKDIR /app
COPY --from=build /out/api /app/api
USER app
ENV PORT=8080 DB_PATH=/app/data/becas.db PDF_DIR=/app/data/pdfs
VOLUME /app/data
EXPOSE 8080
CMD ["/app/api"]
