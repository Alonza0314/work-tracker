import { useEffect } from 'react'
import { CircleAlert, X } from 'lucide-react'
import styles from './errorBox.module.css'

interface ErrorMessage {
  id: string
  message: string
  timestamp: number
}

interface ErrorBoxProps {
  errors: ErrorMessage[]
  onClose: (id: string) => void
  duration?: number
}

function SingleError({ 
  error, 
  onClose, 
  duration = 5000,
  index
}: { 
  error: ErrorMessage
  onClose: (id: string) => void
  duration: number
  index: number
}) {
  useEffect(() => {
    const timer = setTimeout(() => {
      onClose(error.id)
    }, duration)
    return () => clearTimeout(timer)
  }, [error.id, duration, onClose])

  return (
    <div 
      className={styles.errorBox}
      role="alert"
      style={{ 
        top: `${(index * 4.75) + 1.25}rem`,
        animationDelay: `${index * 0.05}s`
      }}
    >
      <span className={styles.icon}><CircleAlert size={18} aria-hidden="true" /></span>
      <span className={styles.message}>{error.message}</span>
      <button 
        type="button" 
        className={styles.closeButton}
        onClick={() => onClose(error.id)}
        aria-label="Close error message"
      >
        <X size={16} aria-hidden="true" />
      </button>
    </div>
  )
}

export default function ErrorBox({ errors, onClose, duration = 5000 }: ErrorBoxProps) {
  return (
    <div className={styles.errorBoxContainer}>
      {errors.map((error, index) => (
        <SingleError
          key={error.id}
          error={error}
          onClose={onClose}
          duration={duration}
          index={index}
        />
      ))}
    </div>
  )
}
