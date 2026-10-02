import { useCallback, useEffect, useState } from 'react'
import type { SaveWorkEntryRequest, WorkRecord, WorkRecordListResponse } from '../../api'
import { api } from '../../apiClient'
import Modal from '../../components/modal/modal'
import NotificationContainer from '../../components/notifications/NotificationContainer'
import Panel from '../../components/panel/Panel'
import WeekNavigator from '../../components/weekNavigator/WeekNavigator'
import { useNotifications } from '../../hooks/useNotifications'
import { useI18n } from '../../i18n/useI18n'
import { periodEnd, periodStart, today } from '../../work/format'
import EditableWorkTable from './EditableWorkTable'
import styles from './work.module.css'

// the records table shows this many weeks, ending with the current one
const WEEKS = 2

// todos share the work record shape
type Dialog =
  | { kind: 'deleteRecord', record: WorkRecord }
  | { kind: 'deleteTodo', todo: WorkRecord }

export default function MyWorkPage() {
  const { t } = useI18n()
  const { errors, successes, addError, addSuccess, removeNotification } = useNotifications()

  const [todos, setTodos] = useState<WorkRecord[]>([])
  const [records, setRecords] = useState<WorkRecordListResponse | null>(null)
  // Monday of the first shown week
  const [week, setWeek] = useState(() => periodStart(today(), WEEKS))
  const [dialog, setDialog] = useState<Dialog | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [completing, setCompleting] = useState<string | null>(null)

  const loadTodos = useCallback(async () => {
    try {
      const response = await api.listMyTodos()
      setTodos(response.data.todos)
    } catch {
      addError(t('work.loadFailed'))
    }
  }, [addError, t])

  const loadRecords = useCallback(async () => {
    try {
      const response = await api.listMyWorkRecords(week, periodEnd(week, WEEKS))
      setRecords(response.data)
    } catch {
      addError(t('work.loadFailed'))
    }
  }, [addError, week, t])

  // show date: reload in place when it is shown, else switch to the period
  // ending with its week (which reloads)
  const showWeekOf = useCallback(async (date: string) => {
    if (date >= week && date <= periodEnd(week, WEEKS)) {
      await loadRecords()
    } else {
      setWeek(periodStart(date, WEEKS))
    }
  }, [loadRecords, week])

  useEffect(() => {
    void loadTodos()
  }, [loadTodos])

  useEffect(() => {
    void loadRecords()
  }, [loadRecords])

  // runs a save from the tables; errors are reported here, the table only
  // needs to know whether to keep or revert its edit
  async function save(action: () => Promise<unknown>, reload: () => Promise<void>): Promise<boolean> {
    try {
      await action()
      await reload()
      return true
    } catch {
      addError(t('work.saveFailed'))
      return false
    }
  }

  // a record added for another week switches to that week, so it stays visible
  function createRecord(request: SaveWorkEntryRequest) {
    return save(() => api.createMyWorkRecord(request), () => showWeekOf(request.date ?? today()))
  }

  function updateRecord(record: WorkRecord, request: SaveWorkEntryRequest) {
    return save(() => api.updateMyWorkRecord(record.id, request), loadRecords)
  }

  function createTodo(request: SaveWorkEntryRequest) {
    return save(() => api.createMyTodo(request), loadTodos)
  }

  function updateTodo(todo: WorkRecord, request: SaveWorkEntryRequest) {
    return save(() => api.updateMyTodo(todo.id, request), loadTodos)
  }

  // runs a dialog action, then closes the dialog and reloads both lists
  async function submit(action: () => Promise<void>, failure: string) {
    setSubmitting(true)
    try {
      await action()
      setDialog(null)
      await Promise.all([loadTodos(), loadRecords()])
    } catch {
      addError(failure)
    } finally {
      setSubmitting(false)
    }
  }

  // the record keeps the todo's date; today only fills in a missing one
  async function completeTodo(todo: WorkRecord) {
    setCompleting(todo.id)
    try {
      const response = await api.completeMyTodo(todo.id, { date: today() })
      addSuccess(t('todos.completed'))
      await Promise.all([loadTodos(), showWeekOf(response.data.record.date)])
    } catch {
      addError(t('todos.completeFailed'))
    } finally {
      setCompleting(null)
    }
  }

  return (
    <div className={styles.page}>
      <NotificationContainer errors={errors} successes={successes} onClose={removeNotification} />

      <Panel title={t('todos.title')} description={t('todos.desc')} flush>
        <EditableWorkTable
          kind="todo"
          rows={todos}
          emptyText={t('todos.empty')}
          onCreate={createTodo}
          onUpdate={updateTodo}
          onDelete={(todo) => setDialog({ kind: 'deleteTodo', todo })}
          onInvalid={addError}
          leading={(todo) => (
            <input
              type="checkbox"
              className={styles.checkbox}
              checked={completing === todo.id}
              disabled={completing !== null}
              onChange={() => void completeTodo(todo)}
              aria-label={`${t('todos.complete')}: ${todo.description}`}
            />
          )}
        />
      </Panel>

      <Panel
        title={t('records.title')}
        description={records ? t('records.weekSummary', { count: records.total, hours: records.totalHours }) : undefined}
        actions={<WeekNavigator start={week} onChange={setWeek} weeks={WEEKS} />}
        flush
      >
        <EditableWorkTable
          kind="record"
          rows={records?.records ?? []}
          emptyText={t('records.empty')}
          onCreate={createRecord}
          onUpdate={updateRecord}
          onDelete={(record) => setDialog({ kind: 'deleteRecord', record })}
          onInvalid={addError}
        />
      </Panel>

      {dialog?.kind === 'deleteRecord' && (
        <Modal
          isOpen
          title={t('records.deleteTitle')}
          onClose={() => setDialog(null)}
          onSubmit={() => void submit(async () => {
            await api.deleteMyWorkRecord(dialog.record.id)
            addSuccess(t('records.deleted'))
          }, t('work.deleteFailed'))}
          submitLabel={t('common.delete')}
          submitVariant="danger"
          submitting={submitting}
        >
          <p>{t('records.deleteConfirm', { date: dialog.record.date })}</p>
        </Modal>
      )}
      {dialog?.kind === 'deleteTodo' && (
        <Modal
          isOpen
          title={t('todos.deleteTitle')}
          onClose={() => setDialog(null)}
          onSubmit={() => void submit(async () => {
            await api.deleteMyTodo(dialog.todo.id)
            addSuccess(t('todos.deleted'))
          }, t('work.deleteFailed'))}
          submitLabel={t('common.delete')}
          submitVariant="danger"
          submitting={submitting}
        >
          <p>{t('todos.deleteConfirm')}</p>
        </Modal>
      )}
    </div>
  )
}
