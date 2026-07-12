import type { Role } from '../lib/types'

export type NavItem = {
  label: string
  path: string
  /** Match path exactly (for `/` hub routes). */
  end?: boolean
}

export type NavGroup = {
  title?: string
  items: NavItem[]
}

export type PortalMeta = {
  id: Role
  name: string
  subtitle: string
  homeLabel: string
  crumb: string
  demoEmail: string
}

/** Demo accounts (seeded; password ChangeMe123!). */
export const PORTALS: PortalMeta[] = [
  {
    id: 'admin',
    name: 'Admin Portal',
    subtitle: 'Full ESG operating system · config · org scores',
    homeLabel: 'Dashboard',
    crumb: 'Admin Portal',
    demoEmail: 'admin@ecosphere.local',
  },
  {
    id: 'dept_head',
    name: 'Department Portal',
    subtitle: 'Operational log · goals · team approvals',
    homeLabel: 'Department Hub',
    crumb: 'Department Portal',
    demoEmail: 'rohan@ecosphere.local',
  },
  {
    id: 'auditor',
    name: 'Auditor Portal',
    subtitle: 'Audit queue · issues · data verification',
    homeLabel: 'Audit Hub',
    crumb: 'Auditor Portal',
    demoEmail: 'kiran@ecosphere.local',
  },
  {
    id: 'employee',
    name: 'Employee Portal',
    subtitle: 'Challenges · CSR · badges · policy sign-off',
    homeLabel: 'My ESG Hub',
    crumb: 'Employee Portal',
    demoEmail: 'employee06@ecosphere.local',
  },
]

export const DEMO_PASSWORD = 'ChangeMe123!'

/** Wireframe-aligned nav per role. */
export const navByRole: Record<Role, NavGroup[]> = {
  admin: [
    { items: [{ label: 'Dashboard', path: '/', end: true }] },
    { title: 'Environmental', items: [{ label: 'Emissions & Goals', path: '/environmental' }] },
    { title: 'Social', items: [{ label: 'CSR & Engagement', path: '/social' }] },
    { title: 'Governance', items: [{ label: 'Policies & Audits', path: '/governance' }] },
    { title: 'Gamification', items: [{ label: 'Challenges', path: '/gamification' }] },
    {
      title: 'Insights',
      items: [
        { label: 'Reports', path: '/reports' },
        { label: 'Settings', path: '/settings' },
      ],
    },
  ],
  dept_head: [
    { items: [{ label: 'Department Hub', path: '/', end: true }] },
    {
      title: 'Operations',
      items: [
        { label: 'Carbon Transactions', path: '/environmental' },
        { label: 'CSR & Team', path: '/social' },
      ],
    },
    {
      title: 'Team',
      items: [{ label: 'Challenges & Approvals', path: '/gamification' }],
    },
    { title: 'Insights', items: [{ label: 'Reports', path: '/reports' }] },
  ],
  auditor: [
    { items: [{ label: 'Audit Hub', path: '/', end: true }] },
    {
      title: 'Audits',
      items: [{ label: 'Audits & Issues', path: '/governance' }],
    },
    {
      title: 'Verification',
      items: [{ label: 'Carbon Data', path: '/environmental' }],
    },
    { title: 'Insights', items: [{ label: 'Reports', path: '/reports' }] },
  ],
  employee: [
    { items: [{ label: 'My ESG Hub', path: '/', end: true }] },
    {
      title: 'Participate',
      items: [
        { label: 'My Challenges', path: '/gamification' },
        { label: 'CSR Activities', path: '/social' },
      ],
    },
    {
      title: 'Rewards',
      items: [{ label: 'Badges & Rewards', path: '/gamification' }],
    },
    {
      title: 'Compliance',
      items: [{ label: 'Policy Sign-off', path: '/governance' }],
    },
  ],
}

/** Which roles may open a given app path. */
export function rolesAllowedForPath(pathname: string): Role[] {
  const path = pathname.replace(/\/+$/, '') || '/'
  if (path === '/settings' || path.startsWith('/settings/')) return ['admin']
  if (path === '/reports' || path.startsWith('/reports/')) return ['admin', 'dept_head', 'auditor']
  if (path === '/environmental' || path.startsWith('/environmental/')) {
    return ['admin', 'dept_head', 'auditor', 'employee']
  }
  if (path === '/social' || path.startsWith('/social/')) return ['admin', 'dept_head', 'employee']
  if (path === '/gamification' || path.startsWith('/gamification/')) {
    return ['admin', 'dept_head', 'employee']
  }
  if (path === '/governance' || path.startsWith('/governance/')) {
    return ['admin', 'dept_head', 'auditor', 'employee']
  }
  // Hub `/` and unknown fallbacks
  return ['admin', 'dept_head', 'auditor', 'employee']
}

export function canAccessPath(role: Role | undefined, pathname: string): boolean {
  if (!role) return false
  return rolesAllowedForPath(pathname).includes(role)
}

/** Post-login home for every role (never force Settings). */
export function homePathForRole(_role: Role): string {
  return '/'
}

export function portalMeta(role: Role | undefined): PortalMeta {
  return PORTALS.find((p) => p.id === role) ?? PORTALS[0]
}

export function crumbForPath(role: Role | undefined, pathname: string): string {
  const meta = portalMeta(role)
  const path = pathname.replace(/\/+$/, '') || '/'
  if (path === '/') return `${meta.crumb} / ${meta.homeLabel}`
  for (const group of navByRole[role ?? 'admin'] ?? []) {
    for (const item of group.items) {
      if (item.path === path) return `${meta.crumb} / ${item.label}`
    }
  }
  const segment = path.split('/').filter(Boolean)[0] ?? 'Home'
  return `${meta.crumb} / ${segment.charAt(0).toUpperCase()}${segment.slice(1)}`
}
