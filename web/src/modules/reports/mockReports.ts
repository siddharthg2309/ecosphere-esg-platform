import type { Report, ReportSection, ReportType } from '../../lib/types'

function id() {
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) return crypto.randomUUID()
  return `mock-${Date.now()}-${Math.random().toString(16).slice(2)}`
}

const deptRanking: Record<string, string>[] = [
  { department: 'Compliance', env: '54', social: '24', gov: '83', total: '54' },
  { department: 'Logistics', env: '54', social: '24', gov: '83', total: '54' },
  { department: 'Human Resources', env: '54', social: '24', gov: '83', total: '54' },
  { department: 'Finance', env: '54', social: '24', gov: '83', total: '54' },
  { department: 'Manufacturing', env: '54', social: '24', gov: '45', total: '42' },
]

const emissionRows: Record<string, string>[] = [
  { source: 'Fleet', co2e: '296 t', share: '58%', trend: '▼ 11% YoY' },
  { source: 'Purchased Energy', co2e: '118 t', share: '23%', trend: '▼ 4% YoY' },
  { source: 'Expense / Travel', co2e: '64 t', share: '12%', trend: '▲ 2% YoY' },
  { source: 'Manufacturing', co2e: '34 t', share: '7%', trend: '▼ 9% YoY' },
]

const goalRows: Record<string, string>[] = [
  { goal: 'Reduce Fleet Emissions 20%', department: 'Logistics', progress: '78%', status: 'On track' },
  { goal: 'Warehouse Solar Retrofit', department: 'Manufacturing', progress: '32%', status: 'At risk' },
  { goal: 'Paperless Dispatch', department: 'Logistics', progress: '100%', status: 'Completed' },
  { goal: 'Cut Idle Energy 15%', department: 'Facilities', progress: '55%', status: 'On track' },
]

const socialRows: Record<string, string>[] = [
  { metric: 'Gender diversity (women)', value: '40%', target: '45%' },
  { metric: 'Leadership diversity', value: '50%', target: '40%' },
  { metric: 'CSR participation YTD', value: '28%', target: '50%' },
  { metric: 'Training completion', value: '72%', target: '90%' },
]

const csrRows: Record<string, string>[] = [
  { activity: 'Tree Plantation', joined: '42', points: '50', evidence: 'Required' },
  { activity: 'Blood Donation', joined: '31', points: '40', evidence: 'Required' },
  { activity: 'Beach Cleanup', joined: '58', points: '60', evidence: 'Optional' },
  { activity: 'ESG Workshop', joined: '76', points: '30', evidence: 'Optional' },
]

const govIssueRows: Record<string, string>[] = [
  { issue: 'Missing MSDS sheets', department: 'Manufacturing', severity: 'High', due: '2026-07-02', status: 'Open · overdue' },
  { issue: 'Vendor CoC lag', department: 'Procurement', severity: 'Medium', due: '2026-07-20', status: 'In progress' },
  { issue: 'Retention policy gap', department: 'IT', severity: 'Low', due: '2026-08-01', status: 'Open' },
]

const auditRows: Record<string, string>[] = [
  { audit: 'Q2 Waste Audit', department: 'Manufacturing', date: '2026-06-12', status: 'Completed', findings: '2 minor' },
  { audit: 'Energy Governance Review', department: 'Facilities', date: '2026-07-22', status: 'Assigned', findings: '—' },
  { audit: 'Vendor Compliance Check', department: 'Procurement', date: '2026-07-10', status: 'Overdue', findings: '1 open issue' },
]

const policyRows: Record<string, string>[] = [
  { policy: 'Environmental Responsibility', version: 'v1', ackRate: '94%', status: 'Published' },
  { policy: 'Supplier Code of Conduct', version: 'v1', ackRate: '88%', status: 'Published' },
  { policy: 'Ethics and Governance', version: 'v1', ackRate: '96%', status: 'Published' },
]

function section(title: string, partial: Partial<ReportSection>): ReportSection {
  return { title, ...partial }
}

function environmentalSections(): ReportSection[] {
  return [
    section('Emissions overview', {
      summary:
        'Verified emissions total 512 t CO₂e YTD across fleet, energy, travel and manufacturing. Overall environmental pillar score is 54 / 100.',
      rows: [
        { metric: 'Verified CO₂e (YTD)', value: '512 t' },
        { metric: 'YoY change', value: '▼ 9%' },
        { metric: 'Active reduction goals', value: '4' },
        { metric: 'Environmental pillar', value: '54 / 100' },
      ],
      metrics: { verifiedEmissions: 512, goalsTracked: 4, pillar: 54 },
    }),
    section('Emissions by source', { rows: emissionRows }),
    section('Environmental goals', { rows: goalRows }),
  ]
}

function socialSections(): ReportSection[] {
  return [
    section('Social overview', {
      summary:
        'Workforce diversity and CSR engagement are tracking mid-range. Leadership diversity exceeds target; training completion needs attention before FY close.',
      rows: socialRows,
      metrics: { genderWomenPct: 40, csrParticipationPct: 28, trainingCompletionPct: 72 },
    }),
    section('CSR activities', { rows: csrRows }),
    section('Training snapshot', {
      rows: [
        { training: 'ESG Fundamentals', assigned: 'All employees', completed: '144 / 200', rate: '72%' },
        { training: 'Anti-Corruption Awareness', assigned: 'All employees', completed: '168 / 200', rate: '84%' },
      ],
    }),
  ]
}

function governanceSections(): ReportSection[] {
  return [
    section('Governance overview', {
      summary:
        'Governance pillar score is 75 / 100. Policy acknowledgements are healthy org-wide; one high-severity compliance issue is overdue.',
      rows: [
        { metric: 'Governance pillar', value: '75 / 100' },
        { metric: 'Open issues', value: '3' },
        { metric: 'Overdue issues', value: '1' },
        { metric: 'Policy ack. rate', value: '94%' },
        { metric: 'Audits (FY)', value: '12' },
      ],
      metrics: { pillar: 75, openIssues: 3, overdue: 1, ackRate: 94 },
    }),
    section('Compliance issues', { rows: govIssueRows }),
    section('Audit trail', { rows: auditRows }),
    section('Policy acknowledgements', { rows: policyRows }),
  ]
}

function summarySections(): ReportSection[] {
  return [
    section('Executive summary (AI-generated)', {
      ai: true,
      summary:
        'Executive summary (prototype narrative): Overall ESG performance sits at 52 / 100 with a 40 / 30 / 30 Env–Social–Gov weighting. Environmental progress is driven by fleet and energy tracking; Social engagement is improving via CSR and training; Governance remains the strongest pillar where policy acknowledgements and audit coverage are healthy. Manufacturing lags on total department score and warrants focused remediation on the overdue MSDS compliance issue. All numeric scores above are system-calculated for this prototype — the narrative is advisory only.',
    }),
    section('Overall ESG Score', {
      rows: [
        { pillar: 'Environmental', score: '54', weight: '40%' },
        { pillar: 'Social', score: '24', weight: '30%' },
        { pillar: 'Governance', score: '75', weight: '30%' },
        { pillar: 'Overall (weighted)', score: '52', weight: '100%' },
      ],
      metrics: { overall: 52, weightEnv: 40, weightSocial: 30, weightGov: 30 },
    }),
    section('Department ranking', { rows: deptRanking }),
    ...environmentalSections().slice(0, 2),
    ...socialSections().slice(0, 1),
    ...governanceSections().slice(0, 2),
  ]
}

function customSections(filters: Record<string, unknown>): ReportSection[] {
  const filterRows = Object.entries(filters)
    .filter(([, v]) => v !== '' && v != null)
    .map(([k, v]) => ({ filter: k, value: String(v) }))
  return [
    section('Custom report scope', {
      summary: 'Prototype custom report assembled from the selected filters. Data below is demo/mock for review.',
      rows:
        filterRows.length > 0
          ? filterRows
          : [
              { filter: 'department', value: 'All' },
              { filter: 'module', value: 'All' },
              { filter: 'period', value: 'FY2026 YTD' },
            ],
    }),
    section('Department ranking', { rows: deptRanking }),
    section('Emissions by source', { rows: emissionRows }),
    section('Compliance issues', { rows: govIssueRows }),
    section('CSR activities', { rows: csrRows }),
  ]
}

/** Rich prototype report for every standard / custom type. */
export function buildMockReport(
  type: ReportType,
  filters: Record<string, unknown> = {},
): Report {
  let sections: ReportSection[]
  switch (type) {
    case 'environmental':
      sections = environmentalSections()
      break
    case 'social':
      sections = socialSections()
      break
    case 'governance':
      sections = governanceSections()
      break
    case 'esg_summary':
      sections = summarySections()
      break
    case 'custom':
    default:
      sections = customSections(filters)
      break
  }
  return {
    id: id(),
    type,
    filters,
    sections,
    generatedAt: new Date().toISOString(),
  }
}

/** Client-side CSV for mock report export. */
export function mockReportToCsv(report: Report): string {
  const lines: string[] = [`type,${report.type}`, `generatedAt,${report.generatedAt}`, '']
  for (const sec of report.sections) {
    lines.push(`section,${JSON.stringify(sec.title)}`)
    if (sec.summary) lines.push(`summary,${JSON.stringify(sec.summary)}`)
    if (sec.rows?.length) {
      const keys = Object.keys(sec.rows[0])
      lines.push(keys.join(','))
      for (const row of sec.rows) {
        lines.push(keys.map((k) => JSON.stringify(row[k] ?? '')).join(','))
      }
    }
    lines.push('')
  }
  return lines.join('\n')
}

export function downloadTextFile(filename: string, content: string, mime: string) {
  const blob = new Blob([content], { type: mime })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}
