# Header Component

**Path:** `dashboard/src/components/layout/header.tsx`
**Last Updated:** 2026-02-17

## Overview
Sticky top navigation header with logo, primary nav links, theme toggle, and mobile menu button. Uses TanStack Router for navigation and active link styling.

## Layout
- **Left:** DevArch logo (Server icon) + title "DevArch" (hidden on mobile)
- **Center:** Primary nav links (hidden below md breakpoint)
- **Right:** Settings link, theme toggle, mobile menu button

## Props
```typescript
{ onMenuClick?: () => void }
```

- `onMenuClick` — Optional callback for mobile menu hamburger button

## Navigation Items

### Primary Links
- Fetched from `navItems` config (navigation utility)
- Excludes settings item (separate)
- Rendered as buttons with links

### Settings Link
- Only shown on md+ screens
- Separate from primary items (always in right section)

### Both Use Active Link Styling
- `activeProps={{ className: 'bg-accent text-accent-foreground' }}`
- Home link: `activeOptions={{ exact: true }}` (only match exact `/` route)
- Other links: match by prefix

## Responsive Behavior

### Mobile (below md)
- Logo title hidden (icon only)
- Primary nav hidden
- Settings link hidden
- Divider separator visible ("|")
- Hamburger menu visible

### Desktop (md+)
- Logo title visible
- Primary nav visible
- Settings link visible
- Hamburger menu hidden
- Divider hidden

## Theme Toggle

### Button
- Cycles between light/dark theme
- Icon changes based on `resolvedTheme`:
  - Dark theme: Sun icon (light mode suggestion)
  - Light theme: Moon icon (dark mode suggestion)

### Hook
- Uses `useTheme()` hook from theme context
- Returns `resolvedTheme` and `setTheme()`

## Styling
- Sticky positioning with z-50 (top navbar layer)
- Border bottom for separation
- Backdrop blur with fallback (glass morphism effect)
- Height: h-12 (mobile), h-14 (desktop via sm:)
- Padding: px-3 (mobile), px-6 (desktop)

## Dependencies
- `Link` — TanStack Router Link component
- `Button` — UI button component (variant="ghost")
- `useTheme()` — Theme context hook
- `navItems` — Navigation configuration
- Icons: Server, Moon, Sun, Menu

## Active Link Behavior
- Home route (`/`): exact match required
- Other routes: prefix match
- Active state adds `bg-accent text-accent-foreground` class
- Provides visual feedback for current page

## Related Components
- Nav drawer/menu (implemented in parent layout)
- Navigation configuration in `@/lib/nav-items`
