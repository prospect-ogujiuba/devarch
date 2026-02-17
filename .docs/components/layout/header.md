# Header Component

**Path:** `dashboard/src/components/layout/header.tsx`
**Last Updated:** 2026-02-17

## Overview
Sticky top navigation header with logo, primary nav links, theme toggle, and mobile menu button. Uses TanStack Router for navigation and active link styling.

## Layout
- **Left:** DevArch logo + title (hidden on mobile)
- **Center:** Primary nav links (hidden on mobile, medium+ screens)
- **Right:** Settings link, theme toggle, mobile menu button

## Navigation Items
- Primary items: fetched from `navItems` config (excludes settings)
- Settings link: separate, hidden on mobile
- Both use TanStack Router `Link` component with activeProps

## Active Link Styling
- **Previous:** `activeProps={{ className: 'bg-accent' }}`
- **Current:** `activeProps={{ className: 'bg-accent text-accent-foreground' }}`
  - Added foreground color for better contrast on light/dark themes

## Features
- **Theme Toggle:** Button cycles light/dark based on `resolvedTheme`
- **Mobile Menu:** Hamburger button calls `onMenuClick` callback
- **Responsive:**
  - Logo title hidden on mobile
  - Nav links hidden below md breakpoint
  - Settings link hidden below md breakpoint
  - Divider separator shown on mobile only

## Dependencies
- `Link` — React Router navigation
- `Button` — UI button with icon support
- `useTheme()` — Theme context hook
- `navItems` — Nav config

## Props
```typescript
{ onMenuClick?: () => void }
```

## Related Components
- Nav drawer (mobile menu implementation in parent layout)
