import type { ReactNode } from 'react'
import styles from './panel.module.css'

interface PanelProps {
  title: string
  description?: string
  actions?: ReactNode
  flush?: boolean
  children: ReactNode
}

export default function Panel({ title, description, actions, flush = false, children }: PanelProps) {
  return (
    <section className={styles.panel}>
      <header className={styles.header}>
        <div>
          <h2 className={styles.title}>{title}</h2>
          {description && <p className={styles.description}>{description}</p>}
        </div>
        {actions && <div className={styles.actions}>{actions}</div>}
      </header>
      <div className={flush ? undefined : styles.body}>{children}</div>
    </section>
  )
}
