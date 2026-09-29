import { useState } from 'react'
import type { CompleteTodoRequest, WorkRecord } from '../../api'
import { Field, SelectInput, TextArea, TextInput } from '../../components/field/Field'
import Modal from '../../components/modal/modal'
import { useI18n } from '../../i18n/useI18n'
import { selectableOptions } from '../../work/format'
import { useWork } from '../../work/useWork'
import { draftFrom, draftToRequest } from './entryDraft'
import styles from './work.module.css'

interface CompleteTodoModalProps {
  todo: WorkRecord
  submitting: boolean
  onClose: () => void
  onSubmit: (request: CompleteTodoRequest) => void
  onInvalid: (message: string) => void
}

// Fills in what a work record needs (category, hours) for a todo that lacks
// it. The date defaults to today and the description is kept.
export default function CompleteTodoModal({ todo, submitting, onClose, onSubmit, onInvalid }: CompleteTodoModalProps) {
  const { t } = useI18n()
  const { categories, projects } = useWork()
  // draftFrom() without an entry gives today's date
  const [draft, setDraft] = useState({ ...draftFrom(todo), date: draftFrom().date })

  function handleSubmit() {
    const result = draftToRequest(draft, 'record')
    if ('error' in result) {
      onInvalid(t(result.error))
      return
    }
    const { date, categoryId, hours, projectId } = result.request
    onSubmit({ date, categoryId, hours, projectId })
  }

  return (
    <Modal
      isOpen
      title={t('todos.completeTitle')}
      onClose={onClose}
      onSubmit={handleSubmit}
      submitLabel={t('todos.completeSubmit')}
      submitting={submitting}
    >
      <p className={styles.notice}>{t('todos.completeHint')}</p>
      <div className={styles.formRow}>
        <Field id="complete-date" label={t('work.date')}>
          <TextInput
            id="complete-date"
            type="date"
            value={draft.date}
            onChange={(event) => setDraft({ ...draft, date: event.target.value })}
            required
          />
        </Field>
        <Field id="complete-hours" label={t('work.hours')}>
          <TextInput
            id="complete-hours"
            type="number"
            inputMode="decimal"
            min={0.5}
            max={24}
            step={0.5}
            value={draft.hours}
            onChange={(event) => setDraft({ ...draft, hours: event.target.value })}
            placeholder="hr"
            required
          />
        </Field>
      </div>
      <Field id="complete-category" label={t('work.category')}>
        <SelectInput
          id="complete-category"
          value={draft.categoryId}
          onChange={(event) => setDraft({ ...draft, categoryId: event.target.value })}
          required
        >
          <option value="" disabled>{t('work.select')}</option>
          {selectableOptions(categories, todo.categoryId).map((option) => (
            <option key={option.id} value={option.id}>{option.name}</option>
          ))}
        </SelectInput>
      </Field>
      <Field id="complete-project" label={`${t('work.project')}（${t('work.optional')}）`}>
        <SelectInput
          id="complete-project"
          value={draft.projectId}
          onChange={(event) => setDraft({ ...draft, projectId: event.target.value })}
        >
          <option value="">{t('work.none')}</option>
          {selectableOptions(projects, todo.projectId).map((option) => (
            <option key={option.id} value={option.id}>{option.name}</option>
          ))}
        </SelectInput>
      </Field>
      <Field id="complete-description" label={t('work.description')}>
        <TextArea id="complete-description" value={draft.description} disabled />
      </Field>
    </Modal>
  )
}
