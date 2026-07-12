import { useMutation } from '@tanstack/react-query'
import { useState } from 'react'
import { Button, Card, Note, Pill } from '../../design/components'
import { api, downloadReport } from '../../lib/apiClient'
import type { Report, ReportType } from '../../lib/types'
import { userFacingError } from '../../lib/userFacingError'
import { useDepartmentsVM } from '../settings/useDepartmentsVM'

const standard: { type: ReportType; title: string; blurb: string }[] = [
  { type: 'environmental', title: 'Environmental', blurb: 'Emissions, goals, source breakdown.' },
  { type: 'social', title: 'Social', blurb: 'Diversity, CSR participation, training.' },
  { type: 'governance', title: 'Governance', blurb: 'Policies, audits, compliance & risk.' },
  { type: 'esg_summary', title: 'ESG Summary', blurb: 'Executive overview + AI narrative.' },
]

export function ReportsPage() {
  const depts = useDepartmentsVM()
  const [preview, setPreview] = useState<Report | null>(null)
  const [error, setError] = useState('')
  const [filters, setFilters] = useState({
    departmentId: '',
    from: '',
    to: '',
    module: '',
    employee: '',
    challenge: '',
    category: '',
  })

  const generate = useMutation({
    mutationFn: api.reports.generate,
    onSuccess: (r) => {
      setPreview(r)
      setError('')
    },
    onError: (e) => setError(userFacingError(e, 'Unable to generate report')),
  })

  function run(type: ReportType) {
    const extra: Record<string, unknown> = {}
    if (filters.module) extra.module = filters.module
    if (filters.employee) extra.employee = filters.employee
    if (filters.challenge) extra.challenge = filters.challenge
    if (filters.category) extra.category = filters.category
    void generate.mutateAsync({
      type,
      departmentId: filters.departmentId || undefined,
      from: filters.from || undefined,
      to: filters.to || undefined,
      filters: extra,
    })
  }

  return (
    <main className="page">
      <div className="content">
        <header className="page-head">
          <div>
            <p className="eyebrow">Reports</p>
            <h1>Reports &amp; Analytics</h1>
            <p className="muted">Generate ESG reports · build custom views · export PDF / Excel / CSV</p>
          </div>
        </header>

        <div className="grid cols-2" style={{ gridTemplateColumns: 'repeat(auto-fit,minmax(220px,1fr))', gap: 16 }}>
          {standard.map((r) => (
            <Card key={r.type}>
              <div className="card-head">
                <h3>{r.title}</h3>
              </div>
              <div className="muted">{r.blurb}</div>
              <div className="card-foot">
                <span className="muted">Standard</span>
                <Button className="primary sm" disabled={generate.isPending} onClick={() => run(r.type)}>
                  Generate
                </Button>
              </div>
            </Card>
          ))}
        </div>

        <Card style={{ marginTop: 24 }}>
          <div className="card-head">
            <h3>Custom report builder</h3>
            <span className="muted">6 filters</span>
          </div>
          <div className="modal-form">
            <div className="field-row">
              <label>
                Department
                <select value={filters.departmentId} onChange={(e) => setFilters({ ...filters, departmentId: e.target.value })}>
                  <option value="">All departments</option>
                  {depts.rows.map((d) => (
                    <option key={d.id} value={d.id}>
                      {d.name}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                Module
                <select value={filters.module} onChange={(e) => setFilters({ ...filters, module: e.target.value })}>
                  <option value="">All modules</option>
                  <option value="environmental">environmental</option>
                  <option value="social">social</option>
                  <option value="governance">governance</option>
                </select>
              </label>
            </div>
            <div className="field-row">
              <label>
                From
                <input type="date" value={filters.from} onChange={(e) => setFilters({ ...filters, from: e.target.value })} />
              </label>
              <label>
                To
                <input type="date" value={filters.to} onChange={(e) => setFilters({ ...filters, to: e.target.value })} />
              </label>
            </div>
            <div className="field-row">
              <label>
                Employee
                <input
                  value={filters.employee}
                  onChange={(e) => setFilters({ ...filters, employee: e.target.value })}
                  placeholder="name or id"
                />
              </label>
              <label>
                Challenge
                <input
                  value={filters.challenge}
                  onChange={(e) => setFilters({ ...filters, challenge: e.target.value })}
                  placeholder="challenge filter"
                />
              </label>
              <label>
                Category
                <input
                  value={filters.category}
                  onChange={(e) => setFilters({ ...filters, category: e.target.value })}
                  placeholder="category filter"
                />
              </label>
            </div>
            <div className="rowflex">
              <Button className="primary" disabled={generate.isPending} onClick={() => run('custom')}>
                Generate custom
              </Button>
            </div>
          </div>
        </Card>

        {error && (
          <div className="form-error" role="alert" style={{ marginTop: 16 }}>
            {error}
          </div>
        )}

        {preview && (
          <Card style={{ marginTop: 24 }}>
            <div className="card-head">
              <h3>Report preview</h3>
              <div className="rowflex">
                <Pill status="plum">{preview.type}</Pill>
                <Button className="secondary sm" onClick={() => void downloadReport(preview.id, 'pdf')}>
                  PDF
                </Button>
                <Button className="secondary sm" onClick={() => void downloadReport(preview.id, 'xlsx')}>
                  Excel
                </Button>
                <Button className="secondary sm" onClick={() => void downloadReport(preview.id, 'csv')}>
                  CSV
                </Button>
              </div>
            </div>
            <p className="muted" style={{ fontSize: 12 }}>
              Generated {String(preview.generatedAt).slice(0, 19).replace('T', ' ')} · id {preview.id.slice(0, 8)}…
            </p>
            {preview.sections.map((sec, i) => (
              <div key={i} style={{ marginTop: 16 }}>
                <h3 style={{ margin: '0 0 6px' }}>
                  {sec.title} {sec.ai && <Pill status="warning">AI advisory</Pill>}
                </h3>
                {sec.ai && (
                  <Note>
                    This section is AI-generated and advisory. All numeric scores remain system-calculated.
                  </Note>
                )}
                {sec.summary && <p style={{ marginTop: 8 }}>{sec.summary}</p>}
                {sec.rows && sec.rows.length > 0 && (
                  <div className="table-wrap" style={{ marginTop: 8 }}>
                    <table>
                      <thead>
                        <tr>
                          {Object.keys(sec.rows[0]).map((k) => (
                            <th key={k}>{k}</th>
                          ))}
                        </tr>
                      </thead>
                      <tbody>
                        {sec.rows.map((row, ri) => (
                          <tr key={ri}>
                            {Object.values(row).map((v, vi) => (
                              <td key={vi}>{v}</td>
                            ))}
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </div>
            ))}
          </Card>
        )}
      </div>
    </main>
  )
}
