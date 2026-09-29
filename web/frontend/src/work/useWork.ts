import { useContext } from 'react'
import { WorkContext, type WorkContextValue } from './context'

export function useWork(): WorkContextValue {
  const context = useContext(WorkContext)
  if (!context) {
    throw new Error('useWork must be used within WorkProvider')
  }
  return context
}
