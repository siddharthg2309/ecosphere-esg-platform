import { useMutation } from '@tanstack/react-query'
import { Button, Note, Pill } from '../../design/components'
import { api } from '../../lib/apiClient'
import type { EvidenceReview } from '../../lib/types'
import { userFacingError } from '../../lib/userFacingError'

/** Advisory AI proof check — never auto-approves. */
export function EvidenceAssist({ proofUrl }: { proofUrl?: string }) {
  const review = useMutation({
    mutationFn: () => api.ai.evidenceReview(proofUrl || ''),
  })

  if (!proofUrl) {
    return <span className="muted" style={{ fontSize: 12 }}>No proof to review</span>
  }

  return (
    <div style={{ marginTop: 6, maxWidth: 280 }}>
      <Button
        className="secondary sm"
        type="button"
        disabled={review.isPending}
        onClick={() => void review.mutateAsync()}
      >
        {review.isPending ? 'AI reviewing…' : 'AI verify'}
      </Button>
      {review.isError && (
        <div className="form-error" style={{ marginTop: 4, fontSize: 12 }}>
          {userFacingError(review.error, 'Evidence assist unavailable')}
        </div>
      )}
      {review.data && <EvidenceAssistResult result={review.data} />}
    </div>
  )
}

export function EvidenceAssistResult({ result }: { result: EvidenceReview }) {
  const pct = Math.round((result.confidence ?? 0) * 100)
  return (
    <div style={{ marginTop: 6 }}>
      <Note>
        <div className="rowflex" style={{ gap: 8, marginBottom: 4 }}>
          <Pill status={result.looksValid ? 'approved' : 'warning'}>
            {result.looksValid ? 'Looks valid' : 'Needs review'}
          </Pill>
          <span className="muted">{pct}% confidence</span>
        </div>
        <div style={{ fontSize: 12 }}>{result.notes}</div>
        <div className="muted" style={{ marginTop: 4, fontSize: 12 }}>
          Advisory only — a human must still approve or reject.
        </div>
      </Note>
    </div>
  )
}
