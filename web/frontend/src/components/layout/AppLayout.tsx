import { useState } from 'react'
import { Outlet, useLocation } from 'react-router-dom'
import NotificationContainer from '../notifications/NotificationContainer'
import { useNotifications } from '../../hooks/useNotifications'
import Sidebar from './Sidebar'
import Topbar from './Topbar'
import { findNavItem } from './navigation'
import { useI18n } from '../../i18n/useI18n'
import styles from './app-layout.module.css'

export default function AppLayout() {
  const { t } = useI18n()
  const { pathname } = useLocation()
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const { errors, successes, addError, removeNotification } = useNotifications()

  const navItem = findNavItem(pathname)
  const title = navItem ? t(navItem.label) : t('app.name')

  return (
    <div className={styles.shell}>
      <NotificationContainer errors={errors} successes={successes} onClose={removeNotification} />
      <Sidebar open={sidebarOpen} onClose={() => setSidebarOpen(false)} onError={addError} />
      <div className={styles.main}>
        <Topbar title={title} onMenuClick={() => setSidebarOpen(true)} onError={addError} />
        <main className={styles.content}>
          <Outlet />
        </main>
      </div>
    </div>
  )
}
