import { useLayoutEffect, useRef, useState, type KeyboardEvent, type ReactNode, type TextareaHTMLAttributes } from 'react'
import { Plus, Trash2 } from 'lucide-react'
import type { SaveWorkEntryRequest, WorkRecord } from '../../api'
import { useI18n } from '../../i18n/useI18n'
import { categoryColorStyle, optionColor, selectableOptions } from '../../work/format'
import { useWork } from '../../work/useWork'
import { draftFrom, draftToRequest, sameDraft, type EntryDraft, type EntryKind } from './entryDraft'
import styles from './work.module.css'

interface EditableWorkTableProps {
  kind: EntryKind
  rows: WorkRecord[]
  // resolve to false when the save failed (the caller reports the error)
  onCreate: (request: SaveWorkEntryRequest) => Promise<boolean>
  onUpdate: (row: WorkRecord, request: SaveWorkEntryRequest) => Promise<boolean>
  onDelete: (row: WorkRecord) => void
  onInvalid: (message: string) => void
  // leading cell per saved row (e.g. the todo checkbox)
  leading?: (row: WorkRecord) => ReactNode
  emptyText: string
}

type CellKeyEvent = KeyboardEvent<HTMLInputElement | HTMLTextAreaElement>

// A spreadsheet-like table: every cell is an input. Text, number and date
// cells save when they lose focus (Enter too; Shift+Enter adds a line in the
// description), selects save on change and Escape restores the saved value.
// The first row adds a new entry.
export default function EditableWorkTable({ kind, rows, onCreate, onUpdate, onDelete, onInvalid, leading, emptyText }: EditableWorkTableProps) {
  const { t } = useI18n()
  const hasLeading = leading !== undefined

  return (
    <div className={styles.tableWrap}>
      <table className={`${styles.table} ${styles.sheet}`}>
        <thead>
          <tr>
            {hasLeading && <th className={styles.leadingCell} aria-hidden="true" />}
            <th className={styles.dateCell}>{t('work.date')}</th>
            <th>{t('work.category')}</th>
            <th className={styles.descriptionCell}>{t('work.description')}</th>
            <th className={styles.hoursCell}>{t('work.hours')}</th>
            <th>{t('work.project')}</th>
            <th className={styles.actionsCell} aria-label={t('common.actions')} />
          </tr>
        </thead>
        <tbody>
          <NewEntryRow kind={kind} hasLeading={hasLeading} onCreate={onCreate} onInvalid={onInvalid} />
          {rows.map((row) => (
            <EntryRow
              // remount when the saved values change, resetting the draft
              key={`${row.id}:${row.date}:${row.categoryId}:${row.hours}:${row.projectId}:${row.description}`}
              kind={kind}
              row={row}
              leading={leading}
              onUpdate={onUpdate}
              onDelete={onDelete}
              onInvalid={onInvalid}
            />
          ))}
          {rows.length === 0 && (
            <tr>
              <td colSpan={hasLeading ? 7 : 6} className={styles.sheetEmpty}>{emptyText}</td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  )
}

interface EntryCellsProps {
  kind: EntryKind
  draft: EntryDraft
  // the saved entry, so inactive options it already uses stay selectable
  row?: WorkRecord
  disabled: boolean
  onChange: (draft: EntryDraft) => void
  // a select changed: saved rows commit right away
  onSelect: (draft: EntryDraft) => void
  onBlur: () => void
  onKeyDown: (event: CellKeyEvent) => void
  autoFocusRef?: React.Ref<HTMLInputElement>
}

function EntryCells({ kind, draft, row, disabled, onChange, onSelect, onBlur, onKeyDown, autoFocusRef }: EntryCellsProps) {
  const { t } = useI18n()
  const { categories, projects } = useWork()
  const isRecord = kind === 'record'

  return (
    <>
      <td className={styles.dateCell}>
        <input
          ref={autoFocusRef}
          type="date"
          className={styles.cellInput}
          value={draft.date}
          disabled={disabled}
          // a record always keeps a date: clearing it restores the last one
          onChange={(event) => onChange({ ...draft, date: event.target.value || (isRecord ? draft.date : '') })}
          onBlur={onBlur}
          onKeyDown={onKeyDown}
          aria-label={t('work.date')}
        />
      </td>
      <td>
        <select
          className={`${styles.cellInput} ${styles.cellSelect} ${draft.categoryId ? styles.categorySelect : ''}`}
          style={categoryColorStyle(optionColor(categories, draft.categoryId))}
          value={draft.categoryId}
          disabled={disabled}
          onChange={(event) => onSelect({ ...draft, categoryId: event.target.value })}
          aria-label={t('work.category')}
        >
          <option value="">—</option>
          {selectableOptions(categories, row?.categoryId).map((option) => (
            <option key={option.id} value={option.id}>{option.name}</option>
          ))}
        </select>
      </td>
      <td className={styles.descriptionCell}>
        <AutoHeightTextArea
          className={`${styles.cellInput} ${styles.cellTextArea}`}
          value={draft.description}
          disabled={disabled}
          placeholder={row ? undefined : t('work.descriptionPlaceholder')}
          onChange={(event) => onChange({ ...draft, description: event.target.value })}
          onBlur={onBlur}
          onKeyDown={onKeyDown}
          aria-label={t('work.description')}
        />
      </td>
      <td className={styles.hoursCell}>
        <input
          type="number"
          inputMode="decimal"
          min={0}
          max={24}
          step={0.5}
          className={`${styles.cellInput} ${styles.cellNumber}`}
          value={draft.hours}
          disabled={disabled}
          placeholder={isRecord ? '0' : '—'}
          onChange={(event) => onChange({ ...draft, hours: event.target.value })}
          onBlur={onBlur}
          onKeyDown={onKeyDown}
          aria-label={t('work.hours')}
        />
      </td>
      <td>
        <select
          className={`${styles.cellInput} ${styles.cellSelect}`}
          value={draft.projectId}
          disabled={disabled}
          onChange={(event) => onSelect({ ...draft, projectId: event.target.value })}
          aria-label={t('work.project')}
        >
          <option value="">—</option>
          {selectableOptions(projects, row?.projectId).map((option) => (
            <option key={option.id} value={option.id}>{option.name}</option>
          ))}
        </select>
      </td>
    </>
  )
}

// a one-line textarea that grows to show every line of its value
function AutoHeightTextArea(props: TextareaHTMLAttributes<HTMLTextAreaElement>) {
  const ref = useRef<HTMLTextAreaElement>(null)

  useLayoutEffect(() => {
    const el = ref.current
    if (el) {
      el.style.height = 'auto'
      el.style.height = `${el.scrollHeight + el.offsetHeight - el.clientHeight}px`
    }
  }, [props.value])

  return <textarea ref={ref} rows={1} {...props} />
}

interface EntryRowProps {
  kind: EntryKind
  row: WorkRecord
  leading?: (row: WorkRecord) => ReactNode
  onUpdate: (row: WorkRecord, request: SaveWorkEntryRequest) => Promise<boolean>
  onDelete: (row: WorkRecord) => void
  onInvalid: (message: string) => void
}

function EntryRow({ kind, row, leading, onUpdate, onDelete, onInvalid }: EntryRowProps) {
  const { t } = useI18n()
  const saved = draftFrom(row, kind)
  const [draft, setDraft] = useState(saved)
  const [saving, setSaving] = useState(false)
  // set by Escape so the blur that follows does not save
  const cancelled = useRef(false)

  async function commit(next: EntryDraft) {
    if (sameDraft(next, saved)) {
      setDraft(saved)
      return
    }
    const result = draftToRequest(next)
    if ('error' in result) {
      onInvalid(t(result.error))
      setDraft(saved)
      return
    }

    setSaving(true)
    const ok = await onUpdate(row, result.request)
    // on success the table remounts this row with the saved values
    if (!ok) {
      setDraft(saved)
      setSaving(false)
    }
  }

  function handleBlur() {
    if (cancelled.current) {
      cancelled.current = false
      setDraft(saved)
      return
    }
    void commit(draft)
  }

  function handleKeyDown(event: CellKeyEvent) {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault()
      event.currentTarget.blur()
    } else if (event.key === 'Escape') {
      cancelled.current = true
      event.currentTarget.blur()
    }
  }

  return (
    <tr className={saving ? styles.rowSaving : undefined}>
      {leading && <td className={styles.leadingCell}>{leading(row)}</td>}
      <EntryCells
        kind={kind}
        draft={draft}
        row={row}
        disabled={saving}
        onChange={setDraft}
        onSelect={(next) => {
          setDraft(next)
          void commit(next)
        }}
        onBlur={handleBlur}
        onKeyDown={handleKeyDown}
      />
      <td className={styles.actionsCell}>
        <button
          type="button"
          className={`${styles.iconButton} ${styles.danger}`}
          onClick={() => onDelete(row)}
          disabled={saving}
          title={t('common.delete')}
          aria-label={`${t('common.delete')} ${row.date} ${row.description}`}
        >
          <Trash2 size={16} aria-hidden="true" />
        </button>
      </td>
    </tr>
  )
}

interface NewEntryRowProps {
  kind: EntryKind
  hasLeading: boolean
  onCreate: (request: SaveWorkEntryRequest) => Promise<boolean>
  onInvalid: (message: string) => void
}

// the blank first row: filled in place, added with Enter or the + button
function NewEntryRow({ kind, hasLeading, onCreate, onInvalid }: NewEntryRowProps) {
  const { t } = useI18n()
  const [draft, setDraft] = useState(() => draftFrom(undefined, kind))
  const [saving, setSaving] = useState(false)
  const firstInput = useRef<HTMLInputElement>(null)

  async function create() {
    const result = draftToRequest(draft)
    if ('error' in result) {
      onInvalid(t(result.error))
      return
    }

    setSaving(true)
    const ok = await onCreate(result.request)
    setSaving(false)
    if (ok) {
      setDraft(draftFrom(undefined, kind))
      firstInput.current?.focus()
    }
  }

  return (
    <tr className={`${styles.newRow} ${saving ? styles.rowSaving : ''}`}>
      {hasLeading && <td className={styles.leadingCell} />}
      <EntryCells
        kind={kind}
        draft={draft}
        disabled={saving}
        onChange={setDraft}
        onSelect={setDraft}
        onBlur={() => {}}
        onKeyDown={(event) => {
          if (event.key === 'Enter' && !event.shiftKey) {
            event.preventDefault()
            void create()
          } else if (event.key === 'Escape') {
            setDraft(draftFrom(undefined, kind))
          }
        }}
        autoFocusRef={firstInput}
      />
      <td className={styles.actionsCell}>
        <button
          type="button"
          className={`${styles.iconButton} ${styles.addButton}`}
          onClick={() => void create()}
          disabled={saving}
          title={kind === 'record' ? t('records.add') : t('todos.add')}
          aria-label={kind === 'record' ? t('records.add') : t('todos.add')}
        >
          <Plus size={16} aria-hidden="true" />
        </button>
      </td>
    </tr>
  )
}
