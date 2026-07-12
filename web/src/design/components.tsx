import type { ButtonHTMLAttributes, PropsWithChildren, ReactNode } from 'react'

export function Button({ className = '', ...props }: ButtonHTMLAttributes<HTMLButtonElement>) {
  return <button className={`button ${className}`} {...props} />
}

type PillStatus =
  | 'active' | 'inactive' | 'pending' | 'approved' | 'rejected' | 'blocked'
  | 'draft' | 'under_review' | 'completed' | 'archived' | 'warning' | 'danger'
  | 'neutral' | 'plum' | 'success' | 'in_progress'

const pillLabel: Partial<Record<PillStatus, string>> = {
  under_review: 'Under Review',
  in_progress: 'In Progress',
  active: 'Active',
  inactive: 'Inactive',
  pending: 'Pending',
  approved: 'Approved',
  rejected: 'Rejected',
  blocked: 'Blocked',
  draft: 'Draft',
  completed: 'Completed',
  archived: 'Archived',
}

export function Pill({ status, children }: { status: PillStatus | string; children?: ReactNode }) {
  const key = String(status).toLowerCase().replace(' ', '_') as PillStatus
  const label = children ?? pillLabel[key] ?? String(status).replace(/_/g, ' ')
  return (
    <span className={`pill ${key}`}>
      <span className="dot" />
      {label}
    </span>
  )
}

export function EmptyState({ children }: PropsWithChildren) {
  return <div className="empty-state">{children}</div>
}

export function Card({ children, className = '' }: PropsWithChildren<{ className?: string }>) {
  return <div className={`card ${className}`}>{children}</div>
}

export function Progress({ value, tone }: { value: number; tone?: 'warning' | 'danger' }) {
  const width = Math.max(0, Math.min(100, value))
  return (
    <div className="progress" aria-valuenow={width} aria-valuemin={0} aria-valuemax={100} role="progressbar">
      <span className={tone} style={{ width: `${width}%` }} />
    </div>
  )
}

export function Modal({
  title,
  children,
  onClose,
  footer,
}: PropsWithChildren<{ title: string; onClose(): void; footer?: ReactNode }>) {
  return (
    <div className="modal-backdrop" role="presentation" onMouseDown={onClose}>
      <section className="modal" role="dialog" aria-modal="true" aria-labelledby="modal-title" onMouseDown={(e) => e.stopPropagation()}>
        <header className="modal-head">
          <h2 id="modal-title">{title}</h2>
          <button className="close-button" onClick={onClose} aria-label="Close" type="button">
            ×
          </button>
        </header>
        {children}
        {footer && <div className="modal-foot">{footer}</div>}
      </section>
    </div>
  )
}

export function StatBar({ items }: { items: { label: string; value: string | number; sub?: string; unit?: string }[] }) {
  return (
    <div className="card statcard">
      <div className="statbar">
        {items.map((item) => (
          <div className="stat" key={item.label}>
            <div className="label">{item.label}</div>
            <div className="num">
              {item.value}
              {item.unit && <small> {item.unit}</small>}
            </div>
            {item.sub && <div className="sub">{item.sub}</div>}
          </div>
        ))}
      </div>
    </div>
  )
}

export function Note({ children }: PropsWithChildren) {
  return (
    <div className="note">
      <span aria-hidden>⚠</span>
      <div>{children}</div>
    </div>
  )
}

export function initials(name: string) {
  return name
    .split(' ')
    .map((v) => v[0])
    .slice(0, 2)
    .join('')
    .toUpperCase()
}
