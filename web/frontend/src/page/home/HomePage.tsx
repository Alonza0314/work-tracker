import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { CalendarDays, CircleCheck, ClipboardList, Clock, PenLine } from 'lucide-react'
import type { WeekDay, WeekSummaryResponse } from '../../api'
import { api } from '../../apiClient'
import Button from '../../components/button/button'
import NotificationContainer from '../../components/notifications/NotificationContainer'
import Panel from '../../components/panel/Panel'
import StatsCard from '../../components/stats/stats-card'
import { useNotifications } from '../../hooks/useNotifications'
import { useI18n } from '../../i18n/useI18n'
import { parseDate, today, weekStart } from '../../work/format'
import { useWork } from '../../work/useWork'
import MissingPanel from './MissingPanel'
import styles from './home-page.module.css'

// This week at a glance for the signed-in user: how much is logged against
// the hours the week requires (8 per workday, holidays taken out).
export default function HomePage() {
  const { t } = useI18n()
  const navigate = useNavigate()
  const { errors, successes, addError, removeNotification } = useNotifications()
  const [summary, setSummary] = useState<WeekSummaryResponse | null>(null)
  const { canViewAll, loaded } = useWork()

  useEffect(() => {
    api.getWeekSummary(weekStart(today()))
      .then((response) => setSummary(response.data))
      .catch(() => addError(t('home.loadFailed')))
  }, [addError, t])

  return (
    <div className={styles.page}>
      <NotificationContainer errors={errors} successes={successes} onClose={removeNotification} />

      <section className={styles.header}>
        <div>
          <h2 className={styles.title}>{t('home.title')}</h2>
          {summary && <p className={styles.subtitle}><WeekRange from={summary.from} to={summary.to} /></p>}
        </div>
        <Button icon={<PenLine aria-hidden="true" />} onClick={() => navigate('/work/me')}>
          {t('home.logWork')}
        </Button>
      </section>

      {summary && (
        <>
          <section className={styles.stats}>
            <StatsCard
              title={t('home.stat.records')}
              value={summary.recordCount}
              unit={t('home.unit.records')}
              icon={<ClipboardList size={16} aria-hidden="true" />}
            />
            <StatsCard
              title={t('home.stat.logged')}
              value={summary.loggedHours}
              unit={t('home.unit.hours')}
              icon={<Clock size={16} aria-hidden="true" />}
            />
            <StatsCard
              title={t('home.stat.required')}
              value={summary.requiredHours}
              unit={t('home.unit.hours')}
              description={summary.daysOff > 0 ? t('home.stat.daysOff', { days: summary.daysOff }) : t('home.stat.fullWeek')}
              icon={<CalendarDays size={16} aria-hidden="true" />}
            />
            <RemainingCard remaining={summary.remainingHours} />
          </section>

          <Panel title={t('home.progress.title')}>
            <ProgressMeter logged={summary.loggedHours} required={summary.requiredHours} />
          </Panel>

          <Panel title={t('home.daily.title')} description={t('home.daily.desc')}>
            <DailyChart days={summary.days} />
          </Panel>
        </>
      )}

      {loaded && canViewAll && <MissingPanel onError={addError} />}
    </div>
  )
}

function WeekRange({ from, to }: { from: string, to: string }) {
  const { locale } = useI18n()
  const format = new Intl.DateTimeFormat(locale, { month: 'long', day: 'numeric', weekday: 'short' })
  return <>{format.format(parseDate(from))} – {format.format(parseDate(to))}</>
}

// the one card whose state matters: done gets an icon and a word, not just color
function RemainingCard({ remaining }: { remaining: number }) {
  const { t } = useI18n()

  if (remaining <= 0) {
    return (
      <div className={`${styles.remaining} ${styles.remainingDone}`}>
        <span className={styles.remainingLabel}>{t('home.stat.remaining')}</span>
        <span className={styles.remainingValue}>
          <CircleCheck size={22} aria-hidden="true" />
          {t('home.stat.done')}
        </span>
        <span className={styles.remainingHint}>{t('home.stat.doneHint')}</span>
      </div>
    )
  }

  return (
    <div className={styles.remaining}>
      <span className={styles.remainingLabel}>{t('home.stat.remaining')}</span>
      <span className={styles.remainingValue}>
        {remaining}
        <span className={styles.remainingUnit}>{t('home.unit.hours')}</span>
      </span>
      <span className={styles.remainingHint}>{t('home.stat.remainingHint')}</span>
    </div>
  )
}

function ProgressMeter({ logged, required }: { logged: number, required: number }) {
  const { t } = useI18n()
  const ratio = required > 0 ? logged / required : 1
  const percent = Math.round(ratio * 100)

  return (
    <div className={styles.meter}>
      <div className={styles.meterText}>
        <span className={styles.meterValue}>
          {t('home.progress.value', { logged, required })}
        </span>
        <span className={styles.meterPercent}>{percent}%</span>
      </div>
      <div
        className={styles.meterTrack}
        role="progressbar"
        aria-valuemin={0}
        aria-valuemax={required}
        aria-valuenow={logged}
        aria-label={t('home.progress.title')}
      >
        <div className={styles.meterFill} style={{ width: `${Math.min(ratio, 1) * 100}%` }} />
      </div>
    </div>
  )
}

const CHART_HEIGHT = 180

// Logged hours per day as columns, with each workday's 8-hour target as a
// dashed tick and days off shaded and named. Hover a day for its numbers.
function DailyChart({ days }: { days: WeekDay[] }) {
  const { t, locale } = useI18n()
  const [hovered, setHovered] = useState<string | null>(null)
  const current = today()

  // headroom above the target; even steps for the gridlines
  const peak = Math.max(10, ...days.map((day) => Math.max(day.loggedHours, day.requiredHours)))
  const max = Math.ceil(peak / 2) * 2
  const gridlines = [0, max / 2, max]
  const y = (hours: number) => (hours / max) * CHART_HEIGHT

  const weekday = new Intl.DateTimeFormat(locale, { weekday: 'short' })
  const monthDay = new Intl.DateTimeFormat(locale, { month: 'numeric', day: 'numeric' })

  return (
    <div className={styles.chart}>
      <div className={styles.legend} aria-hidden="true">
        <span className={styles.legendItem}><span className={styles.legendBar} />{t('home.daily.logged')}</span>
        <span className={styles.legendItem}><span className={styles.legendTarget} />{t('home.daily.target')}</span>
        <span className={styles.legendItem}><span className={styles.legendOff} />{t('home.daily.off')}</span>
      </div>

      <div className={styles.plot} aria-hidden="true">
        <div className={styles.axis} style={{ height: CHART_HEIGHT }}>
          {gridlines.map((hours) => (
            <span key={hours} className={styles.axisLabel} style={{ bottom: y(hours) }}>{hours}</span>
          ))}
        </div>

        <div className={styles.columns} style={{ height: CHART_HEIGHT }}>
          {gridlines.map((hours) => (
            <span key={hours} className={styles.gridline} style={{ bottom: y(hours) }} />
          ))}

          {days.map((day) => {
            const date = parseDate(day.date)
            return (
              <div
                key={day.date}
                className={[
                  styles.column,
                  day.workday ? '' : styles.columnOff,
                  day.date === current ? styles.columnToday : '',
                ].join(' ')}
                onMouseEnter={() => setHovered(day.date)}
                onMouseLeave={() => setHovered(null)}
              >
                {!day.workday && day.holidayName && <span className={styles.offName}>{day.holidayName}</span>}
                {day.requiredHours > 0 && <span className={styles.target} style={{ bottom: y(day.requiredHours) }} />}
                {day.loggedHours > 0 && <span className={styles.bar} style={{ height: y(day.loggedHours) }} />}

                {hovered === day.date && (
                  <div className={styles.tooltip} style={{ bottom: y(Math.max(day.loggedHours, day.requiredHours)) + 10 }}>
                    <strong>{weekday.format(date)} {monthDay.format(date)}</strong>
                    {day.holidayName && <span>{day.holidayName}</span>}
                    <span>{t('home.daily.tooltipLogged', { hours: day.loggedHours })}</span>
                    <span>{day.workday ? t('home.daily.tooltipRequired', { hours: day.requiredHours }) : t('home.daily.off')}</span>
                  </div>
                )}
              </div>
            )
          })}
        </div>
      </div>

      <div className={styles.dayLabels} aria-hidden="true">
        {days.map((day) => {
          const date = parseDate(day.date)
          return (
            <div key={day.date} className={`${styles.dayLabel} ${day.date === current ? styles.dayLabelToday : ''}`}>
              <span>{weekday.format(date)}</span>
              <span className={styles.dayDate}>{monthDay.format(date)}</span>
            </div>
          )
        })}
      </div>

      {/* the same numbers for screen readers */}
      <table className={styles.srOnly}>
        <caption>{t('home.daily.title')}</caption>
        <thead>
          <tr>
            <th>{t('work.date')}</th>
            <th>{t('home.daily.logged')}</th>
            <th>{t('home.daily.target')}</th>
          </tr>
        </thead>
        <tbody>
          {days.map((day) => (
            <tr key={day.date}>
              <td>{day.date}{day.holidayName ? ` ${day.holidayName}` : ''}</td>
              <td>{day.loggedHours}</td>
              <td>{day.requiredHours}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

