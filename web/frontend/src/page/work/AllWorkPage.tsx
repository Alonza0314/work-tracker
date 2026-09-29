import { useCallback, useEffect, useState } from 'react'
import type { WorkMember, WorkRecordListResponse } from '../../api'
import { api } from '../../apiClient'
import Button from '../../components/button/button'
import { Field, SelectInput } from '../../components/field/Field'
import NotificationContainer from '../../components/notifications/NotificationContainer'
import Panel from '../../components/panel/Panel'
import WeekNavigator from '../../components/weekNavigator/WeekNavigator'
import { useNotifications } from '../../hooks/useNotifications'
import { useI18n } from '../../i18n/useI18n'
import { today, weekEnd, weekStart } from '../../work/format'
import { useWork } from '../../work/useWork'
import WorkRecordTable from './WorkRecordTable'
import styles from './work.module.css'

interface Filters {
  account: string
  categoryId: string
  projectId: string
}

const EMPTY_FILTERS: Filters = { account: '', categoryId: '', projectId: '' }

export default function AllWorkPage() {
  const { t } = useI18n()
  const { categories, projects } = useWork()
  const { errors, successes, addError, removeNotification } = useNotifications()

  const [members, setMembers] = useState<WorkMember[]>([])
  const [filters, setFilters] = useState<Filters>(EMPTY_FILTERS)
  // Monday of the shown week
  const [week, setWeek] = useState(() => weekStart(today()))
  const [result, setResult] = useState<WorkRecordListResponse | null>(null)

  useEffect(() => {
    api.listWorkMembers()
      .then((response) => setMembers(response.data.members))
      .catch(() => addError(t('work.loadFailed')))
  }, [addError, t])

  const loadRecords = useCallback(async () => {
    try {
      const { account, categoryId, projectId } = filters
      const response = await api.listAllWorkRecords(
        week,
        weekEnd(week),
        account || undefined,
        categoryId || undefined,
        projectId || undefined,
      )
      setResult(response.data)
    } catch {
      addError(t('work.loadFailed'))
    }
  }, [addError, filters, week, t])

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
        <div className={styles.weekBar}>
          <WeekNavigator start={week} onChange={setWeek} />
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
