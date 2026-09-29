import { createContext } from 'react'
import en, { type MessageKey, type Messages } from './locales/en'
import zhTW from './locales/zh-TW'

export const LOCALES = {
  'zh-TW': { label: '繁體中文', messages: zhTW },
  en: { label: 'English', messages: en },
} as const satisfies Record<string, { label: string, messages: Messages }>

export type Locale = keyof typeof LOCALES

export const DEFAULT_LOCALE: Locale = 'zh-TW'
export const LOCALE_STORAGE_KEY = 'locale'

export interface I18nContextValue {
  locale: Locale
  setLocale: (locale: Locale) => void
  t: (key: MessageKey, params?: Record<string, string | number>) => string
}

export const I18nContext = createContext<I18nContextValue | null>(null)

export function isLocale(value: string | null): value is Locale {
  return value !== null && value in LOCALES
}

export function detectLocale(): Locale {
  const stored = localStorage.getItem(LOCALE_STORAGE_KEY)
  if (isLocale(stored)) {
    return stored
  }
  return navigator.language.toLowerCase().startsWith('en') ? 'en' : DEFAULT_LOCALE
}
