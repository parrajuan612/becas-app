const TOKEN_KEY = 'becas_token'

export const getToken = () => localStorage.getItem(TOKEN_KEY)
export const setToken = (t: string) => localStorage.setItem(TOKEN_KEY, t)
export const clearToken = () => localStorage.removeItem(TOKEN_KEY)

export class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

/** Cliente fetch tipado: añade el JWT, parsea JSON y lanza ApiError con el mensaje del backend. */
export async function api<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers)
  const token = getToken()
  if (token) headers.set('Authorization', `Bearer ${token}`)
  if (options.body && !(options.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json')
  }
  const res = await fetch(`/api${path}`, { ...options, headers })
  if (res.status === 401 && token) {
    clearToken()
    window.location.href = '/login'
  }
  const data = await res.json().catch(() => ({}))
  if (!res.ok) throw new ApiError(res.status, data.error ?? 'Error inesperado')
  return data as T
}

/** Descarga un archivo protegido (el <a href> no envía el JWT): usa fetch + blob. */
export async function descargar(path: string, nombre: string): Promise<void> {
  const res = await fetch(`/api${path}`, { headers: { Authorization: `Bearer ${getToken()}` } })
  if (!res.ok) throw new ApiError(res.status, 'No se pudo descargar el archivo')
  const url = URL.createObjectURL(await res.blob())
  const a = document.createElement('a')
  a.href = url
  a.download = nombre
  a.click()
  URL.revokeObjectURL(url)
}

/** Abre un archivo protegido (p. ej. un PDF) en una pestaña nueva usando fetch + blob. */
export async function abrirBlob(path: string): Promise<void> {
  const res = await fetch(`/api${path}`, { headers: { Authorization: `Bearer ${getToken()}` } })
  if (!res.ok) throw new ApiError(res.status, 'No se pudo abrir el archivo')
  window.open(URL.createObjectURL(await res.blob()), '_blank')
}
