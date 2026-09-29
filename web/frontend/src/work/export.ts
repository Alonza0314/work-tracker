import type { WorkOption, WorkRecord } from '../api'
import { optionName } from './format'

// A table of work records for a spreadsheet: one row per record, hours as a
// number so it can be summed.
export interface ExportTable {
  headers: string[]
  rows: (string | number)[][]
}

export function workRecordTable(
  records: WorkRecord[],
  headers: string[],
  memberName: (account: string) => string,
  categories: WorkOption[],
  projects: WorkOption[],
): ExportTable {
  return {
    headers,
    rows: records.map((record) => [
      record.date,
      memberName(record.account),
      record.account,
      record.categoryId ? optionName(categories, record.categoryId) : '',
      record.description,
      record.hours || '',
      record.projectId ? optionName(projects, record.projectId) : '',
    ]),
  }
}

export function saveBlob(blob: Blob, fileName: string): void {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = fileName
  document.body.appendChild(link)
  link.click()
  link.remove()
  // revoking right away can cut the download off before the browser reads it
  setTimeout(() => URL.revokeObjectURL(url), 1000)
}

function csvField(value: string | number): string {
  const text = String(value)
  return /[",\r\n]/.test(text) ? `"${text.replace(/"/g, '""')}"` : text
}

// UTF-8 with a BOM, so Excel opens Chinese text correctly
export function downloadCsv(table: ExportTable, fileName: string): void {
  const lines = [table.headers, ...table.rows].map((row) => row.map(csvField).join(','))
  saveBlob(new Blob(['﻿' + lines.join('\r\n') + '\r\n'], { type: 'text/csv;charset=utf-8' }), fileName)
}

// the xlsx writer is loaded only when someone exports
export async function downloadXlsx(table: ExportTable, fileName: string, sheet: string): Promise<void> {
  const { default: writeXlsxFile } = await import('write-excel-file/browser')

  const header = table.headers.map((value) => ({ value, fontWeight: 'bold' as const }))
  const rows = table.rows.map((row) => row.map((value) => (
    typeof value === 'number' ? { value, type: Number } : { value, type: String, wrap: true }
  )))

  await writeXlsxFile([header, ...rows], {
    sheet,
    stickyRowsCount: 1,
    columns: [{ width: 12 }, { width: 14 }, { width: 12 }, { width: 14 }, { width: 60 }, { width: 10 }, { width: 18 }],
  }).toFile(fileName)
}
