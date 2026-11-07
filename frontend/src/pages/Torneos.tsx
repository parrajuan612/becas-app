import { useCallback, useEffect, useState, type FormEvent } from 'react'
import { api } from '../api/client'
import { cop, fechaCorta } from '../lib'
import type { Torneo } from '../types'

type Form = { id?: number; nombre: string; valor_inscripcion: string; fecha_limite_pago: string; activo: boolean }
const vacio: Form = { nombre: '', valor_inscripcion: '', fecha_limite_pago: '', activo: true }

export default function Torneos() {
  const [torneos, setTorneos] = useState<Torneo[]>([])
  const [form, setForm] = useState<Form | null>(null)
  const [error, setError] = useState('')

  const cargar = useCallback(() => {
    api<Torneo[]>('/torneos').then(setTorneos).catch((e) => setError(e.message))
  }, [])
  useEffect(cargar, [cargar])

  const guardar = async (e: FormEvent) => {
    e.preventDefault()
    if (!form) return
    setError('')
    const body = JSON.stringify({
      nombre: form.nombre,
      valor_inscripcion: Number(form.valor_inscripcion),
      fecha_limite_pago: form.fecha_limite_pago,
      activo: form.activo,
    })
    try {
      await api(form.id ? `/torneos/${form.id}` : '/torneos', { method: form.id ? 'PUT' : 'POST', body })
      setForm(null)
      cargar()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Error')
    }
  }

  const desactivar = async (t: Torneo) => {
    if (!confirm(`¿Desactivar «${t.nombre}»? No se podrá usar en nuevas cargas.`)) return
    try {
      await api(`/torneos/${t.id}`, { method: 'DELETE' })
      cargar()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Error')
    }
  }

  return (
    <section>
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-xl font-semibold">Torneos</h1>
        <button onClick={() => { setError(''); setForm(vacio) }} className="bg-indigo-600 text-white px-4 py-2 rounded text-sm hover:bg-indigo-700">
          Nuevo torneo
        </button>
      </div>
      {error && !form && <p className="text-sm text-red-600 mb-3">{error}</p>}

      <div className="bg-white rounded shadow overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="bg-slate-100 text-left">
            <tr>
              <th className="p-3">Nombre</th>
              <th className="p-3">Inscripción</th>
              <th className="p-3">Fecha límite de pago</th>
              <th className="p-3">Estado</th>
              <th className="p-3"></th>
            </tr>
          </thead>
          <tbody>
            {torneos.map((t) => (
              <tr key={t.id} className="border-t">
                <td className="p-3">{t.nombre}</td>
                <td className="p-3">{cop(t.valor_inscripcion)}</td>
                <td className="p-3">{fechaCorta(t.fecha_limite_pago)}</td>
                <td className="p-3">
                  <span className={`px-2 py-0.5 rounded text-xs ${t.activo ? 'bg-green-100 text-green-800' : 'bg-slate-200 text-slate-600'}`}>
                    {t.activo ? 'Activo' : 'Inactivo'}
                  </span>
                </td>
                <td className="p-3 text-right space-x-3">
                  <button className="text-indigo-600 hover:underline" onClick={() => { setError(''); setForm({ id: t.id, nombre: t.nombre, valor_inscripcion: String(t.valor_inscripcion), fecha_limite_pago: t.fecha_limite_pago, activo: t.activo }) }}>
                    Editar
                  </button>
                  {t.activo && (
                    <button className="text-red-600 hover:underline" onClick={() => desactivar(t)}>
                      Desactivar
                    </button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {form && (
        <div className="fixed inset-0 bg-black/40 grid place-items-center px-4">
          <form onSubmit={guardar} className="bg-white rounded-lg p-6 w-full max-w-md space-y-3">
            <h2 className="font-semibold">{form.id ? 'Editar torneo' : 'Nuevo torneo'}</h2>
            <label className="block text-sm">
              Nombre
              <input required value={form.nombre} onChange={(e) => setForm({ ...form, nombre: e.target.value })} className="mt-1 w-full border rounded px-3 py-2" />
            </label>
            <label className="block text-sm">
              Valor de inscripción (COP)
              <input required type="number" min={1} value={form.valor_inscripcion} onChange={(e) => setForm({ ...form, valor_inscripcion: e.target.value })} className="mt-1 w-full border rounded px-3 py-2" />
            </label>
            <label className="block text-sm">
              Fecha límite de pago
              <input required type="date" value={form.fecha_limite_pago} onChange={(e) => setForm({ ...form, fecha_limite_pago: e.target.value })} className="mt-1 w-full border rounded px-3 py-2" />
            </label>
            {form.id && (
              <label className="flex items-center gap-2 text-sm">
                <input type="checkbox" checked={form.activo} onChange={(e) => setForm({ ...form, activo: e.target.checked })} />
                Activo
              </label>
            )}
            {error && <p className="text-sm text-red-600">{error}</p>}
            <div className="flex justify-end gap-2 pt-2">
              <button type="button" onClick={() => setForm(null)} className="px-4 py-2 text-sm rounded border">Cancelar</button>
              <button className="px-4 py-2 text-sm rounded bg-indigo-600 text-white hover:bg-indigo-700">Guardar</button>
            </div>
          </form>
        </div>
      )}
    </section>
  )
}
