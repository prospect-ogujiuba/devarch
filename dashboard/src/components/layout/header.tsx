import { Link } from '@tanstack/react-router'
import { Server, LayoutGrid, FolderKanban, Settings, Moon, Sun } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useTheme } from '@/lib/theme'

export function Header() {
  const { resolvedTheme, setTheme } = useTheme()

  return (
    <header className="sticky top-0 z-50 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="flex h-14 items-center gap-4 px-6">
        <Link to="/" className="flex items-center gap-2 font-semibold">
          <Server className="size-5" />
          <span>DevArch</span>
        </Link>

        <nav className="flex items-center gap-1 ml-6">
          <Button variant="ghost" size="sm" asChild>
            <Link to="/" activeProps={{ className: 'bg-accent' }}>
              <LayoutGrid className="size-4" />
              Overview
            </Link>
          </Button>
          <Button variant="ghost" size="sm" asChild>
            <Link to="/services" activeProps={{ className: 'bg-accent' }}>
              <Server className="size-4" />
              Services
            </Link>
          </Button>
          <Button variant="ghost" size="sm" asChild>
            <Link to="/categories" activeProps={{ className: 'bg-accent' }}>
              <FolderKanban className="size-4" />
              Categories
            </Link>
          </Button>
          <Button variant="ghost" size="sm" asChild>
            <Link to="/settings" activeProps={{ className: 'bg-accent' }}>
              <Settings className="size-4" />
              Settings
            </Link>
          </Button>
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
