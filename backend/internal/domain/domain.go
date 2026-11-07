// Package domain define las entidades y estados compartidos.
package domain

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
