import { describe, expect, it } from 'vitest'
import { applyOptimisticRedeem, canRedeem, rollbackPoints } from './redeemVM'

describe('redeemVM', () => {
  it('disables when out of stock', () => {
    expect(canRedeem(0, 200, 1000)).toBe(false)
  })
  it('disables when insufficient points', () => {
    expect(canRedeem(5, 800, 100)).toBe(false)
  })
  it('optimistic decrement and rollback on error', () => {
    const before = 1000
    const next = applyOptimisticRedeem(before, 800)
    expect(next).toBe(200)
    expect(rollbackPoints(before)).toBe(1000)
  })
})
