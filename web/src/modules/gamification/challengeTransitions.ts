/** Keep in sync with internal/modules/gamification/challenge/domain Transitions map. */
export type ChallengeStatus = 'draft' | 'active' | 'under_review' | 'completed' | 'archived'

export const CHALLENGE_TRANSITIONS: Record<ChallengeStatus, ChallengeStatus[]> = {
  draft: ['active', 'archived'],
  active: ['under_review', 'archived'],
  under_review: ['completed', 'active', 'archived'],
  completed: ['archived'],
  archived: [],
}

export function allowedTransitions(from: ChallengeStatus | string): ChallengeStatus[] {
  return CHALLENGE_TRANSITIONS[from as ChallengeStatus] ?? []
}

export function canApproveParticipation(proofUrl: string | undefined, evidenceRequired: boolean): boolean {
  if (!evidenceRequired) return true
  return Boolean(proofUrl && proofUrl.trim())
}
