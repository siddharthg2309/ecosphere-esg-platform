import { describe, expect, it } from 'vitest'

/** Mirrors filter composition used by ReportsPage generate payload. */
export function buildReportRequest(input: {
  type: string
  departmentId?: string
  from?: string
  to?: string
  module?: string
  employee?: string
  challenge?: string
  category?: string
}) {
  const filters: Record<string, unknown> = {}
  if (input.module) filters.module = input.module
  if (input.employee) filters.employee = input.employee
  if (input.challenge) filters.challenge = input.challenge
  if (input.category) filters.category = input.category
  return {
    type: input.type,
    departmentId: input.departmentId || undefined,
    from: input.from || undefined,
    to: input.to || undefined,
    filters,
  }
}

describe('buildReportRequest', () => {
  it('composes filters into the generate request', () => {
    expect(
      buildReportRequest({
        type: 'esg_summary',
        departmentId: 'd1',
        from: '2026-01-01',
        to: '2026-12-31',
        module: 'social',
      }),
    ).toEqual({
      type: 'esg_summary',
      departmentId: 'd1',
      from: '2026-01-01',
      to: '2026-12-31',
      filters: { module: 'social' },
    })
  })
  it('omits empty filters', () => {
    expect(buildReportRequest({ type: 'social' })).toEqual({
      type: 'social',
      departmentId: undefined,
      from: undefined,
      to: undefined,
      filters: {},
    })
  })
})
