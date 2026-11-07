import { useRef, useState, type DragEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { api, descargar } from '../api/client'
import type { Carga } from '../types'

export default function NuevaCarga() {
  const navigate = useNavigate()
  const input = useRef<HTMLInputElement>(null)
  const [archivo, setArchivo] = useState<File | null>(null)
  const [arrastrando, setArrastrando] = useState(false)
  const [error, setError] = useState('')
  const [enviando, setEnviando] = useState(false)

  const elegir = (f?: File) => {
    setError('')
    if (!f) return
    if (!f.name.toLowerCase().endsWith('.xlsx')) return setError('Solo se aceptan archivos .xlsx')
    if (f.size > 5 * 1024 * 1024) return setError('El archivo supera el máximo de 5 MB')
    setArchivo(f)
  }

  const onDrop = (e: DragEvent) => {
    e.preventDefault()
    setArrastrando(false)
    elegir(e.dataTransfer.files[0])
  }

  const validar = async () => {
    if (!archivo) return
    setEnviando(true)
    setError('')
    try {
      const fd = new FormData()
      fd.append('archivo', archivo)
      const c = await api<Carga>('/cargas', { method: 'POST', body: fd })
      navigate(`/cargas/${c.id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'No se pudo validar el archivo')
    } finally {
      setEnviando(false)
    }
  }

  return (
    <section className="max-w-2xl">
      <h1 className="text-xl font-semibold mb-4">Nueva carga</h1>
      <p className="text-sm text-slate-600 mb-3">
        1. Descargue la plantilla, complete los datos y 2. suba el archivo para validarlo antes de enviar.
      </p>
      <button
        onClick={() => descargar('/plantilla', 'plantilla_becas.xlsx').catch((e) => setError(e.message))}
        className="mb-4 text-sm px-4 py-2 rounded border bg-white hover:bg-slate-50"
      >
        Descargar plantilla
      </button>

      <div
        onDragOver={(e) => { e.preventDefault(); setArrastrando(true) }}
        onDragLeave={() => setArrastrando(false)}
        onDrop={onDrop}
        onClick={() => input.current?.click()}
        className={`cursor-pointer border-2 border-dashed rounded-lg p-10 text-center bg-white ${arrastrando ? 'border-indigo-500 bg-indigo-50' : 'border-slate-300'}`}
      >
        <input ref={input} type="file" accept=".xlsx" hidden onChange={(e) => elegir(e.target.files?.[0])} />
        {archivo ? (
          <p className="font-medium">{archivo.name} <span className="text-slate-500 font-normal">({(archivo.size / 1024).toFixed(0)} KB)</span></p>
        ) : (
          <p className="text-slate-500">Arrastre aquí el archivo .xlsx o haga clic para elegirlo</p>
        )}
      </div>

      {error && <p className="mt-3 text-sm text-red-600">{error}</p>}
      <button
        disabled={!archivo || enviando}
        onClick={validar}
        className="mt-4 bg-indigo-600 text-white px-5 py-2 rounded hover:bg-indigo-700 disabled:opacity-50"
      >
        {enviando ? 'Validando…' : 'Validar archivo'}
      </button>
    </section>
  )
}
