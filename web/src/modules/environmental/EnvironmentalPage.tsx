import { useEffect, useState, type FormEvent } from 'react'
import { Button, EmptyState } from '../../design/components'
import { RequestError } from '../../lib/apiClient'
import type { CarbonSource, EmissionFactor, EnvironmentalGoal, GoalStatus } from '../../lib/types'
import { userFacingError } from '../../lib/userFacingError'
import { co2Preview, useEnvironmentalVM } from './useEnvironmentalVM'

type Tab = 'transactions' | 'goals' | 'factors' | 'products'
const sourceLabels: Record<CarbonSource, string> = {
  purchase: 'Purchase / Energy',
  manufacturing: 'Manufacturing',
  expense: 'Expense / Travel',
  fleet: 'Fleet',
}
const statusLabels: Record<GoalStatus, string> = {
  on_track: 'On track',
  at_risk: 'At risk',
  completed: 'Completed',
}

export function EnvironmentalPage() {
  const vm = useEnvironmentalVM()
  const [tab, setTab] = useState<Tab>('transactions')
  const [entry, setEntry] = useState(false)
  const [goal, setGoal] = useState(false)
  const total = Number(vm.summary?.total ?? 0)

  return (
    <main className="page environmental-page">
      <section className="content">
        <header className="page-head">
          <div>
            <p className="eyebrow env-text">Environmental</p>
            <h1 className="page-title">Emission Tracking &amp; Goals</h1>
          </div>
          <div className="environmental-actions">
            {vm.user?.role !== 'auditor' && (
              <Button className="secondary" type="button" onClick={() => setEntry(true)}>
                Log carbon data
              </Button>
            )}
            {(vm.user?.role === 'admin' || vm.user?.role === 'dept_head') && (
              <Button className="primary" type="button" onClick={() => setGoal(true)}>
                New goal
              </Button>
            )}
          </div>
        </header>

        <div className="env-kpis">
          <KPI label="Verified emissions (YTD)" value={`${total.toLocaleString()} kg CO₂`} />
          <KPI label="Active goals" value={String(vm.goals.filter((v) => v.status !== 'completed').length)} />
          <KPI label="Emission factors" value={String(vm.factors.length)} />
          <KPI label="Auto calculation" value="Enabled" />
        </div>

        <div className="tabs" role="tablist">
          {(
            [
              ['transactions', 'Carbon Transactions'],
              ['goals', 'Environmental Goals'],
              ['factors', 'Emission Factors'],
              ['products', 'Product ESG Profiles'],
            ] as [Tab, string][]
          ).map(([id, label]) => (
            <button
              key={id}
              type="button"
              role="tab"
              aria-selected={tab === id}
              className={`tab ${tab === id ? 'active' : ''}`}
              onClick={() => setTab(id)}
            >
              {label}
            </button>
          ))}
        </div>

        {tab === 'transactions' && <Transactions vm={vm} />}
        {tab === 'goals' && <Goals goals={vm.goals} departments={vm.departments} />}
        {tab === 'factors' && <Factors factors={vm.factors} />}
        {tab === 'products' && (
          <EmptyState>
            <h3>Product ESG profiles are managed in Settings</h3>
            
          </EmptyState>
        )}

        <SourceSummary summary={vm.summary?.bySource ?? {}} />

        {entry && (
          <CarbonModal
            vm={vm}
            onClose={() => {
              setEntry(false)
              vm.resetIngest()
            }}
            onSubmitted={() => {
              setEntry(false)
              vm.resetIngest()
              setTab('transactions')
            }}
          />
        )}
        {goal && <GoalModal vm={vm} onClose={() => setGoal(false)} />}
      </section>
    </main>
  )
}

function KPI({ label, value }: { label: string; value: string }) {
  return (
    <article className="env-kpi">
      <span>{label}</span>
      <strong>{value}</strong>
    </article>
  )
}

function Transactions({ vm }: { vm: ReturnType<typeof useEnvironmentalVM> }) {
  if (vm.loading) return <div className="center-state">Loading transactions...</div>
  if (!vm.transactions.length) {
    return (
      <EmptyState>
        <h3>No carbon transactions yet</h3>
        
      </EmptyState>
    )
  }
  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Date</th>
            <th>Source</th>
            <th>Quantity</th>
            <th className="numeric">CO₂ (kg)</th>
            <th>Status</th>
            <th>Action</th>
          </tr>
        </thead>
        <tbody>
          {vm.transactions.map((row) => (
            <tr key={row.id}>
              <td>{row.txnDate.slice(0, 10)}</td>
              <td>{sourceLabels[row.source]}</td>
              <td>{row.quantity}</td>
              <td className="numeric">
                <strong>{row.status === 'verified' ? Number(row.computedCo2).toLocaleString() : 'Pending'}</strong>
              </td>
              <td>
                <span className={`env-pill ${row.status}`}>{row.status}</span>
              </td>
              <td>
                {vm.user?.role === 'dept_head' && row.status === 'draft' ? (
                  <Button className="ghost sm" type="button" onClick={() => void vm.verify(row.id)}>
                    Verify
                  </Button>
                ) : (
                  '—'
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function Goals({
  goals,
  departments,
}: {
  goals: EnvironmentalGoal[]
  departments: { id: string; name: string }[]
}) {
  if (!goals.length) {
    return (
      <EmptyState>
        <h3>No environmental goals yet</h3>
        
      </EmptyState>
    )
  }
  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Goal</th>
            <th>Department</th>
            <th>Target</th>
            <th>Current</th>
            <th>Progress</th>
            <th>Deadline</th>
            <th>Status</th>
          </tr>
        </thead>
        <tbody>
          {goals.map((row) => {
            const progress = Math.min(
              100,
              (Number(row.currentCo2) / Math.max(Number(row.targetCo2), 0.001)) * 100,
            )
            return (
              <tr key={row.id}>
                <td>
                  <strong>{row.name}</strong>
                </td>
                <td>{departments.find((v) => v.id === row.departmentId)?.name ?? 'Department'}</td>
                <td>{row.targetCo2} kg</td>
                <td>{row.currentCo2} kg</td>
                <td>
                  <div className="goal-progress" aria-label={`${progress.toFixed(0)} percent`}>
                    <span style={{ width: `${progress}%` }} />
                  </div>
                </td>
                <td>{row.deadline.slice(0, 10)}</td>
                <td>
                  <span className={`env-pill ${row.status}`}>{statusLabels[row.status]}</span>
                </td>
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}

function Factors({ factors }: { factors: EmissionFactor[] }) {
  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Factor</th>
            <th>Unit</th>
            <th className="numeric">kg CO₂ / unit</th>
            <th>Status</th>
          </tr>
        </thead>
        <tbody>
          {factors.map((row) => (
            <tr key={row.id}>
              <td>
                <strong>{row.name}</strong>
              </td>
              <td>{row.unit}</td>
              <td className="numeric">{row.kgCo2PerUnit}</td>
              <td>
                <span className="env-pill verified">Active</span>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function SourceSummary({ summary }: { summary: Partial<Record<CarbonSource, string>> }) {
  const values = (Object.keys(sourceLabels) as CarbonSource[]).map(
    (source) => [source, Number(summary[source] ?? 0)] as const,
  )
  const max = Math.max(...values.map((v) => v[1]), 1)
  return (
    <section className="source-card">
      <h2>Verified emissions by source</h2>
      {values.map(([source, value]) => (
        <div className="source-row" key={source}>
          <span>{sourceLabels[source]}</span>
          <div>
            <i style={{ width: `${(value / max) * 100}%` }} />
          </div>
          <strong>{value.toLocaleString()} kg</strong>
        </div>
      ))}
    </section>
  )
}

function CarbonModal({
  vm,
  onClose,
  onSubmitted,
}: {
  vm: ReturnType<typeof useEnvironmentalVM>
  onClose(): void
  onSubmitted(): void
}) {
  const [factorID, setFactorID] = useState('')
  const [quantity, setQuantity] = useState('')
  const [source, setSource] = useState<CarbonSource>('expense')
  const [txnDate, setTxnDate] = useState(() => new Date().toISOString().slice(0, 10))
  const [fileName, setFileName] = useState('')
  const [localError, setLocalError] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [uploading, setUploading] = useState(false)

  const factor = vm.factors.find((v) => v.id === factorID)
  const departmentID =
    vm.user?.role === 'admin'
      ? (vm.user.departmentId || vm.departments[0]?.id || '')
      : (vm.user?.departmentId ?? '')

  // Prefill first factor once loaded
  useEffect(() => {
    if (!factorID && vm.factors.length > 0) setFactorID(vm.factors[0].id)
  }, [vm.factors, factorID])

  // Apply AI suggestion when available
  useEffect(() => {
    const suggested = vm.suggestion
    if (!suggested) return
    if (suggested.quantity) setQuantity(String(suggested.quantity))
    if (['purchase', 'manufacturing', 'expense', 'fleet'].includes(suggested.source)) {
      setSource(suggested.source as CarbonSource)
    }
    // match unit to a factor when possible
    if (suggested.unit) {
      const match = vm.factors.find((f) => f.unit.toLowerCase() === suggested.unit.toLowerCase())
      if (match) setFactorID(match.id)
    }
  }, [vm.suggestion, vm.factors])

  async function upload(file?: File) {
    if (!file) return
    setLocalError('')
    setFileName(file.name)
    setUploading(true)
    try {
      await vm.ingestFile(file)
    } catch (err) {
      // Prototype: keep the file name even if AI ingest fails
      setLocalError(
        userFacingError(err, 'Could not auto-categorize document — fill source/quantity manually, then submit.'),
      )
    } finally {
      setUploading(false)
    }
  }

  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setLocalError('')

    if (!departmentID) {
      setLocalError('No department assigned to your account. Ask an admin to set your department.')
      return
    }
    if (!factorID || !factor) {
      setLocalError('Select an emission factor.')
      return
    }
    if (!quantity || Number(quantity) <= 0) {
      setLocalError('Enter a quantity greater than zero.')
      return
    }
    if (!txnDate) {
      setLocalError('Choose a transaction date.')
      return
    }

    setSubmitting(true)
    try {
      // evidenceUrl: use AI upload path, or a prototype placeholder so submit always works
      const evidenceUrl =
        vm.suggestion?.evidenceUrl ||
        (fileName ? `prototype://evidence/${encodeURIComponent(fileName)}` : undefined)

      await vm.createDraft({
        departmentId: departmentID,
        source,
        quantity: String(quantity),
        emissionFactorId: factor.id,
        unit: factor.unit,
        txnDate,
        evidenceUrl,
      })
      onSubmitted()
    } catch (err) {
      setLocalError(userFacingError(err, 'Unable to submit carbon transaction for verification'))
    } finally {
      setSubmitting(false)
    }
  }

  const busy = submitting || uploading || vm.state === 'uploading' || vm.state === 'submitting'
  const preview = co2Preview(quantity, factor?.kgCo2PerUnit ?? 0)

  return (
    <div className="modal-backdrop" role="presentation" onMouseDown={onClose}>
      <section
        className="modal"
        role="dialog"
        aria-modal="true"
        aria-labelledby="carbon-modal-title"
        onMouseDown={(e) => e.stopPropagation()}
      >
        <header className="modal-head">
          <h2 id="carbon-modal-title">Log carbon data</h2>
          <button className="close-button" type="button" onClick={onClose} aria-label="Close">
            ×
          </button>
        </header>

        <form className="modal-form" onSubmit={(e) => void submit(e)}>

          <label>
            Operational document
            <input
              type="file"
              accept="application/pdf,image/jpeg,image/png,.pdf,.jpg,.jpeg,.png"
              disabled={busy}
              onChange={(e) => void upload(e.target.files?.[0] ?? undefined)}
            />
            {fileName && (
              <span className="muted" style={{ fontSize: 12, marginTop: 4 }}>
                {uploading ? 'Uploading…' : `Attached: ${fileName}`}
                {vm.suggestion?.evidenceUrl ? ' · stored for evidence' : ''}
              </span>
            )}
          </label>

          <div className="field-row">
            <label>
              Source
              <select value={source} onChange={(e) => setSource(e.target.value as CarbonSource)} disabled={busy}>
                {Object.entries(sourceLabels).map(([id, label]) => (
                  <option key={id} value={id}>
                    {label}
                  </option>
                ))}
              </select>
            </label>
            <label>
              Emission factor
              <select
                required
                value={factorID}
                onChange={(e) => setFactorID(e.target.value)}
                disabled={busy || vm.factors.length === 0}
              >
                <option value="">Select factor</option>
                {vm.factors.map((v) => (
                  <option value={v.id} key={v.id}>
                    {v.name} · {v.unit}
                  </option>
                ))}
              </select>
            </label>
          </div>

          <div className="field-row">
            <label>
              Quantity
              <input
                required
                type="number"
                min="0.001"
                step="0.001"
                value={quantity}
                onChange={(e) => setQuantity(e.target.value)}
                disabled={busy}
                placeholder="e.g. 6"
              />
            </label>
            <label>
              Transaction date
              <input
                required
                type="date"
                value={txnDate}
                onChange={(e) => setTxnDate(e.target.value)}
                disabled={busy}
              />
            </label>
          </div>

          <div className="co2-preview">
            <span>CO₂ preview</span>
            <strong>{preview} kg</strong>
          </div>

          {(localError || vm.error) && (
            <div className="form-error" role="alert">
              {localError ||
                (vm.error instanceof RequestError
                  ? vm.error.body.message
                  : userFacingError(vm.error, 'Unable to complete the request'))}
            </div>
          )}

          <footer className="modal-foot" style={{ border: 0, paddingTop: 0, marginTop: 4 }}>
            <Button type="button" className="secondary" onClick={onClose} disabled={submitting}>
              Cancel
            </Button>
            <Button type="submit" className="primary" disabled={busy || !factorID || !quantity}>
              {submitting || vm.state === 'submitting' ? 'Submitting…' : 'Submit for verification'}
            </Button>
          </footer>
        </form>
      </section>
    </div>
  )
}

function GoalModal({
  vm,
  onClose,
}: {
  vm: ReturnType<typeof useEnvironmentalVM>
  onClose(): void
}) {
  const [error, setError] = useState('')
  const [pending, setPending] = useState(false)

  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setError('')
    setPending(true)
    const f = new FormData(e.currentTarget)
    try {
      await vm.createGoal({
        name: String(f.get('name')),
        departmentId: String(f.get('departmentId')),
        targetCo2: String(f.get('targetCo2')),
        currentCo2: String(f.get('currentCo2') || '0'),
        deadline: String(f.get('deadline')),
      })
      onClose()
    } catch (err) {
      setError(userFacingError(err, 'Unable to create goal'))
    } finally {
      setPending(false)
    }
  }

  return (
    <div className="modal-backdrop" role="presentation" onMouseDown={onClose}>
      <section className="modal" role="dialog" aria-modal="true" onMouseDown={(e) => e.stopPropagation()}>
        <header className="modal-head">
          <h2>New environmental goal</h2>
          <button className="close-button" type="button" onClick={onClose} aria-label="Close">
            ×
          </button>
        </header>
        <form className="modal-form" onSubmit={(e) => void submit(e)}>
          <label>
            Goal name
            <input name="name" required />
          </label>
          <label>
            Department
            <select name="departmentId" required defaultValue={vm.user?.departmentId ?? ''}>
              {vm.departments.map((v) => (
                <option key={v.id} value={v.id}>
                  {v.name}
                </option>
              ))}
            </select>
          </label>
          <div className="field-row">
            <label>
              Target CO₂ (kg)
              <input name="targetCo2" type="number" min="0.001" step="0.001" required />
            </label>
            <label>
              Current CO₂ (kg)
              <input name="currentCo2" type="number" min="0" step="0.001" defaultValue="0" />
            </label>
          </div>
          <label>
            Deadline
            <input name="deadline" type="date" required />
          </label>
          {error && (
            <div className="form-error" role="alert">
              {error}
            </div>
          )}
          <footer className="modal-foot" style={{ border: 0, paddingTop: 0 }}>
            <Button type="button" className="secondary" onClick={onClose}>
              Cancel
            </Button>
            <Button className="primary" type="submit" disabled={pending}>
              {pending ? 'Creating…' : 'Create goal'}
            </Button>
          </footer>
        </form>
      </section>
    </div>
  )
}
