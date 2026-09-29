import { Menu } from 'lucide-react'
import LanguageSwitcher from '../languageSwitcher/LanguageSwitcher'
import { useI18n } from '../../i18n/useI18n'
import styles from './topbar.module.css'

interface TopbarProps {
  title: string
  onMenuClick: () => void
}

export default function Topbar({ title, onMenuClick }: TopbarProps) {
  const { t } = useI18n()

  return (
    <header className={styles.topbar}>
      <button type="button" className={styles.menuButton} onClick={onMenuClick} aria-label={t('common.openMenu')}>
        <Menu size={20} aria-hidden="true" />
      </button>

      <h1 className={styles.title}>{title}</h1>

      <div className={styles.actions}>
        <LanguageSwitcher />
      </div>
    </header>
  )
}
