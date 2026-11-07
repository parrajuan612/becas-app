CREATE TABLE IF NOT EXISTS usuarios (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  correo TEXT NOT NULL UNIQUE,
  nombre TEXT NOT NULL,
  password_hash TEXT NOT NULL,
  creado_en DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS torneos (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  nombre TEXT NOT NULL UNIQUE COLLATE NOCASE,
  valor_inscripcion INTEGER NOT NULL CHECK (valor_inscripcion > 0),
  fecha_limite_pago DATE NOT NULL,
  activo INTEGER NOT NULL DEFAULT 1,
  creado_en DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS cargas (
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

CREATE TABLE IF NOT EXISTS registros (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  carga_id INTEGER NOT NULL REFERENCES cargas(id) ON DELETE CASCADE,
  fila INTEGER NOT NULL,
  nombre TEXT, numero TEXT, correo TEXT, club TEXT,
  fecha_nacimiento TEXT, torneo TEXT, porcentaje_beca TEXT, autoriza TEXT,
  numero_normalizado TEXT,
  torneo_id INTEGER REFERENCES torneos(id),
  estado TEXT NOT NULL CHECK (estado IN ('valido','invalido','omitido_duplicado','omitido_sin_autorizacion')),
  errores TEXT NOT NULL DEFAULT '[]'
);
CREATE INDEX IF NOT EXISTS idx_registros_carga ON registros(carga_id, estado);

CREATE TABLE IF NOT EXISTS becas (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  registro_id INTEGER NOT NULL UNIQUE REFERENCES registros(id),
  carga_id INTEGER NOT NULL REFERENCES cargas(id),
  torneo_id INTEGER NOT NULL REFERENCES torneos(id),
  codigo TEXT NOT NULL UNIQUE,
  nombre TEXT NOT NULL,
  nombre_normalizado TEXT NOT NULL,
  numero TEXT NOT NULL,
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
  UNIQUE (numero, torneo_id, nombre_normalizado)
);

CREATE TABLE IF NOT EXISTS notificaciones (
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
