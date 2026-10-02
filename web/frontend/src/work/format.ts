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

// a period of `weeks` weeks ending with the week of date
export function periodStart(date: string, weeks: number): string {
  return addDays(weekStart(date), -7 * (weeks - 1))
}

export function periodEnd(start: string, weeks: number): string {
  return addDays(start, 7 * weeks - 1)
}

export type Quarter = 1 | 2 | 3 | 4

export function quarterOf(date: string): Quarter {
  return (Math.floor((parseDate(date).getMonth()) / 3) + 1) as Quarter
}

// the inclusive YYYY-MM-DD range of a quarter, e.g. Q2 2026 = 04-01..06-30
export function quarterRange(year: number, quarter: Quarter): { from: string, to: string } {
  const firstMonth = (quarter - 1) * 3
  const from = toDateString(new Date(year, firstMonth, 1))
  const to = toDateString(new Date(year, firstMonth + 3, 0))
  return { from, to }
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
  return hours >= 0 && hours <= 24 && Number.isInteger(hours * 2)
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
