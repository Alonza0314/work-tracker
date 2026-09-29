import { useRef, useState } from 'react'
import { Download, RotateCcw, Trash2, Upload } from 'lucide-react'
import { api, extractErrorMessage } from '../../apiClient'
import { leaveToLogin } from '../../auth/session'
import { useI18n } from '../../i18n/useI18n'
import { saveBlob } from '../../work/export'
import Button from '../button/button'
import { TextInput } from '../field/Field'
import Modal from '../modal/modal'
import styles from './system-modal.module.css'

const RESTORE_CONFIRM = 'RESTORE'
const RESET_CONFIRM = 'RESET'

interface SystemModalProps {
  onClose: () => void
  onError: (message: string) => void
}

// Admin maintenance: download a backup, restore one, or reset everything.
// Restore and reset replace all data, so they need the word typed in and
// sign everyone out afterwards.
export default function SystemModal({ onClose, onError }: SystemModalProps) {
  const { t } = useI18n()
  const [busy, setBusy] = useState(false)
  const [file, setFile] = useState<File | null>(null)
  const [restoreConfirm, setRestoreConfirm] = useState('')
  const [resetConfirm, setResetConfirm] = useState('')
  const fileInput = useRef<HTMLInputElement>(null)

  async function download() {
    setBusy(true)
    try {
      const response = await api.downloadBackup({ responseType: 'blob' })
      const disposition = String(response.headers['content-disposition'] ?? '')
      const fileName = /filename="([^"]+)"/.exec(disposition)?.[1] ?? 'work-tracker-backup.zip'
      saveBlob(response.data as unknown as Blob, fileName)
    } catch {
      onError(t('system.backupFailed'))
    } finally {
      setBusy(false)
    }
  }

  async function restore() {
    if (!file || restoreConfirm !== RESTORE_CONFIRM) {
      return
    }
    setBusy(true)
    try {
      await api.restoreBackup(file)
      leaveToLogin('restored')
    } catch (error: unknown) {
      onError(extractErrorMessage(error, t('system.restoreFailed')))
      setBusy(false)
    }
  }

  async function reset() {
    if (resetConfirm !== RESET_CONFIRM) {
      return
    }
    setBusy(true)
    try {
      await api.resetSystem({ confirm: RESET_CONFIRM })
      leaveToLogin('reset')
    } catch {
      onError(t('system.resetFailed'))
      setBusy(false)
    }
  }

  return (
    <Modal isOpen title={t('system.title')} description={t('system.desc')} onClose={onClose} cancelLabel={t('common.close')}>
      <div className={styles.sections}>
        <section className={styles.section}>
          <div>
            <h3 className={styles.heading}>{t('system.backup.title')}</h3>
            <p className={styles.text}>{t('system.backup.desc')}</p>
          </div>
          <Button variant="secondary" icon={<Download aria-hidden="true" />} disabled={busy} onClick={() => void download()}>
            {t('system.backup.action')}
          </Button>
        </section>

        <section className={styles.section}>
          <div className={styles.grow}>
            <h3 className={styles.heading}>{t('system.restore.title')}</h3>
            <p className={styles.text}>{t('system.restore.desc')}</p>
            <div className={styles.controls}>
              <input
                ref={fileInput}
                type="file"
                accept=".zip,application/zip"
                className={styles.fileInput}
                onChange={(event) => setFile(event.target.files?.[0] ?? null)}
              />
              <Button variant="secondary" icon={<Upload aria-hidden="true" />} disabled={busy} onClick={() => fileInput.current?.click()}>
                {t('system.restore.choose')}
              </Button>
              <span className={styles.fileName}>{file?.name ?? t('system.restore.noFile')}</span>
            </div>
            {file && (
              <div className={styles.controls}>
                <TextInput
                  value={restoreConfirm}
                  onChange={(event) => setRestoreConfirm(event.target.value)}
                  placeholder={t('system.confirmPlaceholder', { word: RESTORE_CONFIRM })}
                  aria-label={t('system.confirmPlaceholder', { word: RESTORE_CONFIRM })}
                />
                <Button
                  variant="danger"
                  icon={<RotateCcw aria-hidden="true" />}
                  disabled={busy || restoreConfirm !== RESTORE_CONFIRM}
                  onClick={() => void restore()}
                >
                  {t('system.restore.action')}
                </Button>
              </div>
            )}
          </div>
        </section>

        <section className={`${styles.section} ${styles.danger}`}>
          <div className={styles.grow}>
            <h3 className={styles.heading}>{t('system.reset.title')}</h3>
            <p className={styles.text}>{t('system.reset.desc')}</p>
            <div className={styles.controls}>
              <TextInput
                value={resetConfirm}
                onChange={(event) => setResetConfirm(event.target.value)}
                placeholder={t('system.confirmPlaceholder', { word: RESET_CONFIRM })}
                aria-label={t('system.confirmPlaceholder', { word: RESET_CONFIRM })}
              />
              <Button
                variant="danger"
                icon={<Trash2 aria-hidden="true" />}
                disabled={busy || resetConfirm !== RESET_CONFIRM}
                onClick={() => void reset()}
              >
                {t('system.reset.action')}
              </Button>
            </div>
          </div>
        </section>
      </div>
    </Modal>
  )
}
