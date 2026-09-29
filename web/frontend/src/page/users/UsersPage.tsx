import { useCallback, useEffect, useState } from 'react'
import { Pencil, Plus, Trash2 } from 'lucide-react'
import { Role, type UpdateUserRequest, type User } from '../../api'
import { api, errorStatus } from '../../apiClient'
import { useAuth } from '../../auth/useAuth'
import Badge from '../../components/badge/Badge'
import Button from '../../components/button/button'
import Modal from '../../components/modal/modal'
import NotificationContainer from '../../components/notifications/NotificationContainer'
import Panel from '../../components/panel/Panel'
import { useNotifications } from '../../hooks/useNotifications'
import { LOCALES } from '../../i18n/context'
import { useI18n } from '../../i18n/useI18n'
import UserFormModal, { type UserFormValues } from './UserFormModal'
import styles from './users-page.module.css'

type Dialog =
  | { kind: 'create' }
  | { kind: 'edit', user: User }
  | { kind: 'delete', user: User }

export default function UsersPage() {
  const { t } = useI18n()
  const { session } = useAuth()
  const { errors, successes, addError, addSuccess, removeNotification } = useNotifications()

  const [users, setUsers] = useState<User[]>([])
  const [dialog, setDialog] = useState<Dialog | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const loadUsers = useCallback(async () => {
    try {
      const response = await api.listUsers()
      setUsers(response.data.users)
    } catch {
      addError(t('users.loadFailed'))
    }
  }, [addError, t])

  useEffect(() => {
    void loadUsers()
  }, [loadUsers])

  // runs a dialog action, then closes the dialog and reloads the list
  async function submit(action: () => Promise<void>, onError: (error: unknown) => void) {
    setSubmitting(true)
    try {
      await action()
      setDialog(null)
      await loadUsers()
    } catch (error: unknown) {
      onError(error)
    } finally {
      setSubmitting(false)
    }
  }

  function handleCreate(values: UserFormValues) {
    const account = values.account.trim()
    const name = values.name.trim()
    if (!account || !name) {
      addError(t('users.required'))
      return
    }
    void submit(async () => {
      await api.createUser({ account, name, role: values.role, i18n: values.i18n })
      addSuccess(t('users.created', { account }))
    }, (error) => addError(errorStatus(error) === 409 ? t('users.exists') : t('users.createFailed')))
  }

  function handleUpdate(user: User, values: UserFormValues) {
    const request: UpdateUserRequest = { i18n: values.i18n }
    const name = values.name.trim()
    if (!name) {
      addError(t('users.required'))
      return
    }
    if (name !== user.name) {
      request.name = name
    }
    if (values.role !== user.role) {
      request.role = values.role
    }
    if (values.password) {
      request.password = values.password
    }
    void submit(async () => {
      await api.updateUser(user.account, request)
      addSuccess(t('users.updated', { account: user.account }))
    }, () => addError(t('users.updateFailed')))
  }

  function handleDelete(user: User) {
    void submit(async () => {
      await api.deleteUser(user.account)
      addSuccess(t('users.deleted', { account: user.account }))
    }, () => addError(t('users.deleteFailed')))
  }

  return (
    <div className={styles.page}>
      <NotificationContainer errors={errors} successes={successes} onClose={removeNotification} />

      <Panel
        title={t('nav.users')}
        description={t('users.count', { count: users.length })}
        actions={(
          <Button icon={<Plus aria-hidden="true" />} onClick={() => setDialog({ kind: 'create' })}>
            {t('users.create')}
          </Button>
        )}
        flush
      >
        <div className={styles.tableWrap}>
          <table className={styles.table}>
            <thead>
              <tr>
                <th>{t('users.table.name')}</th>
                <th>{t('users.table.account')}</th>
                <th>{t('users.table.role')}</th>
                <th>{t('users.table.language')}</th>
                <th className={styles.actionsCell}>{t('common.actions')}</th>
              </tr>
            </thead>
            <tbody>
              {users.map((user) => {
                const isSelf = user.account === session?.account
                return (
                  <tr key={user.account}>
                    <td>
                      <div className={styles.accountCell}>
                        <span className={styles.avatar} aria-hidden="true">{(user.name || user.account).charAt(0).toUpperCase()}</span>
                        <span className={styles.accountName}>{user.name || '—'}</span>
                        {isSelf && <Badge>{t('user.you')}</Badge>}
                        {user.isSystem && <Badge tone="navy">{t('user.system')}</Badge>}
                      </div>
                    </td>
                    <td className={styles.muted}>{user.account}</td>
                    <td>
                      <Badge tone={user.role === Role.Admin ? 'primary' : 'neutral'}>
                        {t(user.role === Role.Admin ? 'role.admin' : 'role.default')}
                      </Badge>
                    </td>
                    <td className={styles.muted}>{LOCALES[user.i18n].label}</td>
                    <td className={styles.actionsCell}>
                      <div className={styles.rowActions}>
                        <button
                          type="button"
                          className={styles.iconButton}
                          onClick={() => setDialog({ kind: 'edit', user })}
                          disabled={user.isSystem}
                          title={t('common.edit')}
                          aria-label={`${t('common.edit')} ${user.account}`}
                        >
                          <Pencil size={16} aria-hidden="true" />
                        </button>
                        <button
                          type="button"
                          className={`${styles.iconButton} ${styles.danger}`}
                          onClick={() => setDialog({ kind: 'delete', user })}
                          disabled={user.isSystem || isSelf}
                          title={t('common.delete')}
                          aria-label={`${t('common.delete')} ${user.account}`}
                        >
                          <Trash2 size={16} aria-hidden="true" />
                        </button>
                      </div>
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      </Panel>

      {dialog?.kind === 'create' && (
        <UserFormModal
          submitting={submitting}
          onClose={() => setDialog(null)}
          onSubmit={handleCreate}
        />
      )}
      {dialog?.kind === 'edit' && (
        <UserFormModal
          user={dialog.user}
          isSelf={dialog.user.account === session?.account}
          submitting={submitting}
          onClose={() => setDialog(null)}
          onSubmit={(values) => handleUpdate(dialog.user, values)}
        />
      )}
      {dialog?.kind === 'delete' && (
        <Modal
          isOpen
          title={t('users.deleteTitle')}
          onClose={() => setDialog(null)}
          onSubmit={() => handleDelete(dialog.user)}
          submitLabel={t('common.delete')}
          submitVariant="danger"
          submitting={submitting}
        >
          <p>{t('users.deleteConfirm', { account: dialog.user.account })}</p>
        </Modal>
      )}
    </div>
  )
}
