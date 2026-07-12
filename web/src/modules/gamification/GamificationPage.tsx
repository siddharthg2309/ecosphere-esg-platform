import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useState, useEffect, type FormEvent } from 'react'
import { useSearchParams } from 'react-router-dom'
import { RoleGuard } from '../../app/RoleGuard'
import { useAuthStore } from '../../app/authStore'
import { Button, Card, EmptyState, Modal, Note, Pill, Progress, initials } from '../../design/components'
import { api } from '../../lib/apiClient'
import { userFacingError } from '../../lib/userFacingError'
import { queryKeys } from '../../lib/queryKeys'
import type { Category, Challenge, ChallengeParticipation, ChallengeStatus, Reward } from '../../lib/types'
import { EvidenceAssist, ProofUploadField } from '../ai/EvidenceAssist'
import { allowedTransitions, canApproveParticipation } from './challengeTransitions'
import { applyOptimisticRedeem, canRedeem, rollbackPoints } from './redeemVM'

type Tab = 'challenges' | 'participation' | 'badges' | 'rewards' | 'leaderboard'
const tabs: [Tab, string][] = [
  ['challenges', 'Challenges'],
  ['participation', 'Challenge Participation'],
  ['badges', 'Badges'],
  ['rewards', 'Rewards'],
  ['leaderboard', 'Leaderboard'],
]

function isTab(v: string | null): v is Tab {
  return v === 'challenges' || v === 'participation' || v === 'badges' || v === 'rewards' || v === 'leaderboard'
}

export function GamificationPage() {
  const [params, setParams] = useSearchParams()
  const paramTab = params.get('tab')
  const [tab, setTab] = useState<Tab>(() => (isTab(paramTab) ? paramTab : 'challenges'))

  useEffect(() => {
    if (isTab(params.get('tab'))) {
      setTab(params.get('tab') as Tab)
    }
  }, [params])

  function selectTab(next: Tab) {
    setTab(next)
    const nextParams = new URLSearchParams(params)
    nextParams.set('tab', next)
    nextParams.delete('action')
    setParams(nextParams, { replace: true })
  }

  return (
    <main className="page">
      <div className="content">
        <header className="page-head">
          <div>
            <p className="eyebrow">Gamification</p>
            <h1>Challenges, Badges &amp; Leaderboard</h1>
          </div>
        </header>
        <div className="tabs" role="tablist" aria-label="Gamification sections">
          {tabs.map(([id, label]) => (
            <button key={id} role="tab" aria-selected={tab === id} className={`tab ${tab === id ? 'active' : ''}`} onClick={() => selectTab(id)}>
              {label}
            </button>
          ))}
        </div>
        {tab === 'challenges' && <ChallengesPanel />}
        {tab === 'participation' && <ParticipationPanel />}
        {tab === 'badges' && <BadgesPanel />}
        {tab === 'rewards' && <RewardsPanel />}
        {tab === 'leaderboard' && <LeaderboardPanel />}
      </div>
    </main>
  )
}

function ChallengesPanel() {
  const qc = useQueryClient()
  const [params, setParams] = useSearchParams()
  const list = useQuery({ queryKey: queryKeys.challenges, queryFn: api.game.challenges })
  const counts = useQuery({ queryKey: queryKeys.challengeCounts, queryFn: api.game.statusCounts })
  const categories = useQuery({ queryKey: queryKeys.categories, queryFn: () => api.master.list<Category>('categories') })
  const [open, setOpen] = useState(false)
  const [transition, setTransition] = useState<Challenge | null>(null)
  const [join, setJoin] = useState<Challenge | null>(null)
  const [error, setError] = useState<unknown>()

  const action = params.get('action')
  useEffect(() => {
    if (action === 'new-challenge') {
      setOpen(true)
      const nextParams = new URLSearchParams(window.location.search)
      nextParams.delete('action')
      setParams(nextParams, { replace: true })
    }
  }, [action])

  const create = useMutation({
    mutationFn: api.game.createChallenge,
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: queryKeys.challenges })
      void qc.invalidateQueries({ queryKey: queryKeys.challengeCounts })
      setOpen(false)
    },
  })
  const transit = useMutation({
    mutationFn: ({ id, to }: { id: string; to: ChallengeStatus }) => api.game.transition(id, to),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: queryKeys.challenges })
      void qc.invalidateQueries({ queryKey: queryKeys.challengeCounts })
      setTransition(null)
    },
  })
  const participate = useMutation({
    mutationFn: ({ id, proofUrl }: { id: string; proofUrl: string }) => api.game.participate(id, { progress: 100, proofUrl }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: queryKeys.challengeParticipations })
      void qc.invalidateQueries({ queryKey: queryKeys.challenges })
      setJoin(null)
    },
  })
  const c = counts.data ?? {}
  const challengeCats = (categories.data?.items ?? []).filter((x) => x.type === 'challenge')

  async function submitCreate(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setError(undefined)
    const f = new FormData(e.currentTarget)
    try {
      await create.mutateAsync({
        title: String(f.get('title')),
        categoryId: String(f.get('categoryId')),
        description: String(f.get('description')),
        xp: Number(f.get('xp') || 0),
        difficulty: String(f.get('difficulty') || 'medium'),
        evidenceRequired: f.get('evidenceRequired') === 'on',
        deadline: String(f.get('deadline') || '') || undefined,
      })
    } catch (err) {
      setError(err)
    }
  }

  return (
    <>
      <div className="section-head">
        <div className="lifecycle-pills">
          <Pill status="draft">Draft {c.draft ?? 0}</Pill>
          <Pill status="active">Active {c.active ?? 0}</Pill>
          <Pill status="under_review">Under Review {c.under_review ?? 0}</Pill>
          <Pill status="neutral">Completed {c.completed ?? 0}</Pill>
          <Pill status="neutral">Archived {c.archived ?? 0}</Pill>
        </div>
        <RoleGuard roles={['admin', 'dept_head']}>
          <Button className="primary sm" onClick={() => setOpen(true)}>
            + New Challenge
          </Button>
        </RoleGuard>
      </div>
      {list.isLoading ? (
        <div className="center-state">Loading challenges…</div>
      ) : (list.data?.items ?? []).length === 0 ? (
        <EmptyState>
          <h3>No challenges yet</h3>
          
        </EmptyState>
      ) : (
        <div className="grid autofit">
          {(list.data?.items ?? []).map((ch) => (
            <Card key={ch.id}>
              <div className="card-head">
                <h3>{ch.title}</h3>
                <Pill status={ch.status}>{ch.status.replace('_', ' ')}</Pill>
              </div>
              <div className="rowflex" style={{ marginTop: 12 }}>
                <Pill status="warning">{ch.xp} XP</Pill>
                <Pill status="neutral">{ch.difficulty}</Pill>
                {ch.evidenceRequired && <span className="meta">Evidence</span>}
              </div>
              <div className="card-foot">
                <span className="meta">
                  {ch.deadline ? `Ends ${String(ch.deadline).slice(0, 10)}` : ch.pendingCount ? `${ch.pendingCount} pending` : 'Open'}
                </span>
                <div className="rowflex">
                  <RoleGuard roles={['admin', 'dept_head']}>
                    <Button className="secondary sm" onClick={() => setTransition(ch)}>
                      Transition
                    </Button>
                  </RoleGuard>
                  {ch.status === 'active' && (
                    <Button className="primary sm" onClick={() => setJoin(ch)}>
                      Join Challenge
                    </Button>
                  )}
                </div>
              </div>
            </Card>
          ))}
        </div>
      )}
      {open && (
        <Modal
          title="New Challenge"
          onClose={() => setOpen(false)}
          footer={
            <>
              <span className="muted" style={{ marginRight: 'auto' }}>
                <Pill status="draft">Saved as Draft</Pill>
              </span>
              <Button className="secondary" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <Button className="primary" form="create-challenge" disabled={create.isPending}>
                Create Challenge
              </Button>
            </>
          }
        >
          <form id="create-challenge" className="modal-form" onSubmit={submitCreate}>
            <label>
              Challenge Title
              <input name="title" required placeholder="e.g. Zero Waste Week" />
            </label>
            <label>
              Category
              <select name="categoryId" required>
                {challengeCats.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name}
                  </option>
                ))}
              </select>
            </label>
            <label>
              Description
              <textarea name="description" rows={3} />
            </label>
            <div className="field-row">
              <label>
                XP Reward
                <input name="xp" type="number" min={0} defaultValue={120} />
              </label>
              <label>
                Difficulty
                <select name="difficulty" defaultValue="medium">
                  <option value="easy">Easy</option>
                  <option value="medium">Medium</option>
                  <option value="hard">Hard</option>
                </select>
              </label>
            </div>
            <label>
              Deadline
              <input name="deadline" type="date" />
            </label>
            <label className="rowflex" style={{ textTransform: 'none', letterSpacing: 0, fontWeight: 500 }}>
              <input name="evidenceRequired" type="checkbox" defaultChecked /> Require evidence for participation
            </label>
            <ErrorMessage error={error} />
          </form>
        </Modal>
      )}
      {transition && (
        <Modal title="Transition Challenge" onClose={() => setTransition(null)}>
          <p className="muted">
            Current status: <Pill status={transition.status}>{transition.status}</Pill>
          </p>
          <div className="rowflex" style={{ marginTop: 16 }}>
            {allowedTransitions(transition.status).map((to) => (
              <Button key={to} className="primary sm" onClick={() => void transit.mutateAsync({ id: transition.id, to })}>
                → {to.replace('_', ' ')}
              </Button>
            ))}
            {allowedTransitions(transition.status).length === 0 && <span className="muted">No further transitions</span>}
          </div>
        </Modal>
      )}
      {join && (
        <Modal
          title="Join Challenge"
          onClose={() => setJoin(null)}
          footer={
            <>
              <Button className="secondary" onClick={() => setJoin(null)}>
                Cancel
              </Button>
              <Button
                className="primary"
                form="join-challenge"
                disabled={participate.isPending}
              >
                Submit
              </Button>
            </>
          }
        >
          <p>
            <b>{join.title}</b> · {join.xp} XP
          </p>
          <form
            id="join-challenge"
            className="modal-form"
            onSubmit={(e) => {
              e.preventDefault()
              const f = new FormData(e.currentTarget)
              void participate.mutateAsync({ id: join.id, proofUrl: String(f.get('proofUrl') || '') }).catch(setError)
            }}
          >
            <ProofUploadField required={join.evidenceRequired} name="proofUrl" />
            <ErrorMessage error={error} />
          </form>
        </Modal>
      )}
    </>
  )
}

function ParticipationPanel() {
  const qc = useQueryClient()
  const list = useQuery({ queryKey: queryKeys.challengeParticipations, queryFn: api.game.participations })
  const approve = useMutation({
    mutationFn: api.game.approveParticipation,
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: queryKeys.challengeParticipations })
      void qc.invalidateQueries({ queryKey: queryKeys.leaderboard('employee') })
      void qc.invalidateQueries({ queryKey: queryKeys.gameBadges })
    },
  })
  const reject = useMutation({
    mutationFn: api.game.rejectParticipation,
    onSuccess: () => void qc.invalidateQueries({ queryKey: queryKeys.challengeParticipations }),
  })
  return (
    <>
      <Note>Approving a challenge participation awards XP, increments the completed-challenge count, and triggers badge auto-award check.</Note>
      <div className="table-wrap" style={{ marginTop: 14 }}>
        <table>
          <thead>
            <tr>
              <th>Employee</th>
              <th>Challenge</th>
              <th>Progress</th>
              <th>Proof</th>
              <th className="numeric">XP</th>
              <th>Status</th>
              <th>Action</th>
            </tr>
          </thead>
          <tbody>
            {(list.data?.items ?? []).map((row) => (
              <ChallengePartRow
                key={row.id}
                row={row}
                onApprove={() => void approve.mutateAsync(row.id)}
                onReject={() => void reject.mutateAsync(row.id)}
              />
            ))}
          </tbody>
        </table>
      </div>
    </>
  )
}

function ChallengePartRow({
  row,
  onApprove,
  onReject,
}: {
  row: ChallengeParticipation
  onApprove(): void
  onReject(): void
}) {
  const canApprove =
    row.approval === 'pending' && canApproveParticipation(row.proofUrl, Boolean(row.evidenceRequired))
  return (
    <tr>
      <td>
        <span className="avatar-sm">{initials(row.employeeName ?? '?')}</span>
        {row.employeeName}
      </td>
      <td>{row.challengeTitle}</td>
      <td>
        <div className="rowflex">
          <Progress value={row.progress} tone={row.progress < 100 ? 'warning' : undefined} />
          <span className="muted">{row.progress}%</span>
        </div>
      </td>
      <td>
        {row.proofUrl ? (
          row.proofUrl.startsWith('http') ? (
            <a href={row.proofUrl} target="_blank" rel="noreferrer">
              proof
            </a>
          ) : (
            <span className="muted" title={row.proofUrl}>
              {row.proofUrl.replace(/^upload:/, '')}
            </span>
          )
        ) : (
          <Pill status="neutral">No proof yet</Pill>
        )}
        {row.approval === 'pending' ? <EvidenceAssist proofUrl={row.proofUrl} /> : null}
      </td>
      <td className="numeric">{row.approval === 'approved' ? row.xpAwarded : row.challengeXp ?? '—'}</td>
      <td>
        <Pill status={row.approval}>{row.approval.replace('_', ' ')}</Pill>
      </td>
      <td>
        {row.approval === 'pending' || row.approval === 'in_progress' ? (
          row.approval === 'pending' ? (
            <RoleGuard roles={['admin', 'dept_head']}>
              <div className="rowflex">
                <Button className="primary sm" disabled={!canApprove} onClick={onApprove}>
                  Approve
                </Button>
                <Button className="secondary sm" onClick={onReject}>
                  Reject
                </Button>
              </div>
            </RoleGuard>
          ) : (
            <span className="muted">Ongoing</span>
          )
        ) : row.approval === 'approved' ? (
          <span className="muted">+{row.xpAwarded} XP awarded</span>
        ) : (
          <span className="muted">Rejected</span>
        )}
      </td>
    </tr>
  )
}

function BadgesPanel() {
  const list = useQuery({ queryKey: queryKeys.gameBadges, queryFn: api.game.badges })
  return (
    <>
      <p className="muted" style={{ fontSize: 13, marginBottom: 14 }}>
        Auto-awarded when XP or completed-challenge count meets the unlock rule
      </p>
      <div className="grid autofit">
        {(list.data?.items ?? []).map((b) => (
          <Card key={b.id}>
            <div className="rowflex" style={{ gap: 14, alignItems: 'flex-start' }}>
              <span className="tile-ico">★</span>
              <div style={{ flex: 1 }}>
                <div style={{ fontWeight: 700 }}>{b.name}</div>
                <div className="muted" style={{ fontSize: 12, marginTop: 3 }}>
                  Unlock: {b.unlockRule.type === 'xp' ? `${b.unlockRule.value} XP` : `${b.unlockRule.value} challenges`} · {b.earnedCount ?? 0}{' '}
                  employees earned
                </div>
                <div className="rowflex" style={{ marginTop: 8 }}>
                  <Pill status="plum">{b.unlockRule.type === 'xp' ? 'XP threshold' : 'Challenge count'}</Pill>
                </div>
              </div>
            </div>
          </Card>
        ))}
      </div>
    </>
  )
}

function RewardsPanel() {
  const qc = useQueryClient()
  const user = useAuthStore((s) => s.user)
  const [points, setPoints] = useState<number | null>(null)
  const [error, setError] = useState<string>()
  const list = useQuery({ queryKey: queryKeys.gameRewards, queryFn: api.game.rewards })
  const balQ = useQuery({ queryKey: ['me', 'balance'], queryFn: api.game.balance })
  const balance = points ?? balQ.data?.points ?? 0

  const redeem = useMutation({
    mutationFn: api.game.redeem,
    onMutate: async (rewardId) => {
      const reward = (list.data?.items ?? []).find((r) => r.id === rewardId)
      if (!reward) return { prev: balance }
      const prev = balance
      setPoints(applyOptimisticRedeem(prev, reward.pointsRequired))
      setError(undefined)
      return { prev }
    },
    onError: (_err, _id, ctx) => {
      if (ctx?.prev != null) setPoints(rollbackPoints(ctx.prev))
      setError(userFacingError(_err, 'Redeem failed'))
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: queryKeys.gameRewards })
      void balQ.refetch().then((r) => setPoints(r.data?.points ?? null))
    },
  })

  return (
    <>
      <div className="section-head">
        <span className="muted" style={{ fontSize: 13 }}>
          Points redeemed from employee balance · stock-limited · atomic deduction
          {user && (
            <>
              {' '}
              · Your balance: <b>{balance} pts</b>
            </>
          )}
        </span>
      </div>
      {error && (
        <div className="form-error" role="alert">
          {error}
        </div>
      )}
      <Card>
        <div className="card-head">
          <h3>Rewards catalog</h3>
          
        </div>
        <div className="stack">
          {(list.data?.items ?? []).map((r: Reward) => {
            const ok = canRedeem(r.stock, r.pointsRequired, balance)
            return (
              <div className="list-row" key={r.id} style={r.stock <= 0 ? { opacity: 0.55 } : undefined}>
                <div>
                  <b>{r.name}</b> <span className="muted">· Stock: {r.stock}</span>
                </div>
                <div className="rowflex">
                  <Pill status="plum">{r.pointsRequired} pts</Pill>
                  {r.stock <= 0 ? (
                    <Button className="secondary sm" disabled>
                      Out of Stock
                    </Button>
                  ) : (
                    <Button className="primary sm" disabled={!ok || redeem.isPending} onClick={() => void redeem.mutateAsync(r.id)}>
                      Redeem
                    </Button>
                  )}
                </div>
              </div>
            )
          })}
        </div>
      </Card>
    </>
  )
}

function LeaderboardPanel() {
  const employees = useQuery({ queryKey: queryKeys.leaderboard('employee'), queryFn: () => api.game.leaderboard('employee') })
  const departments = useQuery({ queryKey: queryKeys.leaderboard('department'), queryFn: () => api.game.leaderboard('department') })
  return (
    <div className="grid cols-2">
      <Card>
        <div className="card-head">
          <h3>Employee Leaderboard</h3>
          <span className="muted">by XP</span>
        </div>
        <div className="table-wrap" style={{ border: 0 }}>
          <table>
            <thead>
              <tr>
                <th>#</th>
                <th>Employee</th>
                <th className="numeric">XP</th>
                <th className="numeric">Badges</th>
              </tr>
            </thead>
            <tbody>
              {(employees.data?.items ?? []).map((e) => (
                <tr key={e.id}>
                  <td>
                    <b>{e.rank}</b>
                  </td>
                  <td>
                    <span className="avatar-sm">{initials(e.name)}</span>
                    {e.name}
                  </td>
                  <td className="numeric">
                    <b>{e.xp.toLocaleString()}</b>
                  </td>
                  <td className="numeric">{e.badgeCount}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Card>
      <Card>
        <div className="card-head">
          <h3>Department Leaderboard</h3>
          <span className="muted">total XP</span>
        </div>
        <div className="table-wrap" style={{ border: 0 }}>
          <table>
            <thead>
              <tr>
                <th>#</th>
                <th>Department</th>
                <th className="numeric">Total XP</th>
                <th className="numeric">Badges</th>
              </tr>
            </thead>
            <tbody>
              {(departments.data?.items ?? []).map((e) => (
                <tr key={e.id}>
                  <td>
                    <b>{e.rank}</b>
                  </td>
                  <td>
                    <span className="avatar-sm">{initials(e.name)}</span>
                    {e.name}
                  </td>
                  <td className="numeric">
                    <b>{e.xp.toLocaleString()}</b>
                  </td>
                  <td className="numeric">{e.badgeCount}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Card>
    </div>
  )
}

function ErrorMessage({ error }: { error: unknown }) {
  if (!error) return null
  return (
    <div className="form-error" role="alert">
      {userFacingError(error, 'Unable to save changes')}
    </div>
  )
}
