import { Languages } from 'lucide-react'
import { useState } from 'react'
import { useAuth } from '../../auth/useAuth'
import { LOCALES, isLocale } from '../../i18n/context'
import { useI18n } from '../../i18n/useI18n'
import styles from './language-switcher.module.css'

interface LanguageSwitcherProps {
  tone?: 'light' | 'dark'
  onError?: (message: string) => void
}

export default function LanguageSwitcher({ tone = 'light', onError }: LanguageSwitcherProps) {
  const { locale, t } = useI18n()
  const { changeLocale } = useAuth()
  const [saving, setSaving] = useState(false)

  return (
    <label className={`${styles.switcher} ${styles[tone]}`}>
      <Languages size={16} aria-hidden="true" />
      <span className={styles.srOnly}>{t('common.language')}</span>
      <select
        className={styles.select}
        value={locale}
        disabled={saving}
        onChange={async (event) => {
          const next = event.target.value
          if (!isLocale(next)) {
            return
          }
          setSaving(true)
          try {
            await changeLocale(next)
          } catch {
            onError?.(t('common.languageFailed'))
          } finally {
            setSaving(false)
          }
        }}
      >
        {Object.entries(LOCALES).map(([code, { label }]) => (
          <option key={code} value={code}>{label}</option>
        ))}
      </select>
    </label>
  )
}
