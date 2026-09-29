import { Inbox } from 'lucide-react'
import type { WorkRecord } from '../../api'
import { useI18n } from '../../i18n/useI18n'
import CategoryChip from '../../components/categoryChip/CategoryChip'
import { optionColor, optionName } from '../../work/format'
import { useWork } from '../../work/useWork'
import styles from './work.module.css'

interface WorkRecordTableProps {
  records: WorkRecord[]
  // shows a member column
  memberName?: (account: string) => string
  emptyText: string
}

// read-only records table (the everyone's work page); own records are edited
// in EditableWorkTable
export default function WorkRecordTable({ records, memberName, emptyText }: WorkRecordTableProps) {
  const { t } = useI18n()
  const { categories, projects } = useWork()

  if (records.length === 0) {
    return (
      <div className={styles.empty}>
        <span className={styles.emptyIcon}><Inbox size={20} aria-hidden="true" /></span>
        <p>{emptyText}</p>
      </div>
    )
  }

  return (
    <div className={styles.tableWrap}>
      <table className={styles.table}>
        <thead>
          <tr>
            <th>{t('work.date')}</th>
            {memberName && <th>{t('work.member')}</th>}
            <th>{t('work.category')}</th>
            <th className={styles.descriptionCell}>{t('work.description')}</th>
            <th className={styles.numeric}>{t('work.hours')}</th>
            <th>{t('work.project')}</th>
          </tr>
        </thead>
        <tbody>
          {records.map((record) => (
            <tr key={record.id}>
              <td className={styles.date}>{record.date}</td>
              {memberName && <td>{memberName(record.account)}</td>}
              <td>
                {record.categoryId
                  ? <CategoryChip name={optionName(categories, record.categoryId)} color={optionColor(categories, record.categoryId)} />
                  : '—'}
              </td>
              <td className={styles.descriptionCell}>{record.description}</td>
              <td className={styles.numeric}>{record.hours ? t('work.hoursValue', { hours: record.hours }) : '—'}</td>
              <td>{optionName(projects, record.projectId)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
