import { NavLink, Outlet } from 'react-router-dom'
import { useAuthStore } from './authStore'

const navigation = [
  ['Dashboard', '/'],
  ['Environmental', '/environmental'],
  ['Social', '/social'],
  ['Governance', '/governance'],
  ['Gamification', '/gamification'],
  ['Reports', '/reports'],
  ['Settings', '/settings'],
]
export function AppShell() {
  const user = useAuthStore((s) => s.user)
  const logout = useAuthStore((s) => s.logout)
  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="brand">EcoSphere</div>
        <nav aria-label="Primary navigation">
          {navigation.map(([label, path]) => (
            <NavLink key={path} to={path} end={path === '/'} className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}>
              {label}
            </NavLink>
          ))}
        </nav>
      </aside>
      <div className="main">
        <header className="topbar">
          <span className="crumb">EcoSphere ESG</span>
          <span className="spacer" />
          <button className="icon-button" aria-label="Notifications" type="button">
            ●
          </button>
          <span className="avatar" title={user?.name}>
            {user?.name
              .split(' ')
              .map((v) => v[0])
              .slice(0, 2)
              .join('')
              .toUpperCase()}
          </span>
          <button className="button ghost" onClick={logout} type="button">
            Sign out
          </button>
        </header>
        <Outlet />
      </div>
    </div>
  )
}
