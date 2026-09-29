import type { ReactNode } from 'react'
import { X } from 'lucide-react'
import styles from './modal.module.css'
import Button from '../button/button'
import { useI18n } from '../../i18n/useI18n'

interface ModalProps {
  isOpen: boolean
  onClose: () => void
  title: string
  description?: string
  children: ReactNode
  onSubmit?: () => void
  submitLabel?: string
  submitVariant?: 'primary' | 'danger'
  submitting?: boolean
}

export default function Modal({
  isOpen,
  onClose,
  title,
  description,
  children,
  onSubmit,
  submitLabel,
  submitVariant = 'primary',
  submitting = false,
}: ModalProps) {
  const { t } = useI18n()

  if (!isOpen) return null

  return (
    <div className={styles.overlay} onClick={onClose}>
      <form
        className={styles.modal}
        role="dialog"
        aria-modal="true"
        aria-labelledby="modal-title"
        onClick={(event) => event.stopPropagation()}
        onSubmit={(event) => {
          event.preventDefault()
          onSubmit?.()
        }}
      >
        <div className={styles.header}>
          <div>
            <h2 id="modal-title" className={styles.title}>{title}</h2>
            {description && <p className={styles.description}>{description}</p>}
          </div>
          <button type="button" className={styles.closeButton} onClick={onClose} aria-label={t('common.close')}>
            <X size={18} aria-hidden="true" />
          </button>
        </div>
        <div className={styles.body}>
          {children}
        </div>
        <div className={styles.footer}>
          <Button variant="secondary" onClick={onClose} disabled={submitting}>
            {t('common.cancel')}
          </Button>
          {onSubmit && (
            <Button type="submit" variant={submitVariant} disabled={submitting}>
              {submitLabel ?? t('common.submit')}
            </Button>
          )}
        </div>
      </form>
    </div>
  )
}
