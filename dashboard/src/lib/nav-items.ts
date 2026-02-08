import { Server, LayoutGrid, FolderKanban, FolderOpen, Settings, Layers } from 'lucide-react'

export const navItems = [
  { to: '/' as const, icon: LayoutGrid, label: 'Overview' },
  { to: '/stacks' as const, icon: Layers, label: 'Stacks' },
  { to: '/services' as const, icon: Server, label: 'Services' },
  { to: '/categories' as const, icon: FolderKanban, label: 'Categories' },
  { to: '/projects' as const, icon: FolderOpen, label: 'Projects' },
  { to: '/settings' as const, icon: Settings, label: 'Settings' },
]
