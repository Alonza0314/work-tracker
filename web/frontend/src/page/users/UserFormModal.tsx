import { useState } from 'react'
import { I18n, Role, type User } from '../../api'
import { Field, SelectInput, TextInput } from '../../components/field/Field'
import Modal from '../../components/modal/modal'
import { LOCALES } from '../../i18n/context'
import { useI18n } from '../../i18n/useI18n'

export interface UserFormValues {
  account: string
  name: string
  // edit only: a new password, empty keeps the current one
  password: string
  role: Role
  i18n: I18n
}

interface UserFormModalProps {
  // edit mode when set; the account cannot be renamed
  user?: User
  isSelf?: boolean
  submitting: boolean
  onClose: () => void
  onSubmit: (values: UserFormValues) => void
}

export default function UserFormModal({ user, isSelf = false, submitting, onClose, onSubmit }: UserFormModalProps) {
  const { t, locale } = useI18n()
  const [values, setValues] = useState<UserFormValues>({
    account: user?.account ?? '',
    name: user?.name ?? '',
    password: '',
    role: user?.role ?? Role.Default,
    i18n: user?.i18n ?? locale,
  })

  function update<K extends keyof UserFormValues>(key: K, value: UserFormValues[K]) {
    setValues((prev) => ({ ...prev, [key]: value }))
  }

  const isEdit = user !== undefined

  return (
    <Modal
      isOpen
      title={isEdit ? t('users.editTitle', { account: user.account }) : t('users.createTitle')}
      onClose={onClose}
      onSubmit={() => onSubmit(values)}
      submitLabel={isEdit ? t('common.save') : t('users.create')}
      submitting={submitting}
    >
      {!isEdit && (
        <Field id="user-account" label={t('users.form.account')} hint={t('users.form.defaultPassword')}>
          <TextInput
            id="user-account"
            value={values.account}
            onChange={(event) => update('account', event.target.value)}
            autoComplete="off"
            autoFocus
            required
          />
        </Field>
      )}
      <Field id="user-name" label={t('users.form.name')}>
        <TextInput
          id="user-name"
          value={values.name}
          onChange={(event) => update('name', event.target.value)}
          autoComplete="off"
          autoFocus={isEdit}
          required
        />
      </Field>
      {isEdit && (
        <Field id="user-password" label={t('users.form.newPassword')} hint={t('users.form.passwordKeep')}>
          <TextInput
            id="user-password"
            type="password"
            value={values.password}
            onChange={(event) => update('password', event.target.value)}
            autoComplete="new-password"
          />
        </Field>
      )}
      <Field id="user-role" label={t('users.form.role')} hint={isSelf ? t('users.form.selfRole') : undefined}>
        <SelectInput
          id="user-role"
          value={values.role}
          disabled={isSelf}
          onChange={(event) => update('role', event.target.value as Role)}
        >
          <option value={Role.Default}>{t('role.default')}</option>
          <option value={Role.Admin}>{t('role.admin')}</option>
        </SelectInput>
      </Field>
      <Field id="user-language" label={t('users.form.language')}>
        <SelectInput
          id="user-language"
          value={values.i18n}
          onChange={(event) => update('i18n', event.target.value as I18n)}
        >
          {Object.entries(LOCALES).map(([code, { label }]) => (
            <option key={code} value={code}>{label}</option>
          ))}
        </SelectInput>
      </Field>
    </Modal>
  )
}
