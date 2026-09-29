import { House, UserCog, Users, type LucideIcon } from 'lucide-react'
import type { MessageKey } from '../../i18n/locales/en'

export interface NavItem {
  to: string
  label: MessageKey
  icon: LucideIcon
  adminOnly?: boolean
}

export const NAV_ITEMS: NavItem[] = [
  { to: '/', label: 'nav.home', icon: House },
  { to: '/users', label: 'nav.users', icon: Users, adminOnly: true },
  { to: '/profile', label: 'nav.profile', icon: UserCog },
]

export function findNavItem(pathname: string): NavItem | undefined {
  return NAV_ITEMS.find((item) => item.to === pathname)
}
