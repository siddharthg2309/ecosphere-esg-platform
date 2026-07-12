import { useState, type FormEvent } from 'react'
import { Navigate } from 'react-router-dom'
import { userFacingError } from '../lib/userFacingError'
import type { Role } from '../lib/types'
import { useAuthStore } from './authStore'
import { DEMO_PASSWORD, homePathForRole, PORTALS } from './rbac'

export function LoginPage() {
  const user = useAuthStore((s) => s.user)
  const login = useAuthStore((s) => s.login)
  const [email, setEmail] = useState('admin@ecosphere.local')
  const [password, setPassword] = useState(DEMO_PASSWORD)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [selectedPortal, setSelectedPortal] = useState<Role>('admin')

  if (user) return <Navigate to={homePathForRole(user.role)} replace />

  async function submit(event: FormEvent) {
    event.preventDefault()
    setLoading(true)
    setError('')
    try {
      await login(email, password)
    } catch (err) {
      setError(userFacingError(err, 'Unable to sign in'))
    } finally {
      setLoading(false)
    }
  }

  function pickPortal(role: Role) {
    const portal = PORTALS.find((p) => p.id === role)
    if (!portal) return
    setSelectedPortal(role)
    setEmail(portal.demoEmail)
    setPassword(DEMO_PASSWORD)
    setError('')
  }

  return (
    <main className="login-page">
      <div className="login-layout">
        <section className="login-portals" aria-label="Choose a portal">
          <p className="eyebrow">Role-based access</p>
          <h1>EcoSphere portals</h1>
          <p className="muted">
            Four portals with distinct navigation and features. Pick a demo role or sign in with any seeded account.
          </p>
          <div className="portal-grid">
            {PORTALS.map((p) => (
              <button
                key={p.id}
                type="button"
                className={`portal-card ${selectedPortal === p.id ? 'active' : ''}`}
                onClick={() => pickPortal(p.id)}
              >
                <strong>{p.name}</strong>
                <span className="muted">{p.subtitle}</span>
                <code>{p.demoEmail}</code>
              </button>
            ))}
          </div>
        </section>

        <section className="login-card">
          <p className="eyebrow">ESG operating system</p>
          <h1>Sign in</h1>
          <p className="muted">
            {PORTALS.find((p) => p.id === selectedPortal)?.name ?? 'Portal'} · password{' '}
            <code>{DEMO_PASSWORD}</code>
          </p>
          <form onSubmit={submit}>
            <label>
              Email
              <input
                type="email"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                autoComplete="email"
              />
            </label>
            <label>
              Password
              <input
                type="password"
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                autoComplete="current-password"
              />
            </label>
            {error && (
              <div className="form-error" role="alert">
                {error}
              </div>
            )}
            <button className="button primary" type="submit" disabled={loading}>
              {loading ? 'Signing in…' : 'Sign in'}
            </button>
          </form>
        </section>
      </div>
    </main>
  )
}
