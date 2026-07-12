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

export type IssueSeverity = 'low' | 'medium' | 'high'
export type IssueStatus = 'open' | 'in_progress' | 'resolved'
export type AuditStatus = 'draft' | 'under_review' | 'completed'

export interface GovernancePolicy extends Policy {
  acked?: number
  total?: number
  ackRate?: number
}

export interface Audit {
  id: string
  title: string
  departmentId: string
  auditorId: string
  auditDate: string
  findings: string
  status: AuditStatus
  departmentName?: string
  auditorName?: string
}

export interface ComplianceIssue {
  id: string
  auditId?: string
  departmentId: string
  severity: IssueSeverity
  description: string
  ownerId: string
  dueDate: string
  status: IssueStatus
  overdue?: boolean
  ownerName?: string
  departmentName?: string
  auditTitle?: string
}

export interface PolicyAck {
  id: string
  employeeId: string
  policyId: string
  version: number
  acknowledgedAt: string
  employeeName?: string
  departmentName?: string
  policyTitle?: string
}

export interface AppNotification {
  id: string
  userId: string
  type: NotificationEvent
  title: string
  payload: Record<string, unknown>
  readAt?: string
  createdAt: string
}

export interface DepartmentScore {
  departmentId: string
  name?: string
  environmental: number
  social: number
  governance: number
  total: number
  period: string
}

export interface OverallScore {
  overall: number
  environmental: number
  social: number
  governance: number
  weights: { weightEnv: number; weightSocial: number; weightGov: number }
  departments: DepartmentScore[]
}

export type ReportType = 'environmental' | 'social' | 'governance' | 'esg_summary' | 'custom'

export interface ReportSection {
  title: string
  summary?: string
  rows?: Record<string, string>[]
  metrics?: Record<string, unknown>
  ai?: boolean
}

export interface Report {
  id: string
  type: ReportType
  filters: Record<string, unknown>
  sections: ReportSection[]
  generatedAt: string
}

export interface EvidenceReview {
  looksValid: boolean
  confidence: number
  notes: string
}

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

export type CarbonSource='purchase'|'manufacturing'|'expense'|'fleet'
export type CarbonStatus='draft'|'verified'
export interface CarbonSuggestion{source:string;categoryId?:string;quantity:number;unit:string;confidence:number;evidenceUrl:string}
export interface CarbonTransaction{id:string;departmentId:string;source:CarbonSource;quantity:string;emissionFactorId:string;factorValue:string;computedCo2:string;txnDate:string;evidenceUrl?:string;status:CarbonStatus;verifiedBy?:string;verifiedAt?:string;createdAt:string}
export interface CarbonTransactionInput{departmentId:string;source:CarbonSource;quantity:string;emissionFactorId:string;unit:string;txnDate:string;evidenceUrl?:string}
export interface CarbonSummary{total:string;bySource:Partial<Record<CarbonSource,string>>}
export type GoalStatus='on_track'|'at_risk'|'completed'
export interface EnvironmentalGoal{id:string;name:string;departmentId:string;targetCo2:string;currentCo2:string;deadline:string;status:GoalStatus;createdAt:string;updatedAt:string}
export interface EnvironmentalGoalInput{name:string;departmentId:string;targetCo2:string;currentCo2?:string;deadline:string}
