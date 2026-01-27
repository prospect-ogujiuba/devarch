import { Link } from '@tanstack/react-router'
import { Server, LayoutGrid, FolderKanban, FolderOpen, Settings, Moon, Sun, Menu } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useTheme } from '@/lib/theme'

const navItems = [
  { to: '/' as const, icon: LayoutGrid, label: 'Overview' },
  { to: '/services' as const, icon: Server, label: 'Services' },
  { to: '/categories' as const, icon: FolderKanban, label: 'Categories' },
  { to: '/projects' as const, icon: FolderOpen, label: 'Projects' },
  { to: '/settings' as const, icon: Settings, label: 'Settings' },
]

export { navItems }

export function Header({ onMenuClick }: { onMenuClick?: () => void }) {
  const { resolvedTheme, setTheme } = useTheme()

  return (
    <header className="sticky top-0 z-50 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="flex h-14 items-center gap-4 px-6">
        <Button
          variant="ghost"
          size="icon-sm"
          className="md:hidden"
          onClick={onMenuClick}
        >
          <Menu className="size-5" />
        </Button>

        <Link to="/" className="flex items-center gap-2 font-semibold">
          <Server className="size-5" />
          <span>DevArch</span>
        </Link>

        <nav className="hidden md:flex items-center gap-1 ml-6">
          {navItems.map(item => (
            <Button key={item.to} variant="ghost" size="sm" asChild>
              <Link to={item.to} activeProps={{ className: 'bg-accent' }} activeOptions={item.to === '/' ? { exact: true } : undefined}>
                <item.icon className="size-4" />
                {item.label}
              </Link>
            </Button>
          ))}
        </nav>

        <div className="ml-auto flex items-center gap-2">
          <Button
            variant="ghost"
            size="icon-sm"
            onClick={() => setTheme(resolvedTheme === 'dark' ? 'light' : 'dark')}
          >
            {resolvedTheme === 'dark' ? (
              <Sun className="size-4" />
            ) : (
              <Moon className="size-4" />
            )}
          </Button>
        </div>
      </div>
    </header>
  )
}
