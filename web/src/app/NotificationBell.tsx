import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useEffect, useRef, useState } from 'react'
import { api } from '../lib/apiClient'
import { queryKeys } from '../lib/queryKeys'
import type { AppNotification } from '../lib/types'

function relativeTime(iso: string) {
  const ms = Date.now() - new Date(iso).getTime()
  const m = Math.floor(ms / 60000)
  if (m < 1) return 'just now'
  if (m < 60) return `${m}m ago`
  const h = Math.floor(m / 60)
  if (h < 24) return `${h}h ago`
  return `${Math.floor(h / 24)}d ago`
}

function iconFor(type: string) {
  switch (type) {
    case 'compliance_raised':
    case 'compliance_overdue':
      return '!'
    case 'badge_unlock':
      return '★'
    case 'policy_reminder':
      return 'P'
    default:
      return '✓'
  }
}

export function NotificationBell() {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)
  const qc = useQueryClient()
  const list = useQuery({
    queryKey: queryKeys.notifications,
    queryFn: api.notifications.list,
    refetchInterval: 30_000,
  })
  const mark = useMutation({
    mutationFn: api.notifications.markRead,
    onSuccess: () => void qc.invalidateQueries({ queryKey: queryKeys.notifications }),
  })

  useEffect(() => {
    function onDoc(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [])

  const unread = list.data?.unread ?? 0
  const items = list.data?.items ?? []

  return (
    <div className="notif-wrap" ref={ref} style={{ position: 'relative' }}>
      <button
        type="button"
        className="icon-button"
        aria-label="Notifications"
        aria-expanded={open}
        onClick={() => setOpen((v) => !v)}
        style={{ position: 'relative' }}
      >
        ●
        {unread > 0 && (
          <span
            className="bell-badge"
            style={{
              position: 'absolute',
              top: 2,
              right: 2,
              minWidth: 16,
              height: 16,
              borderRadius: 999,
              background: 'var(--danger)',
              color: '#fff',
              fontSize: 10,
              fontWeight: 700,
              display: 'grid',
              placeItems: 'center',
              padding: '0 4px',
            }}
          >
            {unread > 9 ? '9+' : unread}
          </span>
        )}
      </button>
      {open && (
        <div
          className="notif-panel"
          role="dialog"
          aria-label="Notifications"
          style={{
            position: 'absolute',
            right: 0,
            top: 44,
            width: 340,
            maxHeight: 400,
            overflow: 'auto',
            background: '#fff',
            border: '1px solid var(--line)',
            borderRadius: 'var(--radius)',
            boxShadow: 'var(--shadow-lg)',
            zIndex: 40,
          }}
        >
          <div className="notif-hd" style={{ display: 'flex', justifyContent: 'space-between', padding: '12px 14px', borderBottom: '1px solid var(--line)' }}>
            <span style={{ fontWeight: 600 }}>Notifications</span>
            <span className="muted" style={{ fontSize: 12 }}>
              {unread} unread
            </span>
          </div>
          {items.length === 0 && <div className="center-state" style={{ padding: 24 }}>No notifications</div>}
          {items.map((n: AppNotification) => (
            <button
              key={n.id}
              type="button"
              className="notif-item"
              onClick={() => {
                if (!n.readAt) void mark.mutateAsync(n.id)
              }}
              style={{
                display: 'flex',
                gap: 10,
                width: '100%',
                textAlign: 'left',
                border: 0,
                borderBottom: '1px solid var(--line)',
                background: n.readAt ? '#fff' : 'var(--brand-100)',
                padding: '12px 14px',
                cursor: 'pointer',
              }}
            >
              <span
                className="notif-ico"
                style={{
                  width: 28,
                  height: 28,
                  borderRadius: 8,
                  background: 'var(--canvas)',
                  display: 'grid',
                  placeItems: 'center',
                  fontWeight: 700,
                  flexShrink: 0,
                }}
              >
                {iconFor(n.type)}
              </span>
              <div style={{ minWidth: 0 }}>
                <div className="notif-title" style={{ fontWeight: 600, fontSize: 13 }}>
                  {n.title || n.type.replace(/_/g, ' ')}
                </div>
                <div className="notif-meta muted" style={{ fontSize: 12 }}>
                  {n.type.replace(/_/g, ' ')} · {relativeTime(n.createdAt)}
                </div>
              </div>
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
