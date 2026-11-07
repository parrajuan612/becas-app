import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api } from '../api/client'
import type { Carga, EstadoRegistro, PaginaRegistros } from '../types'

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
  const [carga, setCarga] = useState<Carga | null>(null)
  const [pagina, setPagina] = useState<PaginaRegistros | null>(null)
  const [estado, setEstado] = useState<EstadoRegistro | ''>('')
  const [page, setPage] = useState(1)
  const [error, setError] = useState('')

  useEffect(() => {
    api<Carga>(`/cargas/${id}`).then(setCarga).catch((e) => setError(e.message))
  }, [id])

  useEffect(() => {
    api<PaginaRegistros>(`/cargas/${id}/registros?estado=${estado}&page=${page}&size=${TAMANO}`)
      .then(setPagina)
      .catch((e) => setError(e.message))
  }, [id, estado, page])

  if (error) return <p className="text-red-600">{error}</p>
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
      <h1 className="text-xl font-semibold mt-1">Previsualización: {carga.nombre_archivo}</h1>
      <p className="text-sm text-slate-500 mb-4">Estado: {carga.estado}</p>

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
      <p className="mt-6 text-sm text-slate-500">El envío de las becas válidas se habilitará en la fase F3.</p>
    </section>
  )
}
