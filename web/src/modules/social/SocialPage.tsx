import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useState, useEffect, type FormEvent } from 'react'
import { useSearchParams } from 'react-router-dom'
import { RoleGuard } from '../../app/RoleGuard'
import { Button, Card, EmptyState, Modal, Note, Pill, Progress, StatBar, initials } from '../../design/components'
import { api } from '../../lib/apiClient'
import { userFacingError } from '../../lib/userFacingError'
import { queryKeys } from '../../lib/queryKeys'
import type { Category, CSRActivity, CSRParticipation } from '../../lib/types'
import { EvidenceAssist, ProofUploadField } from '../ai/EvidenceAssist'
import { canApproveParticipation } from '../gamification/challengeTransitions'

type Tab = 'activities' | 'participation' | 'training'
const tabs: [Tab, string][] = [
  ['activities', 'CSR Activities'],
  ['participation', 'Employee Participation'],
  ['training', 'Training Completion'],
]

function isTab(v: string | null): v is Tab {
  return v === 'activities' || v === 'participation' || v === 'diversity' || v === 'training'
}

export function SocialPage() {
  const [params, setParams] = useSearchParams()
  const paramTab = params.get('tab')
  const [tab, setTab] = useState<Tab>(() => (isTab(paramTab) ? paramTab : 'activities'))

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

  const diversity = useQuery({ queryKey: queryKeys.diversity, queryFn: api.social.diversity })
  const d = diversity.data
  return (
    <main className="page">

      <div className="content">
        <header className="page-head">
          <div>
            <p className="eyebrow">Social</p>
            <h1>CSR &amp; Employee Engagement</h1>
          </div>
        </header>
        <StatBar
          items={[
            { label: 'Gender Diversity', value: Math.round(d?.genderWomenPct ?? 0), unit: '%', sub: 'women in workforce' },
            { label: 'Leadership Diversity', value: Math.round(d?.leadershipWomenPct ?? 0), unit: '%', sub: 'under-rep. in leadership' },
            { label: 'Training Completion', value: Math.round(d?.trainingCompletionPct ?? 0), unit: '%', sub: 'mandatory ESG training' },
            { label: 'CSR Participation', value: Math.round(d?.csrParticipationPct ?? 0), unit: '%', sub: 'of employees YTD' },
          ]}
        />
        <div className="tabs" role="tablist" aria-label="Social sections">
          {tabs.map(([id, label]) => (
            <button key={id} role="tab" aria-selected={tab === id} className={`tab ${tab === id ? 'active' : ''}`} onClick={() => selectTab(id)}>
              {label}
            </button>
          ))}
        </div>
        {tab === 'activities' && <ActivitiesPanel />}
        {tab === 'participation' && <ParticipationPanel />}
        {tab === 'training' && <TrainingPanel />}
      </div>
    </main>
  )
}

function ActivitiesPanel() {
  const qc = useQueryClient()
  const [params, setParams] = useSearchParams()
  const activities = useQuery({ queryKey: queryKeys.csrActivities, queryFn: api.social.activities })
  const categories = useQuery({ queryKey: queryKeys.categories, queryFn: () => api.master.list<Category>('categories') })
  const [openCreate, setOpenCreate] = useState(false)
  const [join, setJoin] = useState<CSRActivity | null>(null)
  const [view, setView] = useState<CSRActivity | null>(null)
  const [error, setError] = useState<unknown>()

  const action = params.get('action')
  useEffect(() => {
    if (action === 'new-activity') {
      setOpenCreate(true)
      const nextParams = new URLSearchParams(window.location.search)
      nextParams.delete('action')
      setParams(nextParams, { replace: true })
    }
  }, [action])

  const create = useMutation({
    mutationFn: api.social.createActivity,
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: queryKeys.csrActivities })
      setOpenCreate(false)
    },
  })
  const joinMut = useMutation({
    mutationFn: api.social.joinActivity,
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: queryKeys.csrActivities })
      void qc.invalidateQueries({ queryKey: queryKeys.csrParticipations })
      setJoin(null)
    },
  })

  async function submitCreate(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setError(undefined)
    const f = new FormData(e.currentTarget)
    try {
      await create.mutateAsync({
        title: String(f.get('title')),
        categoryId: String(f.get('categoryId')),
        description: String(f.get('description')),
        points: Number(f.get('points') || 0),
        evidenceRequired: f.get('evidenceRequired') === 'on',
        activityDate: String(f.get('activityDate') || '') || undefined,
      })
    } catch (err) {
      setError(err)
    }
  }

  async function submitJoin(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    if (!join) return
    setError(undefined)
    const f = new FormData(e.currentTarget)
    try {
      await joinMut.mutateAsync({
        activityId: join.id,
        proofUrl: String(f.get('proofUrl') || ''),
        notes: String(f.get('notes') || ''),
      })
    } catch (err) {
      setError(err)
    }
  }

  const csrCats = (categories.data?.items ?? []).filter((c) => c.type === 'csr_activity')

  return (
    <>
      <div className="section-head">
        <div>
          <h2>CSR Activities</h2>
        </div>
        <RoleGuard roles={['admin', 'dept_head']}>
          <Button className="primary sm" onClick={() => setOpenCreate(true)}>
            + New Activity
          </Button>
        </RoleGuard>
      </div>
      {activities.isLoading ? (
        <div className="center-state">Loading activities…</div>
      ) : (activities.data?.items ?? []).length === 0 ? (
        <EmptyState>
          <h3>No CSR activities yet</h3>
          
        </EmptyState>
      ) : (
        <div className="grid autofit">
          {(activities.data?.items ?? []).map((a) => (
            <Card key={a.id}>
              <div className="card-head">
                <h3>{a.title}</h3>
                <Pill status="neutral">{csrCats.find((c) => c.id === a.categoryId)?.name ?? 'CSR'}</Pill>
              </div>
              <div className="rowflex" style={{ marginTop: 12 }}>
                {a.evidenceRequired ? <Pill status="warning">Evidence Required</Pill> : <Pill status="neutral">Open</Pill>}
                <span className="meta">{a.joinedCount ?? 0} joined</span>
              </div>
              <div className="card-foot">
                <span className="meta">{a.points} pts</span>
                <div className="rowflex">
                  <Button className="secondary sm" onClick={() => setView(a)}>
                    View
                  </Button>
                  <Button className="primary sm" onClick={() => setJoin(a)}>
                    Join
                  </Button>
                </div>
              </div>
            </Card>
          ))}
        </div>
      )}
      {openCreate && (
        <Modal
          title="New CSR Activity"
          onClose={() => setOpenCreate(false)}
          footer={
            <>
              <Button className="secondary" onClick={() => setOpenCreate(false)}>
                Cancel
              </Button>
              <Button className="primary" form="create-activity" disabled={create.isPending}>
                Create Activity
              </Button>
            </>
          }
        >
          <form id="create-activity" className="modal-form" onSubmit={submitCreate}>
            <label>
              Activity Title
              <input name="title" required placeholder="e.g. Tree Plantation Drive" />
            </label>
            <label>
              Category
              <select name="categoryId" required>
                {csrCats.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name}
                  </option>
                ))}
              </select>
            </label>
            <label>
              Description
              <textarea name="description" rows={3} placeholder="Describe the activity…" />
            </label>
            <div className="field-row">
              <label>
                Points Awarded
                <input name="points" type="number" min={0} defaultValue={50} />
              </label>
              <label>
                Activity Date
                <input name="activityDate" type="date" />
              </label>
            </div>
            <label className="rowflex" style={{ textTransform: 'none', letterSpacing: 0, fontWeight: 500 }}>
              <input name="evidenceRequired" type="checkbox" defaultChecked /> Require evidence for participation
            </label>
            <ErrorMessage error={error} />
          </form>
        </Modal>
      )}
      {view && (
        <Modal
          title="Activity Details"
          onClose={() => setView(null)}
          footer={
            <>
              <Button className="secondary" onClick={() => setView(null)}>
                Close
              </Button>
              <Button
                className="primary"
                onClick={() => {
                  setJoin(view)
                  setView(null)
                }}
              >
                Join Activity
              </Button>
            </>
          }
        >
          <div className="stack">
            <div className="list-row">
              <span className="muted">Activity</span>
              <b>{view.title}</b>
            </div>
            <div className="list-row">
              <span className="muted">Points</span>
              <b>{view.points} pts</b>
            </div>
            <div className="list-row">
              <span className="muted">Evidence Required</span>
              {view.evidenceRequired ? <Pill status="warning">Yes</Pill> : <Pill status="neutral">No</Pill>}
            </div>
            <p className="muted">{view.description}</p>
          </div>
        </Modal>
      )}
      {join && (
        <Modal
          title="Join Activity"
          onClose={() => setJoin(null)}
          footer={
            <>
              <Button className="secondary" onClick={() => setJoin(null)}>
                Cancel
              </Button>
              <Button className="primary" form="join-activity" disabled={joinMut.isPending}>
                Confirm &amp; Submit
              </Button>
            </>
          }
        >
          {join.evidenceRequired && (
            <Note>
              <b>Evidence required.</b> Upload a proof photo or document before submitting.
            </Note>
          )}
          <p style={{ margin: '12px 0' }}>
            <b>{join.title}</b>
            <div className="muted">Confirm your participation. You will earn points after approval.</div>
          </p>
          <form id="join-activity" className="modal-form" onSubmit={submitJoin}>
            <ProofUploadField required={join.evidenceRequired} name="proofUrl" />
            <label>
              Notes (optional)
              <textarea name="notes" rows={2} />
            </label>
            <ErrorMessage error={error} />
          </form>
        </Modal>
      )}
    </>
  )
}

function ParticipationPanel() {
  const qc = useQueryClient()
  const config = useQuery({ queryKey: ['settings', 'config'], queryFn: api.settings.config })
  const list = useQuery({ queryKey: queryKeys.csrParticipations, queryFn: () => api.social.participations() })
  const approve = useMutation({
    mutationFn: api.social.approveParticipation,
    onSuccess: () => void qc.invalidateQueries({ queryKey: queryKeys.csrParticipations }),
  })
  const reject = useMutation({
    mutationFn: api.social.rejectParticipation,
    onSuccess: () => void qc.invalidateQueries({ queryKey: queryKeys.csrParticipations }),
  })
  const requireEvidence = config.data?.requireCsrEvidence ?? true

  return (
    <>
      <Note>
        With <b>&quot;Require evidence for all CSR activities&quot;</b> on, participation cannot be approved without an attached proof file.
      </Note>
      <div className="table-wrap" style={{ marginTop: 16 }}>
        <table>
          <thead>
            <tr>
              <th>Employee</th>
              <th>Activity</th>
              <th>Proof</th>
              <th className="numeric">Points</th>
              <th>Status</th>
              <th>Action</th>
            </tr>
          </thead>
          <tbody>
            {(list.data?.items ?? []).map((row) => {
              const evidenceNeeded = row.evidenceRequired || requireEvidence
              const canApprove = canApproveParticipation(row.proofUrl, evidenceNeeded)
              return (
                <ParticipationRow
                  key={row.id}
                  row={row}
                  canApprove={canApprove}
                  onApprove={() => void approve.mutateAsync(row.id)}
                  onReject={() => void reject.mutateAsync(row.id)}
                />
              )
            })}
          </tbody>
        </table>
        {(list.data?.items ?? []).length === 0 && !list.isLoading && (
          <EmptyState>
            <h3>No participations yet</h3>
            <p className="muted">Employees join activities from the CSR Activities tab.</p>
          </EmptyState>
        )}
      </div>
    </>
  )
}

function ParticipationRow({
  row,
  canApprove,
  onApprove,
  onReject,
}: {
  row: CSRParticipation
  canApprove: boolean
  onApprove(): void
  onReject(): void
}) {
  const blocked = row.approval === 'pending' && !canApprove
  return (
    <tr>
      <td>
        <span className="avatar-sm">{initials(row.employeeName ?? '?')}</span>
        {row.employeeName}
      </td>
      <td>{row.activityTitle}</td>
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
          <Pill status="danger">No proof</Pill>
        )}
        {row.approval === 'pending' ? <EvidenceAssist proofUrl={row.proofUrl} /> : null}
      </td>
      <td className="numeric">{row.approval === 'approved' ? row.pointsEarned : row.activityPoints ?? '—'}</td>
      <td>
        <Pill status={blocked ? 'blocked' : row.approval}>{blocked ? 'Blocked' : row.approval}</Pill>
      </td>
      <td>
        {row.approval === 'pending' ? (
          <RoleGuard roles={['admin', 'dept_head']}>
            <div className="rowflex">
              <Button className="primary sm" disabled={!canApprove || blocked} onClick={onApprove}>
                Approve
              </Button>
              <Button className="secondary sm" onClick={onReject}>
                Reject
              </Button>
            </div>
          </RoleGuard>
        ) : (
          <span className="muted">{row.approval === 'approved' ? 'Approved' : 'Rejected'}</span>
        )}
      </td>
    </tr>
  )
}

function TrainingPanel() {
  const qc = useQueryClient()
  const list = useQuery({ queryKey: queryKeys.trainings, queryFn: api.social.trainings })
  const create = useMutation({
    mutationFn: api.social.createTraining,
    onSuccess: () => void qc.invalidateQueries({ queryKey: queryKeys.trainings }),
  })
  const complete = useMutation({
    mutationFn: api.social.completeTraining,
    onSuccess: () => void qc.invalidateQueries({ queryKey: queryKeys.trainings }),
  })
  return (
    <>
      <div className="section-head">
        <h2>Training Completion</h2>
        <RoleGuard roles={['admin']}>
          <Button
            className="primary sm"
            onClick={() => void create.mutateAsync({ name: 'ESG Fundamentals', assignedTo: 'All employees' })}
          >
            + Seed Training
          </Button>
        </RoleGuard>
      </div>
      <div className="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Training</th>
              <th>Assigned To</th>
              <th className="numeric">Completed</th>
              <th className="numeric">Total</th>
              <th>Completion %</th>
              <th>Status</th>
              <th />
            </tr>
          </thead>
          <tbody>
            {(list.data?.items ?? []).map((t) => {
              const pct = t.total ? Math.round((t.completed / t.total) * 100) : 0
              return (
                <tr key={t.id}>
                  <td>
                    <b>{t.name}</b>
                  </td>
                  <td>{t.assignedTo}</td>
                  <td className="numeric">{t.completed}</td>
                  <td className="numeric">{t.total}</td>
                  <td>
                    <div className="rowflex">
                      <Progress value={pct} tone={pct < 85 ? 'warning' : undefined} />
                      <span className="muted">{pct}%</span>
                    </div>
                  </td>
                  <td>
                    <Pill status="active">{t.status}</Pill>
                  </td>
                  <td>
                    <RoleGuard roles={['employee', 'dept_head', 'admin']}>
                      <Button className="secondary sm" onClick={() => void complete.mutateAsync(t.id)}>
                        Mark complete
                      </Button>
                    </RoleGuard>
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
    </>
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
