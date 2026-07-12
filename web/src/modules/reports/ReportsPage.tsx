import { useState } from 'react'
import { Button, Card, Note, Pill } from '../../design/components'
import { api, downloadReport } from '../../lib/apiClient'
import type { Report, ReportType } from '../../lib/types'
import { userFacingError } from '../../lib/userFacingError'
import { useDepartmentsVM } from '../settings/useDepartmentsVM'
import { buildMockReport, downloadTextFile, mockReportToCsv } from './mockReports'

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
  const [pending, setPending] = useState(false)
  const [usingMock, setUsingMock] = useState(false)
  const [filters, setFilters] = useState({
    departmentId: '',
    from: '',
    to: '',
    module: '',
    employee: '',
    challenge: '',
    category: '',
  })

  async function run(type: ReportType) {
    setPending(true)
    setError('')
    const extra: Record<string, unknown> = {}
    if (filters.module) extra.module = filters.module
    if (filters.employee) extra.employee = filters.employee
    if (filters.challenge) extra.challenge = filters.challenge
    if (filters.category) extra.category = filters.category
    if (filters.departmentId) extra.departmentId = filters.departmentId
    if (filters.from) extra.from = filters.from
    if (filters.to) extra.to = filters.to

    const mock = buildMockReport(type, extra)

    try {
      // Prefer live API when available; still show rich prototype content if API is sparse.
      const live = await api.reports.generate({
        type,
        departmentId: filters.departmentId || undefined,
        from: filters.from || undefined,
        to: filters.to || undefined,
        filters: extra,
      })
      const liveHasBody =
        (live.sections?.length ?? 0) > 0 &&
        live.sections.some((s) => Boolean(s.summary) || (s.rows?.length ?? 0) > 0 || Boolean(s.metrics))

      if (liveHasBody && (live.sections?.length ?? 0) >= 2) {
        setPreview(live)
        setUsingMock(false)
      } else {
        // Prototype: merge API id for export attempts, keep rich mock sections.
        setPreview({ ...mock, id: live.id || mock.id, generatedAt: live.generatedAt || mock.generatedAt })
        setUsingMock(true)
      }
    } catch {
      // Offline / prototype fallback — always show a full mock report.
      setPreview(mock)
      setUsingMock(true)
    } finally {
      setPending(false)
      // scroll preview into view after paint
      requestAnimationFrame(() => {
        document.getElementById('report-preview')?.scrollIntoView({ behavior: 'smooth', block: 'start' })
      })
    }
  }

  async function exportFmt(fmt: 'pdf' | 'xlsx' | 'csv') {
    if (!preview) return
    setError('')
    try {
      if (!usingMock) {
        await downloadReport(preview.id, fmt)
        return
      }
      // Client-side export for mock prototypes
      if (fmt === 'csv') {
        downloadTextFile(`${preview.type}-report.csv`, mockReportToCsv(preview), 'text/csv;charset=utf-8')
        return
      }
      // PDF/Excel: download a readable text/json package for the prototype
      const body = JSON.stringify(preview, null, 2)
      downloadTextFile(
        `${preview.type}-report.${fmt === 'pdf' ? 'json' : 'json'}`,
        body,
        'application/json',
      )
    } catch (e) {
      // last resort mock csv
      try {
        downloadTextFile(`${preview.type}-report.csv`, mockReportToCsv(preview), 'text/csv;charset=utf-8')
      } catch {
        setError(userFacingError(e, 'Unable to export report'))
      }
    }
  }

  return (
    <main className="page">
      <div className="content">
        <header className="page-head">
          <div>
            <p className="eyebrow">Reports</p>
            <h1 className="page-title">Reports &amp; Analytics</h1>
          </div>
        </header>

        <div className="report-cards">
          {standard.map((r) => (
            <article className="report-card" key={r.type}>
              <div className="report-card-body">
                <h3>{r.title}</h3>
              </div>
              <div className="report-card-foot">
                <span />
                <Button
                  className="primary sm"
                  type="button"
                  disabled={pending}
                  onClick={() => void run(r.type)}
                >
                  {pending ? 'Generating…' : 'Generate'}
                </Button>
              </div>
            </article>
          ))}
        </div>

        <Card className="section-block report-builder">
          <div className="card-head">
            <h3>Custom report builder</h3>
          </div>
          <div className="modal-form">
            <div className="field-row">
              <label>
                Department
                <select
                  value={filters.departmentId}
                  onChange={(e) => setFilters({ ...filters, departmentId: e.target.value })}
                >
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
                  <option value="environmental">Environmental</option>
                  <option value="social">Social</option>
                  <option value="governance">Governance</option>
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
            </div>
            <label>
              Category
              <input
                value={filters.category}
                onChange={(e) => setFilters({ ...filters, category: e.target.value })}
                placeholder="category filter"
              />
            </label>
            <div className="rowflex">
              <Button className="primary" type="button" disabled={pending} onClick={() => void run('custom')}>
                {pending ? 'Generating…' : 'Generate custom'}
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
          <section id="report-preview" className="report-preview section-block">
            <div className="card-head">
              <div>
                <h3 style={{ margin: 0 }}>Report preview</h3>
                <p className="muted" style={{ fontSize: 12, margin: '4px 0 0' }}>
                  Generated {String(preview.generatedAt).slice(0, 19).replace('T', ' ')}
                  {usingMock ? ' · prototype data' : ` · id ${preview.id.slice(0, 8)}…`}
                </p>
              </div>
              <div className="rowflex">
                <Pill status="plum">{preview.type.replace(/_/g, ' ')}</Pill>
                <Button className="secondary sm" type="button" onClick={() => void exportFmt('pdf')}>
                  PDF
                </Button>
                <Button className="secondary sm" type="button" onClick={() => void exportFmt('xlsx')}>
                  Excel
                </Button>
                <Button className="secondary sm" type="button" onClick={() => void exportFmt('csv')}>
                  CSV
                </Button>
              </div>
            </div>


            {preview.sections.map((sec, i) => (
              <div key={`${sec.title}-${i}`} className="report-section">
                <h3>
                  {sec.title} {sec.ai ? <Pill status="warning">AI advisory</Pill> : null}
                </h3>
                {sec.ai && (
                  <Note>This section is AI-generated and advisory. Numeric scores remain system-calculated.</Note>
                )}
                {sec.summary && <p className="report-summary">{sec.summary}</p>}
                {sec.metrics && Object.keys(sec.metrics).length > 0 && (
                  <div className="report-metrics">
                    {Object.entries(sec.metrics).map(([k, v]) => (
                      <div className="report-metric" key={k}>
                        <span className="muted">{k}</span>
                        <strong>{formatMetric(v)}</strong>
                      </div>
                    ))}
                  </div>
                )}
                {sec.rows && sec.rows.length > 0 && (
                  <div className="table-wrap" style={{ marginTop: 10 }}>
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
          </section>
        )}
      </div>
    </main>
  )
}

function formatMetric(v: unknown): string {
  if (v == null) return '—'
  if (typeof v === 'object') return JSON.stringify(v)
  return String(v)
}
