import type {
  ApiError, AppNotification, Audit, AuthResponse, CSRActivity, CSRParticipation, Challenge,
  ChallengeParticipation, ChallengeStatus, CarbonSuggestion, CarbonSummary, CarbonTransaction,
  CarbonTransactionInput, ComplianceIssue, Department, DepartmentInput, DepartmentScore,
  DiversityMetrics, EnvironmentalGoal, EnvironmentalGoalInput, ESGConfig, EvidenceReview,
  GameBadge, GovernancePolicy, IssueStatus, LeaderboardEntry, NotificationPreference,
  OverallScore, PageResult, Policy, PolicyAck, Report, ReportType, Reward, Training, User,
} from './types'
import { sanitizeErrorMessage } from './userFacingError'

const API_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

export class RequestError extends Error {
  public body: ApiError
  constructor(public status: number, body: ApiError) {
    const fallback = status >= 500 ? 'Something went wrong. Please try again.' : 'Unable to complete this request'
    const safeMessage = sanitizeErrorMessage(body?.message, fallback)
    const safeBody: ApiError = { ...body, message: safeMessage, code: body?.code || 'request_failed' }
    super(safeMessage)
    this.body = safeBody
  }
}

export async function request<T>(path: string, init: RequestInit = {}) {
  const token = localStorage.getItem('ecosphere.accessToken')
  let response: Response
  try {
    const headers = new Headers(init.headers)
    if (!(init.body instanceof FormData)) headers.set('Content-Type', 'application/json')
    if (token) headers.set('Authorization', `Bearer ${token}`)
    response = await fetch(`${API_URL}${path}`, {
      ...init,
      headers,
    })
  } catch {
    throw new RequestError(0, { code: 'network_error', message: 'Unable to reach the server. Please try again.' })
  }
  if (!response.ok) {
    const body = (await response.json().catch(() => ({
      code: 'request_failed',
      message: response.status >= 500 ? 'Something went wrong. Please try again.' : 'Unable to complete this request',
    }))) as ApiError
    throw new RequestError(response.status, body)
  }
  if (response.status === 204) return undefined as T
  return response.json() as Promise<T>
}

export const api = {
  auth:{ login:(email:string,password:string)=>request<AuthResponse>('/auth/login',{method:'POST',body:JSON.stringify({email,password})}), me:()=>request<User>('/me') },
  departments:{
    list:(limit=20,offset=0)=>request<PageResult<Department>>(`/departments?limit=${limit}&offset=${offset}`),
    create:(input:DepartmentInput)=>request<Department>('/departments',{method:'POST',body:JSON.stringify(input)}),
    update:(id:string,input:DepartmentInput)=>request<Department>(`/departments/${id}`,{method:'PUT',body:JSON.stringify(input)}),
    deactivate:(id:string)=>request<Department>(`/departments/${id}`,{method:'DELETE'}),
  },
  master:{
    list:<T>(entity:string,params='')=>request<PageResult<T>>(`/${entity}?limit=100&offset=0${params}`),
    create:<T>(entity:string,input:unknown)=>request<T>(`/${entity}`,{method:'POST',body:JSON.stringify(input)}),
    update:<T>(entity:string,id:string,input:unknown)=>request<T>(`/${entity}/${id}`,{method:'PUT',body:JSON.stringify(input)}),
    remove:(entity:string,id:string)=>request<void>(`/${entity}/${id}`,{method:'DELETE'}),
  },
  settings:{
    config:()=>request<ESGConfig>('/settings/esg-config'),
    saveConfig:(input:ESGConfig)=>request<ESGConfig>('/settings/esg-config',{method:'PUT',body:JSON.stringify(input)}),
    preferences:()=>request<NotificationPreference[]>('/settings/notification-preferences'),
    savePreferences:(input:NotificationPreference[])=>request<NotificationPreference[]>('/settings/notification-preferences',{method:'PUT',body:JSON.stringify(input)}),
  },
  carbon:{
    ingest:(file:File)=>{const body=new FormData();body.append('file',file);return request<CarbonSuggestion>('/carbon/ingest',{method:'POST',body})},
    create:(input:CarbonTransactionInput)=>request<CarbonTransaction>('/carbon/transactions',{method:'POST',body:JSON.stringify(input)}),
    list:(params='')=>request<PageResult<CarbonTransaction>>(`/carbon/transactions?limit=100&offset=0${params}`),
    verify:(id:string)=>request<CarbonTransaction>(`/carbon/transactions/${id}/verify`,{method:'POST'}),
    summary:(params='')=>request<CarbonSummary>(`/carbon/summary?${params}`),
  },
  goals:{
    list:(params='')=>request<PageResult<EnvironmentalGoal>>(`/goals?limit=100&offset=0${params}`),
    create:(input:EnvironmentalGoalInput)=>request<EnvironmentalGoal>('/goals',{method:'POST',body:JSON.stringify(input)}),
    update:(id:string,input:EnvironmentalGoalInput)=>request<EnvironmentalGoal>(`/goals/${id}`,{method:'PUT',body:JSON.stringify(input)}),
  },
  social:{
    activities:()=>request<PageResult<CSRActivity>>('/csr/activities?limit=100&offset=0'),
    createActivity:(input:Partial<CSRActivity>)=>request<CSRActivity>('/csr/activities',{method:'POST',body:JSON.stringify(input)}),
    joinActivity:(input:{activityId:string;proofUrl?:string;notes?:string})=>request<CSRParticipation>('/csr/participations',{method:'POST',body:JSON.stringify(input)}),
    participations:(approval='')=>request<PageResult<CSRParticipation>>(`/csr/participations?limit=100&offset=0${approval?`&approval=${approval}`:''}`),
    approveParticipation:(id:string)=>request<CSRParticipation>(`/csr/participations/${id}/approve`,{method:'POST'}),
    rejectParticipation:(id:string)=>request<CSRParticipation>(`/csr/participations/${id}/reject`,{method:'POST'}),
    diversity:()=>request<DiversityMetrics>('/diversity'),
    trainings:()=>request<PageResult<Training>>('/trainings'),
    createTraining:(input:{name:string;assignedTo?:string})=>request<Training>('/trainings',{method:'POST',body:JSON.stringify(input)}),
    completeTraining:(id:string)=>request<{status:string}>(`/trainings/${id}/complete`,{method:'POST'}),
  },
  game:{
    challenges:()=>request<PageResult<Challenge>>('/challenges?limit=100&offset=0'),
    createChallenge:(input:Partial<Challenge>)=>request<Challenge>('/challenges',{method:'POST',body:JSON.stringify(input)}),
    transition:(id:string,to:ChallengeStatus)=>request<Challenge>(`/challenges/${id}/transition`,{method:'PUT',body:JSON.stringify({to})}),
    participate:(id:string,input:{progress?:number;proofUrl?:string})=>request<ChallengeParticipation>(`/challenges/${id}/participate`,{method:'POST',body:JSON.stringify(input)}),
    statusCounts:()=>request<Record<string,number>>('/challenges/status-counts'),
    participations:()=>request<PageResult<ChallengeParticipation>>('/challenge-participations?limit=100&offset=0'),
    approveParticipation:(id:string)=>request<ChallengeParticipation>(`/challenge-participations/${id}/approve`,{method:'POST'}),
    rejectParticipation:(id:string)=>request<ChallengeParticipation>(`/challenge-participations/${id}/reject`,{method:'POST'}),
    leaderboard:(scope:'employee'|'department'='employee')=>request<{items:LeaderboardEntry[];scope:string}>(`/leaderboard?scope=${scope}&limit=20`),
    rewards:()=>request<PageResult<Reward>>('/game-rewards'),
    redeem:(id:string)=>request<{reward:Reward;pointsSpent:number}>(`/game-rewards/${id}/redeem`,{method:'POST'}),
    badges:()=>request<PageResult<GameBadge>>('/game-badges'),
    balance:()=>request<{id:string;xp:number;points:number;completedChallenges:number;name:string}>('/me/balance'),
  },
  governance:{
    stats:()=>request<{openIssues:number;overdueIssues:number;auditsFY:number;governanceScore:number}>('/governance/stats'),
    policies:()=>request<PageResult<GovernancePolicy>>('/governance/policies'),
    acknowledgements:()=>request<PageResult<PolicyAck>>('/governance/acknowledgements?limit=100&offset=0'),
    unacknowledged:()=>request<PageResult<Policy>>('/governance/unacknowledged'),
    acknowledge:(id:string)=>request<PolicyAck>(`/governance/policies/${id}/acknowledge`,{method:'POST'}),
    audits:()=>request<PageResult<Audit>>('/audits?limit=100&offset=0'),
    createAudit:(input:Partial<Audit>)=>request<Audit>('/audits',{method:'POST',body:JSON.stringify(input)}),
    issues:(params='')=>request<PageResult<ComplianceIssue>>(`/compliance-issues?limit=100&offset=0${params}`),
    raiseIssue:(input:{departmentId:string;ownerId:string;severity:string;description:string;dueDate:string;auditId?:string})=>
      request<ComplianceIssue>('/compliance-issues',{method:'POST',body:JSON.stringify(input)}),
    updateIssue:(id:string,status:IssueStatus)=>request<ComplianceIssue>(`/compliance-issues/${id}`,{method:'PUT',body:JSON.stringify({status})}),
    bundle:(departmentId:string)=>request<Record<string,unknown[]>>(`/audits/department/${departmentId}/bundle`),
  },
  notifications:{
    list:()=>request<{items:AppNotification[];total:number;unread:number}>('/notifications?limit=30&offset=0'),
    markRead:(id:string)=>request<{status:string}>(`/notifications/${id}/read`,{method:'POST'}),
  },
  scores:{
    overall:(period='')=>request<OverallScore>(`/scores/overall${period?`?period=${period}`:''}`),
    departments:(period='')=>request<PageResult<DepartmentScore>>(`/scores/departments${period?`?period=${period}`:''}`),
    recompute:()=>request<PageResult<DepartmentScore>>('/scores/recompute',{method:'POST'}),
  },
  reports:{
    generate:(input:{type:ReportType;filters?:Record<string,unknown>;departmentId?:string;from?:string;to?:string})=>
      request<Report>('/reports/generate',{method:'POST',body:JSON.stringify(input)}),
    get:(id:string)=>request<Report>(`/reports/${id}`),
    exportUrl:(id:string,fmt:'pdf'|'xlsx'|'csv')=>`${API_URL}/reports/${id}/export?fmt=${fmt}`,
  },
  ai:{
    /** Advisory only. Prefer imageDataUrl for uploads (not stored server-side beyond the request). */
    evidenceReview:(input:string|{proofUrl?:string;imageUrl?:string;imageDataUrl?:string;fileName?:string})=>{
      const body=typeof input==='string'
        ?{proofUrl:input,imageUrl:input}
        :{proofUrl:input.proofUrl,imageUrl:input.imageUrl??input.proofUrl,imageDataUrl:input.imageDataUrl,fileName:input.fileName}
      return request<EvidenceReview>('/ai/evidence-review',{method:'POST',body:JSON.stringify(body)})
    },
  },
}

/** Download a report export with auth header. */
export async function downloadReport(id: string, fmt: 'pdf' | 'xlsx' | 'csv') {
  const token = localStorage.getItem('ecosphere.accessToken')
  const res = await fetch(api.reports.exportUrl(id, fmt), {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  })
  if (!res.ok) throw new RequestError(res.status, { code: 'export_failed', message: 'Unable to export report' })
  const blob = await res.blob()
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `ecosphere-report.${fmt}`
  a.click()
  URL.revokeObjectURL(url)
}
