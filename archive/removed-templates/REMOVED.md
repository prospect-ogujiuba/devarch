# Removed Templates and Creation Scripts

**Removal Date**: 2025-12-05
**Reason**: Redundant with JetBrains IDE project creation capabilities

## Overview

DevArch previously included comprehensive application templates and creation scripts for multiple frameworks. These have been archived as JetBrains IDEs (PHPStorm, WebStorm, PyCharm, GoLand, Rider) now provide superior project scaffolding capabilities.

## What Was Removed

### Scripts

- **create-app.sh** - Main template orchestrator for all frameworks
  - Location: `scripts/create-app.sh`
  - Function: Interactive/CLI app creation from templates
  - Replacement: JetBrains IDE "New Project" wizards

### Templates

All framework templates except WordPress:

#### PHP Templates (archived)
- **laravel/** - Laravel framework template
  - Replacement: PHPStorm → New Project → Laravel
  - Port range: 8100-8199

- **vanilla/** - Generic PHP application template
  - Replacement: PHPStorm → New Project → PHP Empty Project
  - Manual structure creation

#### Node.js Templates (archived)
- **react-vite/** - React + Vite SPA template
  - Replacement: WebStorm → New Project → React App (Vite)
  - Port range: 8200-8299

- **nextjs/** - Next.js framework template
  - Replacement: WebStorm → New Project → Next.js
  - Port range: 8200-8299

- **vue/** - Vue.js SPA template
  - Replacement: WebStorm → New Project → Vue.js
  - Port range: 8200-8299

- **express/** - Express.js server template
  - Replacement: WebStorm → New Project → Express App
  - Port range: 8200-8299

#### Python Templates (archived)
- **django/** - Django framework template
  - Replacement: PyCharm → New Project → Django
  - Port range: 8300-8399

- **flask/** - Flask application template
  - Replacement: PyCharm → New Project → Flask
  - Port range: 8300-8399

- **fastapi/** - FastAPI template
  - Replacement: PyCharm → New Project → FastAPI
  - Port range: 8300-8399

#### Go Templates (archived)
- **gin/** - Gin framework template
  - Replacement: GoLand → New Project → Go (manual Gin setup)
  - Port range: 8400-8499

- **echo/** - Echo framework template
  - Replacement: GoLand → New Project → Go (manual Echo setup)
  - Port range: 8400-8499

#### .NET Templates (archived)
- **aspnet-core/** - ASP.NET Core template
  - Replacement: Rider → New Solution → ASP.NET Core Web Application
  - Port range: 8600-8699

### Documentation

- **TEMPLATES.md** - Template usage guide
  - Location: `docs/TEMPLATES.md`
  - Replacement: JetBrains IDE documentation + `docs/jetbrains/` guides

- **MIGRATION_GUIDE.md** - Guide for migrating apps to `public/` structure
  - Location: `docs/MIGRATION_GUIDE.md`
  - Replacement: No longer needed (JetBrains IDEs create correct structure)

## What Was Preserved

### WordPress Tooling (FULLY PRESERVED)

WordPress installation remains in DevArch because PHPStorm's WordPress project creation cannot handle:

1. **Custom Repository Integration**
   - Makermaker plugin from private repos
   - Makerblocks custom blocks plugin
   - Makerstarter custom theme
   - TypeRocket Pro v6 framework (mu-plugins)

2. **Galaxy Configuration Management**
   - Custom configuration system integration
   - Environment-specific WordPress setup

3. **Preset-Based Installation**
   - bare, clean, custom, loaded, starred presets
   - Automated plugin/theme installation from custom repos

### Preserved Files
- `scripts/wordpress/install-wordpress.sh` - WordPress installation with custom repo support
- `templates/php/wordpress/` - WordPress template (if needed for install script)
- All WordPress-specific utilities in `scripts/wordpress/`

## Migration Path

### For New Projects

Instead of `create-app.sh`, use JetBrains IDEs:

1. **Open IDE** (PHPStorm, WebStorm, PyCharm, GoLand, Rider)
2. **New Project** → Select framework
3. **Location**: `/home/fhcadmin/projects/devarch/apps/{app-name}`
4. **Configure** framework-specific options
5. **Follow** JetBrains setup guides in `docs/jetbrains/`

### For WordPress

WordPress workflow unchanged:

```bash
# Still works exactly as before
./scripts/wordpress/install-wordpress.sh my-wp-site \
  --preset clean \
  --title "My WordPress Site"
```

## Why This Change?

### Before (DevArch Templates)
- Maintained custom templates for 10+ frameworks
- Manual updates when frameworks change
- Basic scaffolding only
- Limited IDE integration

### After (JetBrains IDEs)
- Industry-standard project creation
- Auto-updated by JetBrains
- Full IDE integration (debugging, linting, etc.)
- Superior framework-specific features
- Better long-term maintenance

### DevArch Focus
DevArch now focuses on:
- Container orchestration
- Service management
- Database infrastructure
- Observability stack
- WordPress custom workflows (where JetBrains can't replace)

## Archive Contents

```
archive/removed-templates/
├── REMOVED.md                    (this file)
├── create-app.sh                 (main creation script)
├── TEMPLATES.md                  (template documentation)
├── MIGRATION_GUIDE.md            (migration guide)
└── templates/
    ├── README.md                 (template overview)
    ├── node/
    │   ├── react-vite/
    │   ├── nextjs/
    │   ├── vue/
    │   └── express/
    ├── python/
    │   ├── django/
    │   ├── flask/
    │   └── fastapi/
    ├── go/
    │   ├── gin/
    │   └── echo/
    └── dotnet/
        └── aspnet-core/
```

## Restoration (If Needed)

If you need to restore any archived template:

```bash
# Restore create-app.sh
cp archive/removed-templates/create-app.sh scripts/

# Restore specific template
cp -r archive/removed-templates/templates/node/react-vite templates/node/

# Restore documentation
cp archive/removed-templates/TEMPLATES.md docs/
```

## Questions

**Q: Can I still use the old create-app.sh?**
A: Yes, it's in `archive/removed-templates/create-app.sh`. Copy it back to `scripts/` if needed.

**Q: What if JetBrains IDE doesn't support my framework?**
A: Use archived templates or create manually. Consider contributing to JetBrains IDE support.

**Q: Will WordPress installation change?**
A: No. WordPress workflow fully preserved and enhanced with PHPStorm integration.

**Q: What about the `public/` directory standard?**
A: Still enforced. JetBrains IDEs create correct structure by default for most frameworks.

## References

- JetBrains IDE guides: `docs/jetbrains/`
- WordPress workflow: `docs/jetbrains/phpstorm-wordpress.md`
- DevArch architecture: `CLAUDE.md`
- App structure standard: `APP_STRUCTURE.md` (updated)
