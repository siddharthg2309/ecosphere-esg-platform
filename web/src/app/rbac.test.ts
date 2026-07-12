import { describe, expect, it } from 'vitest'
import { canAccessPath, homePathForRole, navByRole } from './rbac'

describe('rbac portals', () => {
  it('sends every role home to / not settings', () => {
    for (const role of ['admin', 'dept_head', 'auditor', 'employee'] as const) {
      expect(homePathForRole(role)).toBe('/')
    }
  })
  it('blocks settings for non-admins', () => {
    expect(canAccessPath('admin', '/settings')).toBe(true)
    expect(canAccessPath('employee', '/settings')).toBe(false)
    expect(canAccessPath('auditor', '/settings')).toBe(false)
    expect(canAccessPath('dept_head', '/settings')).toBe(false)
  })
  it('gives each role a distinct nav', () => {
    const labels = (role: keyof typeof navByRole) =>
      navByRole[role].flatMap((g) => g.items.map((i) => i.label)).join('|')
    expect(labels('admin')).toContain('Settings')
    expect(labels('employee')).toContain('My ESG Hub')
    expect(labels('employee')).not.toContain('Settings')
    expect(labels('auditor')).toContain('Audit Hub')
    expect(labels('dept_head')).toContain('Department Hub')
  })
})
