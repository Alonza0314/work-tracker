import type { ReactNode } from 'react'
import styles from './stats-card.module.css'

interface StatsCardProps {
  title: string
  value: number | string
  unit?: string
  description?: string
  icon?: ReactNode
  valueColor?: string
}

export default function StatsCard({ title, value, unit, description, icon, valueColor }: StatsCardProps) {
  return (
    <div className={styles.card}>
      <div className={styles.header}>
        <span className={styles.title}>{title}</span>
        {icon && <span className={styles.icon}>{icon}</span>}
      </div>
      <div className={styles.valueRow}>
        <span className={styles.value} style={valueColor ? { color: valueColor } : undefined}>{value}</span>
        {unit && <span className={styles.unit}>{unit}</span>}
      </div>
      {description && <div className={styles.description}>{description}</div>}
    </div>
  )
}
