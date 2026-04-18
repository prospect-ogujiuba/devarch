import { Activity, Boxes, FolderKanban, Settings2 } from 'lucide-react'
import { NavLink, Outlet } from 'react-router-dom'
import { classNames } from '../lib/utils'

const navItems = [
  { to: '/workspaces', label: 'Workspaces', icon: Boxes },
  { to: '/catalog', label: 'Catalog', icon: FolderKanban },
  { to: '/activity', label: 'Activity', icon: Activity },
  { to: '/settings', label: 'Settings', icon: Settings2 },
]

export function AppShell() {
  return (
    <div className="app-shell">
      <header className="topbar">
        <div className="topbar__brand">
          <span className="topbar__logo">DA</span>
          <div>
            <div className="topbar__title">DevArch V2</div>
            <div className="topbar__subtitle">Workspace-first control plane</div>
          </div>
        </div>
        <nav className="topbar__nav" aria-label="Primary">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) => classNames('nav-link', isActive && 'nav-link--active')}
            >
              <item.icon size={16} />
              <span>{item.label}</span>
            </NavLink>
          ))}
        </nav>
      </header>
      <main className="page-shell">
        <Outlet />
      </main>
    </div>
  )
}
