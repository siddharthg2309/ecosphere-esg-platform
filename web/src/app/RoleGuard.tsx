import type { PropsWithChildren, ReactNode } from 'react'
import type { Role } from '../lib/types'
import { useAuthStore } from './authStore'

export function RoleGuard({roles,children,fallback=null}:PropsWithChildren<{roles:Role[];fallback?:ReactNode}>){const role=useAuthStore(s=>s.user?.role);return role&&roles.includes(role)?children:fallback}
