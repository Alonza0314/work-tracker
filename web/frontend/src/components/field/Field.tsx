import type { InputHTMLAttributes, ReactNode, SelectHTMLAttributes } from 'react'
import styles from './field.module.css'

interface FieldProps {
  id: string
  label: string
  hint?: string
  children: ReactNode
}

export function Field({ id, label, hint, children }: FieldProps) {
  return (
    <div className={styles.field}>
      <label className={styles.label} htmlFor={id}>{label}</label>
      {children}
      {hint && <p className={styles.hint}>{hint}</p>}
    </div>
  )
}

export function TextInput(props: InputHTMLAttributes<HTMLInputElement>) {
  return <input {...props} className={styles.control} />
}

export function SelectInput(props: SelectHTMLAttributes<HTMLSelectElement>) {
  return <select {...props} className={`${styles.control} ${styles.select}`} />
}
