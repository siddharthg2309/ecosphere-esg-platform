import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { RoleGuard } from '../../app/RoleGuard'
import { Button, Card, Pill, StatBar } from '../../design/components'
import { api } from '../../lib/apiClient'
import { queryKeys } from '../../lib/queryKeys'
import { userFacingError } from '../../lib/userFacingError'

export function DashboardPage() {
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
    <main className="page">
      <div className="page-shell">
        <header className="page-head">
          <div>
            <p className="eyebrow">Dashboard</p>
            <h1>Executive Overview</h1>
            <p className="muted">Organization-wide ESG performance · live scores</p>
          </div>
          <div className="rowflex">
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
          </div>
        </header>
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

        <div className="grid cols-2" style={{ marginTop: 24 }}>
          <Card>
            <div className="card-head">
              <h3>Emissions summary</h3>
              <Pill status="plum">Verified t CO₂e</Pill>
            </div>
            <div className="stat" style={{ border: 0, padding: '8px 0' }}>
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
            <div className="table-wrap" style={{ border: 0 }}>
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

        <Card style={{ marginTop: 24 }}>
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
      </div>
    </main>
  )
}
