import { useEffect, useState } from 'react'
import { CircleAlert, CircleCheck } from 'lucide-react'
import type { MissingEntriesResponse } from '../../api'
import { api } from '../../apiClient'
import Panel from '../../components/panel/Panel'
import { useI18n } from '../../i18n/useI18n'
import { addDays, parseDate, today } from '../../work/format'
import { useWork } from '../../work/useWork'
import styles from './home-page.module.css'

interface MissingPanelProps {
  onError: (message: string) => void
}

// Who missed logging work in the 30 days up to yesterday: workdays without a
// single record. Shown to whoever may view everyone's work.
export default function MissingPanel({ onError }: MissingPanelProps) {
  const { t, locale } = useI18n()
  const { startDate } = useWork()
  const [result, setResult] = useState<MissingEntriesResponse | null>(null)

  useEffect(() => {
    api.listMissingEntries(addDays(today(), -1))
      .then((response) => setResult(response.data))
      .catch(() => onError(t('missing.loadFailed')))
  }, [onError, t])

  if (!result) {
    return null
  }

  const rangeFormat = new Intl.DateTimeFormat(locale, { month: 'numeric', day: 'numeric' })
  const dateFormat = new Intl.DateTimeFormat(locale, { month: '2-digit', day: '2-digit', weekday: 'short' })
  const totalMissing = result.members.reduce((sum, member) => sum + member.missingCount, 0)
  const maxMissing = Math.max(1, ...result.members.map((member) => member.missingCount))

  return (
    <Panel
      title={t('missing.title')}
      description={`${t('missing.desc', {
        from: rangeFormat.format(parseDate(result.from)),
        to: rangeFormat.format(parseDate(result.to)),
        workdays: result.workdays,
      })}${result.from === startDate ? t('missing.fromStart') : ''}`}
      actions={(
        <div className={styles.missingSummary}>
          <div className={styles.missingSummaryItem}>
            <span className={styles.missingSummaryLabel}>{t('missing.membersMissing')}</span>
            <span className={styles.missingSummaryValue}>{result.members.length} / {result.checkedCount}</span>
          </div>
          <div className={styles.missingSummaryItem}>
            <span className={styles.missingSummaryLabel}>{t('missing.daysMissing')}</span>
            <span className={styles.missingSummaryValue}>{totalMissing}</span>
          </div>
        </div>
      )}
      flush
    >
      {result.members.length === 0 ? (
        <p className={styles.missingDone}>
          <CircleCheck size={18} aria-hidden="true" />
          {t('missing.allDone')}
        </p>
      ) : (
        <ul className={styles.missingList}>
          {result.members.map((member) => (
            <li key={member.account} className={styles.missingRow}>
              <div className={styles.missingWho}>
                <span className={styles.missingAvatar} aria-hidden="true">{(member.name || member.account).charAt(0).toUpperCase()}</span>
                <div className={styles.missingNames}>
                  <span className={styles.missingName}>{member.name || member.account}</span>
                  <span className={styles.missingAccount}>{member.account}</span>
                </div>
              </div>

              <div className={styles.missingCount}>
                <span className={styles.missingCountText}>
                  <CircleAlert size={14} aria-hidden="true" />
                  {t('missing.days', { count: member.missingCount })}
                </span>
                <span className={styles.missingTrack} aria-hidden="true">
                  <span className={styles.missingBar} style={{ width: `${(member.missingCount / maxMissing) * 100}%` }} />
                </span>
              </div>

              <ul className={styles.missingDates} aria-label={t('missing.datesOf', { name: member.name || member.account })}>
                {member.missingDates.map((date) => (
                  <li key={date} className={styles.missingDate}>{dateFormat.format(parseDate(date))}</li>
                ))}
              </ul>
            </li>
          ))}
        </ul>
      )}
    </Panel>
  )
}
