import type { SaveWorkEntryRequest, WorkRecord } from '../../api'
import type { MessageKey } from '../../i18n/locales/en'
import { isValidHours, today } from '../../work/format'

// nothing is required: records always carry a date (today by default),
// todos may have none, and empty hours count as 0
export type EntryKind = 'record' | 'todo'

// the editable form of an entry: every field as its input's string value
export interface EntryDraft {
  date: string
  categoryId: string
  description: string
  hours: string
  projectId: string
}

// a new record starts on today, a new todo without a date
export function draftFrom(entry: WorkRecord | undefined, kind: EntryKind): EntryDraft {
  return {
    date: entry ? entry.date : (kind === 'record' ? today() : ''),
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

export function draftToRequest(draft: EntryDraft): DraftResult {
  let hours: number | undefined
  if (draft.hours.trim() !== '') {
    hours = Number(draft.hours)
    if (!isValidHours(hours)) {
      return { error: 'work.hoursStep' }
    }
  }

  return {
    request: {
      date: draft.date || undefined,
      categoryId: draft.categoryId || undefined,
      description: draft.description.trim(),
      hours,
      projectId: draft.projectId || undefined,
    },
  }
}
