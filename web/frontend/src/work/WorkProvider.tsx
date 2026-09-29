import { useCallback, useEffect, useMemo, useState, type ReactNode } from 'react'
import type { WorkOption } from '../api'
import { api } from '../apiClient'
import { useAuth } from '../auth/useAuth'
import { WorkContext } from './context'

// Loads the categories, projects and work table setting once per session;
// the settings page calls reload() after changing them.
export default function WorkProvider({ children }: { children: ReactNode }) {
  const { isAdmin } = useAuth()
  const [categories, setCategories] = useState<WorkOption[]>([])
  const [projects, setProjects] = useState<WorkOption[]>([])
  const [allowViewAll, setAllowViewAll] = useState(false)
  const [loaded, setLoaded] = useState(false)

  const reload = useCallback(async () => {
    const response = await api.getWorkOptions()
    setCategories(response.data.categories)
    setProjects(response.data.projects)
    setAllowViewAll(response.data.allowViewAll)
  }, [])

  useEffect(() => {
    reload()
      .catch(() => {
        // pages that need the options show their own load errors
      })
      .finally(() => setLoaded(true))
  }, [reload])

  const value = useMemo(() => ({
    categories,
    projects,
    allowViewAll,
    loaded,
    canViewAll: isAdmin || allowViewAll,
    reload,
  }), [categories, projects, allowViewAll, loaded, isAdmin, reload])

  return <WorkContext.Provider value={value}>{children}</WorkContext.Provider>
}
