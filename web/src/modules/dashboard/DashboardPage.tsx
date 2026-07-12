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
  actions,
  children,
}: {
  eyebrow: string
  title: string
  sub?: string
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
          </div>
          {actions ? <div className="rowflex">{actions}</div> : null}
        </header>
        {children}
      </div>
    </main>
  )
}

/* ── Sparkline SVG (Emissions Trend) ── */
function EmissionsSparkline() {
  return (
    <svg
      className="admin-spark"
      viewBox="0 0 520 150"
      preserveAspectRatio="none"
      aria-hidden="true"
    >
      <defs>
        <linearGradient id="spark-grad" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" stopColor="#714B67" stopOpacity=".16" />
          <stop offset="1" stopColor="#714B67" stopOpacity="0" />
        </linearGradient>
      </defs>
      <line x1="0" y1="37" x2="520" y2="37" stroke="#E9ECEF" />
      <line x1="0" y1="75" x2="520" y2="75" stroke="#E9ECEF" />
      <line x1="0" y1="113" x2="520" y2="113" stroke="#E9ECEF" />
      <path
        d="M0,40 L47,52 L94,44 L141,66 L188,60 L235,82 L282,74 L329,96 L376,88 L423,104 L470,100 L520,118 L520,150 L0,150 Z"
        fill="url(#spark-grad)"
      />
      <path
        d="M0,40 L47,52 L94,44 L141,66 L188,60 L235,82 L282,74 L329,96 L376,88 L423,104 L470,100 L520,118"
        fill="none"
        stroke="#714B67"
        strokeWidth="2.5"
        strokeLinejoin="round"
      />
    </svg>
  )
}

/* ── Department bar chart ── */
function DeptBarChart({
  departments,
}: {
  departments: { name?: string; total: number | string }[]
}) {
  // Use live data if available, otherwise show wireframe placeholders
  const bars =
    departments.length > 0
      ? departments.slice(0, 5).map((d) => ({
          label: d.name?.slice(0, 4) || 'Dept',
          value: Number(d.total) || 0,
        }))
      : [
          { label: 'Corp', value: 88 },
          { label: 'R&D', value: 82 },
          { label: 'Mfg', value: 74 },
          { label: 'Logi', value: 61 },
          { label: 'Sales', value: 47 },
        ]

  const max = Math.max(...bars.map((b) => b.value), 1)

  function barClass(value: number) {
    if (value < 55) return 'danger'
    if (value < 70) return 'warning'
    return ''
  }

  return (
    <div className="admin-bars">
      {bars.map((b) => (
        <div key={b.label} className="admin-bar-col">
          <em>{b.value}</em>
          <div
            className={`admin-bar ${barClass(b.value)}`}
            style={{ height: `${Math.round((b.value / max) * 100)}%` }}
          />
          <span>{b.label}</span>
        </div>
      ))}
    </div>
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

  const recentActivity = (notifs.data?.items ?? []).slice(0, 4)

  return (
    <PageFrame
      eyebrow="Dashboard"
      title="Executive Overview"
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
          <Link to="/environmental?action=log-carbon">
            <Button className="secondary sm">Log Carbon Data</Button>
          </Link>
          <Link to="/reports">
            <Button className="primary sm">+ New Report</Button>
          </Link>
        </>
      }
    >
      {recompute.isError && (
        <div className="form-error" role="alert">
          {userFacingError(recompute.error, 'Unable to recompute scores')}
        </div>
      )}

      {/* ── Stat strip ── */}
      <StatBar
        items={[
          {
            label: 'Environmental',
            value: s?.environmental ?? '—',
            unit: '/ 100',
            sub: '+4.1 vs last quarter',
          },
          {
            label: 'Social',
            value: s?.social ?? '—',
            unit: '/ 100',
            sub: '+2.3 vs last quarter',
          },
          {
            label: 'Governance',
            value: s?.governance ?? '—',
            unit: '/ 100',
            sub: '-1.2 vs last quarter',
          },
          {
            label: 'Overall ESG Score',
            value: s?.overall ?? '—',
            unit: '/ 100',
            sub: weights
              ? `${weights.weightEnv} / ${weights.weightSocial} / ${weights.weightGov} weighting`
              : '40 / 30 / 30 weighting',
          },
        ]}
      />

      {/* ── Trend + Ranking row ── */}
      <div className="grid cols-2 section-block">
        {/* Emissions Trend */}
        <Card>
          <div className="card-head">
            <h3>
              Emissions Trend{' '}
              <span className="muted" style={{ fontWeight: 400 }}>
                · 12 months (t CO₂)
              </span>
            </h3>
            <Pill status="plum">-18% YoY</Pill>
          </div>
          <EmissionsSparkline />
          <div className="admin-spark-labels">
            <span>Aug</span>
            <span>Oct</span>
            <span>Dec</span>
            <span>Feb</span>
            <span>Apr</span>
            <span>Jun</span>
          </div>
        </Card>

        {/* Department ESG Ranking */}
        <Card>
          <div className="card-head">
            <h3>Department ESG Ranking</h3>
            <Link to="/reports" className="muted" style={{ fontSize: 13 }}>
              View all
            </Link>
          </div>
          <DeptBarChart departments={s?.departments ?? []} />
        </Card>
      </div>

      {/* ── Activity + Quick Actions row ── */}
      <div className="grid cols-2 section-block">
        {/* Recent Activity */}
        <Card>
          <div className="card-head">
            <h3>Recent Activity</h3>
          </div>
          <div className="stack" style={{ gap: 0 }}>
            {recentActivity.length > 0
              ? recentActivity.map((n) => (
                  <div className="list-row admin-activity-row" key={n.id}>
                    <span className="avatar-sm">{n.title.slice(0, 2).toUpperCase()}</span>
                    <div style={{ flex: 1 }}>
                      <b>{n.title}</b>
                      <div className="muted" style={{ fontSize: 12 }}>
                        {n.type.replace(/_/g, ' ')} ·{' '}
                        {String(n.createdAt).slice(0, 10)}
                      </div>
                    </div>
                    {!n.readAt && <Pill status="warning">Unread</Pill>}
                  </div>
                ))
              : (
                  <>
                    <div className="list-row admin-activity-row">
                      <span className="avatar-sm">PS</span>
                      <div style={{ flex: 1 }}>
                        Priya S. completed <b>Zero Waste Week</b>{' '}
                        <span className="pill plum admin-xp-pill">+200 XP</span>
                        <div className="muted" style={{ fontSize: 12 }}>2h ago · Manufacturing</div>
                      </div>
                    </div>
                    <div className="list-row admin-activity-row">
                      <span className="tile-ico admin-tile-dng">⚠</span>
                      <div style={{ flex: 1 }}>
                        New compliance issue in <b>Logistics</b>{' '}
                        <Pill status="danger">High</Pill>
                        <div className="muted" style={{ fontSize: 12 }}>5h ago · Auditor K. Menon</div>
                      </div>
                    </div>
                    <div className="list-row admin-activity-row">
                      <span className="tile-ico">↻</span>
                      <div style={{ flex: 1 }}>
                        <b>42</b> new carbon transactions (Fleet)
                        <div className="muted" style={{ fontSize: 12 }}>Today · auto-calculated</div>
                      </div>
                    </div>
                    <div className="list-row admin-activity-row">
                      <span className="tile-ico">✓</span>
                      <div style={{ flex: 1 }}>
                        <b>R&D</b> acknowledged Anti-Corruption Policy
                        <div className="muted" style={{ fontSize: 12 }}>Yesterday · 41/41 employees</div>
                      </div>
                    </div>
                  </>
                )}
          </div>
        </Card>

        {/* Quick Actions */}
        <Card>
          <div className="card-head">
            <h3>Quick Actions</h3>
          </div>
          <div className="grid cols-2" style={{ gap: 12 }}>
            <Link to="/environmental?action=log-carbon">
              <Button className="secondary admin-qa-btn">↑ Log Carbon Data</Button>
            </Link>
            <Link to="/gamification?action=new-challenge">
              <Button className="secondary admin-qa-btn">◎ Create Challenge</Button>
            </Link>
            <Link to="/settings?tab=master&kind=policies">
              <Button className="secondary admin-qa-btn">☰ Publish Policy</Button>
            </Link>
            <Link to="/reports">
              <Button className="secondary admin-qa-btn">↑ Generate Report</Button>
            </Link>
          </div>
          <div className="note" style={{ marginTop: 16 }}>
            <span>⚠</span>
            <div>
              <b>3 goals</b> due this quarter · <b>2 compliance issues</b> approaching due date.
            </div>
          </div>
          <div className="admin-budget-row">
            <span className="muted">ESG Budget utilization</span>
            <div className="rowflex">
              <div className="progress" style={{ width: 140 }}>
                <span style={{ width: '64%' }} />
              </div>
              <b>64%</b>
            </div>
          </div>
        </Card>
      </div>
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
      actions={
        <>
          <Link to="/social?tab=activities">
            <Button className="secondary sm">Browse CSR</Button>
          </Link>
          <Link to="/gamification?tab=challenges">
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
              <Link to="/governance?tab=policies">
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
            <div className="rowflex" style={{ marginTop: 12 }}>
              <Progress value={40} />
              <span className="muted">{c.xp} XP</span>
            </div>
            <div className="card-foot">
              <span className="meta">
                {c.deadline ? `Ends ${String(c.deadline).slice(0, 10)}` : 'No deadline'} · {c.difficulty}
              </span>
              <Link to="/gamification?tab=challenges">
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
            <div className="card-foot">
              <span className="meta">{a.points} pts</span>
              <Link to="/social?tab=activities">
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
      actions={
        <>
          <Link to="/environmental?action=log-carbon">
            <Button className="secondary sm">Log Carbon</Button>
          </Link>
          <Link to="/environmental?tab=goals&action=new-goal">
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
            <Link to="/environmental?tab=transactions" className="muted">
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
            <Link to="/environmental?tab=goals" className="muted">
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
            <Link to="/social?tab=participation">
              <Button className="sm secondary">CSR queue</Button>
            </Link>
            <Link to="/gamification?tab=participation">
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
      actions={
        <>
          <Link className="button secondary sm" to="/governance?tab=audits&action=new-audit">
            New Audit
          </Link>
          <Link className="button primary sm" to="/governance?tab=issues&action=raise-issue">
            Raise Issue
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
                  <Link className="button sm secondary" to="/governance?tab=audits">
                    Open
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
