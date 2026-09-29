import { NavLink, useNavigate } from 'react-router-dom'
import { LogOut, X } from 'lucide-react'
import Logo from '../logo/Logo'
import { useAuth } from '../../auth/useAuth'
import { useI18n } from '../../i18n/useI18n'
import { useWork } from '../../work/useWork'
import { NAV_ITEMS, canAccess } from './navigation'
import styles from './sidebar.module.css'

interface SidebarProps {
  open: boolean
  onClose: () => void
}

export default function Sidebar({ open, onClose }: SidebarProps) {
  const navigate = useNavigate()
  const { t } = useI18n()
  const { session, isAdmin, signOut } = useAuth()
  const { canViewAll } = useWork()
  const name = session?.name ?? ''

  async function handleLogout() {
    await signOut()
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
            {NAV_ITEMS.filter((item) => canAccess(item, isAdmin, canViewAll)).map(({ to, label, icon: Icon }) => (
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
            <span className={styles.avatar} aria-hidden="true">{name.charAt(0).toUpperCase()}</span>
            <div className={styles.userMeta}>
              <span className={styles.userName}>{name}</span>
              <span className={styles.userRole}>{t(isAdmin ? 'role.admin' : 'role.default')}</span>
            </div>
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
