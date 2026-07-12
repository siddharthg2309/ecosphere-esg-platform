import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useMemo, useState, type FormEvent } from 'react'
import { RoleGuard } from '../../app/RoleGuard'
import { Button, EmptyState, Modal, Note, Pill, Progress, StatBar, initials } from '../../design/components'
import { api } from '../../lib/apiClient'
import { queryKeys } from '../../lib/queryKeys'
import type { Audit, ComplianceIssue, Employee, IssueSeverity } from '../../lib/types'
import { userFacingError } from '../../lib/userFacingError'
import { useDepartmentsVM } from '../settings/useDepartmentsVM'
import { canSubmitIssue } from './raiseIssueValidation'

type Tab = 'policies' | 'acknowledgements' | 'audits' | 'issues'
const tabs: [Tab, string][] = [
  ['policies', 'Policies'],
  ['acknowledgements', 'Policy Acknowledgements'],
  ['audits', 'Audits'],
  ['issues', 'Compliance Issues'],
]

export function GovernancePage() {
  const [tab, setTab] = useState<Tab>('audits')
  const stats = useQuery({ queryKey: queryKeys.governanceStats, queryFn: api.governance.stats })
  const s = stats.data
  return (
    <main className="page">
      <div className="page-shell">
        <header className="page-head">
          <div>
            <p className="eyebrow">Governance</p>
            <h1>Policies, Audits &amp; Compliance</h1>
            <p className="muted">Governance policies · acknowledgements · audit trail · issue tracking</p>
          </div>
        </header>
        <StatBar
          items={[
            { label: 'Governance Score', value: s?.governanceScore ?? '—', unit: '/ 100', sub: 'org roll-up' },
            { label: 'Open Issues', value: s?.openIssues ?? 0, sub: `${s?.overdueIssues ?? 0} overdue` },
            { label: 'Policy Ack. Rate', value: '—', unit: '%', sub: 'see Policies tab' },
            { label: 'Audits (FY)', value: s?.auditsFY ?? 0, sub: 'this year' },
          ]}
        />
        <div className="tabs" role="tablist" aria-label="Governance sections">
          {tabs.map(([id, label]) => (
            <button key={id} role="tab" aria-selected={tab === id} className={`tab ${tab === id ? 'active' : ''}`} onClick={() => setTab(id)}>
              {label}
            </button>
          ))}
        </div>
        {tab === 'policies' && <PoliciesPanel />}
        {tab === 'acknowledgements' && <AcksPanel />}
        {tab === 'audits' && <AuditsPanel />}
        {tab === 'issues' && <IssuesPanel />}
      </div>
    </main>
  )
}

function PoliciesPanel() {
  const list = useQuery({ queryKey: queryKeys.governancePolicies, queryFn: api.governance.policies })
  return (
    <>
      <p className="muted" style={{ fontSize: 13, marginBottom: 14 }}>
        Active policies · version-controlled · ack rates from employee acknowledgements
      </p>
      <div className="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Policy</th>
              <th>Version</th>
              <th>Effective Date</th>
              <th>Ack. Rate</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            {(list.data?.items ?? []).map((p) => (
              <tr key={p.id}>
                <td>
                  <b>{p.title}</b>
                </td>
                <td>v{p.version}</td>
                <td>{String(p.effectiveDate).slice(0, 10)}</td>
                <td>
                  <div className="rowflex">
                    <Progress value={p.ackRate ?? 0} />
                    <span className="muted">{Math.round(p.ackRate ?? 0)}%</span>
                  </div>
                </td>
                <td>
                  <Pill status="active">Published</Pill>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {(list.data?.items ?? []).length === 0 && !list.isLoading && (
          <EmptyState>
            <h3>No policies</h3>
            <p className="muted">Publish policies from Settings → Master Data.</p>
          </EmptyState>
        )}
      </div>
    </>
  )
}

function AcksPanel() {
  const list = useQuery({ queryKey: queryKeys.acknowledgements, queryFn: api.governance.acknowledgements })
  const unacked = useQuery({ queryKey: queryKeys.unacknowledged, queryFn: api.governance.unacknowledged })
  return (
    <>
      {(unacked.data?.total ?? 0) > 0 && (
        <Note>
          <b>{unacked.data?.total}</b> policies still need your acknowledgement (see login modal / Policies).
        </Note>
      )}
      <div className="table-wrap" style={{ marginTop: 14 }}>
        <table>
          <thead>
            <tr>
              <th>Employee</th>
              <th>Department</th>
              <th>Policy</th>
              <th>Status</th>
              <th>Acknowledged At</th>
            </tr>
          </thead>
          <tbody>
            {(list.data?.items ?? []).map((a) => (
              <tr key={a.id}>
                <td>
                  <span className="avatar-sm">{initials(a.employeeName ?? '?')}</span>
                  {a.employeeName}
                </td>
                <td>{a.departmentName || '—'}</td>
                <td>
                  {a.policyTitle} v{a.version}
                </td>
                <td>
                  <Pill status="approved">Acknowledged</Pill>
                </td>
                <td>{String(a.acknowledgedAt).slice(0, 10)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </>
  )
}

function AuditsPanel() {
  const qc = useQueryClient()
  const depts = useDepartmentsVM()
  const employees = useQuery({ queryKey: queryKeys.employees, queryFn: () => api.master.list<Employee>('employees') })
  const list = useQuery({ queryKey: queryKeys.audits, queryFn: api.governance.audits })
  const [open, setOpen] = useState(false)
  const [error, setError] = useState('')
  const create = useMutation({
    mutationFn: api.governance.createAudit,
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: queryKeys.audits })
      void qc.invalidateQueries({ queryKey: queryKeys.governanceStats })
      setOpen(false)
    },
  })
  const auditors = (employees.data?.items ?? []).filter((e) => e.role === 'auditor' || e.role === 'admin')

  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setError('')
    const f = new FormData(e.currentTarget)
    try {
      await create.mutateAsync({
        title: String(f.get('title')),
        departmentId: String(f.get('departmentId')),
        auditorId: String(f.get('auditorId')),
        auditDate: String(f.get('auditDate')),
        findings: String(f.get('findings') || ''),
        status: String(f.get('status') || 'draft') as Audit['status'],
      })
    } catch (err) {
      setError(userFacingError(err))
    }
  }

  return (
    <>
      <div className="section-head">
        <h2 style={{ margin: 0 }}>Audit trail</h2>
        <RoleGuard roles={['auditor', 'admin']}>
          <Button className="primary sm" onClick={() => setOpen(true)}>
            + New Audit
          </Button>
        </RoleGuard>
      </div>
      <div className="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Audit</th>
              <th>Department</th>
              <th>Auditor</th>
              <th>Date</th>
              <th>Findings</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            {(list.data?.items ?? []).map((a) => (
              <tr key={a.id}>
                <td>
                  <b>{a.title}</b>
                </td>
                <td>{a.departmentName}</td>
                <td>
                  <span className="avatar-sm">{initials(a.auditorName ?? '?')}</span>
                  {a.auditorName}
                </td>
                <td>{String(a.auditDate).slice(0, 10)}</td>
                <td className="muted">{a.findings ? a.findings.slice(0, 60) : '—'}</td>
                <td>
                  <Pill status={a.status === 'under_review' ? 'under_review' : a.status === 'completed' ? 'completed' : 'draft'}>
                    {a.status.replace('_', ' ')}
                  </Pill>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {open && (
        <Modal
          title="New Audit"
          onClose={() => setOpen(false)}
          footer={
            <>
              <Button className="secondary" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <Button className="primary" form="create-audit" disabled={create.isPending}>
                Create Audit
              </Button>
            </>
          }
        >
          <form id="create-audit" className="modal-form" onSubmit={submit}>
            <label>
              Audit Title
              <input name="title" required placeholder="e.g. Q3 Waste Management Audit" />
            </label>
            <div className="field-row">
              <label>
                Department
                <select name="departmentId" required>
                  {depts.rows.map((d) => (
                    <option key={d.id} value={d.id}>
                      {d.name}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                Auditor
                <select name="auditorId" required>
                  {auditors.map((u) => (
                    <option key={u.id} value={u.id}>
                      {u.name}
                    </option>
                  ))}
                </select>
              </label>
            </div>
            <label>
              Audit Date
              <input name="auditDate" type="date" required />
            </label>
            <label>
              Scope / Findings
              <textarea name="findings" rows={3} />
            </label>
            <label>
              Status
              <select name="status" defaultValue="draft">
                <option value="draft">Draft</option>
                <option value="under_review">Under Review</option>
                <option value="completed">Completed</option>
              </select>
            </label>
            {error && (
              <div className="form-error" role="alert">
                {error}
              </div>
            )}
          </form>
        </Modal>
      )}
    </>
  )
}

function IssuesPanel() {
  const qc = useQueryClient()
  const depts = useDepartmentsVM()
  const employees = useQuery({ queryKey: queryKeys.employees, queryFn: () => api.master.list<Employee>('employees') })
  const audits = useQuery({ queryKey: queryKeys.audits, queryFn: api.governance.audits })
  const list = useQuery({ queryKey: queryKeys.complianceIssues, queryFn: () => api.governance.issues() })
  const [open, setOpen] = useState(false)
  const [form, setForm] = useState({
    description: '',
    departmentId: '',
    ownerId: '',
    dueDate: '',
    severity: 'high' as IssueSeverity,
    auditId: '',
  })
  const [error, setError] = useState('')
  const raise = useMutation({
    mutationFn: api.governance.raiseIssue,
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: queryKeys.complianceIssues })
      void qc.invalidateQueries({ queryKey: queryKeys.governanceStats })
      void qc.invalidateQueries({ queryKey: queryKeys.notifications })
      setOpen(false)
      setForm({ description: '', departmentId: '', ownerId: '', dueDate: '', severity: 'high', auditId: '' })
    },
  })
  const update = useMutation({
    mutationFn: ({ id, status }: { id: string; status: ComplianceIssue['status'] }) => api.governance.updateIssue(id, status),
    onSuccess: () => void qc.invalidateQueries({ queryKey: queryKeys.complianceIssues }),
  })

  const valid = useMemo(() => canSubmitIssue(form), [form])
  const owners = employees.data?.items ?? []

  return (
    <>
      <div className="section-head">
        <h2 style={{ margin: 0 }}>
          Compliance issues <span className="muted" style={{ fontSize: 13, fontWeight: 400 }}>· owner + due date required · overdue auto-flagged</span>
        </h2>
        <RoleGuard roles={['auditor', 'dept_head', 'admin']}>
          <Button className="primary sm" onClick={() => setOpen(true)}>
            + Raise Issue
          </Button>
        </RoleGuard>
      </div>
      <div className="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Issue</th>
              <th>Department</th>
              <th>Severity</th>
              <th>Owner</th>
              <th>Due Date</th>
              <th>Status</th>
              <th />
            </tr>
          </thead>
          <tbody>
            {(list.data?.items ?? []).map((i) => (
              <tr key={i.id}>
                <td>
                  <b>{i.description.slice(0, 48)}</b>
                  {i.auditTitle && <div className="muted">from {i.auditTitle}</div>}
                </td>
                <td>{i.departmentName}</td>
                <td>
                  <Pill status={i.severity === 'high' ? 'danger' : i.severity === 'medium' ? 'warning' : 'neutral'}>{i.severity}</Pill>
                </td>
                <td>
                  <span className="avatar-sm">{initials(i.ownerName ?? '?')}</span>
                  {i.ownerName}
                </td>
                <td style={i.overdue ? { color: 'var(--danger)', fontWeight: 600 } : undefined}>
                  {String(i.dueDate).slice(0, 10)}
                  {i.overdue ? ' · overdue' : ''}
                </td>
                <td>
                  <Pill status={i.status === 'open' && i.overdue ? 'danger' : i.status === 'resolved' ? 'approved' : 'pending'}>
                    {i.status.replace('_', ' ')}
                  </Pill>
                </td>
                <td>
                  {i.status !== 'resolved' && (
                    <RoleGuard roles={['auditor', 'dept_head', 'admin']}>
                      <Button className="secondary sm" onClick={() => void update.mutateAsync({ id: i.id, status: 'resolved' })}>
                        Resolve
                      </Button>
                    </RoleGuard>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {open && (
        <Modal
          title="Raise Compliance Issue"
          onClose={() => setOpen(false)}
          footer={
            <>
              <Button className="secondary" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <Button
                className="primary"
                disabled={!valid || raise.isPending}
                onClick={() => {
                  setError('')
                  void raise
                    .mutateAsync({
                      description: form.description,
                      departmentId: form.departmentId,
                      ownerId: form.ownerId,
                      dueDate: form.dueDate,
                      severity: form.severity,
                      auditId: form.auditId || undefined,
                    })
                    .catch((e) => setError(userFacingError(e)))
                }}
              >
                Raise Issue &amp; Notify Owner
              </Button>
            </>
          }
        >
          <Note>Every issue <b>must have an owner and a due date</b>. Overdue-open issues are automatically flagged and trigger notifications.</Note>
          <form className="modal-form" style={{ marginTop: 14 }} onSubmit={(e) => e.preventDefault()}>
            <label>
              Issue Description
              <textarea
                required
                rows={3}
                value={form.description}
                onChange={(e) => setForm({ ...form, description: e.target.value })}
                placeholder="Describe the compliance issue clearly…"
              />
            </label>
            <div className="field-row">
              <label>
                Department
                <select required value={form.departmentId} onChange={(e) => setForm({ ...form, departmentId: e.target.value })}>
                  <option value="">Select…</option>
                  {depts.rows.map((d) => (
                    <option key={d.id} value={d.id}>
                      {d.name}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                Severity
                <select value={form.severity} onChange={(e) => setForm({ ...form, severity: e.target.value as IssueSeverity })}>
                  <option value="high">High</option>
                  <option value="medium">Medium</option>
                  <option value="low">Low</option>
                </select>
              </label>
            </div>
            <div className="field-row">
              <label>
                Owner (required)
                <select required value={form.ownerId} onChange={(e) => setForm({ ...form, ownerId: e.target.value })}>
                  <option value="">Select…</option>
                  {owners.map((u) => (
                    <option key={u.id} value={u.id}>
                      {u.name}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                Due Date (required)
                <input type="date" required value={form.dueDate} onChange={(e) => setForm({ ...form, dueDate: e.target.value })} />
              </label>
            </div>
            <label>
              Linked Audit (optional)
              <select value={form.auditId} onChange={(e) => setForm({ ...form, auditId: e.target.value })}>
                <option value="">None</option>
                {(audits.data?.items ?? []).map((a) => (
                  <option key={a.id} value={a.id}>
                    {a.title}
                  </option>
                ))}
              </select>
            </label>
            {!valid && <p className="muted compact">Submit stays disabled until description, department, owner, and due date are set.</p>}
            {error && (
              <div className="form-error" role="alert">
                {error}
              </div>
            )}
          </form>
        </Modal>
      )}
    </>
  )
}
