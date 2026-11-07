import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api/client'
import type { ConteoCanal, Metricas } from '../types'

const estadoClase: Record<string, string> = {
  completada: 'bg-green-100 text-green-800',
  procesando: 'bg-blue-100 text-blue-800',
  previsualizada: 'bg-amber-100 text-amber-800',
  cancelada: 'bg-slate-200 text-slate-600',
}

export default function Panel() {
  const [m, setM] = useState<Metricas | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    api<Metricas>('/metricas').then(setM).catch((e) => setError(e.message))
  }, [])

  if (error) return <p className="text-red-600">{error}</p>
  if (!m) return <p className="text-slate-500">Cargando…</p>

  const canales = Object.entries(m.por_canal) as ['whatsapp' | 'correo', ConteoCanal][]
  const sum = (f: (c: ConteoCanal) => number) => canales.reduce((a, [, c]) => a + f(c), 0)
  const tarjetas = [
    { label: 'Becas generadas', valor: m.total_becas, clase: 'text-slate-800' },
    { label: 'Notificaciones enviadas', valor: sum((c) => c.enviado), clase: 'text-green-700' },
    { label: 'Fallidas', valor: sum((c) => c.fallido), clase: 'text-red-700' },
    { label: 'Pendientes', valor: sum((c) => c.pendiente), clase: 'text-slate-600' },
    { label: 'Omitidas por duplicado', valor: m.omitidas_duplicado, clase: 'text-amber-700' },
    { label: 'Omitidas sin autorización', valor: m.omitidas_sin_autorizacion, clase: 'text-amber-700' },
    { label: 'Tasa de éxito', valor: `${m.tasa_exito.toFixed(1)} %`, clase: 'text-indigo-700' },
  ]

  return (
    <section>
      <h1 className="text-xl font-semibold mb-4">Panel</h1>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-3 mb-6">
        {tarjetas.map((t) => (
          <div key={t.label} className="bg-white rounded shadow p-4">
            <div className={`text-2xl font-bold ${t.clase}`}>{t.valor}</div>
            <div className="text-xs text-slate-500">{t.label}</div>
          </div>
        ))}
      </div>

      <div className="bg-white rounded shadow p-4 mb-6">
        <h2 className="font-semibold mb-3">Éxito por canal</h2>
        {canales.map(([canal, c]) => {
          const total = c.enviado + c.fallido + c.pendiente
          const pct = (n: number) => (total ? (n * 100) / total : 0)
          return (
            <div key={canal} className="mb-3 last:mb-0">
              <div className="flex justify-between text-sm mb-1">
                <span className="capitalize">{canal}</span>
                <span className="text-slate-500">
                  {c.enviado} enviadas · {c.fallido} fallidas · {c.pendiente} pendientes
                </span>
              </div>
              <div className="h-3 bg-slate-200 rounded overflow-hidden flex">
                <div className="bg-green-500" style={{ width: `${pct(c.enviado)}%` }} />
                <div className="bg-red-500" style={{ width: `${pct(c.fallido)}%` }} />
              </div>
            </div>
          )
        })}
      </div>

      <div className="bg-white rounded shadow overflow-x-auto">
        <h2 className="font-semibold p-4 pb-2">Últimas cargas</h2>
        <table className="w-full text-sm">
          <thead className="bg-slate-100 text-left">
            <tr>
              <th className="p-3">Archivo</th><th className="p-3">Estado</th><th className="p-3">Válidas</th>
              <th className="p-3">Enviadas / Fallidas</th><th className="p-3">Duración</th><th className="p-3">ms por registro</th>
            </tr>
          </thead>
          <tbody>
            {m.ultimas_cargas.map((c) => (
              <tr key={c.id} className="border-t">
                <td className="p-3"><Link className="text-indigo-600 hover:underline" to={`/cargas/${c.id}`}>#{c.id} · {c.nombre_archivo}</Link></td>
                <td className="p-3"><span className={`px-2 py-0.5 rounded text-xs ${estadoClase[c.estado]}`}>{c.estado}</span></td>
                <td className="p-3">{c.validas}</td>
                <td className="p-3">{c.notif_enviadas} / {c.notif_fallidas}</td>
                <td className="p-3">{c.duracion_ms !== null ? `${(c.duracion_ms / 1000).toFixed(1)} s` : '—'}</td>
                <td className="p-3">{c.ms_por_registro !== null ? c.ms_por_registro.toFixed(0) : '—'}</td>
              </tr>
            ))}
            {m.ultimas_cargas.length === 0 && (
              <tr><td colSpan={6} className="p-6 text-center text-slate-500">Aún no hay cargas. <Link className="text-indigo-600 hover:underline" to="/cargas/nueva">Cree la primera</Link>.</td></tr>
            )}
          </tbody>
        </table>
      </div>
    </section>
  )
}
