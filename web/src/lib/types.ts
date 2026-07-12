export type Role = 'employee' | 'dept_head' | 'auditor' | 'admin'
export type Status = 'active' | 'inactive'

export interface User { id:string; name:string; email:string; role:Role; departmentId?:string }
export interface AuthResponse { accessToken:string; refreshToken:string; user:User }
export interface PageResult<T> { items:T[]; total:number }
export interface ApiError { code:string; message:string; fields?:Record<string,string> }
export interface Department { id:string; name:string; code:string; headId?:string; parentId?:string; employeeCount:number; status:Status; createdAt:string; updatedAt:string }
export type DepartmentInput = Pick<Department,'name'|'code'|'employeeCount'|'status'> & {headId?:string;parentId?:string}
export type CategoryType='csr_activity'|'challenge'
export interface Category{id:string;name:string;type:CategoryType;status:Status;createdAt:string;updatedAt:string}
export interface EmissionFactor{id:string;name:string;categoryId:string;unit:string;kgCo2PerUnit:string;status:Status}
export interface ProductProfile{id:string;product:string;attributes:Record<string,unknown>;emissionFactorId?:string}
export interface Policy{id:string;title:string;body:string;version:number;effectiveDate:string}
export interface Badge{id:string;name:string;description:string;icon:string;unlockRule:{type:'xp'|'challenges';value:number}}
export interface Reward{id:string;name:string;description:string;pointsRequired:number;stock:number;status:Status}
export interface Employee extends User{xp:number;points:number;completedChallenges:number;status:Status;createdAt:string}
export interface ESGConfig{autoEmissionCalc:boolean;requireCsrEvidence:boolean;autoAwardBadges:boolean;notifyComplianceEmail:boolean;weightEnv:number;weightSocial:number;weightGov:number}
export type NotificationEvent='compliance_raised'|'approval_decision'|'policy_reminder'|'badge_unlock'|'compliance_overdue'
export interface NotificationPreference{eventType:NotificationEvent;inAppEnabled:boolean;emailEnabled:boolean}

export type ApprovalStatus = 'pending' | 'approved' | 'rejected' | 'in_progress'
export type ChallengeStatus = 'draft' | 'active' | 'under_review' | 'completed' | 'archived'

export interface CSRActivity {
  id: string
  title: string
  categoryId: string
  description: string
  points: number
  evidenceRequired: boolean
  status: Status
  activityDate?: string
  joinedCount?: number
  createdAt: string
}

export interface CSRParticipation {
  id: string
  employeeId: string
  activityId: string
  proofUrl: string
  notes?: string
  approval: ApprovalStatus
  pointsEarned: number
  completionDate?: string
  employeeName?: string
  activityTitle?: string
  activityPoints?: number
  evidenceRequired?: boolean
}

export interface DiversityMetrics {
  genderWomenPct: number
  genderMenPct: number
  genderNonBinaryPct: number
  leadershipWomenPct: number
  diverseLeadersPct: number
  leadershipTargetPct: number
  trainingCompletionPct: number
  csrParticipationPct: number
}

export interface Training {
  id: string
  name: string
  assignedTo: string
  status: string
  completed: number
  total: number
}

export interface Challenge {
  id: string
  title: string
  categoryId: string
  description: string
  xp: number
  difficulty: string
  evidenceRequired: boolean
  deadline?: string
  status: ChallengeStatus
  pendingCount?: number
}

export interface ChallengeParticipation {
  id: string
  challengeId: string
  employeeId: string
  progress: number
  proofUrl: string
  approval: ApprovalStatus
  xpAwarded: number
  employeeName?: string
  challengeTitle?: string
  challengeXp?: number
  evidenceRequired?: boolean
}

export interface LeaderboardEntry {
  rank: number
  id: string
  name: string
  xp: number
  badgeCount: number
}

export interface GameBadge extends Badge {
  earnedCount?: number
}
