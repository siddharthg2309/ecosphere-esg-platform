const FALLBACK = 'Something went wrong. Please try again.'

/** Detect SQL/driver dumps that must never be shown in the UI. */
export function looksTechnical(message: string): boolean {
  if (!message || !message.trim()) return true
  const lower = message.toLowerCase()
  const needles = [
    'sqlstate',
    'pq:',
    'pgx',
    'error:',
    'fatal:',
    'detail:',
    'hint:',
    'duplicate key',
    'violates unique constraint',
    'violates foreign key',
    'violates check constraint',
    'violates not-null',
    'relation "',
    'column "',
    'syntax error',
    'connection refused',
    'failed to connect',
    'dial tcp',
    'driver:',
    'could not connect',
    'server closed the connection',
    'ssl is not enabled',
    'password authentication failed',
    'context deadline exceeded',
    'i/o timeout',
    'broken pipe',
    'prepared statement',
    'econnrefused',
    'enotfound',
    'networkerror',
    'failed to fetch',
  ]
  if (needles.some((n) => lower.includes(n))) return true
  if (message.includes('\n') || message.includes('\r')) return true
  if (message.length > 160 && (message.match(/_/g) || []).length > 6) return true
  return false
}

export function sanitizeErrorMessage(message: string | undefined, fallback = FALLBACK): string {
  if (!message || looksTechnical(message)) return fallback
  return message
}

/** Prefer this for all user-visible error copy. Avoids importing RequestError (no circular deps). */
export function userFacingError(err: unknown, fallback = FALLBACK): string {
  if (err && typeof err === 'object') {
    const body = (err as { body?: { message?: string } }).body
    if (body && typeof body.message === 'string') {
      return sanitizeErrorMessage(body.message, fallback)
    }
  }
  if (err instanceof Error) {
    return sanitizeErrorMessage(err.message, fallback)
  }
  if (typeof err === 'string') {
    return sanitizeErrorMessage(err, fallback)
  }
  return fallback
}
