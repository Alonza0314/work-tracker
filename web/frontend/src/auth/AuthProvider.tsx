import { useCallback, useEffect, useMemo, useState, type ReactNode } from 'react'
import { Role } from '../api'
import { api } from '../apiClient'
import type { Locale } from '../i18n/context'
import { useI18n } from '../i18n/useI18n'
import { AuthContext } from './context'
import { clearToken, getToken, parseToken, setToken } from './session'

export default function AuthProvider({ children }: { children: ReactNode }) {
  const { setLocale } = useI18n()
  const [session, setSession] = useState(() => parseToken(getToken()))

  // the account's language wins over the one remembered by the browser
  useEffect(() => {
    if (session) {
      setLocale(session.i18n)
    }
  }, [session, setLocale])

  const signIn = useCallback((token: string) => {
    setToken(token)
    setSession(parseToken(token))
  }, [])

  const signOut = useCallback(async () => {
    try {
      await api.logout()
    } catch {
      // logging out locally is enough; the token is stateless
    }
    clearToken()
    setSession(null)
  }, [])

  const changeLocale = useCallback(async (locale: Locale) => {
    if (!session) {
      setLocale(locale)
      return
    }
    const response = await api.updateMe({ i18n: locale })
    signIn(response.data.token)
  }, [session, setLocale, signIn])

  const value = useMemo(() => ({
    session,
    isAdmin: session?.role === Role.Admin,
    signIn,
    signOut,
    changeLocale,
  }), [session, signIn, signOut, changeLocale])

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}
