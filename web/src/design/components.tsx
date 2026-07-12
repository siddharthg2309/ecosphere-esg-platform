import type { ButtonHTMLAttributes, PropsWithChildren } from 'react'
export function Button({className='',...props}:ButtonHTMLAttributes<HTMLButtonElement>){return <button className={`button ${className}`} {...props}/>}
export function Pill({status}: {status:'active'|'inactive'}){return <span className={`pill ${status}`}>{status==='active'?'Active':'Inactive'}</span>}
export function EmptyState({children}:PropsWithChildren){return <div className="empty-state">{children}</div>}
