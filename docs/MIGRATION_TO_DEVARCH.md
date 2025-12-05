# Migration to DevArch Command

## Summary

The DevArch service manager has been completely rewritten as a minimal, transparent wrapper around podman-compose. This eliminates over-engineering while preserving essential functionality.

## What Changed

### Old System (service-manager.sh)
- **1,666 lines** of complex bash code
- **898 lines** of configuration in config.sh
- Complex dependency management
- Parallel execution options
- Health check monitoring
- Extensive cleanup flags
- Service discovery with fallbacks
- Heavy abstraction layer

### New System (devarch)
- **~600 lines** of simple bash code
- **No external dependencies** (everything inline)
- Direct podman-compose execution
- Transparent command output (shows what runs)
- Minimal abstraction
- UX-friendly command names

## Philosophy Change

**Before:** Abstract away complexity, provide many options, automate everything

**After:** Simple transparent wrapper, show users what's happening, let podman do its job

## Command Migration Guide

### Basic Commands

| Old | New |
|-----|-----|
| `./scripts/service-manager.sh up postgres` | `devarch start postgres` |
| `./scripts/service-manager.sh down postgres` | `devarch stop postgres` |
| `./scripts/service-manager.sh restart postgres` | `devarch restart postgres` |
| `./scripts/service-manager.sh logs postgres --follow` | `devarch logs postgres -f` |
| `./scripts/service-manager.sh status` | `devarch status` |
| `./scripts/service-manager.sh ps` | `devarch ps` |
| `./scripts/service-manager.sh list` | `devarch list` |

### Bulk Commands

| Old | New |
|-----|-----|
| `./scripts/service-manager.sh start database` | `devarch start-db` |
| `./scripts/service-manager.sh start backend` | `devarch start-backend` |
| `./scripts/service-manager.sh start-all` | `devarch start-all` |
| `./scripts/service-manager.sh stop-all` | `devarch stop-all` |

### Container Access

| Old | New |
|-----|-----|
| `podman exec -it php bash` | `devarch exec php bash` |
| `podman exec -it php wp plugin list` | `devarch exec php wp plugin list` |

## Removed Features

These features were intentionally removed. Use podman/podman-compose directly for these operations:

### 1. Parallel Execution
**Old:**
```bash
./scripts/service-manager.sh start-all --parallel
```

**Why removed:** Sequential is fast enough and more predictable. If you need parallel, run multiple terminals.

**Alternative:** Run services in different terminals simultaneously

---

### 2. Health Check Waiting
**Old:**
```bash
./scripts/service-manager.sh start-all --wait-healthy --health-timeout 120
```

**Why removed:** Podman-compose handles health checks. Users can check logs if needed.

**Alternative:**
```bash
devarch start-all
devarch logs <service> -f  # Check if service is healthy
```

---

### 3. Force Recreate
**Old:**
```bash
./scripts/service-manager.sh up postgres --force
```

**Why removed:** Use podman-compose directly for advanced flags.

**Alternative:**
```bash
podman-compose -f compose/database/postgres.yml up -d --force-recreate
```

---

### 4. Rebuild with Cache Control
**Old:**
```bash
./scripts/service-manager.sh rebuild php --no-cache
```

**Why removed:** Rebuilding is advanced operation, use podman-compose directly.

**Alternative:**
```bash
podman-compose -f compose/backend/php.yml down
podman-compose -f compose/backend/php.yml build --no-cache
podman-compose -f compose/backend/php.yml up -d
```

---

### 5. Volume Removal Flags
**Old:**
```bash
./scripts/service-manager.sh down postgres --remove-volumes
```

**Why removed:** Volume management should be explicit and deliberate.

**Alternative:**
```bash
devarch stop postgres
podman volume rm postgres_data  # Explicit volume removal
```

Or:
```bash
podman-compose -f compose/database/postgres.yml down --volumes
```

---

### 6. Service-Specific Filtering
**Old:**
```bash
./scripts/service-manager.sh start --services postgres,redis
./scripts/service-manager.sh start database --except-services mysql
```

**Why removed:** Just start the services you want individually.

**Alternative:**
```bash
devarch start postgres
devarch start redis
```

---

### 7. Cleanup Operations
**Old:**
```bash
./scripts/service-manager.sh stop-all --cleanup-orphans
./scripts/service-manager.sh stop-all --cleanup-older-than 7d
./scripts/service-manager.sh stop-all --cleanup-large-volumes
```

**Why removed:** Cleanup should be explicit and separate from stopping services.

**Alternative:**
```bash
devarch stop-all

# Then cleanup explicitly
podman container prune
podman volume prune
podman image prune
podman system prune
```

---

### 8. Category Exclusions
**Old:**
```bash
./scripts/service-manager.sh start-all --exclude analytics,messaging
```

**Why removed:** Just start what you need.

**Alternative:**
```bash
devarch start-db
devarch start-backend
devarch start nginx-proxy-manager
# Don't start analytics or messaging
```

---

### 9. Automatic Dependency Resolution
**Old:**
- Script tracked dependencies
- Ensured services started in correct order
- Failed if dependencies missing

**Why removed:** Document the order instead. Users should understand dependencies.

**Alternative:**
- Use `devarch start-all` (respects documented order)
- Or start manually: db → proxy → backend → etc.
- See docs/SERVICE_MANAGER.md for startup order

---

## Benefits of New System

### 1. Transparency
Every command shows the exact podman operation:

```bash
$ devarch start postgres
🔄 Starting service: postgres
   → podman compose -f /home/fhcadmin/projects/devarch/compose/database/postgres.yml up -d
```

### 2. Simplicity
- No config.sh dependency
- No .env parsing in script
- Everything inline or auto-detected
- Easy to understand and modify

### 3. Learning
Users see and learn podman commands instead of hiding behind abstraction.

### 4. Maintainability
- ~600 lines vs ~2,500 lines
- Single file, no external dependencies
- Easy to modify and extend
- Less code = fewer bugs

### 5. Flexibility
For advanced operations, users can run podman-compose directly with full control.

## What Stayed

### Core Functionality
- Start/stop services
- View logs
- Container execution
- Status checking
- Service listing
- Network management

### Convenience
- start-db (all databases)
- start-backend (all runtimes)
- start-all (everything in order)
- stop-all (stop everything)

### Auto-detection
- Podman vs Docker
- Rootless vs rootful
- Sudo requirement
- Network creation

## File Changes

### New Files
- `/home/fhcadmin/projects/devarch/scripts/devarch` - New minimal manager
- `/home/fhcadmin/projects/devarch/docs/SERVICE_MANAGER.md` - Complete documentation

### Backed Up Files
- `/home/fhcadmin/projects/devarch/scripts/service-manager.sh.backup` - Old manager
- `/home/fhcadmin/projects/devarch/scripts/config.sh.backup` - Old config

### Updated Files
- `/home/fhcadmin/projects/devarch/CLAUDE.md` - Updated to use devarch commands

## Installation

### Add to PATH
```bash
# Option 1: Add to shell config
echo 'export PATH="$PATH:/home/fhcadmin/projects/devarch/scripts"' >> ~/.bashrc
source ~/.bashrc

# Option 2: Create symlink
sudo ln -s /home/fhcadmin/projects/devarch/scripts/devarch /usr/local/bin/devarch
```

Now you can run `devarch` from anywhere.

## Testing Migration

### 1. Test Help
```bash
devarch help
```

### 2. Test List
```bash
devarch list
```

### 3. Test Status
```bash
devarch status
```

### 4. Test Start/Stop
```bash
devarch start postgres
devarch logs postgres
devarch stop postgres
```

### 5. Test Bulk Operations
```bash
devarch start-db
devarch status
devarch stop-all
```

## Rollback

If you need to revert to the old system:

```bash
cd /home/fhcadmin/projects/devarch/scripts

# Restore old files
cp service-manager.sh.backup service-manager.sh
cp config.sh.backup config.sh

# Use old commands
./scripts/service-manager.sh status
```

## Getting Help

### Quick Reference
```bash
devarch help
```

### Full Documentation
- [docs/SERVICE_MANAGER.md](/home/fhcadmin/projects/devarch/docs/SERVICE_MANAGER.md) - Complete guide
- [CLAUDE.md](/home/fhcadmin/projects/devarch/CLAUDE.md) - Project overview

### Direct Podman Commands
The new system encourages learning podman. Every devarch command shows its podman equivalent.

Example:
```bash
devarch start postgres
# Shows: podman-compose -f compose/database/postgres.yml up -d

# You can run this directly too:
podman-compose -f compose/database/postgres.yml up -d
```

## FAQ

### Q: Why remove all the features?
**A:** Simplicity and transparency. Over-abstraction hides what's happening. Direct commands are clearer.

### Q: What if I need parallel startup?
**A:** Open multiple terminals and run `devarch start <service>` in each. Or use podman-compose directly.

### Q: How do I rebuild services now?
**A:** Use podman-compose directly. The script shows you the exact commands.

### Q: Can I still use the old service-manager.sh?
**A:** Yes, it's backed up. But the new system is recommended for simplicity.

### Q: What about volume management?
**A:** Use podman commands directly:
```bash
podman volume ls
podman volume rm <volume-name>
podman volume prune
```

### Q: How do I add a new service?
**A:**
1. Create compose file in appropriate category directory
2. Add service name to the script's `CATEGORY_SERVICES` array
3. Service available immediately

### Q: Does this work with Docker?
**A:** Yes, auto-detects and uses `docker compose`.

### Q: What happened to config.sh?
**A:** No longer needed. Everything is inline or auto-detected.

## Examples

### WordPress Development
```bash
# Old way
./scripts/service-manager.sh start database backend proxy
./scripts/service-manager.sh logs php --follow

# New way
devarch start-db
devarch start-backend
devarch start nginx-proxy-manager
devarch logs php -f
```

### Monitoring Stack
```bash
# Old way
./scripts/service-manager.sh start database exporters analytics --wait-healthy

# New way
devarch start-db
sleep 5
devarch start node-exporter
devarch start postgres-exporter
devarch start redis-exporter
devarch start prometheus
devarch start grafana
```

### Full Environment
```bash
# Old way
./scripts/service-manager.sh start-all --parallel --wait-healthy

# New way
devarch start-all  # Sequential, respects dependencies
```

## Conclusion

The new `devarch` command is simpler, more transparent, and easier to maintain. It removes abstraction that hid podman operations and encourages users to understand the tools they're using.

For most operations, devarch commands are sufficient. For advanced needs, use podman-compose directly with full control.
