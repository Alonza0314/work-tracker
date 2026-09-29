// English is the source of truth for message keys; every other locale must
// provide the same keys (enforced by the Messages type).
const en = {
  'app.name': 'Work Tracker',

  'common.language': 'Language',
  'common.cancel': 'Cancel',
  'common.submit': 'Submit',
  'common.close': 'Close',
  'common.openMenu': 'Open navigation',
  'common.closeMenu': 'Close navigation',

  'login.title': 'Welcome back',
  'login.subtitle': 'Sign in to continue to Work Tracker.',
  'login.username': 'Username',
  'login.usernamePlaceholder': 'Enter your username',
  'login.password': 'Password',
  'login.passwordPlaceholder': 'Enter your password',
  'login.showPassword': 'Show password',
  'login.hidePassword': 'Hide password',
  'login.submit': 'Sign in',
  'login.submitting': 'Signing in...',
  'login.success': 'Login successful',
  'login.failed': 'Login failed',

  'nav.home': 'Home',
  'nav.logout': 'Sign out',
} as const

export type MessageKey = keyof typeof en
export type Messages = Record<MessageKey, string>

export default en as Messages
