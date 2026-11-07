export interface Usuario {
  id: number
  correo: string
  nombre: string
}

export interface Torneo {
  id: number
  nombre: string
  valor_inscripcion: number
  fecha_limite_pago: string // AAAA-MM-DD
  activo: boolean
}

export interface Carga {
  id: number
  nombre_archivo: string
  estado: 'previsualizada' | 'procesando' | 'completada' | 'cancelada'
  total_filas: number
  validas: number
  invalidas: number
  omitidas_duplicado: number
  omitidas_sin_autorizacion: number
  inicio_proceso: string | null
  fin_proceso: string | null
  creado_en: string
}

export type EstadoRegistro = 'valido' | 'invalido' | 'omitido_duplicado' | 'omitido_sin_autorizacion'

export interface Registro {
  id: number
  fila: number
  nombre: string
  numero: string
  correo: string
  club: string
  fecha_nacimiento: string
  torneo: string
  porcentaje_beca: string
  autoriza: string
  estado: EstadoRegistro
  errores: string[]
}

export interface PaginaRegistros {
  registros: Registro[]
  total: number
  page: number
  size: number
}

export interface CargaDetalle extends Carga {
  notif_total: number
  notif_pendientes: number
  notif_enviadas: number
  notif_fallidas: number
}

export type EstadoNotif = 'pendiente' | 'enviado' | 'fallido'

export interface Notificacion {
  id: number
  beca_id: number
  carga_id: number
  deportista: string
  torneo: string
  canal: 'whatsapp' | 'correo'
  destino: string
  mensaje: string
  estado: EstadoNotif
  intentos: number
  ultimo_error: string
  id_externo: string
  enviado_en: string | null
}

export interface PaginaNotificaciones {
  notificaciones: Notificacion[]
  total: number
  page: number
  size: number
}
