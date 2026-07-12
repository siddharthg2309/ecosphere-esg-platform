import { useMutation } from '@tanstack/react-query'
import { useState } from 'react'
import { Button, Note, Pill } from '../../design/components'
import { api } from '../../lib/apiClient'
import type { EvidenceReview } from '../../lib/types'
import { userFacingError } from '../../lib/userFacingError'
import { fileToDataURL, isImageFile, MAX_PROOF_BYTES } from './proofUpload'

/** Advisory AI proof check — never auto-approves. Supports URL or one-shot upload (not stored). */
export function EvidenceAssist({ proofUrl }: { proofUrl?: string }) {
  const [localFile, setLocalFile] = useState<File | null>(null)
  const [preview, setPreview] = useState<string>('')
  const review = useMutation({
    mutationFn: async () => {
      if (localFile) {
        const dataUrl = await fileToDataURL(localFile)
        return api.ai.evidenceReview({
          imageDataUrl: dataUrl,
          fileName: localFile.name,
          proofUrl: proofUrl || `upload:${localFile.name}`,
        })
      }
      return api.ai.evidenceReview({ proofUrl: proofUrl || '', imageUrl: proofUrl || '' })
    },
  })

  const canReview =
    Boolean(localFile) ||
    Boolean(proofUrl && (proofUrl.startsWith('http') || proofUrl.startsWith('data:')))

  return (
    <div style={{ marginTop: 6, maxWidth: 300 }}>
      <label className="muted" style={{ fontSize: 11, display: 'block', marginBottom: 4 }}>
        AI verify (image not stored)
        <input
          type="file"
          accept="image/jpeg,image/png,image/webp,image/gif"
          style={{ display: 'block', marginTop: 4, fontSize: 12 }}
          onChange={(e) => {
            const f = e.target.files?.[0]
            setLocalFile(null)
            setPreview('')
            review.reset()
            if (!f) return
            if (!isImageFile(f)) return
            if (f.size > MAX_PROOF_BYTES) return
            setLocalFile(f)
            void fileToDataURL(f).then(setPreview)
          }}
        />
      </label>
      {preview && (
        <img src={preview} alt="Proof preview" style={{ maxWidth: '100%', maxHeight: 80, borderRadius: 6, marginBottom: 6 }} />
      )}
      <Button
        className="secondary sm"
        type="button"
        disabled={review.isPending || !canReview}
        onClick={() => void review.mutateAsync()}
      >
        {review.isPending ? 'AI reviewing…' : 'AI verify'}
      </Button>
      {!canReview && proofUrl && !proofUrl.startsWith('http') && (
        <div className="muted" style={{ fontSize: 11, marginTop: 4 }}>
          Proof was uploaded without storage — pick the image again to verify.
        </div>
      )}
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
          Advisory only — image is not stored. A human must still approve or reject.
        </div>
      </Note>
    </div>
  )
}

/** Proof field for join modals: upload → optional AI verify → stores only upload:filename */
export function ProofUploadField({
  required,
  name = 'proofUrl',
  onMarkerChange,
}: {
  required?: boolean
  name?: string
  onMarkerChange?: (marker: string, dataUrl: string, fileName: string) => void
}) {
  const [fileName, setFileName] = useState('')
  const [marker, setMarker] = useState('')
  const [preview, setPreview] = useState('')
  const [err, setErr] = useState('')
  const [ai, setAi] = useState<EvidenceReview | null>(null)
  const review = useMutation({
    mutationFn: (dataUrl: string) =>
      api.ai.evidenceReview({ imageDataUrl: dataUrl, fileName, proofUrl: marker }),
    onSuccess: setAi,
  })

  return (
    <div className="field" style={{ display: 'grid', gap: 8 }}>
      <div className="lbl" style={{ fontSize: 11, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.06em', color: 'var(--ink-muted)' }}>
        Proof photo {required ? '(required)' : '(optional)'}
      </div>
      <div className="upload-area" style={{ position: 'relative' }}>
        <span className="ico ico-upload" />
        <div>
          {fileName ? (
            <>
              <b>{fileName}</b>
              <div style={{ fontSize: 12, marginTop: 4 }}>Image stays in the browser — only a name is saved</div>
            </>
          ) : (
            <>
              Drop or <b style={{ color: 'var(--brand)' }}>browse</b> a proof photo
              <div style={{ marginTop: 4, fontSize: 12 }}>JPG, PNG, WebP · max 3 MB · not stored on server</div>
            </>
          )}
        </div>
        <input
          type="file"
          accept="image/jpeg,image/png,image/webp,image/gif"
          required={required && !marker}
          style={{ position: 'absolute', inset: 0, opacity: 0, cursor: 'pointer' }}
          onChange={(e) => {
            const f = e.target.files?.[0]
            setAi(null)
            setErr('')
            if (!f) {
              setFileName('')
              setMarker('')
              setPreview('')
              onMarkerChange?.('', '', '')
              return
            }
            if (!isImageFile(f)) {
              setErr('Please choose an image file (JPG/PNG/WebP).')
              return
            }
            if (f.size > MAX_PROOF_BYTES) {
              setErr('Image must be under 3 MB.')
              return
            }
            void fileToDataURL(f).then((dataUrl) => {
              const m = `upload:${f.name}`
              setFileName(f.name)
              setMarker(m)
              setPreview(dataUrl)
              onMarkerChange?.(m, dataUrl, f.name)
            })
          }}
        />
      </div>
      {/* Only the lightweight marker is submitted — never the image bytes */}
      <input type="hidden" name={name} value={marker} required={required} />
      {preview && (
        <img src={preview} alt="Preview" style={{ maxHeight: 120, maxWidth: '100%', borderRadius: 8, objectFit: 'cover' }} />
      )}
      {marker && (
        <Button
          type="button"
          className="secondary sm"
          disabled={review.isPending || !preview}
          onClick={() => void review.mutateAsync(preview)}
        >
          {review.isPending ? 'AI reviewing…' : 'AI verify (not stored)'}
        </Button>
      )}
      {err && (
        <div className="form-error" role="alert">
          {err}
        </div>
      )}
      {review.isError && (
        <div className="form-error" role="alert">
          {userFacingError(review.error, 'Evidence assist unavailable')}
        </div>
      )}
      {ai && <EvidenceAssistResult result={ai} />}
    </div>
  )
}
