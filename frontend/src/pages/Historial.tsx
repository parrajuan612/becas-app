import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { abrirBlob, api } from '../api/client'
import type { Carga, Notificacion, PaginaNotificaciones, Torneo } from '../types'

const TAMANO = 25

const clase: Record<string, string> = {
  enviado: 'bg-green-100 text-green-800',
  fallido: 'bg-red-100 text-red-800',
  pendiente: 'bg-slate-200 text-slate-700',
}

export default function Historial() {
  const [params, setParams] = useSearchParams()
  const [pagina, setPagina] = useState<PaginaNotificaciones | null>(null)
  const [cargas, setCargas] = useState<Carga[]>([])
  const [torneos, setTorneos] = useState<Torneo[]>([])
  const [mensaje, setMensaje] = useState<Notificacion | null>(null)
  const [error, setError] = useState('')
  const [recarga, setRecarga] = useState(0)

  const filtro = (k: string) => params.get(k) ?? ''
  const page = Number(filtro('page') || 1)

  useEffect(() => {
    api<Carga[]>('/cargas').then(setCargas).catch(() => {})
    api<Torneo[]>('/torneos').then(setTorneos).catch(() => {})
  }, [])

  useEffect(() => {
    const q = new URLSearchParams()
    for (const k of ['carga_id', 'torneo_id', 'canal', 'estado']) if (filtro(k)) q.set(k, filtro(k))
    q.set('page', String(page))
    q.set('size', String(TAMANO))
    api<PaginaNotificaciones>(`/notificaciones?${q}`).then(setPagina).catch((e) => setError(e.message))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [params, recarga])

  const cambiar = (k: string, v: string) => {
    const n = new URLSearchParams(params)
    if (v) n.set(k, v)
    else n.delete(k)
    n.delete('page')
    setParams(n)
  }

  const reintentar = async (n: Notificacion) => {
    setError('')
    try {
      await api(`/notificaciones/${n.id}/reintentar`, { method: 'POST' })
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Error')
    }
    setRecarga((x) => x + 1)
  }

  const verPdf = (n: Notificacion) => abrirBlob(`/becas/${n.beca_id}/pdf`).catch((e) => setError(e.message))
  const paginas = pagina ? Math.max(1, Math.ceil(pagina.total / TAMANO)) : 1

  const select = 'border rounded px-2 py-1.5 text-sm bg-white'
  return (
    <section>
      <h1 className="text-xl font-semibold mb-4">Historial de notificaciones</h1>
      <div className="flex flex-wrap gap-2 mb-4">
        <select className={select} value={filtro('carga_id')} onChange={(e) => cambiar('carga_id', e.target.value)}>
          <option value="">Todas las cargas</option>
          {cargas.map((c) => <option key={c.id} value={c.id}>#{c.id} · {c.nombre_archivo}</option>)}
        </select>
        <select className={select} value={filtro('torneo_id')} onChange={(e) => cambiar('torneo_id', e.target.value)}>
          <option value="">Todos los torneos</option>
          {torneos.map((t) => <option key={t.id} value={t.id}>{t.nombre}</option>)}
        </select>
        <select className={select} value={filtro('canal')} onChange={(e) => cambiar('canal', e.target.value)}>
          <option value="">Todos los canales</option>
          <option value="whatsapp">WhatsApp</option>
          <option value="correo">Correo</option>
        </select>
        <select className={select} value={filtro('estado')} onChange={(e) => cambiar('estado', e.target.value)}>
          <option value="">Todos los estados</option>
          <option value="enviado">Enviado</option>
          <option value="fallido">Fallido</option>
          <option value="pendiente">Pendiente</option>
        </select>
      </div>
      {error && <p className="text-sm text-red-600 mb-3">{error}</p>}

      <div className="bg-white rounded shadow overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="bg-slate-100 text-left">
            <tr>
              <th className="p-3">Deportista</th><th className="p-3">Canal</th><th className="p-3">Destino</th>
              <th className="p-3">Estado</th><th className="p-3">Intentos</th><th className="p-3">Error</th><th className="p-3"></th>
            </tr>
          </thead>
          <tbody>
            {pagina?.notificaciones.map((n) => (
              <tr key={n.id} className="border-t">
                <td className="p-3">{n.deportista}<div className="text-xs text-slate-500">{n.torneo}</div></td>
                <td className="p-3">{n.canal === 'whatsapp' ? 'WhatsApp' : 'Correo'}</td>
                <td className="p-3">{n.destino}</td>
                <td className="p-3"><span className={`px-2 py-0.5 rounded text-xs ${clase[n.estado]}`}>{n.estado}</span></td>
                <td className="p-3">{n.intentos}</td>
                <td className="p-3 text-red-700">{n.ultimo_error}</td>
                <td className="p-3 text-right whitespace-nowrap space-x-3">
                  <button className="text-indigo-600 hover:underline" onClick={() => setMensaje(n)}>Mensaje</button>
                  <button className="text-indigo-600 hover:underline" onClick={() => verPdf(n)}>Ver PDF</button>
                  {n.estado === 'fallido' && n.intentos < 3 && (
                    <button className="text-amber-700 hover:underline" onClick={() => reintentar(n)}>Reintentar</button>
                  )}
                </td>
              </tr>
            ))}
            {pagina?.notificaciones.length === 0 && (
              <tr><td colSpan={7} className="p-6 text-center text-slate-500">No hay notificaciones con estos filtros</td></tr>
            )}
          </tbody>
        </table>
      </div>

      <div className="flex items-center justify-between mt-3 text-sm">
        <span className="text-slate-500">{pagina?.total ?? 0} notificaciones</span>
        <div className="flex items-center gap-2">
          <button disabled={page <= 1} onClick={() => setParams((p) => { p.set('page', String(page - 1)); return p })} className="px-3 py-1 border rounded bg-white disabled:opacity-40">Anterior</button>
          <span>{page} / {paginas}</span>
          <button disabled={page >= paginas} onClick={() => setParams((p) => { p.set('page', String(page + 1)); return p })} className="px-3 py-1 border rounded bg-white disabled:opacity-40">Siguiente</button>
        </div>
      </div>

      {mensaje && (
        <div className="fixed inset-0 bg-black/40 grid place-items-center px-4" onClick={() => setMensaje(null)}>
          <div className="bg-white rounded-lg p-6 w-full max-w-lg" onClick={(e) => e.stopPropagation()}>
            <h2 className="font-semibold mb-1">Mensaje a {mensaje.destino}</h2>
            <p className="text-xs text-slate-500 mb-3">Plantilla notificacion_beca_v1 · {mensaje.canal}{mensaje.id_externo && ` · id ${mensaje.id_externo}`}</p>
            <pre className="whitespace-pre-wrap text-sm bg-slate-50 rounded p-3">{mensaje.mensaje}</pre>
            <div className="text-right mt-4">
              <button onClick={() => setMensaje(null)} className="px-4 py-2 text-sm rounded border">Cerrar</button>
            </div>
          </div>
        </div>
      )}
    </section>
  )
}
