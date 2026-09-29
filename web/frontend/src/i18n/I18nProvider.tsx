import { useCallback, useEffect, useMemo, useState, type ReactNode } from 'react'
import type { MessageKey } from './locales/en'
import { I18nContext, LOCALES, LOCALE_STORAGE_KEY, detectLocale, type Locale } from './context'

export default function I18nProvider({ children }: { children: ReactNode }) {
  const [locale, setLocaleState] = useState<Locale>(detectLocale)

  useEffect(() => {
    document.documentElement.lang = locale
  }, [locale])

  const setLocale = useCallback((next: Locale) => {
    localStorage.setItem(LOCALE_STORAGE_KEY, next)
    setLocaleState(next)
  }, [])

  const t = useCallback((key: MessageKey, params?: Record<string, string | number>) => {
    const message = LOCALES[locale].messages[key]
    if (!params) {
      return message
    }
    return message.replace(/\{(\w+)\}/g, (match, name: string) => (name in params ? String(params[name]) : match))
  }, [locale])

  const value = useMemo(() => ({ locale, setLocale, t }), [locale, setLocale, t])

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>
}
