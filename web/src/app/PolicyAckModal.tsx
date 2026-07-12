import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { type CSSProperties, useState } from 'react'
import { Button, Modal, Note } from '../design/components'
import { api } from '../lib/apiClient'
import { queryKeys } from '../lib/queryKeys'
import { userFacingError } from '../lib/userFacingError'
import { useAuthStore } from './authStore'

/** Blocks employees when they have unacknowledged active policies. */
export function PolicyAckModal() {
  const role = useAuthStore((s) => s.user?.role)
  const qc = useQueryClient()
  const [error, setError] = useState('')
  const [dismissed, setDismissed] = useState(false)
  const enabled = role === 'employee' || role === 'dept_head'
  const unacked = useQuery({
    queryKey: queryKeys.unacknowledged,
    queryFn: api.governance.unacknowledged,
    refetchOnWindowFocus: true,
    enabled,
  })
  const ack = useMutation({
    mutationFn: api.governance.acknowledge,
    onSuccess: () => void qc.invalidateQueries({ queryKey: queryKeys.unacknowledged }),
  })

  if (!enabled || dismissed) return null
  const items = unacked.data?.items ?? []
  if (unacked.isLoading || items.length === 0) return null
  const current = items[0]

  const metaCell: CSSProperties = {
    padding: '10px 16px',
    borderRight: '1px solid var(--line)',
  }

  return (
    <Modal
      title="Policy acknowledgement required"
      onClose={() => setDismissed(true)}
      footer={
        <>
          <Button className="secondary" disabled={ack.isPending} onClick={() => setDismissed(true)}>
            Cancel
          </Button>
          <Button
            className="primary"
            disabled={ack.isPending}
            onClick={() => {
              setError('')
              void ack.mutateAsync(current.id).catch((e) => setError(userFacingError(e, 'Unable to acknowledge')))
            }}
          >
            I acknowledge this policy
          </Button>
        </>
      }
    >
      <Note>
        You must acknowledge active policies before continuing. {items.length > 1 ? `${items.length} policies pending.` : null}
      </Note>
      <div style={{ marginTop: 16 }}>
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'minmax(0, 1fr) auto auto',
            border: '1px solid var(--line)',
            borderRadius: 'var(--radius-sm)',
            overflow: 'hidden',
          }}
        >
          <div style={metaCell}>
            <div className="muted" style={{ fontSize: 11, textTransform: 'uppercase', letterSpacing: '.05em' }}>
              Policy
            </div>
            <b>{current.title}</b>
          </div>
          <div style={metaCell}>
            <div className="muted" style={{ fontSize: 11, textTransform: 'uppercase', letterSpacing: '.05em' }}>
              Version
            </div>
            <span>v{current.version}</span>
          </div>
          <div style={{ ...metaCell, borderRight: 0 }}>
            <div className="muted" style={{ fontSize: 11, textTransform: 'uppercase', letterSpacing: '.05em' }}>
              Effective
            </div>
            <span>{String(current.effectiveDate).slice(0, 10)}</span>
          </div>
        </div>
        <div className="divider" style={{ height: 1, background: 'var(--line)', margin: '16px 0' }} />
        <div style={{ fontSize: 13.5, lineHeight: 1.7, maxHeight: 200, overflow: 'auto' }}>{current.body}</div>
        {error && (
          <div className="form-error" role="alert" style={{ marginTop: 12 }}>
            {error}
          </div>
        )}
      </div>
    </Modal>
  )
}
