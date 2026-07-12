import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Button, Modal, Note } from '../design/components'
import { api } from '../lib/apiClient'
import { queryKeys } from '../lib/queryKeys'
import { userFacingError } from '../lib/userFacingError'
import { useState } from 'react'

/** Blocks interaction when the user has unacknowledged active policies. */
export function PolicyAckModal() {
  const qc = useQueryClient()
  const [error, setError] = useState('')
  const unacked = useQuery({
    queryKey: queryKeys.unacknowledged,
    queryFn: api.governance.unacknowledged,
    refetchOnWindowFocus: true,
  })
  const ack = useMutation({
    mutationFn: api.governance.acknowledge,
    onSuccess: () => void qc.invalidateQueries({ queryKey: queryKeys.unacknowledged }),
  })

  const items = unacked.data?.items ?? []
  if (unacked.isLoading || items.length === 0) return null
  const current = items[0]

  return (
    <Modal
      title="Policy acknowledgement required"
      blocking
      onClose={() => undefined}
      footer={
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
      }
    >
      <Note>
        You must acknowledge active policies before continuing. {items.length > 1 ? `${items.length} policies pending.` : null}
      </Note>
      <div style={{ marginTop: 16 }}>
        <div className="list-row">
          <span className="muted">Policy</span>
          <b>{current.title}</b>
        </div>
        <div className="list-row">
          <span className="muted">Version</span>
          <span>v{current.version}</span>
        </div>
        <div className="list-row">
          <span className="muted">Effective</span>
          <span>{String(current.effectiveDate).slice(0, 10)}</span>
        </div>
        <div className="divider" style={{ height: 1, background: 'var(--line)', margin: '12px 0' }} />
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
