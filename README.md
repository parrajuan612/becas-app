# BecasApp (prototipo académico)

Notificación automatizada de becas deportivas. Ver `docs/SDD.md`.

## Ejecución manual (desarrollo)
```bash
cp .env.example .env
cd backend && go run ./cmd/api          # API en :8080
cd frontend && npm install && npm run dev   # UI en :5173
```
Usuario inicial: el definido en `ADMIN_CORREO` / `ADMIN_PASSWORD`.

Estado: F3 (confirmación, becas, PDF, envío simulado, historial y reintentos) completada.
