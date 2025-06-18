# DevArch Enhancement Roadmap - Actionable Steps

## 🎯 Phase 1: Individual Service Management (High Priority)

### Goal: Replace manual `sudo podman compose -f path_to_file down/build/up -d` with smart commands

### 1.1 Create Enhanced Service Controller Script
**File:** `scripts/service-manager.sh`

**Commands to implement:**
```bash
./scripts/service-manager.sh up adminer              # Start single service
./scripts/service-manager.sh down postgres           # Stop single service  
./scripts/service-manager.sh rebuild nginx           # Rebuild single service
./scripts/service-manager.sh restart grafana         # Restart single service
./scripts/service-manager.sh logs -f postgres        # Follow logs
./scripts/service-manager.sh status database         # Status of category
./scripts/service-manager.sh ps                      # List all services
```

**Features:**
- Auto-resolve service file paths using existing `resolve_service_path()`
- Support both individual services and categories
- Smart compose command building (leverage existing logic)
- Integration with current error handling and status functions

### 1.2 Enhance start-services.sh and stop-services.sh
**Goal:** Stream individual service operations within category workflows

**New capabilities:**
```bash
# Enhanced start-services with individual control
./scripts/start-services.sh --services adminer,postgres,grafana
./scripts/start-services.sh --categories database --except redis
./scripts/start-services.sh --rebuild --services nginx

# Enhanced stop-services with smart cleanup  
./scripts/stop-services.sh --services postgres --preserve-volumes
./scripts/stop-services.sh --categories analytics --cleanup-images
```

---

## 🧹 Phase 2: Smart Cleanup System (Medium Priority)

### Goal: Replace "wipe everything" with intelligent cleanup

### 2.1 Create Intelligent Cleanup Functions
**Add to:** `config.sh`

**Smart cleanup categories:**
- **Service-specific cleanup:** Only remove containers/images/volumes for specified services
- **Category-based cleanup:** Clean up entire service categories  
- **Orphan cleanup:** Remove unused containers/images/volumes not managed by compose files
- **Age-based cleanup:** Remove containers/images older than X days
- **Size-based cleanup:** Remove largest unused volumes when disk space low

### 2.2 Enhance Existing Cleanup Flags
**Files:** `stop-services.sh`, `service-manager.sh`

**New smart flags:**
```bash
--cleanup-service-images     # Only remove images for stopped services
--cleanup-service-volumes    # Only remove volumes for stopped services  
--cleanup-orphans           # Remove unmanaged containers/images/volumes
--cleanup-older-than 7d     # Remove resources older than 7 days
--preserve-data             # Never touch named volumes with data
```

---

## 🔐 Phase 3: SSL/Certificate Management Overhaul (Medium Priority)

### Goal: Simplify and modernize certificate management

### 3.1 Decision Point: Certificate Strategy
**Option A: Enhanced setup-ssl.sh (Recommended)**
- Check for and install `mkcert` automatically
- Use `mkcert` for local development certificates  
- Fallback to OpenSSL if mkcert unavailable
- Remove `trust-host.sh` entirely (mkcert handles trust automatically)

**Option B: Traefik Let's Encrypt Integration**
- Configure Traefik to handle Let's Encrypt certificates automatically
- Remove both `setup-ssl.sh` and `trust-host.sh`
- Add Traefik ACME configuration templates

### 3.2 Recommended: Option A Implementation
**Enhance:** `setup-ssl.sh`

**New capabilities:**
- Auto-detect and install `mkcert` on Linux/macOS/WSL
- Use `mkcert` for certificate generation and automatic trust installation
- Fallback to current OpenSSL method if mkcert fails
- Remove dependency on `trust-host.sh`
- Add certificate renewal checking

**Remove:** `trust-host.sh` (759 lines → 0 lines)

---

## 🛠️ Phase 4: Native Podman Command Integration (High Priority)

### Goal: Make scripts feel like enhanced native Podman commands

### 4.1 Create Podman Command Wrapper
**File:** `scripts/podman-wrapper.sh` or integrate into `service-manager.sh`

**Enhanced commands that mirror native Podman:**
```bash
# Native: podman compose -f file.yml up -d service
# Enhanced: ./service-manager.sh up service
# Result: Auto-finds file, adds smart options, better feedback

# Native: podman compose -f file.yml down
# Enhanced: ./service-manager.sh down service --preserve-data
# Result: Smart cleanup, preserves important volumes

# Native: podman compose -f file.yml build
# Enhanced: ./service-manager.sh rebuild service --no-cache
# Result: Cleans up old images, optimizes build
```

### 4.2 Add Smart Defaults and Context Awareness
- **Auto-detect service context** (if run from apps/service-name/, auto-target that service)
- **Preserve important data** by default (databases, uploads, etc.)
- **Smart resource management** (cleanup unused images after builds)
- **Better progress feedback** than native commands

---

## 📋 Implementation Priority & Dependencies

### Sprint 1 (Week 1): Individual Service Management
1. ✅ Create `service-manager.sh` with basic up/down/rebuild for individual services
2. ✅ Test with existing services (adminer, postgres, nginx)
3. ✅ Update documentation and help text

### Sprint 2 (Week 2): Enhanced Start/Stop Scripts  
1. ✅ Add individual service support to `start-services.sh`
2. ✅ Add smart cleanup options to `stop-services.sh`
3. ✅ Maintain backward compatibility with existing category-based workflows

### Sprint 3 (Week 3): SSL Simplification
1. ✅ Implement mkcert integration in `setup-ssl.sh`
2. ✅ Remove `trust-host.sh`
3. ✅ Update `install.sh` to remove trust-host references

### Sprint 4 (Week 4): Smart Cleanup System
1. ✅ Implement intelligent cleanup functions
2. ✅ Add age-based and size-based cleanup options
3. ✅ Test cleanup scenarios without data loss

---

## 🔧 Technical Integration Points

### Files That Need Updates:
- `config.sh` - Add service resolution and cleanup functions
- `start-services.sh` - Add individual service support  
- `stop-services.sh` - Add smart cleanup options
- `install.sh` - Remove trust-host.sh references
- `setup-ssl.sh` - Add mkcert integration
- Documentation files - Update command examples

### New Files to Create:
- `service-manager.sh` - Individual service management
- Potentially `podman-wrapper.sh` if we want deeper integration

### Files to Remove:
- `trust-host.sh` (after SSL overhaul)
- `setup-podman-socket.sh` (superseded by socket-manager.sh)

---

## 🎯 Success Criteria

### Developer Experience Goals:
1. **Single service operations** should be as easy as category operations
2. **Cleanup operations** should be safe by default, destructive by choice
3. **SSL setup** should be one command with automatic trust installation
4. **All operations** should feel like enhanced native Podman commands
5. **Backward compatibility** maintained for existing workflows

### Technical Goals:
1. **Reduce script count** by ~2-3 files
2. **Reduce total lines of code** while adding functionality
3. **Improve error handling** and user feedback
4. **Maintain existing architecture** patterns and conventions

Ready to start with Sprint 1? Which aspect would you like to tackle first?