import { ClipboardList, House, Settings2, UserCog, Users, UsersRound, type LucideIcon } from 'lucide-react'
import type { MessageKey } from '../../i18n/locales/en'

export interface NavItem {
  to: string
  label: MessageKey
  icon: LucideIcon
  // who sees the entry: everyone (default), admins, or whoever can view all work
  access?: 'admin' | 'viewAll'
}

export const NAV_ITEMS: NavItem[] = [
  { to: '/', label: 'nav.home', icon: House },
  { to: '/work/all', label: 'nav.allWork', icon: UsersRound, access: 'viewAll' },
  { to: '/work/me', label: 'nav.myWork', icon: ClipboardList },
  { to: '/users', label: 'nav.users', icon: Users, access: 'admin' },
  { to: '/work/settings', label: 'nav.workSettings', icon: Settings2, access: 'admin' },
  { to: '/profile', label: 'nav.profile', icon: UserCog },
]

export function canAccess(item: NavItem, isAdmin: boolean, canViewAll: boolean): boolean {
  switch (item.access) {
    case 'admin':
      return isAdmin
    case 'viewAll':
      return canViewAll
    default:
      return true
  }
}

export function findNavItem(pathname: string): NavItem | undefined {
  return NAV_ITEMS.find((item) => item.to === pathname)
}
