import type { ReactNode } from 'react'
import styles from './badge.module.css'

interface BadgeProps {
  tone?: 'neutral' | 'primary' | 'navy'
  children: ReactNode
}

export default function Badge({ tone = 'neutral', children }: BadgeProps) {
  return <span className={`${styles.badge} ${styles[tone]}`}>{children}</span>
}
