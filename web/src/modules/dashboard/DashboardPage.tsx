import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { RoleGuard } from '../../app/RoleGuard'
import { useAuthStore } from '../../app/authStore'
import { portalMeta } from '../../app/rbac'
import { Button, Card, Pill, Progress, StatBar } from '../../design/components'
import { api } from '../../lib/apiClient'
import { queryKeys } from '../../lib/queryKeys'
import { userFacingError } from '../../lib/userFacingError'

export function DashboardPage() {
  const role = useAuthStore((s) => s.user?.role)
  if (role === 'employee') return <EmployeeHub />
  if (role === 'dept_head') return <DepartmentHub />
  if (role === 'auditor') return <AuditorHub />
  return <AdminHub />
}

function PageFrame({
  eyebrow,
  title,
  sub,
  actions,
  children,
}: {
  eyebrow: string
  title: string
  sub: string
  actions?: ReactNode
  children: ReactNode
}) {
  return (
    <main className="page">
      <div className="content">
        <header className="page-head">
          <div>
            <p className="eyebrow">{eyebrow}</p>
            <h1 className="page-title">{title}</h1>
            <p className="muted sub-line">{sub}</p>
          </div>
          {actions ? <div className="rowflex">{actions}</div> : null}
        </header>
        {children}
      </div>
    </main>
  )
}

function AdminHub() {
  const qc = useQueryClient()
  const scores = useQuery({ queryKey: queryKeys.scoresOverall, queryFn: () => api.scores.overall() })
  const carbon = useQuery({ queryKey: queryKeys.carbonSummary, queryFn: () => api.carbon.summary() })
  const notifs = useQuery({ queryKey: queryKeys.notifications, queryFn: api.notifications.list })
  const recompute = useMutation({
    mutationFn: api.scores.recompute,
    onSuccess: () => void qc.invalidateQueries({ queryKey: queryKeys.scoresOverall }),
  })
  const s = scores.data
  const weights = s?.weights

  return (
    <PageFrame
      eyebrow="Dashboard"
      title="Executive Overview"
      sub="Organization-wide ESG performance · live scores"
      actions={
        <>
          <RoleGuard roles={['admin']}>
            <Button
              className="secondary sm"
              disabled={recompute.isPending}
              onClick={() => void recompute.mutateAsync()}
            >
              {recompute.isPending ? 'Recomputing…' : 'Recompute scores'}
            </Button>
          </RoleGuard>
          <Link to="/environmental">
            <Button className="secondary sm">Log Carbon Data</Button>
          </Link>
          <Link to="/reports">
            <Button className="primary sm">New Report</Button>
          </Link>
        </>
      }
    >
      {recompute.isError && (
        <div className="form-error" role="alert">
          {userFacingError(recompute.error, 'Unable to recompute scores')}
        </div>
      )}
      <StatBar
        items={[
          { label: 'Environmental', value: s?.environmental ?? '—', unit: '/ 100', sub: 'pillar score' },
          { label: 'Social', value: s?.social ?? '—', unit: '/ 100', sub: 'pillar score' },
          { label: 'Governance', value: s?.governance ?? '—', unit: '/ 100', sub: 'pillar score' },
          {
            label: 'Overall ESG Score',
            value: s?.overall ?? '—',
            unit: '/ 100',
            sub: weights
              ? `${weights.weightEnv} / ${weights.weightSocial} / ${weights.weightGov} weighting`
              : 'weighted roll-up',
          },
        ]}
      />
      <div className="grid cols-2 section-block">
        <Card>
          <div className="card-head">
            <h3>Emissions summary</h3>
            <Pill status="plum">Verified t CO₂e</Pill>
          </div>
          <div className="stat flat">
            <div className="label">Total verified</div>
            <div className="num">{carbon.data?.total ?? '0'}</div>
          </div>
          <div className="stack">
            {Object.entries(carbon.data?.bySource ?? {}).map(([source, value]) => (
              <div className="list-row" key={source}>
                <span>{source}</span>
                <b>{String(value)}</b>
              </div>
            ))}
            {!carbon.data?.bySource || Object.keys(carbon.data.bySource).length === 0 ? (
              <p className="muted">No verified emissions yet — log and verify carbon transactions.</p>
            ) : null}
          </div>
        </Card>
        <Card>
          <div className="card-head">
            <h3>Department ESG ranking</h3>
            <span className="muted">by total score</span>
          </div>
          <div className="table-wrap flat">
            <table>
              <thead>
                <tr>
                  <th>#</th>
                  <th>Department</th>
                  <th className="numeric">Env</th>
                  <th className="numeric">Social</th>
                  <th className="numeric">Gov</th>
                  <th className="numeric">Total</th>
                </tr>
              </thead>
              <tbody>
                {(s?.departments ?? []).map((d, i) => (
                  <tr key={d.departmentId}>
                    <td>
                      <b>{i + 1}</b>
                    </td>
                    <td>{d.name || d.departmentId.slice(0, 8)}</td>
                    <td className="numeric">{d.environmental}</td>
                    <td className="numeric">{d.social}</td>
                    <td className="numeric">{d.governance}</td>
                    <td className="numeric">
                      <b>{d.total}</b>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            {(s?.departments ?? []).length === 0 && (
              <p className="muted" style={{ padding: 12 }}>
                Scores will appear after recompute (triggered by ESG events).
              </p>
            )}
          </div>
        </Card>
      </div>
      <Card className="section-block">
        <div className="card-head">
          <h3>Recent activity</h3>
          <span className="muted">from notifications</span>
        </div>
        {(notifs.data?.items ?? []).slice(0, 6).map((n) => (
          <div className="list-row" key={n.id}>
            <div>
              <b>{n.title}</b>
              <div className="muted" style={{ fontSize: 12 }}>
                {n.type.replace(/_/g, ' ')} · {String(n.createdAt).slice(0, 16).replace('T', ' ')}
              </div>
            </div>
            {!n.readAt && <Pill status="warning">Unread</Pill>}
          </div>
        ))}
        {(notifs.data?.items ?? []).length === 0 && <p className="muted">No recent notifications.</p>}
      </Card>
    </PageFrame>
  )
}

function EmployeeHub() {
  const user = useAuthStore((s) => s.user)
  const meta = portalMeta('employee')
  const balance = useQuery({ queryKey: queryKeys.balance, queryFn: api.game.balance })
  const challenges = useQuery({ queryKey: queryKeys.challenges, queryFn: api.game.challenges })
  const badges = useQuery({ queryKey: queryKeys.gameBadges, queryFn: api.game.badges })
  const leaderboard = useQuery({
    queryKey: queryKeys.leaderboard('employee'),
    queryFn: () => api.game.leaderboard('employee'),
  })
  const unacked = useQuery({ queryKey: queryKeys.unacknowledged, queryFn: api.governance.unacknowledged })
  const activities = useQuery({ queryKey: queryKeys.csrActivities, queryFn: api.social.activities })

  const xp = balance.data?.xp ?? 0
  const points = balance.data?.points ?? 0
  const activeChallenges = (challenges.data?.items ?? []).filter((c) => c.status === 'active')
  const rankIdx = (leaderboard.data?.items ?? []).findIndex((e) => e.id === user?.id || e.name === user?.name)
  const myRank = rankIdx >= 0 ? rankIdx + 1 : '—'
  const badgeCount = badges.data?.items?.length ?? 0
  const pendingPolicies = unacked.data?.items?.length ?? 0

  return (
    <PageFrame
      eyebrow={meta.homeLabel}
      title={`Welcome back, ${user?.name?.split(' ')[0] ?? 'there'}`}
      sub="Your challenges, badges, CSR participation and policy sign-offs"
      actions={
        <>
          <Link to="/social">
            <Button className="secondary sm">Browse CSR</Button>
          </Link>
          <Link to="/gamification">
            <Button className="primary sm">Join Challenge</Button>
          </Link>
        </>
      }
    >
      <StatBar
        items={[
          { label: 'My XP', value: xp, sub: `${points} redeemable points` },
          { label: 'Badges Earned', value: badgeCount, sub: 'catalog badges shown' },
          { label: 'Active Challenges', value: activeChallenges.length, sub: 'open to join or progress' },
          { label: 'My Rank', value: myRank === 0 ? '—' : myRank, sub: 'org leaderboard' },
        ]}
      />

      {pendingPolicies > 0 && (
        <div className="note section-block">
          <span aria-hidden>⚠</span>
          <div>
            <b>
              {pendingPolicies} polic{pendingPolicies === 1 ? 'y awaits' : 'ies await'} your acknowledgement
            </b>{' '}
            — open Policy Sign-off to stay compliant.
            <div style={{ marginTop: 8 }}>
              <Link to="/governance">
                <Button className="sm primary">Policy Sign-off</Button>
              </Link>
            </div>
          </div>
        </div>
      )}

      <h2 className="section-title">My active challenges</h2>
      <div className="grid autofit">
        {activeChallenges.slice(0, 3).map((c) => (
          <Card key={c.id}>
            <div className="card-head">
              <h3>{c.title}</h3>
              <Pill status="active">Active</Pill>
            </div>
            <p className="muted">{c.description}</p>
            <div className="rowflex" style={{ marginTop: 12 }}>
              <Progress value={40} />
              <span className="muted">{c.xp} XP</span>
            </div>
            <div className="card-foot">
              <span className="meta">
                {c.deadline ? `Ends ${String(c.deadline).slice(0, 10)}` : 'No deadline'} · {c.difficulty}
              </span>
              <Link to="/gamification">
                <Button className="sm primary">Open</Button>
              </Link>
            </div>
          </Card>
        ))}
        {activeChallenges.length === 0 && (
          <p className="muted">No active challenges right now — check Gamification for drafts going live.</p>
        )}
      </div>

      <h2 className="section-title">Open CSR activities</h2>
      <div className="grid autofit">
        {(activities.data?.items ?? []).slice(0, 3).map((a) => (
          <Card key={a.id}>
            <div className="card-head">
              <h3>{a.title}</h3>
              <Pill status={a.evidenceRequired ? 'warning' : 'neutral'}>
                {a.evidenceRequired ? 'Evidence' : 'Open'}
              </Pill>
            </div>
            <p className="muted">{a.description}</p>
            <div className="card-foot">
              <span className="meta">{a.points} pts</span>
              <Link to="/social">
                <Button className="sm">View</Button>
              </Link>
            </div>
          </Card>
        ))}
      </div>
    </PageFrame>
  )
}

function DepartmentHub() {
  const user = useAuthStore((s) => s.user)
  const scores = useQuery({ queryKey: queryKeys.scoresOverall, queryFn: () => api.scores.overall() })
  const carbon = useQuery({ queryKey: queryKeys.carbonSummary, queryFn: () => api.carbon.summary() })
  const goals = useQuery({ queryKey: queryKeys.goals, queryFn: () => api.goals.list() })
  const csrParts = useQuery({
    queryKey: queryKeys.csrParticipations,
    queryFn: () => api.social.participations('pending'),
  })
  const challengeParts = useQuery({
    queryKey: queryKeys.challengeParticipations,
    queryFn: api.game.participations,
  })

  const deptScore = (scores.data?.departments ?? []).find((d) => d.departmentId === user?.departmentId)
  const pendingCsr = (csrParts.data?.items ?? []).filter((p) => p.approval === 'pending').length
  const pendingCh = (challengeParts.data?.items ?? []).filter((p) => p.approval === 'pending').length
  const activeGoals = (goals.data?.items ?? []).filter((g) => g.status !== 'completed')

  return (
    <PageFrame
      eyebrow="Department Hub"
      title="Department operations"
      sub="Operational data · department goals · team participation"
      actions={
        <>
          <Link to="/environmental">
            <Button className="secondary sm">Log Carbon</Button>
          </Link>
          <Link to="/environmental">
            <Button className="primary sm">New Goal</Button>
          </Link>
        </>
      }
    >
      <StatBar
        items={[
          {
            label: 'Department ESG Score',
            value: deptScore?.total ?? scores.data?.overall ?? '—',
            unit: '/ 100',
            sub: 'your department roll-up',
          },
          {
            label: 'Emissions (verified)',
            value: carbon.data?.total ?? '0',
            sub: 'org summary · filter in Environmental',
          },
          { label: 'Active Goals', value: activeGoals.length, sub: 'reduction targets' },
          {
            label: 'Pending Approvals',
            value: pendingCsr + pendingCh,
            sub: `${pendingCsr} CSR · ${pendingCh} challenges`,
          },
        ]}
      />

      <div className="grid cols-2 section-block">
        <Card>
          <div className="card-head">
            <h3>Emissions by source</h3>
            <Link to="/environmental" className="muted">
              Open
            </Link>
          </div>
          <div className="stack">
            {Object.entries(carbon.data?.bySource ?? {}).map(([source, value]) => (
              <div className="list-row" key={source}>
                <span>{source}</span>
                <b>{String(value)}</b>
              </div>
            ))}
            {!carbon.data?.bySource || Object.keys(carbon.data.bySource).length === 0 ? (
              <p className="muted">No verified emissions for the current summary window.</p>
            ) : null}
          </div>
        </Card>
        <Card>
          <div className="card-head">
            <h3>Department goals</h3>
            <Link to="/environmental" className="muted">
              View all
            </Link>
          </div>
          <div className="stack">
            {activeGoals.slice(0, 4).map((g) => {
              const target = Number(g.targetCo2) || 1
              const current = Number(g.currentCo2) || 0
              const pct = Math.min(100, Math.round((current / target) * 100))
              return (
                <div className="list-row" key={g.id}>
                  <div style={{ flex: 1 }}>
                    <b>{g.name}</b>
                    <div className="muted" style={{ fontSize: 12 }}>
                      Due {String(g.deadline).slice(0, 10)}
                    </div>
                  </div>
                  <div className="rowflex">
                    <Progress value={pct} tone={g.status === 'at_risk' ? 'danger' : undefined} />
                    <Pill status={g.status === 'at_risk' ? 'danger' : 'plum'}>{pct}%</Pill>
                  </div>
                </div>
              )
            })}
            {activeGoals.length === 0 && <p className="muted">No active goals — create one from Environmental.</p>}
          </div>
        </Card>
      </div>

      <Card className="section-block">
        <div className="card-head">
          <h3>Team approvals queue</h3>
          <div className="rowflex">
            <Link to="/social">
              <Button className="sm secondary">CSR queue</Button>
            </Link>
            <Link to="/gamification">
              <Button className="sm primary">Challenge queue</Button>
            </Link>
          </div>
        </div>
        <p className="muted">
          {pendingCsr + pendingCh === 0
            ? 'No pending participation approvals.'
            : `${pendingCsr + pendingCh} items waiting for your review.`}
        </p>
      </Card>
    </PageFrame>
  )
}

function AuditorHub() {
  const issues = useQuery({
    queryKey: queryKeys.complianceIssues,
    queryFn: () => api.governance.issues(),
  })
  const audits = useQuery({ queryKey: queryKeys.audits, queryFn: () => api.governance.audits() })
  const acks = useQuery({
    queryKey: queryKeys.acknowledgements,
    queryFn: () => api.governance.acknowledgements(),
  })
  const carbon = useQuery({ queryKey: queryKeys.carbonSummary, queryFn: () => api.carbon.summary() })

  const openIssues = (issues.data?.items ?? []).filter((i) => i.status !== 'resolved')
  const overdue = openIssues.filter((i) => i.overdue).length
  const auditsQ = audits.data?.items ?? []
  const underReview = auditsQ.filter((a) => a.status === 'under_review' || a.status === 'draft').length

  return (
    <PageFrame
      eyebrow="Audit Hub"
      title="Audit & Compliance Workspace"
      sub="Audit trail · issue tracking · acknowledgements · data verification"
      actions={
        <>
          <Link to="/governance">
            <Button className="secondary sm">New Audit</Button>
          </Link>
          <Link to="/governance">
            <Button className="primary sm">Raise Issue</Button>
          </Link>
        </>
      }
    >
      <StatBar
        items={[
          { label: 'Open Issues', value: openIssues.length, sub: `${overdue} overdue` },
          { label: 'Audits', value: auditsQ.length, sub: `${underReview} open / review` },
          { label: 'Policy Acknowledgements', value: acks.data?.items?.length ?? 0, sub: 'recorded acks' },
          { label: 'Verified Carbon', value: carbon.data?.total ?? '0', sub: 't CO₂e summary' },
        ]}
      />

      <h2 className="section-title">Assigned audit queue</h2>
      <div className="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Audit</th>
              <th>Department</th>
              <th>Date</th>
              <th>Status</th>
              <th>Action</th>
            </tr>
          </thead>
          <tbody>
            {auditsQ.slice(0, 6).map((a) => (
              <tr key={a.id}>
                <td>
                  <b>{a.title}</b>
                </td>
                <td>{a.departmentName || a.departmentId.slice(0, 8)}</td>
                <td>{String(a.auditDate).slice(0, 10)}</td>
                <td>
                  <Pill status={a.status}>{a.status.replace(/_/g, ' ')}</Pill>
                </td>
                <td>
                  <Link to="/governance">
                    <Button className="sm secondary">Open</Button>
                  </Link>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {auditsQ.length === 0 && <p className="muted" style={{ padding: 12 }}>No audits yet.</p>}
      </div>

      <h2 className="section-title">Open compliance issues</h2>
      <div className="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Description</th>
              <th>Severity</th>
              <th>Due</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            {openIssues.slice(0, 6).map((i) => (
              <tr key={i.id}>
                <td>
                  <b>{i.description}</b>
                  {i.overdue ? <div className="danger-text" style={{ fontSize: 12 }}>Overdue</div> : null}
                </td>
                <td>
                  <Pill status={i.severity === 'high' ? 'danger' : i.severity === 'medium' ? 'warning' : 'neutral'}>
                    {i.severity}
                  </Pill>
                </td>
                <td>{String(i.dueDate).slice(0, 10)}</td>
                <td>
                  <Pill status={i.status}>{i.status.replace(/_/g, ' ')}</Pill>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {openIssues.length === 0 && <p className="muted" style={{ padding: 12 }}>No open issues.</p>}
      </div>
    </PageFrame>
  )
}
