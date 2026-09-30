import { useCallback, useEffect, useState } from 'react'
import { Check, Copy, KeyRound, Plus, Trash2 } from 'lucide-react'
import type { ApiTokenInfo } from '../../api'
import { api, errorStatus } from '../../apiClient'
import Badge from '../../components/badge/Badge'
import Button from '../../components/button/button'
import { Field, SelectInput, TextInput } from '../../components/field/Field'
import Modal from '../../components/modal/modal'
import Panel from '../../components/panel/Panel'
import { useI18n } from '../../i18n/useI18n'
import styles from './profile-page.module.css'

const EXPIRY_DAYS = [365, 180, 60, 30] as const

interface ApiTokenPanelProps {
  onError: (message: string) => void
  onSuccess: (message: string) => void
}

// Personal API tokens for scripts and Claude skills. A token is shown once,
// right after it is created; the server keeps only its hash.
export default function ApiTokenPanel({ onError, onSuccess }: ApiTokenPanelProps) {
  const { t, locale } = useI18n()
  const [tokens, setTokens] = useState<ApiTokenInfo[]>([])
  const [creating, setCreating] = useState(false)
  const [name, setName] = useState('')
  const [days, setDays] = useState<number>(365)
  const [created, setCreated] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)
  const [revoking, setRevoking] = useState<ApiTokenInfo | null>(null)
  const [busy, setBusy] = useState(false)

  const load = useCallback(async () => {
    try {
      const response = await api.listMyApiTokens()
      setTokens(response.data.tokens)
    } catch {
      onError(t('token.loadFailed'))
    }
  }, [onError, t])

  useEffect(() => {
    void load()
  }, [load])

  async function create() {
    const trimmed = name.trim()
    if (!trimmed) {
      onError(t('token.nameRequired'))
      return
    }
    setBusy(true)
    try {
      const response = await api.createMyApiToken({ name: trimmed, expiresInDays: days as 30 | 60 | 180 | 365 })
      setCreating(false)
      setName('')
      setDays(365)
      setCopied(false)
      setCreated(response.data.token)
      await load()
    } catch (error: unknown) {
      onError(errorStatus(error) === 409 ? t('token.limit') : t('token.createFailed'))
    } finally {
      setBusy(false)
    }
  }

  async function revoke() {
    if (!revoking) {
      return
    }
    setBusy(true)
    try {
      await api.deleteMyApiToken(revoking.id)
      onSuccess(t('token.revoked', { name: revoking.name }))
      setRevoking(null)
      await load()
    } catch {
      onError(t('token.revokeFailed'))
    } finally {
      setBusy(false)
    }
  }

  async function copy() {
    if (!created) {
      return
    }
    try {
      await navigator.clipboard.writeText(created)
      setCopied(true)
    } catch {
      onError(t('token.copyFailed'))
    }
  }

  const dateFormat = new Intl.DateTimeFormat(locale, { dateStyle: 'medium' })
  const dateTimeFormat = new Intl.DateTimeFormat(locale, { dateStyle: 'medium', timeStyle: 'short' })
  const now = Date.now()

  return (
    <Panel
      title={t('token.title')}
      description={t('token.desc')}
      actions={(
        <Button variant="secondary" icon={<Plus aria-hidden="true" />} onClick={() => setCreating(true)}>
          {t('token.create')}
        </Button>
      )}
      flush
    >
      {tokens.length === 0 ? (
        <p className={styles.tokenEmpty}>
          <KeyRound size={16} aria-hidden="true" />
          {t('token.empty')}
        </p>
      ) : (
        <ul className={styles.tokenList}>
          {tokens.map((token) => {
            const expired = new Date(token.expiresAt).getTime() <= now
            return (
              <li key={token.id} className={styles.tokenItem}>
                <div className={styles.tokenMain}>
                  <div className={styles.tokenTitle}>
                    <span className={styles.tokenName}>{token.name}</span>
                    <code className={styles.tokenPrefix}>{token.prefix}…</code>
                    {expired && <Badge>{t('token.expired')}</Badge>}
                  </div>
                  <p className={styles.tokenMeta}>
                    {t('token.meta', {
                      created: dateFormat.format(new Date(token.createdAt)),
                      expires: dateFormat.format(new Date(token.expiresAt)),
                      used: token.lastUsedAt ? dateTimeFormat.format(new Date(token.lastUsedAt)) : t('token.neverUsed'),
                    })}
                  </p>
                </div>
                <button
                  type="button"
                  className={styles.tokenRevoke}
                  onClick={() => setRevoking(token)}
                  title={t('token.revoke')}
                  aria-label={`${t('token.revoke')} ${token.name}`}
                >
                  <Trash2 size={16} aria-hidden="true" />
                </button>
              </li>
            )
          })}
        </ul>
      )}

      {creating && (
        <Modal
          isOpen
          title={t('token.createTitle')}
          onClose={() => setCreating(false)}
          onSubmit={() => void create()}
          submitLabel={t('token.create')}
          submitting={busy}
        >
          <Field id="token-name" label={t('token.name')}>
            <TextInput
              id="token-name"
              value={name}
              onChange={(event) => setName(event.target.value)}
              placeholder={t('token.namePlaceholder')}
              autoFocus
              required
            />
          </Field>
          <Field id="token-expiry" label={t('token.expiry')}>
            <SelectInput id="token-expiry" value={days} onChange={(event) => setDays(Number(event.target.value))}>
              {EXPIRY_DAYS.map((option) => (
                <option key={option} value={option}>{t('token.days', { days: option })}</option>
              ))}
            </SelectInput>
          </Field>
        </Modal>
      )}

      {created && (
        <Modal isOpen title={t('token.createdTitle')} onClose={() => setCreated(null)} cancelLabel={t('common.close')}>
          <p className={styles.tokenWarning}>{t('token.createdWarning')}</p>
          <div className={styles.tokenValue}>
            <TextInput value={created} readOnly onFocus={(event) => event.currentTarget.select()} aria-label={t('token.title')} />
            <Button variant="secondary" icon={copied ? <Check aria-hidden="true" /> : <Copy aria-hidden="true" />} onClick={() => void copy()}>
              {copied ? t('token.copied') : t('token.copy')}
            </Button>
          </div>
          <p className={styles.tokenUsage}>{t('token.usage')}</p>
          <code className={styles.tokenExample}>Authorization: Bearer {created.slice(0, 10)}…</code>
        </Modal>
      )}

      {revoking && (
        <Modal
          isOpen
          title={t('token.revokeTitle')}
          onClose={() => setRevoking(null)}
          onSubmit={() => void revoke()}
          submitLabel={t('token.revoke')}
          submitVariant="danger"
          submitting={busy}
        >
          <p>{t('token.revokeConfirm', { name: revoking.name })}</p>
        </Modal>
      )}
    </Panel>
  )
}
