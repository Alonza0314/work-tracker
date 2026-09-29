import { useState, type FormEvent } from 'react'
import { Pencil, Trash2 } from 'lucide-react'
import type { CategoryColor, WorkOption } from '../../api'
import { api, errorStatus } from '../../apiClient'
import Button from '../../components/button/button'
import ColorPicker from '../../components/colorPicker/ColorPicker'
import { Field, TextInput } from '../../components/field/Field'
import Modal from '../../components/modal/modal'
import NotificationContainer from '../../components/notifications/NotificationContainer'
import Panel from '../../components/panel/Panel'
import Switch from '../../components/switch/switch'
import { useNotifications } from '../../hooks/useNotifications'
import { useI18n } from '../../i18n/useI18n'
import type { MessageKey } from '../../i18n/locales/en'
import { useWork } from '../../work/useWork'
import HolidayPanel from './HolidayPanel'
import styles from './work.module.css'

interface OptionApi {
  create: (name: string) => Promise<unknown>
  update: (id: string, change: { name?: string, active?: boolean, color?: CategoryColor }) => Promise<unknown>
  // resolves to how many work records and todos lost the option
  remove: (id: string) => Promise<number>
}

const CATEGORY_API: OptionApi = {
  create: (name) => api.createCategory({ name }),
  update: (id, change) => api.updateCategory(id, change),
  remove: async (id) => (await api.deleteCategory(id)).data.cleared,
}

const PROJECT_API: OptionApi = {
  create: (name) => api.createProject({ name }),
  update: (id, change) => api.updateProject(id, change),
  remove: async (id) => (await api.deleteProject(id)).data.cleared,
}

export default function WorkSettingsPage() {
  const { t } = useI18n()
  const { categories, projects, allowViewAll, reload } = useWork()
  const { errors, successes, addError, addSuccess, removeNotification } = useNotifications()
  const [savingSetting, setSavingSetting] = useState(false)

  async function handleAllowViewAll(checked: boolean) {
    setSavingSetting(true)
    try {
      await api.updateWorkSetting({ allowViewAll: checked })
      await reload()
      addSuccess(t('settings.saved'))
    } catch {
      addError(t('settings.saveFailed'))
    } finally {
      setSavingSetting(false)
    }
  }

  return (
    <div className={styles.page}>
      <NotificationContainer errors={errors} successes={successes} onClose={removeNotification} />

      <Panel title={t('settings.view.title')} description={t('settings.view.desc')}>
        <Switch
          checked={allowViewAll}
          onChange={handleAllowViewAll}
          disabled={savingSetting}
          label={t('settings.view.label')}
        />
      </Panel>

      <div className={styles.settingsGrid}>
        <OptionPanel
          title="settings.categories.title"
          colored
          options={categories}
          optionApi={CATEGORY_API}
          onChanged={reload}
          onError={addError}
          onSuccess={addSuccess}
        />
        <OptionPanel
          title="settings.projects.title"
          options={projects}
          optionApi={PROJECT_API}
          onChanged={reload}
          onError={addError}
          onSuccess={addSuccess}
        />
      </div>

      <HolidayPanel onError={addError} onSuccess={addSuccess} />
    </div>
  )
}

interface OptionPanelProps {
  title: MessageKey
  // categories have a color, projects do not
  colored?: boolean
  options: WorkOption[]
  optionApi: OptionApi
  onChanged: () => Promise<void>
  onError: (message: string) => void
  onSuccess: (message: string) => void
}

function OptionPanel({ title, colored = false, options, optionApi, onChanged, onError, onSuccess }: OptionPanelProps) {
  const { t } = useI18n()
  const [newName, setNewName] = useState('')
  const [renaming, setRenaming] = useState<WorkOption | null>(null)
  const [renameValue, setRenameValue] = useState('')
  const [deleting, setDeleting] = useState<WorkOption | null>(null)
  const [busy, setBusy] = useState(false)

  async function run(action: () => Promise<unknown>, success: string): Promise<boolean> {
    setBusy(true)
    try {
      await action()
      await onChanged()
      onSuccess(success)
      return true
    } catch (error: unknown) {
      onError(errorStatus(error) === 409 ? t('settings.option.exists') : t('settings.saveFailed'))
      return false
    } finally {
      setBusy(false)
    }
  }

  async function handleAdd(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const name = newName.trim()
    if (!name) {
      return
    }
    if (await run(() => optionApi.create(name), t('settings.option.added', { name }))) {
      setNewName('')
    }
  }

  async function handleRename() {
    const name = renameValue.trim()
    if (!renaming || !name) {
      return
    }
    if (await run(() => optionApi.update(renaming.id, { name }), t('settings.option.renamed', { name }))) {
      setRenaming(null)
    }
  }

  async function handleDelete() {
    if (!deleting) {
      return
    }
    const { id, name } = deleting
    setBusy(true)
    try {
      const cleared = await optionApi.remove(id)
      await onChanged()
      onSuccess(cleared > 0 ? t('settings.option.deletedCleared', { name, count: cleared }) : t('settings.option.deleted', { name }))
      setDeleting(null)
    } catch {
      onError(t('settings.option.deleteFailed'))
    } finally {
      setBusy(false)
    }
  }

  return (
    <Panel title={t(title)} description={t('settings.option.desc')} flush>
      <form className={styles.optionForm} onSubmit={handleAdd}>
        <TextInput
          value={newName}
          onChange={(event) => setNewName(event.target.value)}
          placeholder={t('settings.option.placeholder')}
          aria-label={t('settings.option.placeholder')}
        />
        <Button type="submit" disabled={busy || !newName.trim()}>{t('settings.option.add')}</Button>
      </form>

      {options.length === 0 ? (
        <p className={styles.empty}>{t('settings.option.empty')}</p>
      ) : (
        <ul className={styles.optionList}>
          {options.map((option) => (
            <li key={option.id} className={styles.optionItem}>
              {colored && (
                <ColorPicker
                  value={option.color}
                  disabled={busy}
                  label={t('settings.option.color', { name: option.name })}
                  onChange={(color) => void run(() => optionApi.update(option.id, { color }), t('settings.saved'))}
                />
              )}
              <span className={`${styles.optionName} ${option.active ? '' : styles.optionInactive}`}>{option.name}</span>
              <Switch
                checked={option.active}
                disabled={busy}
                label={t('settings.option.active')}
                onChange={(active) => void run(() => optionApi.update(option.id, { active }), t('settings.saved'))}
              />
              <button
                type="button"
                className={styles.iconButton}
                onClick={() => {
                  setRenaming(option)
                  setRenameValue(option.name)
                }}
                title={t('settings.option.rename')}
                aria-label={`${t('settings.option.rename')} ${option.name}`}
              >
                <Pencil size={16} aria-hidden="true" />
              </button>
              <button
                type="button"
                className={`${styles.iconButton} ${styles.danger}`}
                onClick={() => setDeleting(option)}
                title={t('common.delete')}
                aria-label={`${t('common.delete')} ${option.name}`}
              >
                <Trash2 size={16} aria-hidden="true" />
              </button>
            </li>
          ))}
        </ul>
      )}

      {deleting && (
        <Modal
          isOpen
          title={t('settings.option.deleteTitle')}
          onClose={() => setDeleting(null)}
          onSubmit={() => void handleDelete()}
          submitLabel={t('common.delete')}
          submitVariant="danger"
          submitting={busy}
        >
          <p>{t('settings.option.deleteConfirm', { name: deleting.name })}</p>
        </Modal>
      )}
      {renaming && (
        <Modal
          isOpen
          title={t('settings.option.renameTitle', { name: renaming.name })}
          onClose={() => setRenaming(null)}
          onSubmit={() => void handleRename()}
          submitLabel={t('common.save')}
          submitting={busy}
        >
          <Field id="option-rename" label={t('settings.option.name')}>
            <TextInput
              id="option-rename"
              value={renameValue}
              onChange={(event) => setRenameValue(event.target.value)}
              autoFocus
              required
            />
          </Field>
        </Modal>
      )}
    </Panel>
  )
}
