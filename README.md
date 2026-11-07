# BecasApp — prototipo de notificación automatizada de becas deportivas

Proyecto universitario (Proyecto de Software — Actividad 4). Prototipo funcional con una academia **ficticia** ("Academia Deportiva Horizonte"), datos inventados y envíos **simulados** de WhatsApp y correo.

**Flujo:** Excel estandarizado → validación → previsualización → PDF por deportista → envío simulado (WhatsApp + correo) → historial, reintentos y métricas.

**Stack:** Go (chi, excelize, fpdf, SQLite sin CGO, JWT) · React + Vite + TypeScript + Tailwind · Docker Compose.

| Documento | Contenido |
|---|---|
| [`docs/SDD.md`](docs/SDD.md) | Diseño: requisitos, reglas, arquitectura, API y plan por fases |
| [`docs/EVALUACION.md`](docs/EVALUACION.md) | Resultados de los criterios de aceptación y de los requisitos no funcionales |
| [`docs/capturas/`](docs/capturas) | Capturas de la interfaz (escritorio y móvil) y del PDF |

## Instalación con Docker (recomendada)

```bash
cp .env.example .env          # opcional: cambie JWT_SECRET y ADMIN_PASSWORD
docker compose up --build
```

Abra <http://localhost:8081> (puerto configurable con `APP_PORT` en `.env`). Los datos (SQLite y PDF) se guardan en el volumen `becas-data`.

## Instalación manual (Go 1.22+ y Node 20+)

```bash
cp .env.example .env
cd backend && go run ./cmd/api              # terminal 1 — API en :8080
cd frontend && npm install && npm run dev   # terminal 2 — interfaz en http://localhost:5173
```

Si el puerto 8080 está ocupado, cambie `PORT` en `.env` y arranque el frontend con `VITE_API_TARGET=http://localhost:<PORT> npm run dev`.

**Acceso inicial:** el definido en `.env` (`ADMIN_CORREO` / `ADMIN_PASSWORD`; por defecto `admin@horizonte.test` / `Admin123*`). Cámbielos antes de desplegar: el servidor emite una advertencia si usa los valores por defecto.

## Cómo probarlo

1. Genere el Excel de ejemplo (200 filas, reproducible): `cd backend && go run ./cmd/generar-ejemplo` → `ejemplos/becas_200.xlsx` (ya incluido).
2. Entre a **Nueva carga**, suba `becas_200.xlsx` y pulse **Validar archivo**. Resultado esperado: **170 válidas, 20 inválidas, 5 duplicadas y 5 sin autorización**.
3. Pulse **Confirmar y enviar**. Se generan 170 PDF y 340 notificaciones simuladas (≈ 10 s). Revise el **Historial**, use **Reintentar fallidos** y mire el **Panel**.
4. Suba el mismo archivo otra vez: 0 válidas y 175 duplicadas (no se vuelve a notificar a nadie).

Con `SIM_FAILURE_RATE=0` todo se envía; con `0.05` (por defecto) aparecen fallos que se recuperan al reintentar (máximo 3 intentos por notificación).

## Pruebas

```bash
cd backend && go vet ./... && go test ./...     # 32 pruebas
cd frontend && npx tsc --noEmit -p tsconfig.app.json && npm run build
```

## Configuración (`.env`)

| Variable | Descripción | Por defecto |
|---|---|---|
| `PORT` | Puerto de la API | `8080` |
| `DB_PATH`, `PDF_DIR` | Base SQLite y carpeta de PDF | `./data/becas.db`, `./data/pdfs` |
| `JWT_SECRET` | Secreto de firma (token de 8 h) | valor de desarrollo |
| `ADMIN_CORREO`, `ADMIN_PASSWORD` | Administrador creado en el primer arranque | `admin@horizonte.test` / `Admin123*` |
| `SIM_FAILURE_RATE` | Probabilidad de fallo del envío simulado | `0.05` |
| `SIM_MIN_MS`, `SIM_MAX_MS` | Latencia simulada | `50`, `200` |
| `WORKERS` | Goroutines del pool de envío | `5` |
| `CORS_ORIGIN` | Origen permitido en desarrollo | `http://localhost:5173` |
| `APP_PORT` | Solo Docker: puerto publicado | `8081` |

## Estructura

```
backend/   API en Go (cmd/, internal/{config,db,domain,repository,service,notifier,pdf,excel,http,ejemplo}, assets/)
frontend/  React + Vite + TypeScript (src/{api,pages,components,types})
ejemplos/  becas_200.xlsx
docs/      SDD, evaluación y capturas
```

## Fuera del alcance (siguiente fase)

WhatsApp Cloud API de Meta (basta implementar la interfaz `Notifier`), SMTP real, Chatwoot, despliegue en VPS, múltiples roles y recuperación de contraseña.

> Prototipo académico: no contiene datos, código ni credenciales de ninguna empresa real.
