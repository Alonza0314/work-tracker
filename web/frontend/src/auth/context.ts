import { createContext } from 'react'
import type { Locale } from '../i18n/context'
import type { Session } from './session'

export interface AuthContextValue {
  session: Session | null
  isAdmin: boolean
  // stores a token from login/updateMe and switches to its i18n
  signIn: (token: string) => void
  signOut: () => Promise<void>
  // switches the UI language; when signed in it is also saved to the account
  changeLocale: (locale: Locale) => Promise<void>
}

export const AuthContext = createContext<AuthContextValue | null>(null)
