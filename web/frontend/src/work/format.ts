import type { CSSProperties } from 'react'
import type { CategoryColor, WorkOption } from '../api'

// dates are YYYY-MM-DD strings in the user's time zone

function toDateString(date: Date): string {
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${date.getFullYear()}-${month}-${day}`
}

export function parseDate(date: string): Date {
  const [year, month, day] = date.split('-').map(Number)
  return new Date(year, month - 1, day)
}

export function today(): string {
  return toDateString(new Date())
}

export function addDays(date: string, days: number): string {
  const result = parseDate(date)
  result.setDate(result.getDate() + days)
  return toDateString(result)
}

// weeks run Monday to Sunday
export function weekStart(date: string): string {
  const day = parseDate(date).getDay()
  return addDays(date, -((day + 6) % 7))
}

export function weekEnd(start: string): string {
  return addDays(start, 6)
}

export function optionName(options: WorkOption[], id: string): string {
  return options.find((option) => option.id === id)?.name ?? '—'
}

// options selectable for an entry: active ones plus the one it already uses
export function selectableOptions(options: WorkOption[], currentId?: string): WorkOption[] {
  return options.filter((option) => option.active || option.id === currentId)
}

// hours are positive half-hour steps up to 24
export function isValidHours(hours: number): boolean {
  return hours > 0 && hours <= 24 && Number.isInteger(hours * 2)
}

export function optionColor(options: WorkOption[], id: string): CategoryColor | undefined {
  return options.find((option) => option.id === id)?.color
}

// CSS variables of a category color, for inline styles
export function categoryColorStyle(color?: CategoryColor): CSSProperties | undefined {
  if (!color) {
    return undefined
  }
  return {
    '--chip-bg': `var(--cat-${color}-bg)`,
    '--chip-fg': `var(--cat-${color}-fg)`,
    '--chip-dot': `var(--cat-${color}-dot)`,
  } as CSSProperties
}
