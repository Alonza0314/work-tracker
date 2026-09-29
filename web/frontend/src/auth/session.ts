import { I18n, Role } from '../api'

const TOKEN_STORAGE_KEY = 'token'
// set when an expired session sends the user back to /login
const SESSION_EXPIRED_KEY = 'sessionExpired'

export interface Session {
  token: string
  account: string
  name: string
  role: Role
  i18n: I18n
  expiresAt: number
}

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_STORAGE_KEY)
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_STORAGE_KEY, token)
}

export function clearToken(): void {
  localStorage.removeItem(TOKEN_STORAGE_KEY)
}

// the login page tells the user to sign in again, once
export function markSessionExpired(): void {
  sessionStorage.setItem(SESSION_EXPIRED_KEY, '1')
}

export function takeSessionExpired(): boolean {
  const expired = sessionStorage.getItem(SESSION_EXPIRED_KEY) !== null
  sessionStorage.removeItem(SESSION_EXPIRED_KEY)
  return expired
}

function decodeBase64Url(segment: string): string {
  const base64 = segment.replace(/-/g, '+').replace(/_/g, '/')
  const padded = base64 + '='.repeat((4 - (base64.length % 4)) % 4)
  const bytes = Uint8Array.from(atob(padded), (char) => char.charCodeAt(0))
  return new TextDecoder().decode(bytes)
}

// Reads the claims the backend puts in the JWT (sub, name, role, i18n, exp). The
// signature is not verified here; the backend re-checks every request.
export function parseToken(token: string | null): Session | null {
  if (!token) {
    return null
  }

  try {
    const claims = JSON.parse(decodeBase64Url(token.split('.')[1] ?? ''))
    const session: Session = {
      token,
      // accounts are upper case; tokens issued before that may carry lower case
      account: String(claims.sub ?? '').toUpperCase(),
      name: String(claims.name || claims.sub || ''),
      role: claims.role === Role.Admin ? Role.Admin : Role.Default,
      i18n: claims.i18n === I18n.En ? I18n.En : I18n.ZhTw,
      expiresAt: Number(claims.exp ?? 0) * 1000,
    }
    if (!session.account || session.expiresAt <= Date.now()) {
      return null
    }
    return session
  } catch {
    return null
  }
}
