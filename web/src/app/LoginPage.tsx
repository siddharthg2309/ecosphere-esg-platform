import { useState, type FormEvent } from 'react'
import { Navigate } from 'react-router-dom'
import { userFacingError } from '../lib/userFacingError'
import { useAuthStore } from './authStore'

export function LoginPage() {
  const user = useAuthStore((s) => s.user)
  const login = useAuthStore((s) => s.login)
  const [email, setEmail] = useState('admin@ecosphere.local')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  if (user) return <Navigate to="/settings" replace />
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
  return (
    <main className="login-page">
      <section className="login-card">
        <p className="eyebrow">ESG operating system</p>
        <h1>Welcome to EcoSphere</h1>
        <p className="muted">Sign in to manage sustainability data and settings.</p>
        <form onSubmit={submit}>
          <label>
            Email
            <input type="email" required value={email} onChange={(e) => setEmail(e.target.value)} autoComplete="email" />
          </label>
          <label>
            Password
            <input type="password" required value={password} onChange={(e) => setPassword(e.target.value)} autoComplete="current-password" />
          </label>
          {error && (
            <div className="form-error" role="alert">
              {error}
            </div>
          )}
          <button className="button primary" disabled={loading}>
            {loading ? 'Signing in…' : 'Sign in'}
          </button>
        </form>
      </section>
    </main>
  )
}
