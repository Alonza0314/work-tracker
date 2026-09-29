import type { SaveWorkEntryRequest, WorkRecord } from '../../api'
import type { MessageKey } from '../../i18n/locales/en'
import { isValidHours, today } from '../../work/format'

// record: category and hours required; todo: only date and description
export type EntryKind = 'record' | 'todo'

// the editable form of an entry: every field as its input's string value
export interface EntryDraft {
  date: string
  categoryId: string
  description: string
  hours: string
  projectId: string
}

export function draftFrom(entry?: WorkRecord): EntryDraft {
  return {
    date: entry?.date ?? today(),
    categoryId: entry?.categoryId ?? '',
    description: entry?.description ?? '',
    hours: entry?.hours ? String(entry.hours) : '',
    projectId: entry?.projectId ?? '',
  }
}

export function sameDraft(a: EntryDraft, b: EntryDraft): boolean {
  return a.date === b.date &&
    a.categoryId === b.categoryId &&
    a.description.trim() === b.description.trim() &&
    Number(a.hours || 0) === Number(b.hours || 0) &&
    a.projectId === b.projectId
}

export type DraftResult =
  | { request: SaveWorkEntryRequest }
  | { error: MessageKey }

export function draftToRequest(draft: EntryDraft, kind: EntryKind): DraftResult {
  const required = kind === 'record'
  const description = draft.description.trim()
  if (!draft.date || !description || (required && !draft.categoryId)) {
    return { error: 'work.required' }
  }

  let hours: number | undefined
  if (draft.hours.trim() !== '') {
    hours = Number(draft.hours)
    if (!isValidHours(hours)) {
      return { error: 'work.hoursStep' }
    }
  } else if (required) {
    return { error: 'work.required' }
  }

  return {
    request: {
      date: draft.date,
      categoryId: draft.categoryId || undefined,
      description,
      hours,
      projectId: draft.projectId || undefined,
    },
  }
}
