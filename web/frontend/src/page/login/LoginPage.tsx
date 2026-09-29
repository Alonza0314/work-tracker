import { useState, type FormEvent } from 'react'
import { Eye, EyeOff, Lock, User } from 'lucide-react'
import Button from '../../components/button/button'
import Logo from '../../components/logo/Logo'
import LanguageSwitcher from '../../components/languageSwitcher/LanguageSwitcher'
import NotificationContainer from '../../components/notifications/NotificationContainer'
import { useNotifications } from '../../hooks/useNotifications'
import { useI18n } from '../../i18n/useI18n'
import { api, errorStatus } from '../../apiClient'
import { useAuth } from '../../auth/useAuth'
import { Navigate, useNavigate } from 'react-router-dom'
import styles from './login-page.module.css'

export default function LoginPage() {
  const navigate = useNavigate()
  const { t } = useI18n()
  const { session, signIn } = useAuth()
  const [account, setAccount] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [isLoading, setIsLoading] = useState(false)

  const { errors, successes, addError, removeNotification } = useNotifications()

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setIsLoading(true)

    try {
      const response = await api.login({ account, password })
      signIn(response.data.token)
      navigate('/', { replace: true })
    } catch (error: unknown) {
      addError(errorStatus(error) === 401 ? t('login.invalid') : t('login.failed'))
      setIsLoading(false)
    }
  }

  if (session) {
    return <Navigate to="/" replace />
  }

  return (
    <div className={styles.page}>
      <NotificationContainer
        errors={errors}
        successes={successes}
        onClose={removeNotification}
      />

      <aside className={styles.hero}>
        <div className={styles.heroGrid} aria-hidden="true" />
        <p className={styles.heroTitle}>{t('app.name')}</p>
      </aside>

      <main className={styles.formPane}>
        <div className={styles.formTopbar}>
          <div className={styles.mobileBrand}>
            <Logo size={28} />
            <span>{t('app.name')}</span>
          </div>
          <LanguageSwitcher />
        </div>

        <div className={styles.formWrap}>
          <div className={styles.headerBlock}>
            <h2 className={styles.title}>{t('login.title')}</h2>
            <p className={styles.subtitle}>{t('login.subtitle')}</p>
          </div>

          <form className={styles.form} onSubmit={handleSubmit}>
            <div className={styles.field}>
              <label className={styles.label} htmlFor="account">{t('login.account')}</label>
              <div className={styles.inputWrap}>
                <User size={16} className={styles.inputIcon} aria-hidden="true" />
                <input
                  id="account"
                  className={styles.input}
                  value={account}
                  onChange={(event) => setAccount(event.target.value)}
                  placeholder={t('login.accountPlaceholder')}
                  autoComplete="username"
                  autoFocus
                  required
                />
              </div>
            </div>

            <div className={styles.field}>
              <label className={styles.label} htmlFor="password">{t('login.password')}</label>
              <div className={styles.inputWrap}>
                <Lock size={16} className={styles.inputIcon} aria-hidden="true" />
                <input
                  id="password"
                  type={showPassword ? 'text' : 'password'}
                  className={styles.input}
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                  placeholder={t('login.passwordPlaceholder')}
                  autoComplete="current-password"
                  required
                />
                <button
                  type="button"
                  className={styles.revealButton}
                  onClick={() => setShowPassword((value) => !value)}
                  aria-label={showPassword ? t('login.hidePassword') : t('login.showPassword')}
                  aria-pressed={showPassword}
                >
                  {showPassword ? <EyeOff size={16} aria-hidden="true" /> : <Eye size={16} aria-hidden="true" />}
                </button>
              </div>
            </div>

            <Button type="submit" size="lg" fullWidth disabled={isLoading}>
              {isLoading && <span className={styles.spinner} aria-hidden="true" />}
              {isLoading ? t('login.submitting') : t('login.submit')}
            </Button>
          </form>

        </div>
      </main>
    </div>
  )
}
