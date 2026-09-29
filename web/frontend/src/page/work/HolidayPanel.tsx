import { useCallback, useEffect, useState, type FormEvent } from 'react'
import { RefreshCw, Trash2 } from 'lucide-react'
import { HolidaySource, HolidayType, type Holiday } from '../../api'
import { api } from '../../apiClient'
import Badge from '../../components/badge/Badge'
import Button from '../../components/button/button'
import { SelectInput, TextInput } from '../../components/field/Field'
import Modal from '../../components/modal/modal'
import Panel from '../../components/panel/Panel'
import { useI18n } from '../../i18n/useI18n'
import { parseDate, today } from '../../work/format'
import styles from './work.module.css'

interface HolidayPanelProps {
  onError: (message: string) => void
  onSuccess: (message: string) => void
}

// The holiday calendar used for the weekly required hours: government
// entries are synced daily, admins add their own on top (manual wins).
export default function HolidayPanel({ onError, onSuccess }: HolidayPanelProps) {
  const { t, locale } = useI18n()
  const currentYear = new Date().getFullYear()
  const [year, setYear] = useState(currentYear)
  const [holidays, setHolidays] = useState<Holiday[]>([])
  const [lastSyncedAt, setLastSyncedAt] = useState<string | undefined>()
  const [busy, setBusy] = useState(false)
  const [deleting, setDeleting] = useState<Holiday | null>(null)

  const [date, setDate] = useState(today())
  const [name, setName] = useState('')
  const [type, setType] = useState<HolidayType>(HolidayType.Holiday)

  const load = useCallback(async () => {
    try {
      const response = await api.listHolidays(String(year))
      setHolidays(response.data.holidays)
      setLastSyncedAt(response.data.lastSyncedAt)
    } catch {
      onError(t('holiday.loadFailed'))
    }
  }, [onError, t, year])

  useEffect(() => {
    void load()
  }, [load])

  async function run(action: () => Promise<unknown>, success: string, failure: string) {
    setBusy(true)
    try {
      await action()
      await load()
      onSuccess(success)
      return true
    } catch {
      onError(failure)
      return false
    } finally {
      setBusy(false)
    }
  }

  async function handleAdd(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const trimmed = name.trim()
    if (!date || !trimmed) {
      onError(t('holiday.required'))
      return
    }
    if (await run(() => api.saveHoliday(date, { name: trimmed, type }), t('holiday.saved'), t('holiday.saveFailed'))) {
      setName('')
    }
  }

  const dateFormat = new Intl.DateTimeFormat(locale, { month: '2-digit', day: '2-digit', weekday: 'short' })
  const syncedFormat = new Intl.DateTimeFormat(locale, { dateStyle: 'medium', timeStyle: 'short' })

  return (
    <Panel
      title={t('holiday.title')}
      description={lastSyncedAt
        ? t('holiday.lastSynced', { time: syncedFormat.format(new Date(lastSyncedAt)) })
        : t('holiday.neverSynced')}
      actions={(
        <>
          <SelectInput value={year} onChange={(event) => setYear(Number(event.target.value))} aria-label={t('holiday.year')}>
            {[currentYear - 1, currentYear, currentYear + 1].map((option) => (
              <option key={option} value={option}>{option}</option>
            ))}
          </SelectInput>
          <Button
            variant="secondary"
            icon={<RefreshCw aria-hidden="true" />}
            disabled={busy}
            onClick={() => void run(() => api.syncHolidays(), t('holiday.synced'), t('holiday.syncFailed'))}
          >
            {t('holiday.sync')}
          </Button>
        </>
      )}
      flush
    >
      <p className={styles.holidayHint}>{t('holiday.desc')}</p>

      <form className={styles.holidayForm} onSubmit={handleAdd}>
        <TextInput type="date" value={date} onChange={(event) => setDate(event.target.value)} aria-label={t('work.date')} required />
        <TextInput value={name} onChange={(event) => setName(event.target.value)} placeholder={t('holiday.namePlaceholder')} aria-label={t('holiday.name')} />
        <SelectInput value={type} onChange={(event) => setType(event.target.value as HolidayType)} aria-label={t('holiday.type')}>
          <option value={HolidayType.Holiday}>{t('holiday.type.holiday')}</option>
          <option value={HolidayType.Workday}>{t('holiday.type.workday')}</option>
        </SelectInput>
        <Button type="submit" disabled={busy}>{t('holiday.add')}</Button>
      </form>

      {holidays.length === 0 ? (
        <p className={styles.empty}>{t('holiday.empty')}</p>
      ) : (
        <ul className={`${styles.optionList} ${styles.holidayList}`}>
          {holidays.map((holiday) => (
            <li key={holiday.date} className={styles.optionItem}>
              <span className={styles.holidayDate}>{dateFormat.format(parseDate(holiday.date))}</span>
              <span className={styles.optionName}>{holiday.name}</span>
              <Badge tone={holiday.type === HolidayType.Holiday ? 'neutral' : 'primary'}>
                {t(holiday.type === HolidayType.Holiday ? 'holiday.type.holiday' : 'holiday.type.workday')}
              </Badge>
              <Badge tone={holiday.source === HolidaySource.Manual ? 'navy' : 'neutral'}>
                {t(holiday.source === HolidaySource.Manual ? 'holiday.source.manual' : 'holiday.source.gov')}
              </Badge>
              {holiday.source === HolidaySource.Manual ? (
                <button
                  type="button"
                  className={`${styles.iconButton} ${styles.danger}`}
                  onClick={() => setDeleting(holiday)}
                  title={t('common.delete')}
                  aria-label={`${t('common.delete')} ${holiday.date} ${holiday.name}`}
                >
                  <Trash2 size={16} aria-hidden="true" />
                </button>
              ) : (
                <span className={styles.iconSpacer} aria-hidden="true" />
              )}
            </li>
          ))}
        </ul>
      )}

      {deleting && (
        <Modal
          isOpen
          title={t('holiday.deleteTitle')}
          onClose={() => setDeleting(null)}
          onSubmit={() => void run(() => api.deleteHoliday(deleting.date), t('holiday.deleted'), t('holiday.deleteFailed'))
            .then((ok) => ok && setDeleting(null))}
          submitLabel={t('common.delete')}
          submitVariant="danger"
          submitting={busy}
        >
          <p>{t('holiday.deleteConfirm', { date: deleting.date, name: deleting.name })}</p>
        </Modal>
      )}
    </Panel>
  )
}
