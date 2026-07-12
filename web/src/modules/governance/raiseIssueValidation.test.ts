import { describe, expect, it } from 'vitest'
import { canSubmitIssue } from './raiseIssueValidation'

describe('canSubmitIssue', () => {
  it('disables submit until owner and due date are set', () => {
    expect(
      canSubmitIssue({
        description: 'Missing MSDS',
        departmentId: 'd1',
        severity: 'high',
      }),
    ).toBe(false)
    expect(
      canSubmitIssue({
        description: 'Missing MSDS',
        departmentId: 'd1',
        ownerId: 'u1',
        severity: 'high',
      }),
    ).toBe(false)
    expect(
      canSubmitIssue({
        description: 'Missing MSDS',
        departmentId: 'd1',
        ownerId: 'u1',
        dueDate: '2026-08-01',
        severity: 'high',
      }),
    ).toBe(true)
  })
})
