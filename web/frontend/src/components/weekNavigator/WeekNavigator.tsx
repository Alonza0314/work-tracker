import { ChevronLeft, ChevronRight } from 'lucide-react'
import { useI18n } from '../../i18n/useI18n'
import { addDays, parseDate, today, weekEnd, weekStart } from '../../work/format'
import styles from './week-navigator.module.css'

interface WeekNavigatorProps {
  // Monday of the shown week (YYYY-MM-DD)
  start: string
  onChange: (start: string) => void
}

export default function WeekNavigator({ start, onChange }: WeekNavigatorProps) {
  const { t, locale } = useI18n()
  const current = weekStart(today())
  // e.g. 2026/09/28 – 10/04; the end repeats the year only across a new year
  const from = parseDate(start)
  const to = parseDate(weekEnd(start))
  const full = new Intl.DateTimeFormat(locale, { year: 'numeric', month: '2-digit', day: '2-digit' })
  const short = new Intl.DateTimeFormat(locale, { month: '2-digit', day: '2-digit' })
  const label = `${full.format(from)} – ${(from.getFullYear() === to.getFullYear() ? short : full).format(to)}`

  return (
    <div className={styles.navigator}>
      <button
        type="button"
        className={styles.arrow}
        onClick={() => onChange(addDays(start, -7))}
        aria-label={t('week.prev')}
        title={t('week.prev')}
      >
        <ChevronLeft size={16} aria-hidden="true" />
      </button>
      <span className={styles.label} aria-live="polite">{label}</span>
      <button
        type="button"
        className={styles.arrow}
        onClick={() => onChange(addDays(start, 7))}
        aria-label={t('week.next')}
        title={t('week.next')}
      >
        <ChevronRight size={16} aria-hidden="true" />
      </button>
      <button
        type="button"
        className={styles.current}
        onClick={() => onChange(current)}
        disabled={start === current}
      >
        {t('week.current')}
      </button>
    </div>
  )
}
