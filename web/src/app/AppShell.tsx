import { NavLink, Navigate, Outlet, useLocation } from 'react-router-dom'
import { useAuthStore } from './authStore'
import { NotificationBell } from './NotificationBell'
import { PolicyAckModal } from './PolicyAckModal'
import { canAccessPath, crumbForPath, homePathForRole, navByRole, portalMeta } from './rbac'

export function AppShell() {
  const user = useAuthStore((s) => s.user)
  const logout = useAuthStore((s) => s.logout)
  const location = useLocation()
  const role = user?.role
  const groups = role ? navByRole[role] : []
  const meta = portalMeta(role)

  if (role && !canAccessPath(role, location.pathname)) {
    return <Navigate to={homePathForRole(role)} replace />
  }

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="brand">
          <span>EcoSphere</span>
          <span className="portal-tag">{meta.name.replace(' Portal', '')}</span>
        </div>
        <nav className="sidebar-nav" aria-label="Primary navigation">
          {groups.map((group, gi) => (
            <div className="nav-group" key={group.title ?? `g-${gi}`}>
              {group.title ? <div className="nav-title">{group.title}</div> : null}
              {group.items.map((item) => (
                <NavLink
                  key={`${item.path}-${item.label}`}
                  to={item.path}
                  end={item.end === true}
                  className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}
                >
                  {item.label}
                </NavLink>
              ))}
            </div>
          ))}
        </nav>
        <div className="sidebar-foot">
          <div className="sidebar-user">
            <span className="avatar-sm" aria-hidden>
              {initials(user?.name ?? '?')}
            </span>
            <div>
              <div className="sidebar-user-name">{user?.name}</div>
              <div className="sidebar-user-role">{roleLabel(role)}</div>
            </div>
          </div>
        </div>
      </aside>
      <div className="main">
        <header className="topbar">
          <span className="crumb">{crumbForPath(role, location.pathname)}</span>
          <span className="spacer" />
          <NotificationBell />
          <span className="avatar" title={user?.name}>
            {initials(user?.name ?? '?')}
          </span>
          <button className="button ghost sm" onClick={logout} type="button">
            Sign out
          </button>
        </header>
        <div className="main-scroll">
          <Outlet />
        </div>
        <PolicyAckModal />
      </div>
    </div>
  )
}

function initials(name: string) {
  return name
    .split(' ')
    .map((v) => v[0])
    .slice(0, 2)
    .join('')
    .toUpperCase()
}

function roleLabel(role: string | undefined) {
  switch (role) {
    case 'admin':
      return 'Administrator'
    case 'dept_head':
      return 'Department head'
    case 'auditor':
      return 'Auditor'
    case 'employee':
      return 'Employee'
    default:
      return role ?? ''
  }
}
