# Evaluación del prototipo (Fase F5)

**Fecha:** 7 de noviembre de 2026 · **Archivo de prueba:** `ejemplos/becas_200.xlsx` (200 filas, semilla fija)
**Entorno:** Linux, Go 1.26, SQLite (`modernc.org/sqlite`), latencia simulada por defecto (50–200 ms), `WORKERS=5`.
**Método:** API real levantada desde cero (base de datos vacía) en dos escenarios, más un recorrido manual de la interfaz en Chrome (capturas en `docs/capturas/`).

- Escenario A: `SIM_FAILURE_RATE=0`
- Escenario B: `SIM_FAILURE_RATE=0.05`

## 1. Criterios de aceptación (SDD §3.3)

| # | Criterio | Resultado obtenido | Estado |
|---|---|---|---|
| 1 | La previsualización muestra exactamente los conteos esperados | 200 filas: **170 válidas, 20 inválidas, 5 duplicadas, 5 sin autorización** (A y B) | Cumple |
| 2 | Cada fila válida tiene un PDF y dos notificaciones | 170 becas, **170 PDF** en disco, **340 notificaciones**; 100 de 100 PDF de una muestra se descargan por la API (HTTP 200) | Cumple |
| 3 | Con tasa de fallo 0, el 100 % queda `enviado` | **340 / 340 enviadas**, 0 fallidas | Cumple |
| 4 | Con tasa 0,05 aparecen fallidos y al reintentar bajan y `intentos` aumenta | Fallidas por ronda de reintento: **11 → 2 → 0**; el máximo de intentos observado fue **3** (nunca se superó) | Cumple |
| 5 | Subir el mismo archivo otra vez produce 0 envíos nuevos | **0 válidas**, 175 `omitido_duplicado`, 20 inválidas y 5 sin autorización (conservan su estado); el total de becas siguió en **170** | Cumple |
| 6 | El panel muestra el tiempo de procesamiento y es menor de 60 s | **9,9 s** (≈ 58 ms por registro), visible en el panel | Cumple |

> Nota sobre el criterio 5: el SDD v1.0 decía "todas las filas quedan `omitido_duplicado`". Se corrigió en la v1.1, porque las inválidas y las sin autorización no cambian de estado.

## 2. Requisitos no funcionales

| ID | Requisito | Evidencia | Estado |
|---|---|---|---|
| RNF-01 | 200 registros en menos de 60 s | 9,9 s con latencia simulada activa | Cumple |
| RNF-02 | Seguridad | bcrypt costo 12, JWT de 8 h, `.env` ignorado por git y `.env.example` incluido, 0 archivos de secretos versionados | Cumple con una observación (ver §4) |
| RNF-03 | Usabilidad | Interfaz en español, errores en lenguaje claro (capturas 02, 05, 08) | Cumple |
| RNF-04 | Portabilidad | Driver SQLite sin CGO (`modernc.org/sqlite`); no se probó en Windows ni macOS | Parcial (sin verificar en otros SO) |
| RNF-05 | Mantenibilidad | Capas handler → servicio → repositorio; envío tras la interfaz `Notifier` | Cumple |
| RNF-06 | Trazabilidad | Cada notificación guarda intentos, último error, fecha de envío y mensaje | Cumple |
| RNF-07 | Calidad | `go vet` sin errores, `go test ./...` con **28 pruebas, 0 fallos**, `tsc --noEmit` sin errores | Cumple |
| RNF-08 | Instalación | `docker compose up --build` levanta todo (verificado en F7: API, nginx con proxy de `/api`, flujo completo 170/20/5/5 y datos persistentes tras reiniciar el backend). Instalación manual con 4 comandos (README) | Cumple |

### Cobertura de pruebas (Go)

| Paquete | Cobertura |
|---|---|
| `service` (validación, duplicados, valores, código, envío, métricas) | 85,4 % |
| `excel` | 93,2 % |
| `pdf` | 98,2 % |
| `ejemplo` | 98,7 % |
| `http`, `repository`, `notifier`, `config`, `db` | 0 % directa (se ejercitan indirectamente desde `service` y con las pruebas contra la API) |

## 3. Prueba manual de la interfaz (Chrome)

Recorrido automatizado con el Chrome del equipo: login con clave incorrecta y correcta, torneo duplicado, carga de `becas_200.xlsx`, previsualización con pestañas, confirmación con barra de progreso, reintento de fallidos, historial con filtros y modal del mensaje, y panel. No hubo errores de consola salvo el 401 (login incorrecto) y el 409 (torneo duplicado), ambos esperados. En esa corrida: 9,4 s de procesamiento, 323 enviadas y 17 fallidas al primer pase, y 339 enviadas y 1 fallida tras un reintento (99,7 % de éxito).

## 4. Hallazgos y su resolución (Fase F6)

| # | Hallazgo | Resolución |
|---|---|---|
| 1 | Secreto JWT y clave de administrador por defecto sin aviso | El servidor registra una `ADVERTENCIA` al arrancar si usa los valores por defecto (`config.Advertencias()`, con prueba) |
| 2 | Sin pruebas a nivel HTTP | Nuevo `router_test.go`: `/health` y `/verificar` públicos, 8 rutas protegidas responden 401 sin token o con token falso, login correcto e incorrecto, JSON inválido (400), validación de torneo (400) y nombre duplicado (409) |
| 3 | El Excel de ejemplo dependía del año actual | Los años de nacimiento se generan entre hoy−18 y hoy−7 años: siempre dentro del rango 5–20. Sigue siendo reproducible (misma semilla) dentro del mismo año |
| 4 | Faltaban capturas de "Ver PDF" y de la vista móvil | Añadidas (`16-pdf-beca-1.png` y `m1`–`m5`). Se verificó que ninguna pantalla tiene scroll horizontal de página en 390 px |
| 5 | Menú poco usable en móvil | La barra de navegación pasa a una fila propia con scroll horizontal; el correo se oculta en pantallas pequeñas |
| 6 | Detalles de interfaz | "WhatsApp" bien escrito, separación en las barras por canal y atributos `autocomplete` en el login (la consola ya no emite avisos) |

**Resuelto en F7:** Docker Compose, README con instalación completa (Docker y manual) y SDD actualizado. **Sigue pendiente:** verificar en Windows y macOS (RNF-04).
