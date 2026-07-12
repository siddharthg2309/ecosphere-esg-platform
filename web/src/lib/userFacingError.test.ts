import { describe, expect, it } from 'vitest'
import { RequestError } from './apiClient'
import { looksTechnical, sanitizeErrorMessage, userFacingError } from './userFacingError'

describe('looksTechnical', () => {
  it('flags postgres dumps', () => {
    expect(looksTechnical('ERROR: duplicate key value violates unique constraint "users_email_key" (SQLSTATE 23505)')).toBe(true)
    expect(looksTechnical('pq: relation "csr_activities" does not exist')).toBe(true)
  })
  it('allows human messages', () => {
    expect(looksTechnical('proof file required')).toBe(false)
    expect(looksTechnical('Already joined this activity')).toBe(false)
  })
})

describe('userFacingError', () => {
  it('softens technical API messages', () => {
    const err = new RequestError(500, {
      code: 'internal',
      message: 'ERROR: column "foo" does not exist',
    })
    expect(userFacingError(err)).toBe('Something went wrong. Please try again.')
  })
  it('keeps friendly API messages', () => {
    const err = new RequestError(409, {
      code: 'duplicate_participation',
      message: 'Already joined this activity',
    })
    expect(userFacingError(err)).toBe('Already joined this activity')
  })
  it('softens network failures', () => {
    expect(userFacingError(new TypeError('Failed to fetch'))).toBe('Something went wrong. Please try again.')
  })
})

describe('sanitizeErrorMessage', () => {
  it('returns fallback for empty', () => {
    expect(sanitizeErrorMessage('')).toBe('Something went wrong. Please try again.')
  })
})
