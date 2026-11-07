import { useCallback, useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api } from '../api/client'
import type { CargaDetalle, EstadoRegistro, PaginaRegistros } from '../types'

const TAMANO = 50

const pestanas: { clave: EstadoRegistro | ''; label: string }[] = [
  { clave: '', label: 'Todas' },
  { clave: 'valido', label: 'Válidas' },
  { clave: 'invalido', label: 'Inválidas' },
  { clave: 'omitido_duplicado', label: 'Duplicadas' },
  { clave: 'omitido_sin_autorizacion', label: 'Sin autorización' },
]

const etiqueta: Record<EstadoRegistro, { texto: string; clase: string }> = {
  valido: { texto: 'Válida', clase: 'bg-green-100 text-green-800' },
  invalido: { texto: 'Inválida', clase: 'bg-red-100 text-red-800' },
  omitido_duplicado: { texto: 'Duplicada', clase: 'bg-amber-100 text-amber-800' },
  omitido_sin_autorizacion: { texto: 'Sin autorización', clase: 'bg-amber-100 text-amber-800' },
}

export default function DetalleCarga() {
  const { id } = useParams()
  const [carga, setCarga] = useState<CargaDetalle | null>(null)
  const [pagina, setPagina] = useState<PaginaRegistros | null>(null)
  const [estado, setEstado] = useState<EstadoRegistro | ''>('')
  const [page, setPage] = useState(1)
  const [error, setError] = useState('')
  const [reintentando, setReintentando] = useState(false)

  const cargarCarga = useCallback(
    () => api<CargaDetalle>(`/cargas/${id}`).then(setCarga).catch((e) => setError(e.message)),
    [id],
  )

  useEffect(() => {
    cargarCarga()
  }, [cargarCarga])

  // Mientras se envía (o se reintentan fallidas) se consulta el estado cada 1 s.
  const enCurso = carga?.estado === 'procesando' || (carga?.notif_pendientes ?? 0) > 0 || reintentando
  useEffect(() => {
    if (!enCurso) return
    const t = setInterval(cargarCarga, 1000)
    return () => clearInterval(t)
  }, [enCurso, cargarCarga])

  useEffect(() => {
    api<PaginaRegistros>(`/cargas/${id}/registros?estado=${estado}&page=${page}&size=${TAMANO}`)
      .then(setPagina)
      .catch((e) => setError(e.message))
  }, [id, estado, page, carga?.estado])

  const accion = async (ruta: string) => {
    setError('')
    try {
      await api(`/cargas/${id}/${ruta}`, { method: 'POST' })
      if (ruta === 'reintentar') {
        setReintentando(true)
        setTimeout(() => setReintentando(false), 4000)
      }
      await cargarCarga()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Error')
    }
  }

  if (error && !carga) return <p className="text-red-600">{error}</p>
  if (!carga) return <p className="text-slate-500">Cargando…</p>

  const tarjetas = [
    { label: 'Total de filas', valor: carga.total_filas, clase: 'text-slate-800' },
    { label: 'Válidas', valor: carga.validas, clase: 'text-green-700' },
    { label: 'Inválidas', valor: carga.invalidas, clase: 'text-red-700' },
    { label: 'Duplicadas', valor: carga.omitidas_duplicado, clase: 'text-amber-700' },
    { label: 'Sin autorización', valor: carga.omitidas_sin_autorizacion, clase: 'text-amber-700' },
  ]
  const paginas = pagina ? Math.max(1, Math.ceil(pagina.total / TAMANO)) : 1

  return (
    <section>
      <Link to="/cargas/nueva" className="text-sm text-indigo-600 hover:underline">← Nueva carga</Link>
      <h1 className="text-xl font-semibold mt-1">
        {carga.estado === 'previsualizada' ? 'Previsualización' : 'Carga'}: {carga.nombre_archivo}
      </h1>
      <p className="text-sm text-slate-500 mb-4">Estado: {carga.estado}</p>
      {error && <p className="text-sm text-red-600 mb-3">{error}</p>}

      {carga.estado === 'previsualizada' && (
        <div className="flex gap-2 mb-4 items-center flex-wrap">
          <button
            disabled={carga.validas === 0}
            onClick={() => accion('confirmar')}
            className="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700 disabled:opacity-50"
          >
            Confirmar y enviar ({carga.validas})
          </button>
          <button onClick={() => accion('cancelar')} className="px-4 py-2 rounded border bg-white hover:bg-slate-50">
            Cancelar
          </button>
          <span className="text-sm text-slate-500">Solo se procesarán las filas válidas.</span>
        </div>
      )}

      {carga.estado !== 'previsualizada' && carga.estado !== 'cancelada' && <Progreso c={carga} onReintentar={() => accion('reintentar')} />}

      <div className="grid grid-cols-2 md:grid-cols-5 gap-3 mb-6">
        {tarjetas.map((t) => (
          <div key={t.label} className="bg-white rounded shadow p-4">
            <div className={`text-2xl font-bold ${t.clase}`}>{t.valor}</div>
            <div className="text-xs text-slate-500">{t.label}</div>
          </div>
        ))}
      </div>

      <div className="flex gap-1 mb-3 flex-wrap">
        {pestanas.map((p) => (
          <button
            key={p.clave}
            onClick={() => { setEstado(p.clave); setPage(1) }}
            className={`px-3 py-1.5 rounded text-sm ${estado === p.clave ? 'bg-indigo-600 text-white' : 'bg-white border hover:bg-slate-50'}`}
          >
            {p.label}
          </button>
        ))}
      </div>

      <div className="bg-white rounded shadow overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="bg-slate-100 text-left">
            <tr>
              <th className="p-3">Fila</th><th className="p-3">Nombre</th><th className="p-3">Número</th>
              <th className="p-3">Torneo</th><th className="p-3">%</th><th className="p-3">Estado</th><th className="p-3">Errores</th>
            </tr>
          </thead>
          <tbody>
            {pagina?.registros.map((r) => (
              <tr key={r.id} className={`border-t ${r.estado === 'invalido' ? 'bg-red-50' : ''}`}>
                <td className="p-3">{r.fila}</td>
                <td className="p-3">{r.nombre}</td>
                <td className="p-3">{r.numero}</td>
                <td className="p-3">{r.torneo}</td>
                <td className="p-3">{r.porcentaje_beca}</td>
                <td className="p-3">
                  <span className={`px-2 py-0.5 rounded text-xs whitespace-nowrap ${etiqueta[r.estado].clase}`}>{etiqueta[r.estado].texto}</span>
                </td>
                <td className="p-3 text-red-700">
                  {r.errores.map((e) => <div key={e}>{e}</div>)}
                </td>
              </tr>
            ))}
            {pagina?.registros.length === 0 && (
              <tr><td colSpan={7} className="p-6 text-center text-slate-500">Sin filas en esta categoría</td></tr>
            )}
          </tbody>
        </table>
      </div>

      <div className="flex items-center justify-between mt-3 text-sm">
        <span className="text-slate-500">{pagina?.total ?? 0} filas</span>
        <div className="flex items-center gap-2">
          <button disabled={page <= 1} onClick={() => setPage(page - 1)} className="px-3 py-1 border rounded bg-white disabled:opacity-40">Anterior</button>
          <span>{page} / {paginas}</span>
          <button disabled={page >= paginas} onClick={() => setPage(page + 1)} className="px-3 py-1 border rounded bg-white disabled:opacity-40">Siguiente</button>
        </div>
      </div>
    </section>
  )
}

function Progreso({ c, onReintentar }: { c: CargaDetalle; onReintentar: () => void }) {
  const hechas = c.notif_enviadas + c.notif_fallidas
  const total = Math.max(c.notif_total, 2 * c.validas)
  const pct = total ? Math.round((hechas / total) * 100) : 0
  const terminada = c.estado === 'completada' && c.notif_pendientes === 0
  const seg = c.inicio_proceso && c.fin_proceso ? (new Date(c.fin_proceso).getTime() - new Date(c.inicio_proceso).getTime()) / 1000 : null
  return (
    <div className="bg-white rounded shadow p-4 mb-6">
      <div className="flex justify-between text-sm mb-1">
        <span>{terminada ? 'Procesamiento completado' : 'Enviando notificaciones…'}</span>
        <span>{hechas} / {total} ({pct}%)</span>
      </div>
      <div className="h-3 bg-slate-200 rounded overflow-hidden">
        <div className="h-full bg-green-500 transition-all" style={{ width: `${pct}%` }} />
      </div>
      <div className="flex flex-wrap items-center gap-x-6 gap-y-1 mt-3 text-sm">
        <span className="text-green-700">Enviadas: {c.notif_enviadas}</span>
        <span className="text-red-700">Fallidas: {c.notif_fallidas}</span>
        {seg !== null && <span className="text-slate-600">Duración: {seg.toFixed(1)} s</span>}
        {terminada && c.notif_fallidas > 0 && (
          <button onClick={onReintentar} className="ml-auto px-3 py-1.5 rounded bg-indigo-600 text-white hover:bg-indigo-700">
            Reintentar fallidos
          </button>
        )}
      </div>
      {terminada && <Link to={`/historial?carga_id=${c.id}`} className="inline-block mt-3 text-sm text-indigo-600 hover:underline">Ver historial de esta carga →</Link>}
    </div>
  )
}
