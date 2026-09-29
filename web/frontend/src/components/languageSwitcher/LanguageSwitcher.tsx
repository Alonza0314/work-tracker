import { Languages } from 'lucide-react'
import { LOCALES, isLocale } from '../../i18n/context'
import { useI18n } from '../../i18n/useI18n'
import styles from './language-switcher.module.css'

interface LanguageSwitcherProps {
  tone?: 'light' | 'dark'
}

export default function LanguageSwitcher({ tone = 'light' }: LanguageSwitcherProps) {
  const { locale, setLocale, t } = useI18n()

  return (
    <label className={`${styles.switcher} ${styles[tone]}`}>
      <Languages size={16} aria-hidden="true" />
      <span className={styles.srOnly}>{t('common.language')}</span>
      <select
        className={styles.select}
        value={locale}
        onChange={(event) => {
          if (isLocale(event.target.value)) {
            setLocale(event.target.value)
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
