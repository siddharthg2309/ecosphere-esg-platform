import { describe, expect, it } from 'vitest'
import { allowedTransitions, canApproveParticipation } from './challengeTransitions'

describe('allowedTransitions', () => {
  it('blocks draft → completed', () => {
    expect(allowedTransitions('draft')).not.toContain('completed')
  })
  it('allows active → under_review', () => {
    expect(allowedTransitions('active')).toEqual(['under_review', 'archived'])
  })
})

describe('canApproveParticipation', () => {
  it('disables approve when proof missing and evidence required', () => {
    expect(canApproveParticipation('', true)).toBe(false)
    expect(canApproveParticipation(undefined, true)).toBe(false)
  })
  it('allows approve with proof or when evidence not required', () => {
    expect(canApproveParticipation('photo.jpg', true)).toBe(true)
    expect(canApproveParticipation('', false)).toBe(true)
  })
})
