import { ChevronLeft, ChevronRight } from 'lucide-react'
import { useI18n } from '../../i18n/useI18n'
import { addDays, parseDate, periodEnd, periodStart, today } from '../../work/format'
import styles from './week-navigator.module.css'

interface WeekNavigatorProps {
  // Monday of the first shown week (YYYY-MM-DD)
  start: string
  onChange: (start: string) => void
  // how many weeks are shown at once; the arrows move by that many
  weeks?: number
}

export default function WeekNavigator({ start, onChange, weeks = 1 }: WeekNavigatorProps) {
  const { t, locale } = useI18n()
  const current = periodStart(today(), weeks)
  // e.g. 2026/09/28 – 10/04; the end repeats the year only across a new year
  const from = parseDate(start)
  const to = parseDate(periodEnd(start, weeks))
  const full = new Intl.DateTimeFormat(locale, { year: 'numeric', month: '2-digit', day: '2-digit' })
  const short = new Intl.DateTimeFormat(locale, { month: '2-digit', day: '2-digit' })
  const label = `${full.format(from)} – ${(from.getFullYear() === to.getFullYear() ? short : full).format(to)}`
  const prev = weeks === 1 ? t('week.prev') : t('week.prevN', { count: weeks })
  const next = weeks === 1 ? t('week.next') : t('week.nextN', { count: weeks })

  return (
    <div className={styles.navigator}>
      <button
        type="button"
        className={styles.arrow}
        onClick={() => onChange(addDays(start, -7 * weeks))}
        aria-label={prev}
        title={prev}
      >
        <ChevronLeft size={16} aria-hidden="true" />
      </button>
      <span className={styles.label} aria-live="polite">{label}</span>
      <button
        type="button"
        className={styles.arrow}
        onClick={() => onChange(addDays(start, 7 * weeks))}
        aria-label={next}
        title={next}
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
