import { createContext } from 'react'
import type { WorkOption } from '../api'

export interface WorkContextValue {
  categories: WorkOption[]
  projects: WorkOption[]
  allowViewAll: boolean
  // false until the first load finishes (successfully or not)
  loaded: boolean
  // admins always see everyone's table; others only when allowViewAll is on
  canViewAll: boolean
  reload: () => Promise<void>
}

export const WorkContext = createContext<WorkContextValue | null>(null)
