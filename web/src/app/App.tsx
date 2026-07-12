import { useEffect } from 'react'
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { DashboardPage } from '../modules/dashboard/DashboardPage'
import { EnvironmentalPage } from '../modules/environmental/EnvironmentalPage'
import { GamificationPage } from '../modules/gamification/GamificationPage'
import { GovernancePage } from '../modules/governance/GovernancePage'
import { ReportsPage } from '../modules/reports/ReportsPage'
import { DepartmentsPage } from '../modules/settings/DepartmentsPage'
import { SocialPage } from '../modules/social/SocialPage'
import { AppShell } from './AppShell'
import { LoginPage } from './LoginPage'
import { Providers } from './Providers'
import { useAuthStore } from './authStore'
import { homePathForRole } from './rbac'
import type { Role } from '../lib/types'

function RoleRoute({ roles, children }: { roles: Role[]; children: React.ReactNode }) {
  const role = useAuthStore((s) => s.user?.role)
  if (!role || !roles.includes(role)) {
    return <Navigate to={role ? homePathForRole(role) : '/login'} replace />
  }
  return children
}

function ProtectedApp() {
  const user = useAuthStore((s) => s.user)
  const initialized = useAuthStore((s) => s.initialized)
  const restore = useAuthStore((s) => s.restore)
  useEffect(() => {
    void restore()
  }, [restore])
  if (!initialized) return <div className="center-state">Loading EcoSphere…</div>
  if (!user) return <Navigate to="/login" replace />
  return (
    <Routes>
      <Route element={<AppShell />}>
        <Route index element={<DashboardPage />} />
        <Route
          path="environmental"
          element={
            <RoleRoute roles={['admin', 'dept_head', 'auditor', 'employee']}>
              <EnvironmentalPage />
            </RoleRoute>
          }
        />
        <Route
          path="social"
          element={
            <RoleRoute roles={['admin', 'dept_head', 'employee']}>
              <SocialPage />
            </RoleRoute>
          }
        />
        <Route
          path="governance"
          element={
            <RoleRoute roles={['admin', 'dept_head', 'auditor', 'employee']}>
              <GovernancePage />
            </RoleRoute>
          }
        />
        <Route
          path="gamification"
          element={
            <RoleRoute roles={['admin', 'dept_head', 'employee']}>
              <GamificationPage />
            </RoleRoute>
          }
        />
        <Route
          path="reports"
          element={
            <RoleRoute roles={['admin', 'dept_head', 'auditor']}>
              <ReportsPage />
            </RoleRoute>
          }
        />
        <Route
          path="settings"
          element={
            <RoleRoute roles={['admin']}>
              <DepartmentsPage />
            </RoleRoute>
          }
        />
        <Route
          path="*"
          element={
            <div className="page">
              <div className="content">
                <h1 className="page-title">Page not found</h1>
                <p className="muted">This route is not available for your portal.</p>
              </div>
            </div>
          }
        />
      </Route>
    </Routes>
  )
}

export function App() {
  return (
    <Providers>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/*" element={<ProtectedApp />} />
        </Routes>
      </BrowserRouter>
    </Providers>
  )
}
