import { useEffect, useState } from 'react'
import { KeyRound } from 'lucide-react'
import type { User } from '../../api'
import { api, errorStatus } from '../../apiClient'
import { useAuth } from '../../auth/useAuth'
import Badge from '../../components/badge/Badge'
import Button from '../../components/button/button'
import { Field, SelectInput, TextInput } from '../../components/field/Field'
import NotificationContainer from '../../components/notifications/NotificationContainer'
import Panel from '../../components/panel/Panel'
import { useNotifications } from '../../hooks/useNotifications'
import { LOCALES, isLocale } from '../../i18n/context'
import { useI18n } from '../../i18n/useI18n'
import ApiTokenPanel from './ApiTokenPanel'
import styles from './profile-page.module.css'

export default function ProfilePage() {
  const { t, locale } = useI18n()
  const { changeLocale } = useAuth()
  const { errors, successes, addError, addSuccess, removeNotification } = useNotifications()

  const [user, setUser] = useState<User | null>(null)
  const [savingLocale, setSavingLocale] = useState(false)

  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [savingPassword, setSavingPassword] = useState(false)

  useEffect(() => {
    let cancelled = false
    api.getMe()
      .then((response) => {
        if (!cancelled) {
          setUser(response.data.user)
        }
      })
      .catch(() => {
        if (!cancelled) {
          addError(t('profile.loadFailed'))
        }
      })
    return () => {
      cancelled = true
    }
  }, [addError, t])

  async function handleLocaleChange(next: string) {
    if (!isLocale(next)) {
      return
    }
    setSavingLocale(true)
    try {
      await changeLocale(next)
      addSuccess(LOCALES[next].messages['profile.language.saved'])
    } catch {
      addError(t('common.languageFailed'))
    } finally {
      setSavingLocale(false)
    }
  }

  async function handlePasswordSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (newPassword !== confirmPassword) {
      addError(t('profile.password.mismatch'))
      return
    }

    setSavingPassword(true)
    try {
      await api.changeMyPassword({ oldPassword, newPassword })
      setOldPassword('')
      setNewPassword('')
      setConfirmPassword('')
      addSuccess(t('profile.password.success'))
    } catch (error: unknown) {
      addError(errorStatus(error) === 403 ? t('profile.password.wrongOld') : t('profile.password.failed'))
    } finally {
      setSavingPassword(false)
    }
  }

  return (
    <div className={styles.page}>
      <NotificationContainer errors={errors} successes={successes} onClose={removeNotification} />

      <Panel title={t('profile.info.title')}>
        <dl className={styles.info}>
          <div className={styles.infoRow}>
            <dt>{t('profile.info.account')}</dt>
            <dd>{user?.account ?? '—'}</dd>
          </div>
          <div className={styles.infoRow}>
            <dt>{t('profile.info.name')}</dt>
            <dd>{user?.name || '—'}</dd>
          </div>
          <div className={styles.infoRow}>
            <dt>{t('profile.info.role')}</dt>
            <dd className={styles.badges}>
              {user && (
                <Badge tone={user.role === 'admin' ? 'primary' : 'neutral'}>
                  {t(user.role === 'admin' ? 'role.admin' : 'role.default')}
                </Badge>
              )}
              {user?.isSystem && <Badge tone="navy">{t('user.system')}</Badge>}
            </dd>
          </div>
        </dl>
      </Panel>

      <Panel title={t('profile.language.title')} description={t('profile.language.desc')}>
        <div className={styles.narrow}>
          <Field id="profile-language" label={t('common.language')}>
            <SelectInput
              id="profile-language"
              value={locale}
              disabled={savingLocale}
              onChange={(event) => handleLocaleChange(event.target.value)}
            >
              {Object.entries(LOCALES).map(([code, { label }]) => (
                <option key={code} value={code}>{label}</option>
              ))}
            </SelectInput>
          </Field>
        </div>
      </Panel>

      <Panel title={t('profile.password.title')}>
        {user?.isSystem ? (
          <p className={styles.note}>
            <KeyRound size={16} aria-hidden="true" />
            {t('profile.password.system')}
          </p>
        ) : (
          <form className={styles.narrow} onSubmit={handlePasswordSubmit}>
            <Field id="old-password" label={t('profile.password.old')}>
              <TextInput
                id="old-password"
                type="password"
                autoComplete="current-password"
                value={oldPassword}
                onChange={(event) => setOldPassword(event.target.value)}
                required
              />
            </Field>
            <Field id="new-password" label={t('profile.password.new')}>
              <TextInput
                id="new-password"
                type="password"
                autoComplete="new-password"
                value={newPassword}
                onChange={(event) => setNewPassword(event.target.value)}
                required
              />
            </Field>
            <Field id="confirm-password" label={t('profile.password.confirm')}>
              <TextInput
                id="confirm-password"
                type="password"
                autoComplete="new-password"
                value={confirmPassword}
                onChange={(event) => setConfirmPassword(event.target.value)}
                required
              />
            </Field>
            <div className={styles.formActions}>
              <Button type="submit" disabled={savingPassword || !user}>
                {t('profile.password.submit')}
              </Button>
            </div>
          </form>
        )}
      </Panel>

      <ApiTokenPanel onError={addError} onSuccess={addSuccess} />
    </div>
  )
}
