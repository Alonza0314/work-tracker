import LoginPage from './page/login/LoginPage'
import { Navigate, Route, Routes } from 'react-router-dom'
import AppLayout from './components/layout/AppLayout'
import HomePage from './page/home/HomePage'
import ProfilePage from './page/profile/ProfilePage'
import UsersPage from './page/users/UsersPage'
import { useAuth } from './auth/useAuth'

function RequireAuth({ children }: { children: React.ReactNode }) {
  const { session } = useAuth()
  if (!session) {
    return <Navigate to="/login" replace />
  }

  return <>{children}</>
}

// UI guard only; the backend enforces the admin role on every request
function RequireAdmin({ children }: { children: React.ReactNode }) {
  const { isAdmin } = useAuth()
  if (!isAdmin) {
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
            <AppLayout />
          </RequireAuth>
        )}
      >
        <Route path="/" element={<HomePage />} />
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
