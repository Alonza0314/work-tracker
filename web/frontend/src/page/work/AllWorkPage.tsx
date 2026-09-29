import { useCallback, useEffect, useMemo, useState } from 'react'
import type { WorkMember, WorkRecordListResponse } from '../../api'
import { api } from '../../apiClient'
import Button from '../../components/button/button'
import { Field, SelectInput } from '../../components/field/Field'
import NotificationContainer from '../../components/notifications/NotificationContainer'
import Panel from '../../components/panel/Panel'
import WeekNavigator from '../../components/weekNavigator/WeekNavigator'
import { useNotifications } from '../../hooks/useNotifications'
import { useI18n } from '../../i18n/useI18n'
import { parseDate, quarterOf, quarterRange, today, weekEnd, weekStart, type Quarter } from '../../work/format'
import { useWork } from '../../work/useWork'
import WorkRecordTable from './WorkRecordTable'
import styles from './work.module.css'

interface Filters {
  account: string
  categoryId: string
  projectId: string
}

const EMPTY_FILTERS: Filters = { account: '', categoryId: '', projectId: '' }

type PeriodMode = 'week' | 'quarter'

const QUARTERS: Quarter[] = [1, 2, 3, 4]

export default function AllWorkPage() {
  const { t, locale } = useI18n()
  const { categories, projects, startDate } = useWork()
  const { errors, successes, addError, removeNotification } = useNotifications()

  const [members, setMembers] = useState<WorkMember[]>([])
  const [filters, setFilters] = useState<Filters>(EMPTY_FILTERS)
  const [mode, setMode] = useState<PeriodMode>('week')
  // Monday of the shown week
  const [week, setWeek] = useState(() => weekStart(today()))
  const [year, setYear] = useState(() => parseDate(today()).getFullYear())
  const [quarter, setQuarter] = useState<Quarter>(() => quarterOf(today()))
  const [result, setResult] = useState<WorkRecordListResponse | null>(null)

  const range = useMemo(
    () => (mode === 'week' ? { from: week, to: weekEnd(week) } : quarterRange(year, quarter)),
    [mode, week, year, quarter],
  )

  // from the work start year (or two years back) to this year
  const currentYear = parseDate(today()).getFullYear()
  const firstYear = startDate ? Math.min(parseDate(startDate).getFullYear(), currentYear) : currentYear - 2
  const years = Array.from({ length: currentYear - firstYear + 1 }, (_, i) => currentYear - i)

  useEffect(() => {
    api.listWorkMembers()
      .then((response) => setMembers(response.data.members))
      .catch(() => addError(t('work.loadFailed')))
  }, [addError, t])

  const loadRecords = useCallback(async () => {
    try {
      const { account, categoryId, projectId } = filters
      const response = await api.listAllWorkRecords(
        range.from,
        range.to,
        account || undefined,
        categoryId || undefined,
        projectId || undefined,
      )
      setResult(response.data)
    } catch {
      addError(t('work.loadFailed'))
    }
  }, [addError, filters, range, t])

  useEffect(() => {
    void loadRecords()
  }, [loadRecords])

  function updateFilter<K extends keyof Filters>(key: K, value: Filters[K]) {
    setFilters((prev) => ({ ...prev, [key]: value }))
  }

  function memberName(account: string): string {
    return members.find((member) => member.account === account)?.name || account
  }

  const isFiltered = Object.values(filters).some((value) => value !== '')

  return (
    <div className={styles.page}>
      <NotificationContainer errors={errors} successes={successes} onClose={removeNotification} />

      <Panel
        title={t('nav.allWork')}
        actions={result && (
          <div className={styles.summary}>
            <div className={styles.summaryItem}>
              <span className={styles.summaryLabel}>{t('all.records')}</span>
              <span className={styles.summaryValue}>{result.total}</span>
            </div>
            <div className={styles.summaryItem}>
              <span className={styles.summaryLabel}>{t('all.totalHours')}</span>
              <span className={styles.summaryValue}>{t('work.hoursValue', { hours: result.totalHours })}</span>
            </div>
          </div>
        )}
        flush
      >
        <div className={styles.periodBar}>
          <div className={styles.segmented} role="group" aria-label={t('period.label')}>
            {(['week', 'quarter'] as const).map((option) => (
              <button
                key={option}
                type="button"
                className={mode === option ? styles.segmentActive : styles.segment}
                aria-pressed={mode === option}
                onClick={() => setMode(option)}
              >
                {t(option === 'week' ? 'period.week' : 'period.quarter')}
              </button>
            ))}
          </div>

          {mode === 'week' ? (
            <WeekNavigator start={week} onChange={setWeek} />
          ) : (
            <div className={styles.quarterPicker}>
              <SelectInput value={year} onChange={(event) => setYear(Number(event.target.value))} aria-label={t('period.year')}>
                {years.map((option) => (
                  <option key={option} value={option}>{option}</option>
                ))}
              </SelectInput>
              <div className={styles.segmented} role="group" aria-label={t('period.quarter')}>
                {QUARTERS.map((option) => (
                  <button
                    key={option}
                    type="button"
                    className={quarter === option ? styles.segmentActive : styles.segment}
                    aria-pressed={quarter === option}
                    onClick={() => setQuarter(option)}
                  >
                    Q{option}
                  </button>
                ))}
              </div>
              <span className={styles.periodRange}>
                {new Intl.DateTimeFormat(locale, { month: '2-digit', day: '2-digit' }).format(parseDate(range.from))}
                {' – '}
                {new Intl.DateTimeFormat(locale, { month: '2-digit', day: '2-digit' }).format(parseDate(range.to))}
              </span>
            </div>
          )}
        </div>
        <div className={styles.filters}>
          <Field id="filter-member" label={t('work.member')}>
            <SelectInput id="filter-member" value={filters.account} onChange={(event) => updateFilter('account', event.target.value)}>
              <option value="">{t('all.allMembers')}</option>
              {members.map((member) => (
                <option key={member.account} value={member.account}>{member.name || member.account}</option>
              ))}
            </SelectInput>
          </Field>
          <Field id="filter-category" label={t('work.category')}>
            <SelectInput id="filter-category" value={filters.categoryId} onChange={(event) => updateFilter('categoryId', event.target.value)}>
              <option value="">{t('all.allCategories')}</option>
              {categories.map((option) => (
                <option key={option.id} value={option.id}>{option.name}</option>
              ))}
            </SelectInput>
          </Field>
          <Field id="filter-project" label={t('work.project')}>
            <SelectInput id="filter-project" value={filters.projectId} onChange={(event) => updateFilter('projectId', event.target.value)}>
              <option value="">{t('all.allProjects')}</option>
              {projects.map((option) => (
                <option key={option.id} value={option.id}>{option.name}</option>
              ))}
            </SelectInput>
          </Field>
          <Button
            variant="ghost"
            disabled={!isFiltered}
            onClick={() => setFilters(EMPTY_FILTERS)}
          >
            {t('all.reset')}
          </Button>
        </div>

        <WorkRecordTable records={result?.records ?? []} memberName={memberName} emptyText={t('records.empty')} />
      </Panel>
    </div>
  )
}
