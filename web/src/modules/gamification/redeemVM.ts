/** Pure helpers for redeem optimistic UI — testable without React. */

export function canRedeem(stock: number, pointsRequired: number, userPoints: number): boolean {
  return stock > 0 && userPoints >= pointsRequired
}

export function applyOptimisticRedeem(userPoints: number, pointsRequired: number): number {
  return userPoints - pointsRequired
}

export function rollbackPoints(previous: number): number {
  return previous
}
