import { House, type LucideIcon } from 'lucide-react'
import type { MessageKey } from '../../i18n/locales/en'

export interface NavItem {
  to: string
  label: MessageKey
  icon: LucideIcon
}

export const NAV_ITEMS: NavItem[] = [
  { to: '/', label: 'nav.home', icon: House },
]

export function findNavItem(pathname: string): NavItem | undefined {
  return NAV_ITEMS.find((item) => item.to === pathname)
}
