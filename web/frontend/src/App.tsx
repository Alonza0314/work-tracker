import LoginPage from './page/login/LoginPage'
import { Navigate, Route, Routes } from 'react-router-dom'
import AppLayout from './components/layout/AppLayout'
import HomePage from './page/home/HomePage'
import ProfilePage from './page/profile/ProfilePage'
import UsersPage from './page/users/UsersPage'
import AllWorkPage from './page/work/AllWorkPage'
import MyWorkPage from './page/work/MyWorkPage'
import WorkSettingsPage from './page/work/WorkSettingsPage'
import { useAuth } from './auth/useAuth'
import WorkProvider from './work/WorkProvider'
import { useWork } from './work/useWork'

function RequireAuth({ children }: { children: React.ReactNode }) {
  const { session } = useAuth()
  if (!session) {
    return <Navigate to="/login" replace />
  }

  return <>{children}</>
}

// UI guards only; the backend enforces the same rules on every request

function RequireAdmin({ children }: { children: React.ReactNode }) {
  const { isAdmin } = useAuth()
  if (!isAdmin) {
    return <Navigate to="/" replace />
  }

  return <>{children}</>
}

function RequireViewAll({ children }: { children: React.ReactNode }) {
  const { canViewAll, loaded } = useWork()
  if (!loaded) {
    return null
  }
  if (!canViewAll) {
    return <Navigate to="/" replace />
  }

  return <>{children}</>
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route
        element={(
          <RequireAuth>
            <WorkProvider>
              <AppLayout />
            </WorkProvider>
          </RequireAuth>
        )}
      >
        <Route path="/" element={<HomePage />} />
        <Route path="/work/me" element={<MyWorkPage />} />
        <Route
          path="/work/all"
          element={(
            <RequireViewAll>
              <AllWorkPage />
            </RequireViewAll>
          )}
        />
        <Route
          path="/work/settings"
          element={(
            <RequireAdmin>
              <WorkSettingsPage />
            </RequireAdmin>
          )}
        />
        <Route path="/profile" element={<ProfilePage />} />
        <Route
          path="/users"
          element={(
            <RequireAdmin>
              <UsersPage />
            </RequireAdmin>
          )}
        />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
