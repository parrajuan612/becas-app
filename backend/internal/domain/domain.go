// Package domain define las entidades y estados compartidos.
package domain

import "time"

type Usuario struct {
	ID     int64  `json:"id"`
	Correo string `json:"correo"`
	Nombre string `json:"nombre"`
}

// Estados de un registro del Excel.
const (
	RegistroValido                 = "valido"
	RegistroInvalido               = "invalido"
	RegistroOmitidoDuplicado       = "omitido_duplicado"
	RegistroOmitidoSinAutorizacion = "omitido_sin_autorizacion"
)

// Estados de una carga.
const (
	CargaPrevisualizada = "previsualizada"
	CargaProcesando     = "procesando"
	CargaCompletada     = "completada"
	CargaCancelada      = "cancelada"
)

// Estados de una notificación.
const (
	NotifPendiente = "pendiente"
	NotifEnviado   = "enviado"
	NotifFallido   = "fallido"
)

type Torneo struct {
	ID               int64  `json:"id"`
	Nombre           string `json:"nombre"`
	ValorInscripcion int    `json:"valor_inscripcion"`
	FechaLimitePago  string `json:"fecha_limite_pago"` // AAAA-MM-DD
	Activo           bool   `json:"activo"`
}

type Carga struct {
	ID                      int64      `json:"id"`
	NombreArchivo           string     `json:"nombre_archivo"`
	Estado                  string     `json:"estado"`
	TotalFilas              int        `json:"total_filas"`
	Validas                 int        `json:"validas"`
	Invalidas               int        `json:"invalidas"`
	OmitidasDuplicado       int        `json:"omitidas_duplicado"`
	OmitidasSinAutorizacion int        `json:"omitidas_sin_autorizacion"`
	InicioProceso           *time.Time `json:"inicio_proceso"`
	FinProceso              *time.Time `json:"fin_proceso"`
	CreadoEn                time.Time  `json:"creado_en"`
}

type Registro struct {
	ID              int64    `json:"id"`
	Fila            int      `json:"fila"`
	Nombre          string   `json:"nombre"`
	Numero          string   `json:"numero"`
	Correo          string   `json:"correo"`
	Club            string   `json:"club"`
	FechaNacimiento string   `json:"fecha_nacimiento"`
	Torneo          string   `json:"torneo"`
	PorcentajeBeca  string   `json:"porcentaje_beca"`
	Autoriza        string   `json:"autoriza"`
	Estado          string   `json:"estado"`
	Errores         []string `json:"errores"`
}

type Beca struct {
	ID               int64
	CargaID          int64
	TorneoID         int64
	Codigo           string
	Nombre           string
	Numero           string
	Correo           string
	Club             string
	FechaNacimiento  time.Time
	Porcentaje       int
	ValorInscripcion int
	Descuento        int
	ValorAPagar      int
	FechaLimitePago  time.Time
	RutaPDF          string
}

// NotificacionItem es una fila del historial.
type NotificacionItem struct {
	ID          int64      `json:"id"`
	BecaID      int64      `json:"beca_id"`
	CargaID     int64      `json:"carga_id"`
	Deportista  string     `json:"deportista"`
	Torneo      string     `json:"torneo"`
	Canal       string     `json:"canal"`
	Destino     string     `json:"destino"`
	Mensaje     string     `json:"mensaje"`
	Estado      string     `json:"estado"`
	Intentos    int        `json:"intentos"`
	UltimoError string     `json:"ultimo_error"`
	IDExterno   string     `json:"id_externo"`
	EnviadoEn   *time.Time `json:"enviado_en"`
}

// Progreso del envío de una carga.
type Progreso struct {
	Total      int `json:"notif_total"`
	Pendientes int `json:"notif_pendientes"`
	Enviadas   int `json:"notif_enviadas"`
	Fallidas   int `json:"notif_fallidas"`
}
