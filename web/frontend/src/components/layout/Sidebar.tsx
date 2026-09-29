import { NavLink, useNavigate } from 'react-router-dom'
import { LogOut, X } from 'lucide-react'
import Logo from '../logo/Logo'
import { useI18n } from '../../i18n/useI18n'
import { NAV_ITEMS } from './navigation'
import styles from './sidebar.module.css'

interface SidebarProps {
  open: boolean
  onClose: () => void
}

export default function Sidebar({ open, onClose }: SidebarProps) {
  const navigate = useNavigate()
  const { t } = useI18n()
  const username = localStorage.getItem('username') || 'User'

  function handleLogout() {
    localStorage.removeItem('token')
    localStorage.removeItem('username')
    navigate('/login', { replace: true })
  }

  return (
    <>
      <div className={`${styles.backdrop} ${open ? styles.backdropVisible : ''}`} onClick={onClose} aria-hidden="true" />

      <aside className={`${styles.sidebar} ${open ? styles.open : ''}`}>
        <div className={styles.brandRow}>
          <Logo size={32} />
          <span className={styles.brand}>{t('app.name')}</span>
          <button type="button" className={styles.closeButton} onClick={onClose} aria-label={t('common.closeMenu')}>
            <X size={18} aria-hidden="true" />
          </button>
        </div>

        <nav className={styles.nav}>
          <ul className={styles.list}>
            {NAV_ITEMS.map(({ to, label, icon: Icon }) => (
              <li key={to}>
                <NavLink
                  to={to}
                  end
                  onClick={onClose}
                  className={({ isActive }) => `${styles.navItem} ${isActive ? styles.navItemActive : ''}`}
                >
                  <Icon size={18} aria-hidden="true" />
                  <span>{t(label)}</span>
                </NavLink>
              </li>
            ))}
          </ul>
        </nav>

        <div className={styles.footer}>
          <div className={styles.user}>
            <span className={styles.avatar} aria-hidden="true">{username.charAt(0).toUpperCase()}</span>
            <span className={styles.userName}>{username}</span>
          </div>
          <button type="button" className={styles.logoutButton} onClick={handleLogout} title={t('nav.logout')}>
            <LogOut size={18} aria-hidden="true" />
            <span className={styles.srOnly}>{t('nav.logout')}</span>
          </button>
        </div>
      </aside>
    </>
  )
}
