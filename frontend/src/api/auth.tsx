import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import type { Usuario } from '../types'
import { api, clearToken, getToken, setToken } from './client'

interface AuthCtx {
  usuario: Usuario | null
  cargando: boolean
  login: (correo: string, password: string) => Promise<void>
  logout: () => void
}

const Ctx = createContext<AuthCtx | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [usuario, setUsuario] = useState<Usuario | null>(null)
  const [cargando, setCargando] = useState(!!getToken())

  useEffect(() => {
    if (!getToken()) return
    api<Usuario>('/auth/me')
      .then(setUsuario)
      .catch(() => clearToken())
      .finally(() => setCargando(false))
  }, [])

  const login = async (correo: string, password: string) => {
    const r = await api<{ token: string; usuario: Usuario }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ correo, password }),
    })
    setToken(r.token)
    setUsuario(r.usuario)
  }

  const logout = () => {
    clearToken()
    setUsuario(null)
  }

  return <Ctx.Provider value={{ usuario, cargando, login, logout }}>{children}</Ctx.Provider>
}

export function useAuth() {
  const c = useContext(Ctx)
  if (!c) throw new Error('useAuth fuera de AuthProvider')
  return c
}
