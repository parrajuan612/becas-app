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
