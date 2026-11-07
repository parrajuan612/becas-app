# SDD — Prototipo funcional "BecasApp"
**Software Design Document / Especificación técnica para desarrollo guiado (Claude Code)**

| Campo | Valor |
|---|---|
| Proyecto | Sistema de notificación automatizada de becas deportivas (prototipo funcional) |
| Curso / Actividad | Proyecto de Software — Actividad 4: Diseño del prototipo |
| Autor | Juan Camilo Parra Mancera |
| Versión | 1.2 (implementada) |
| Fecha | 7 de noviembre de 2026 |
| Stack | Go (backend) + React + Vite + TypeScript (frontend) + SQLite. Versiones reales: React 19, TypeScript 7, Tailwind 3 |

> **Nota de contexto:** este prototipo es una versión académica y simplificada de la solución formulada en la Actividad 2. Usa una academia **ficticia** ("Academia Deportiva Horizonte"), **datos inventados** y **envíos simulados** de WhatsApp y correo. No contiene código, datos ni credenciales de ninguna empresa real.

---

## 1. Propósito y alcance

### 1.1 Problema que valida el prototipo
Una academia deportiva otorga entre 100 y 200 becas por evento (visoria). Hoy el proceso es manual: se transcriben los datos, se elaboran los PDF uno a uno y se envían por WhatsApp personal, lo que toma de 2 a 3 días, provoca errores de digitación y causa bloqueos por spam.

### 1.2 Objetivo del prototipo
Demostrar de forma funcional el flujo completo:

**Excel estandarizado → validación → previsualización → generación de PDF por deportista → envío simulado (WhatsApp + correo) → historial, reintentos y métricas.**

### 1.3 Dentro del alcance
- Login de un administrador.
- CRUD del catálogo de torneos.
- Descarga de la plantilla Excel oficial.
- Carga de Excel, validación por fila y previsualización con errores marcados.
- Confirmación y procesamiento solo de las filas válidas.
- Generación de un PDF de beca por deportista (logo, datos, valores, fecha límite y código de verificación).
- Envío **simulado** por WhatsApp (con plantilla de mensaje) y por correo, con tasa de fallo configurable.
- Control de duplicados (mismo número + mismo torneo).
- Control de consentimiento (columna `autoriza`).
- Historial de notificaciones con estados, filtros y reintento de fallidos.
- Panel de métricas.
- Pruebas unitarias en Go y un Excel de ejemplo con 200 registros.
- Docker Compose más una instalación manual documentada.

### 1.4 Prioridad si falta tiempo (se pueden omitir)
Docker/nginx, `POST /cargas/{id}/cancelar`, `GET /auth/me`, promedio de ms por registro y barra de éxito por canal.

### 1.5 Fuera del alcance (fase siguiente, se documenta pero no se construye)
- Integración real con la WhatsApp Cloud API de Meta (queda una interfaz `Notifier` lista para implementarla).
- Envío real por SMTP.
- Chatwoot y el despliegue en VPS.
- Múltiples roles y recuperación de contraseña.

---

## 2. Glosario
| Término | Definición |
|---|---|
| Carga | Cada archivo Excel subido al sistema. Agrupa sus filas. |
| Registro | Una fila del Excel ya validada (válida o con errores). |
| Beca | Registro válido y confirmado, con su código de verificación y su PDF. |
| Notificación | Intento de comunicar una beca por un canal (WhatsApp o correo). |
| Plantilla de mensaje | Texto fijo con variables, equivalente a una plantilla aprobada por Meta (`notificacion_beca_v1`). |

---

## 3. Requisitos

### 3.1 Requisitos funcionales
| ID | Requisito | Prioridad |
|---|---|---|
| RF-01 | El sistema permite iniciar sesión con correo y contraseña de administrador. Las rutas de la API, excepto `/api/auth/login`, `/api/health` y `/api/verificar/{codigo}`, exigen autenticación. | Alta |
| RF-02 | El administrador puede crear, listar, editar y desactivar torneos (nombre único, valor de inscripción en COP y fecha límite de pago). | Alta |
| RF-03 | El administrador puede descargar la plantilla Excel oficial con los encabezados correctos y una fila de ejemplo. | Media |
| RF-04 | El administrador puede subir un archivo `.xlsx` (máximo 5 MB y 1.000 filas). | Alta |
| RF-05 | El sistema valida cada fila según las reglas de la sección 5 y guarda los errores de cada fila. | Alta |
| RF-06 | El sistema muestra una previsualización con el total de filas, las válidas, las que tienen error y las omitidas, y el motivo de cada error. | Alta |
| RF-07 | El administrador confirma la carga y solo se procesan las filas válidas. | Alta |
| RF-08 | Por cada beca se genera un PDF con el contenido de la sección 7. | Alta |
| RF-09 | Por cada beca se crean dos notificaciones (WhatsApp y correo) y se envían por los notificadores simulados. | Alta |
| RF-10 | No se notifica dos veces una beca con el mismo número, el mismo torneo y el mismo deportista (nombre normalizado), ni dentro del mismo archivo ni frente a cargas anteriores. Dos hermanos con el mismo celular de acudiente en un torneo **no** son duplicados. La verificación se repite al confirmar. | Alta |
| RF-11 | Si `autoriza = NO`, el registro no se notifica y queda como `omitido_sin_autorizacion`. | Alta |
| RF-12 | El historial lista las notificaciones con filtros por carga, torneo, canal y estado. | Alta |
| RF-13 | El administrador puede reintentar las notificaciones fallidas, de forma individual o en bloque por carga. Máximo 3 intentos por notificación (el envío inicial cuenta como el intento 1, es decir, 2 reintentos). | Media |
| RF-14 | El administrador puede descargar o ver el PDF de cualquier beca. | Media |
| RF-15 | El panel muestra métricas globales y por carga (sección 8.4). | Media |
| RF-16 | Existe un endpoint público `GET /api/verificar/{codigo}` que confirma si un código de beca es auténtico. | Baja |

### 3.2 Requisitos no funcionales
| ID | Requisito | Criterio verificable |
|---|---|---|
| RNF-01 | Rendimiento | 200 registros validados, con PDF generado y notificaciones simuladas, en **menos de 60 s** (con latencia simulada activa). |
| RNF-02 | Seguridad | Contraseñas con bcrypt (costo ≥ 10). JWT con expiración de 8 h. Ningún secreto en el repositorio; se usa `.env` y se incluye `.env.example`. |
| RNF-03 | Usabilidad | Interfaz en español, utilizable por personal no técnico. Todo error se explica en lenguaje claro. |
| RNF-04 | Portabilidad | Se ejecuta en Windows, macOS y Linux. SQLite con un driver **sin CGO** (`modernc.org/sqlite`). |
| RNF-05 | Mantenibilidad | Arquitectura por capas (handler → servicio → repositorio). El envío depende de una interfaz `Notifier` intercambiable. |
| RNF-06 | Trazabilidad | Cada notificación guarda intentos, último error, fecha de envío y el mensaje generado. |
| RNF-07 | Calidad | `go test ./...` pasa. Hay cobertura de validadores, duplicados y cálculo de valores. `go vet` y `tsc --noEmit` sin errores. |
| RNF-08 | Instalación | `docker compose up --build` levanta todo. La instalación manual se hace con menos de 6 comandos documentados. |

### 3.3 Criterios de aceptación (prueba con el Excel de ejemplo)
1. Al subir `ejemplos/becas_200.xlsx`, la previsualización muestra exactamente los conteos de la sección 10.2.
2. Al confirmar, cada fila válida tiene un PDF y dos notificaciones.
3. Con `SIM_FAILURE_RATE=0`, el 100 % de las notificaciones válidas quedan `enviado`.
4. Con `SIM_FAILURE_RATE=0.05`, aparecen fallidos. Al reintentar, los fallidos bajan y `intentos` aumenta.
5. Subir el mismo archivo por segunda vez produce 0 filas válidas y 0 envíos nuevos: las 170 válidas y las 5 duplicadas quedan `omitido_duplicado`; las inválidas y las sin autorización conservan su estado.
6. El panel muestra el tiempo de procesamiento de la carga y es menor de 60 s.

---

## 4. Plantilla Excel

Hoja: `becas`. La fila 1 son los encabezados (exactos y sin importar mayúsculas). Los datos empiezan en la fila 2.

| Columna | Tipo | Ejemplo | Obligatorio |
|---|---|---|---|
| `nombre` | texto | Santiago Gómez Ruiz | Sí |
| `numero` | texto | 3001234567 | Sí |
| `correo` | texto | acudiente@gmail.com | Sí |
| `club` | texto | Club Los Andes | Sí |
| `fecha_nacimiento` | fecha `DD/MM/AAAA` o fecha nativa de Excel | 14/03/2012 | Sí |
| `torneo` | texto (debe existir en el catálogo) | Copa Horizonte 2027 | Sí |
| `porcentaje_beca` | entero | 50 | Sí |
| `autoriza` | SI / NO | SI | Sí |

- Si falta una columna obligatoria o hay un encabezado desconocido, se **rechaza el archivo completo** con un mensaje claro que lista las columnas esperadas.
- Las filas completamente vacías se ignoran.

---

## 5. Reglas de validación y negocio

### 5.1 Validación por campo (se reportan **todos** los errores de una fila, no solo el primero)
| Campo | Regla | Mensaje de error |
|---|---|---|
| nombre | No vacío, entre 3 y 100 caracteres, solo letras Unicode (`\p{L}`), espacios, apóstrofe y guion | "Nombre vacío o inválido" |
| numero | Se normaliza quitando espacios, guiones, paréntesis y `+`; luego se quita el prefijo `57` solo si quedan 12 dígitos. El resultado debe tener **10 dígitos y empezar por 3** (celular colombiano). Se guarda en formato E.164: `+573001234567` | "Número de celular inválido (debe tener 10 dígitos y empezar por 3)" |
| correo | Válido según `net/mail.ParseAddress` **y** `addr.Address` igual a la entrada (rechaza `Juan <a@b.com>`), con dominio que contenga un punto | "Correo electrónico inválido" |
| club | No vacío, entre 2 y 80 caracteres | "Club vacío" |
| fecha_nacimiento | Fecha válida que no esté en el futuro. **Edad entre 5 y 20 años** a la fecha de la carga *(supuesto)* | "Fecha de nacimiento inválida o fuera del rango de edad (5–20 años)" |
| torneo | Coincide con un torneo **activo** del catálogo (comparación en Go con `strings.EqualFold` y `TrimSpace`; el `NOCASE` de SQLite solo cubre ASCII) | "El torneo no existe o está inactivo" |
| porcentaje_beca | Entero entre **1 y 100** | "El porcentaje debe ser un entero entre 1 y 100" |
| autoriza | `SI`, `SÍ`, `NO` (sin importar mayúsculas) | "Autoriza debe ser SI o NO" |

### 5.2 Clasificación de cada fila
Las reglas se aplican en este orden:
1. Tiene errores de validación → `invalido`.
2. `autoriza = NO` → `omitido_sin_autorizacion`.
3. Ya existe otra fila **en el mismo archivo** con el mismo número, torneo y nombre normalizado (minúsculas, sin tildes ni espacios extra) → la primera es válida y las siguientes quedan `omitido_duplicado`.
4. Ya existe una **beca en la base de datos** con el mismo número, torneo y nombre normalizado → `omitido_duplicado`.
5. En cualquier otro caso → `valido`.

### 5.3 Cálculo de valores
- `valor_inscripcion` viene del torneo (COP, entero).
- `descuento = (valor_inscripcion × porcentaje + 50) / 100` (aritmética entera, redondeo half-up; sin flotantes)
- `valor_a_pagar = valor_inscripcion − descuento`
- `fecha_limite_pago` viene del torneo.
- Los valores se muestran en formato colombiano: `$ 1.250.000`.

### 5.4 Código de verificación
Formato: `BEC-AAAA-XXXXXX`, donde AAAA es el año y XXXXXX son 6 caracteres alfanuméricos en mayúscula que excluyen 0, O, 1 e I. Es único (índice UNIQUE) y se genera con `crypto/rand`.

---

## 6. Envío simulado

### 6.1 Interfaz
```go
type Canal string // "whatsapp" | "correo"

type Mensaje struct {
    Destino   string            // +573001234567 o correo
    Plantilla string            // "notificacion_beca_v1"
    Variables map[string]string // nombre, torneo, porcentaje, valor_a_pagar, fecha_limite, codigo
    AdjuntoPDF string           // ruta del PDF
}

type Notifier interface {
    Canal() Canal
    Enviar(ctx context.Context, m Mensaje) (idExterno string, err error)
}
```
Implementaciones: `WhatsAppSimulado` y `CorreoSimulado`. Un futuro `WhatsAppMeta` implementaría la misma interfaz sin cambiar el resto del sistema.

### 6.2 Comportamiento de la simulación
- Latencia aleatoria entre `SIM_MIN_MS` y `SIM_MAX_MS` (por defecto de 50 a 200 ms).
- Falla con probabilidad `SIM_FAILURE_RATE` (por defecto `0.05`). Los errores se eligen al azar entre: `"timeout del proveedor"`, `"límite de tasa excedido (429)"` y `"número no registrado en WhatsApp"` (este último solo aplica a WhatsApp).
- En caso de éxito devuelve un id externo simulado (`wamid.SIM...` o `mail.SIM...`).
- Cada envío escribe en el log del backend: `[SIMULADO][whatsapp] -> +5730... plantilla=notificacion_beca_v1 OK id=...`.

### 6.3 Plantilla de mensaje `notificacion_beca_v1`
```
Hola. La Academia Deportiva Horizonte le informa que {{nombre}}
ha obtenido una beca del {{porcentaje}}% para el torneo {{torneo}}.
Valor a pagar: {{valor_a_pagar}}. Fecha límite de pago: {{fecha_limite}}.
Código de verificación: {{codigo}}. Adjuntamos el documento oficial de la beca.
```
El correo usa el asunto `"Beca deportiva – {{torneo}}"` y el mismo cuerpo. El mensaje final generado se guarda en la notificación.

### 6.4 Procesamiento concurrente
- Al confirmar una carga, el servicio procesa las becas con un **worker pool** de `WORKERS` goroutines (por defecto 5) para simular un rate limit respetuoso.
- SQLite se abre en modo WAL con `busy_timeout` y las escrituras pasan por una única conexión (`SetMaxOpenConns(1)`) para evitar `database is locked`.
- Al arrancar el servidor, las cargas en `procesando` se reanudan: se reenvían sus notificaciones `pendiente`.
- Antes de crear cada beca se re-verifica el duplicado; si el `UNIQUE` falla (otra carga ganó), el registro pasa a `omitido_duplicado`.
- El progreso se deriva del conteo de notificaciones no pendientes de la carga.
- El procesamiento corre en segundo plano. El frontend consulta el estado de la carga cada 1 s (`GET /api/cargas/{id}`) hasta que pase a `completada`.

### 6.5 Estados de una notificación
`pendiente → enviado` · `pendiente → fallido` · `fallido → (reintento) → enviado | fallido`.
Una notificación con 3 intentos fallidos (el inicial más 2 reintentos) queda `fallido` y ya no se puede reintentar.

---

## 7. PDF de la beca
Se genera con `github.com/go-pdf/fpdf`, en tamaño carta y vertical. Las fuentes core de fpdf no soportan UTF-8: usar `UnicodeTranslatorFromDescriptor("")` (o una fuente TTF) para que tildes y ñ se vean bien. Se guarda en `data/pdfs/{codigo}.pdf`.

Contenido:
1. **Encabezado:** logo ficticio (PNG incluido en `backend/assets/logo.png`, generado de forma simple: un círculo con las iniciales "AH") y el nombre "Academia Deportiva Horizonte".
2. **Título:** "Certificado de Beca Deportiva".
3. **Datos del deportista:** nombre, fecha de nacimiento, edad, club de origen.
4. **Datos de la beca:** torneo, porcentaje de beca.
5. **Tabla de valores:** valor de la inscripción, descuento, **valor a pagar** (resaltado), fecha límite de pago.
6. **Código de verificación** visible, con el texto "Verifique este código en /api/verificar/{codigo}".
7. **Pie:** fecha de emisión y la leyenda "Documento generado automáticamente — prototipo académico".

---

## 8. Arquitectura

### 8.1 Vista general
```
[React + Vite + TS]  --HTTP/JSON (JWT)-->  [API Go (chi)]
                                              ├── handlers  (HTTP)
                                              ├── services  (validación, becas, envío, métricas)
                                              ├── notifier  (WhatsAppSimulado, CorreoSimulado)
                                              ├── pdf       (generación fpdf)
                                              ├── excel     (lectura/plantilla excelize)
                                              └── repository (SQLite, modernc.org/sqlite)
                                                       └── data/becas.db, data/pdfs/
```

### 8.2 Tecnologías
| Capa | Tecnología |
|---|---|
| Backend | Go 1.22+, `go-chi/chi/v5`, `xuri/excelize/v2`, `go-pdf/fpdf`, `modernc.org/sqlite`, `golang-jwt/jwt/v5`, `golang.org/x/crypto/bcrypt` |
| Frontend | React 18, Vite, TypeScript (modo strict), React Router, Tailwind CSS |
| Base de datos | SQLite (archivo `data/becas.db`). Las migraciones se ejecutan automáticamente al arrancar. |
| Infraestructura | Docker Compose: `backend` (Go) y `frontend` (build de Vite servido por nginx, con proxy de `/api` al backend) |

### 8.3 Modelo de datos (SQLite)
```sql
CREATE TABLE usuarios (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  correo TEXT NOT NULL UNIQUE,
  nombre TEXT NOT NULL,
  password_hash TEXT NOT NULL,
  creado_en DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE torneos (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  nombre TEXT NOT NULL UNIQUE COLLATE NOCASE,
  valor_inscripcion INTEGER NOT NULL CHECK (valor_inscripcion > 0),
  fecha_limite_pago DATE NOT NULL,
  activo INTEGER NOT NULL DEFAULT 1,
  creado_en DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE cargas (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  nombre_archivo TEXT NOT NULL,
  usuario_id INTEGER NOT NULL REFERENCES usuarios(id),
  estado TEXT NOT NULL CHECK (estado IN ('previsualizada','procesando','completada','cancelada')),
  total_filas INTEGER NOT NULL DEFAULT 0,
  validas INTEGER NOT NULL DEFAULT 0,
  invalidas INTEGER NOT NULL DEFAULT 0,
  omitidas_duplicado INTEGER NOT NULL DEFAULT 0,
  omitidas_sin_autorizacion INTEGER NOT NULL DEFAULT 0,
  inicio_proceso DATETIME,
  fin_proceso DATETIME,
  creado_en DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE registros (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  carga_id INTEGER NOT NULL REFERENCES cargas(id) ON DELETE CASCADE,
  fila INTEGER NOT NULL,                 -- número de fila en el Excel
  nombre TEXT, numero TEXT, correo TEXT, club TEXT,
  fecha_nacimiento TEXT, torneo TEXT, porcentaje_beca TEXT, autoriza TEXT,  -- valores originales
  numero_normalizado TEXT,
  torneo_id INTEGER REFERENCES torneos(id),
  estado TEXT NOT NULL CHECK (estado IN ('valido','invalido','omitido_duplicado','omitido_sin_autorizacion')),
  errores TEXT NOT NULL DEFAULT '[]'     -- JSON array de strings
);

CREATE TABLE becas (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  registro_id INTEGER NOT NULL UNIQUE REFERENCES registros(id),
  carga_id INTEGER NOT NULL REFERENCES cargas(id),
  torneo_id INTEGER NOT NULL REFERENCES torneos(id),
  codigo TEXT NOT NULL UNIQUE,
  nombre TEXT NOT NULL,
  numero TEXT NOT NULL,                  -- E.164
  correo TEXT NOT NULL,
  club TEXT NOT NULL,
  fecha_nacimiento DATE NOT NULL,
  porcentaje INTEGER NOT NULL,
  valor_inscripcion INTEGER NOT NULL,
  descuento INTEGER NOT NULL,
  valor_a_pagar INTEGER NOT NULL,
  fecha_limite_pago DATE NOT NULL,
  ruta_pdf TEXT,
  creado_en DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  nombre_normalizado TEXT NOT NULL,      -- minúsculas, sin tildes
  UNIQUE (numero, torneo_id, nombre_normalizado)  -- regla anti-duplicados (RF-10)
);

CREATE TABLE notificaciones (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  beca_id INTEGER NOT NULL REFERENCES becas(id) ON DELETE CASCADE,
  canal TEXT NOT NULL CHECK (canal IN ('whatsapp','correo')),
  destino TEXT NOT NULL,
  plantilla TEXT NOT NULL,
  mensaje TEXT NOT NULL,
  estado TEXT NOT NULL CHECK (estado IN ('pendiente','enviado','fallido')),
  intentos INTEGER NOT NULL DEFAULT 0,
  ultimo_error TEXT,
  id_externo TEXT,
  enviado_en DATETIME,
  actualizado_en DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (beca_id, canal)
);
```

### 8.4 Métricas del panel
- Globales: total de becas, notificaciones `enviado`, `fallido` y `pendiente` por canal, registros omitidos por duplicado y por falta de autorización, tasa de éxito (%).
- Por carga: los conteos anteriores, la duración (`fin_proceso − inicio_proceso`) y el promedio de ms por registro.
- Últimas 5 cargas con su estado.

### 8.5 API REST (prefijo `/api`, JSON, errores con el formato `{ "error": "mensaje" }`)
| Método | Ruta | Descripción | Auth |
|---|---|---|---|
| GET | `/health` | Estado del servicio | No |
| POST | `/auth/login` | `{correo, password}` → `{token, usuario}` | No |
| GET | `/auth/me` | Usuario actual | Sí |
| GET | `/torneos` | Listar (`?activos=true`) | Sí |
| POST | `/torneos` | Crear | Sí |
| PUT | `/torneos/{id}` | Editar | Sí |
| DELETE | `/torneos/{id}` | Desactivar (borrado lógico) | Sí |
| GET | `/plantilla` | Descarga `plantilla_becas.xlsx` (mismo mecanismo de blob) | Sí |
| POST | `/cargas` | multipart `archivo` (límite 5 MB con `http.MaxBytesReader`) → valida y devuelve la carga con su resumen (estado `previsualizada`) | Sí |
| GET | `/cargas` | Listar cargas | Sí |
| GET | `/cargas/{id}` | Detalle, conteos y estado | Sí |
| GET | `/cargas/{id}/registros` | Registros (`?estado=`, paginado `?page=&size=`) | Sí |
| POST | `/cargas/{id}/confirmar` | Inicia el procesamiento asíncrono → 202 | Sí |
| POST | `/cargas/{id}/cancelar` | Cancela una carga en previsualización | Sí |
| POST | `/cargas/{id}/reintentar` | Reintenta todas las fallidas de la carga | Sí |
| GET | `/notificaciones` | Historial (`?carga_id=&torneo_id=&canal=&estado=&page=&size=`) | Sí |
| POST | `/notificaciones/{id}/reintentar` | Reintento individual | Sí |
| GET | `/becas/{id}/pdf` | Descarga el PDF (el frontend usa `fetch` con el JWT y abre un blob; un `<a href>` no envía el header) | Sí |
| GET | `/metricas` | Métricas globales | Sí |
| GET | `/verificar/{codigo}` | `{valido, nombre, torneo, porcentaje}`; el nombre se enmascara (nombre de pila + inicial del apellido) por ser menores | No |

### 8.6 Estructura del repositorio
```
majestic/                                # raíz del repositorio
├── backend/
│   ├── cmd/
│   │   ├── api/main.go              # servidor
│   │   └── generar-ejemplo/main.go  # crea ejemplos/becas_200.xlsx
│   ├── internal/
│   │   ├── config/      # lectura de .env
│   │   ├── db/          # conexión + migraciones + seed
│   │   ├── domain/      # structs y estados
│   │   ├── repository/
│   │   ├── service/     # validacion.go, cargas.go, envio.go, metricas.go
│   │   ├── notifier/    # whatsapp_simulado.go, correo_simulado.go
│   │   ├── pdf/
│   │   ├── excel/
│   │   └── http/        # router, handlers, middleware JWT
│   ├── assets/logo.png
│   ├── Dockerfile
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── api/         # cliente fetch tipado
│   │   ├── pages/       # Login, Panel, Torneos, NuevaCarga, DetalleCarga, Historial
│   │   ├── components/
│   │   └── types/
│   ├── Dockerfile
│   └── nginx.conf
├── ejemplos/becas_200.xlsx
├── docs/ (SDD.md, capturas/)
├── docker-compose.yml
├── .env.example
└── README.md
```

### 8.7 Configuración (`.env.example`)
```
PORT=8080
DB_PATH=./data/becas.db
PDF_DIR=./data/pdfs
JWT_SECRET=cambia-esto-en-tu-entorno
ADMIN_CORREO=admin@horizonte.test
ADMIN_PASSWORD=Admin123*
SIM_FAILURE_RATE=0.05
SIM_MIN_MS=50
SIM_MAX_MS=200
WORKERS=5
CORS_ORIGIN=http://localhost:5173
```

### 8.8 Datos semilla (al primer arranque)
- Usuario administrador tomado de `ADMIN_CORREO` y `ADMIN_PASSWORD`.
- Torneos:
  - Copa Horizonte 2027 — $ 350.000 — fecha límite 15/02/2027
  - Torneo Interclubes Sub-15 — $ 280.000 — fecha límite 28/02/2027
  - Festival Infantil de Verano — $ 150.000 — fecha límite 31/01/2027

---

## 9. Interfaz de usuario (pantallas)
| # | Pantalla | Contenido |
|---|---|---|
| P1 | Login | Logo, correo, contraseña y mensaje de error. |
| P2 | Panel | Tarjetas con métricas, barra de éxito por canal y tabla de las últimas cargas. |
| P3 | Torneos | Tabla y formulario modal para crear o editar; botón para desactivar. |
| P4 | Nueva carga | Botón para descargar la plantilla, zona de arrastrar y soltar el `.xlsx` y botón "Validar archivo". |
| P5 | Previsualización / detalle de carga | Tarjetas de conteo (válidas, inválidas, duplicadas, sin autorización). Tabla con pestañas por estado; las filas inválidas se resaltan en rojo con sus errores. Botones "Confirmar y enviar" y "Cancelar". Durante el proceso se muestra una barra de progreso; al terminar, la duración y el botón "Reintentar fallidos". |
| P6 | Historial | Filtros (carga, torneo, canal, estado), tabla de notificaciones (destino, canal, estado, intentos, error), botones "Ver PDF" y "Reintentar" por fila, y un modal con el mensaje enviado. |

Diseño limpio, responsive, en español, con colores de estado (verde: enviado, rojo: fallido/inválido, ámbar: omitido, gris: pendiente).

---

## 10. Pruebas y evidencias

### 10.1 Pruebas unitarias en Go (obligatorias)
- `validacion_test.go`: tabla de casos para el número (con +57, con espacios, 9 dígitos, que empieza por 2), el correo, la fecha (futura, edad 4, edad 21, formato Excel serial), el porcentaje (0, 101, "50%", 50) y autoriza (sí, Sí, no, x).
- `duplicados_test.go`: duplicado dentro del archivo, duplicado frente a la BD, mismo número en otro torneo (no es duplicado) y hermanos con el mismo celular y distinto nombre (no es duplicado).
- `valores_test.go`: cálculo del descuento y del valor a pagar, incluido el redondeo.
- `codigo_test.go`: formato del código (la unicidad la garantiza el `UNIQUE` con reintento al generar).
- `envio_test.go`: con `SIM_FAILURE_RATE=0` todo queda enviado; con `1` todo queda fallido; un reintento incrementa los intentos y respeta el máximo de 3.

### 10.2 Excel de ejemplo `ejemplos/becas_200.xlsx`
Lo genera `go run ./cmd/generar-ejemplo` con nombres colombianos ficticios y una semilla fija (resultado reproducible). Composición exacta:
| Tipo de fila | Cantidad |
|---|---|
| Válidas | 170 |
| Número inválido | 8 |
| Correo inválido | 5 |
| Porcentaje fuera de rango | 3 |
| Torneo inexistente | 4 |
| `autoriza = NO` | 5 |
| Duplicadas (mismo número + torneo de otra fila válida) | 5 |
| **Total** | **200** |

Resultado esperado en la previsualización: 170 válidas, 20 inválidas, 5 omitidas por duplicado y 5 omitidas sin autorización.

---

## 11. Plan de implementación por fases (ciclo de vida del prototipo)
Cada fase termina con un commit y con capturas guardadas en `docs/capturas/` para el informe.

| Fase | Etapa del ciclo de vida | Entregable |
|---|---|---|
| F0 | Recolección de requisitos | Este SDD (secciones 3 a 5). |
| F1 | Diseño rápido | Esqueleto del repo, migraciones, seed, `/health`, login y layout del frontend con navegación. |
| F2 | Construcción I | CRUD de torneos, plantilla Excel, carga, validación y previsualización (P3, P4, P5 sin envío). Pruebas de validación. |
| F3 | Construcción II | Confirmación, becas, PDF, notifier simulado, worker pool, historial y reintentos (P5 completa, P6). |
| F4 | Construcción III | Métricas y panel (P2), endpoint de verificación, generador del Excel de ejemplo y resto de pruebas. |
| F5 | Evaluación | Ejecutar los criterios de aceptación de la sección 3.3 y anotar los resultados (tiempos y conteos). |
| F6 | Refinamiento | Correcciones según la evaluación, mejoras de UX y mensajes de error. |
| F7 | Entrega | Docker Compose (opcional si falta tiempo; la instalación manual es obligatoria), README con instalación (Docker y manual), capturas finales y subida a GitHub. |

---

## 12. Instrucciones para Claude Code
Pegar en la terminal, dentro de la raíz del repositorio (`majestic/`, ya con `git init`) con este archivo en `docs/SDD.md`:

> Lee `docs/SDD.md` completo. Implementa el proyecto **fase por fase** según la sección 11, empezando por F1. Al terminar cada fase: ejecuta `go vet ./...`, `go test ./...` y `npm run build` en frontend, corrige los errores, haz un commit con el mensaje `F<n>: <resumen>` y detente para que yo revise antes de seguir. No agregues funcionalidades fuera de la sección 1.3. Usa `modernc.org/sqlite` (sin CGO). No escribas secretos en el código. Si algo del SDD es ambiguo, pregúntame antes de decidir.

---

## 13. Supuestos a confirmar
1. Edad válida entre 5 y 20 años.
2. Solo números celulares colombianos (10 dígitos que empiezan por 3).
3. El mensaje se dirige al número y correo registrados (del acudiente); no existe un campo separado para el nombre del acudiente.
4. Máximo 3 intentos por notificación (incluye el envío inicial).
6. Duplicado = mismo número + torneo + nombre normalizado (permite hermanos con el mismo acudiente).
5. Academia ficticia: "Academia Deportiva Horizonte".

---

## 14. Desviaciones respecto al diseño original (resultado de la implementación)
1. **Versiones:** el scaffold de Vite instaló React 19 y TypeScript 7 (el diseño decía React 18).
2. **Mensaje del notificador:** `Mensaje` incluye además `Asunto` y `Cuerpo` (el texto final generado) para el log y el correo.
3. **`GET /api/cargas/{id}`** devuelve también el progreso (`notif_total`, `notif_pendientes`, `notif_enviadas`, `notif_fallidas`), que el frontend consulta cada 1 s.
4. **`GET /api/metricas`** incluye las últimas 5 cargas con `duracion_ms` y `ms_por_registro`. La *tasa de éxito* se define como enviadas ÷ (enviadas + fallidas), y los omitidos solo cuentan cargas confirmadas.
5. **Duplicados:** la clave es número + torneo + nombre normalizado (permite hermanos con el mismo celular de acudiente).
6. **Reanudación:** al arrancar, las cargas en `procesando` se reanudan automáticamente.
7. **Desarrollo:** el proxy de Vite apunta a `VITE_API_TARGET` (por defecto `http://localhost:8080`). En Docker, nginx hace el proxy de `/api`.
8. **Excel de ejemplo:** los años de nacimiento son relativos al año actual, para que siga dentro del rango de edad válido.
9. **Docker:** el puerto publicado se configura con `APP_PORT` (por defecto 8081) y los datos viven en el volumen `becas-data`.
10. **Opcionales (§1.4):** se implementaron todos (cancelar carga, `/auth/me`, Docker/nginx, ms por registro y barras por canal).
