# Phase 3: Service Instances - Research

**Researched:** 2026-02-03
**Domain:** Copy-on-write instance overrides with Go API + React dashboard
**Confidence:** HIGH

## Summary

Phase 3 requires implementing full copy-on-write override semantics for service instances. Research reveals that this project already has strong patterns to follow: the service handlers show separate override tables per resource type (ports, volumes, env_vars, etc.), and the dashboard already has editable components for these resources.

**Key findings:**
- Separate override tables (not EAV or JSONB) match existing codebase patterns and provide best query performance
- Go has no deep merge in stdlib; manual field-by-field merge for override resolution is standard approach
- React controlled forms with placeholder patterns already implemented in existing service detail components
- ON DELETE CASCADE for foreign keys ensures override cleanup when instances deleted

**Primary recommendation:** Extend existing service CRUD patterns to instance overrides. Use separate `instance_*` override tables mirroring `service_*` tables. Implement merge resolver as explicit function that combines template + overrides. Reuse dashboard's editable components with placeholder text showing template values.

## Standard Stack

### Core Libraries (Already In Use)

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| lib/pq | stable | PostgreSQL driver | Direct SQL with no ORM overhead, matches existing handlers |
| go-chi/chi | v5 | HTTP router | Already used throughout API, stable, fast |
| TanStack Router | v1 | React routing | File-based routing in dashboard, type-safe |
| TanStack Query | v5 | Server state | Used in dashboard for all API queries, caching built-in |
| Radix UI | current | Headless components | Used for shadcn/ui primitives |
| Zod | v3 | Schema validation | Client-side validation in dashboard |
| CodeMirror | v6 | Code editor | Already used for config files (service detail) |

### Supporting (For Merge Operations)

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| github.com/imdario/mergo | latest | Struct merging | IF complex nested merge needed (probably not) |
| encoding/json | stdlib | JSON marshaling | Effective config endpoint (template + overrides → JSON) |
| gopkg.in/yaml.v3 | v3 | YAML generation | Already used in compose generator |

### NOT Needed

- Deep copy libraries (jinzhu/copier, mitchellh/copystructure): instances reference templates by FK, not by copying data
- EAV libraries: existing codebase uses separate tables per resource type
- ORM (GORM, ent): existing codebase uses direct SQL with lib/pq

**Installation:**
```bash
# No new Go dependencies needed — use existing stack
# Dashboard dependencies already installed
```

## Architecture Patterns

### Database Schema: Separate Override Tables

**Pattern:** One override table per resource type, mirroring service tables

```sql
-- Instance metadata
CREATE TABLE service_instances (
    id SERIAL PRIMARY KEY,
    stack_id INTEGER REFERENCES stacks(id) ON DELETE CASCADE,
    instance_id VARCHAR(63) NOT NULL,
    template_service_id INTEGER REFERENCES services(id),
    description TEXT DEFAULT '',
    enabled BOOLEAN DEFAULT true,
    UNIQUE(stack_id, instance_id)
);

-- Override tables (separate per resource type)
CREATE TABLE instance_port_overrides (
    id SERIAL PRIMARY KEY,
    instance_id INTEGER REFERENCES service_instances(id) ON DELETE CASCADE,
    host_ip VARCHAR(45),
    host_port INTEGER,
    container_port INTEGER,
    protocol VARCHAR(8)
);

CREATE TABLE instance_volume_overrides (
    -- same pattern as service_volumes
);

CREATE TABLE instance_env_var_overrides (
    -- same pattern as service_env_vars
);

-- Repeat for: healthcheck, labels, domains, config_files
```

**Why this pattern:**
- Matches existing service table structure exactly
- Type safety: port numbers are integers, not strings in JSONB
- Efficient queries: no JSON extraction, standard indexes work
- Easy joins: instance → template → overrides in single query
- Foreign key constraints enforce referential integrity

**When overrides exist:** Query instance_port_overrides, fall back to template service_ports
**When no overrides:** Just return template values
**Merge logic:** Explicit Go function, not database views

### Config Merge Pattern: Explicit Resolution Function

**Pattern:** Template + Overrides → Effective Config

```go
type EffectiveConfig struct {
    Ports       []ServicePort
    Volumes     []ServiceVolume
    EnvVars     []ServiceEnvVar
    // ... all override types
}

func (h *InstanceHandler) ResolveEffectiveConfig(instance *ServiceInstance) (*EffectiveConfig, error) {
    // Load template service
    template, err := h.loadService(instance.TemplateServiceID)
    if err != nil {
        return nil, err
    }

    effective := &EffectiveConfig{}

    // Ports: overrides OR template (not merge, full replace per field type)
    overridePorts, _ := h.loadInstancePortOverrides(instance.ID)
    if len(overridePorts) > 0 {
        effective.Ports = overridePorts
    } else {
        effective.Ports = template.Ports
    }

    // Repeat for each resource type
    // EnvVars: merge by key (override wins on key collision)
    templateEnv := h.indexEnvByKey(template.EnvVars)
    overrideEnv, _ := h.loadInstanceEnvOverrides(instance.ID)
    for _, override := range overrideEnv {
        templateEnv[override.Key] = override
    }
    effective.EnvVars = h.envMapToSlice(templateEnv)

    return effective, nil
}
```

**Precedence rules:**
- Ports, Volumes: Full replacement (if ANY override exists, use ALL overrides, ignore template)
- EnvVars, Labels: Merge by key (override wins on collision, template provides missing keys)
- Healthcheck: Full replacement (override OR template, not partial merge)
- Dependencies: Template-only (Phase 3 spec says not user-editable)

**Why explicit over library:**
- Different merge semantics per resource type (replacement vs key-based merge)
- Clear control flow, no "magic" merge behavior
- Matches existing handler patterns (service.go has similar field access patterns)

### React Form Pattern: Controlled with Placeholder Template Values

**Pattern:** Show template value as placeholder, override value as actual value

```tsx
function PortOverrideEditor({ instancePort, templatePort }: Props) {
  const [override, setOverride] = useState(instancePort ?? null)
  const hasOverride = override !== null

  return (
    <div className={cn("border-l-2", hasOverride && "border-l-blue-500")}>
      <Input
        value={override?.host_port ?? ''}
        placeholder={templatePort.host_port.toString()}
        onChange={(e) => setOverride({ ...templatePort, host_port: parseInt(e.value) })}
      />
      {hasOverride && (
        <Button size="sm" onClick={() => setOverride(null)}>
          Reset to template
        </Button>
      )}
    </div>
  )
}
```

**Key patterns from existing codebase:**
- `/dashboard/src/routes/services/$name.tsx` already has tabbed detail page
- `/dashboard/src/components/services/editable-*.tsx` already implement save-on-click pattern
- Zod schemas already used for client validation
- Colored left border for visual distinction (reuse this pattern for overrides)

**State management:**
- TanStack Query for server state (fetching template + instance)
- React useState for form state (dirty overrides before save)
- Explicit save button triggers PUT to `/stacks/{name}/instances/{id}/*` endpoints

### API Route Structure

**Pattern:** Nested routes matching resource hierarchy

```
POST   /api/v1/stacks/{stack}/instances          # Create instance
GET    /api/v1/stacks/{stack}/instances          # List instances in stack
GET    /api/v1/stacks/{stack}/instances/{id}     # Get instance detail
PUT    /api/v1/stacks/{stack}/instances/{id}     # Update instance metadata
DELETE /api/v1/stacks/{stack}/instances/{id}     # Delete instance

PUT    /api/v1/stacks/{stack}/instances/{id}/ports       # Override ports
PUT    /api/v1/stacks/{stack}/instances/{id}/volumes     # Override volumes
PUT    /api/v1/stacks/{stack}/instances/{id}/env-vars    # Override env vars
PUT    /api/v1/stacks/{stack}/instances/{id}/healthcheck # Override healthcheck
PUT    /api/v1/stacks/{stack}/instances/{id}/labels      # Override labels
PUT    /api/v1/stacks/{stack}/instances/{id}/domains     # Override domains

GET    /api/v1/stacks/{stack}/instances/{id}/effective-config  # Merged config
```

**Why this structure:**
- Matches existing `/services/{name}/*` pattern exactly
- Scoped to stack context (instances can't exist without stack)
- Granular endpoints per override type (follows existing service handler pattern)

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Deep struct merging | Custom recursive merger | Manual field-by-field merge OR github.com/imdario/mergo if needed | Different semantics per field (replace vs merge), explicit control better than magic |
| Database migrations | Custom SQL runner | Existing migrate runner in `cmd/migrate` | Already works, supports up/down migrations |
| Form validation | Custom validators | Zod (client) + Go validation (server) | Already used in codebase, type-safe |
| JSON serialization | Manual marshaling | encoding/json stdlib + NullableJSON helper | Existing models.NullableJSON handles null properly |
| Container name generation | String concatenation | Existing container.ContainerName() function | Already handles naming convention |
| Soft delete | Custom deleted_at logic | Copy stack handler pattern (WHERE deleted_at IS NULL) | Consistent with existing soft delete |

**Key insight:** This codebase has mature patterns already. Instance overrides are "service CRUD but scoped to stack + template reference." Reuse, don't reinvent.

## Common Pitfalls

### Pitfall 1: Inconsistent Merge Semantics Across Resource Types

**What goes wrong:** Treating all overrides the same (e.g., always merge, or always replace) leads to confusing UX. User overrides one port expecting it to replace all template ports, but system merges instead → two ports bound to same host port → conflict.

**Why it happens:** Assuming config merge is always "deep merge" without considering resource-specific semantics.

**How to avoid:**
- Ports, Volumes: Full replacement semantics (any override → ignore all template values for that resource type)
- EnvVars, Labels: Key-based merge (override wins on key match, template provides missing keys)
- Healthcheck: Full replacement (can't have "half a healthcheck")
- Document merge behavior per resource type in API docs and UI help text

**Warning signs:**
- Port conflicts when instance starts
- Unexpected env vars from template appearing despite overrides
- User confusion about "why is template value still showing?"

### Pitfall 2: N+1 Query Problem When Loading Instance List

**What goes wrong:** Loading instances for stack detail page: query instances, then for each instance query template service, then for each instance query override counts → 1 + N + N queries.

**Why it happens:** Following naive ORM pattern or sequential handler logic.

**How to avoid:**
```sql
-- Single query with JOINs and aggregates
SELECT
    si.id, si.instance_id, si.template_service_id,
    s.name as template_name, s.image_name, s.image_tag,
    COUNT(DISTINCT ipo.id) + COUNT(DISTINCT ivo.id) +
    COUNT(DISTINCT ieo.id) as override_count
FROM service_instances si
JOIN services s ON s.id = si.template_service_id
LEFT JOIN instance_port_overrides ipo ON ipo.instance_id = si.id
LEFT JOIN instance_volume_overrides ivo ON ivo.instance_id = si.id
LEFT JOIN instance_env_var_overrides ieo ON ieo.instance_id = si.id
WHERE si.stack_id = $1
GROUP BY si.id, s.id;
```

**Warning signs:**
- Slow stack detail page with many instances
- Database connection pool exhaustion
- Logs showing repeated similar queries

### Pitfall 3: Not Using ON DELETE CASCADE for Override Tables

**What goes wrong:** Deleting instance leaves orphaned override rows in instance_port_overrides, instance_env_var_overrides, etc. Database grows unbounded, queries slow down, foreign key references prevent proper cleanup.

**Why it happens:** Forgetting to add ON DELETE CASCADE to foreign key constraints during migration.

**How to avoid:**
- ALWAYS add `ON DELETE CASCADE` to foreign keys in override tables:
```sql
CREATE TABLE instance_port_overrides (
    id SERIAL PRIMARY KEY,
    instance_id INTEGER REFERENCES service_instances(id) ON DELETE CASCADE,
    ...
);
```
- Test: create instance with overrides, delete instance, verify override rows deleted

**Warning signs:**
- Growing override table sizes despite deleting instances
- Foreign key constraint errors when trying to delete stacks
- Orphaned data in JOIN queries

### Pitfall 4: Template Deletion Breaking Instances

**What goes wrong:** User deletes template service that has active instances. Either deletion fails (foreign key constraint), or instances become invalid with NULL template_service_id.

**Why it happens:** Not considering cascade behavior or instance lifecycle when template changes.

**How to avoid:**
- Option A: RESTRICT deletion of templates with active instances
```sql
-- In instance handler, before allowing service deletion:
SELECT COUNT(*) FROM service_instances WHERE template_service_id = $1 AND enabled = true;
-- If count > 0, return error with instance list
```
- Option B: ON DELETE SET NULL + handle null templates gracefully
```sql
template_service_id INTEGER REFERENCES services(id) ON DELETE SET NULL
-- UI shows "Template deleted" badge, effective config uses overrides only
```
- Recommended: Option A for Phase 3 (simpler, clearer)

**Warning signs:**
- Null pointer panics when loading instance effective config
- Foreign key constraint errors when deleting services
- User confusion about why service won't delete

### Pitfall 5: Controlled Form State Out of Sync with Server

**What goes wrong:** User edits override field, navigation happens before save, changes lost. Or concurrent user edits same instance, later save overwrites earlier save.

**Why it happens:** React controlled forms hold local state, no "dirty state" tracking or optimistic updates.

**How to avoid:**
- Warn on navigation if form dirty (react-router beforeunload)
- Show save/cancel buttons prominently
- Disable navigation during save operation
- Use TanStack Query mutations with optimistic updates:
```tsx
const updatePorts = useMutation({
  mutationFn: (ports) => api.put(`/stacks/${stack}/instances/${id}/ports`, { ports }),
  onMutate: async (newPorts) => {
    await queryClient.cancelQueries(['instance', id])
    const prev = queryClient.getQueryData(['instance', id])
    queryClient.setQueryData(['instance', id], (old) => ({ ...old, ports: newPorts }))
    return { prev }
  },
  onError: (err, vars, context) => {
    queryClient.setQueryData(['instance', id], context.prev)
  },
})
```

**Warning signs:**
- User reports "lost my changes"
- Inconsistent data after save
- Multiple users stepping on each other's edits

### Pitfall 6: Placeholder Pattern Breaking Accessibility

**What goes wrong:** Using placeholder to show template value seems intuitive, but screen readers may not announce it, and low-contrast placeholder text fails WCAG guidelines.

**Why it happens:** Overloading placeholder for dual purpose (hint + default value display).

**How to avoid:**
- Use placeholder for template value BUT also show as helper text:
```tsx
<div>
  <Label>Host Port</Label>
  <Input
    placeholder={templatePort.toString()}
    value={overridePort ?? ''}
    aria-describedby="port-help"
  />
  <p id="port-help" className="text-muted-foreground text-xs">
    Template default: {templatePort}
  </p>
</div>
```
- Or use disabled input showing template value side-by-side with override input
- Ensure sufficient color contrast (WCAG AA: 4.5:1 for text)

**Warning signs:**
- Accessibility audit failures
- User confusion about what placeholder represents
- Screen reader users can't determine template value

## Code Examples

Verified patterns from existing codebase and best practices:

### Instance Handler Setup (Following Service Handler Pattern)

```go
// Source: api/internal/api/handlers/service.go pattern
type InstanceHandler struct {
    db              *sql.DB
    containerClient *container.Client
}

func NewInstanceHandler(db *sql.DB, cc *container.Client) *InstanceHandler {
    return &InstanceHandler{
        db:              db,
        containerClient: cc,
    }
}

func (h *InstanceHandler) Create(w http.ResponseWriter, r *http.Request) {
    stackName := chi.URLParam(r, "stack")

    var req struct {
        InstanceID        string `json:"instance_id"`
        TemplateServiceID int    `json:"template_service_id"`
        Description       string `json:"description"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Validate instance_id using existing pattern
    if err := container.ValidateName(req.InstanceID); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Get stack ID
    var stackID int
    err := h.db.QueryRow(`SELECT id FROM stacks WHERE name = $1 AND deleted_at IS NULL`, stackName).Scan(&stackID)
    if err == sql.ErrNoRows {
        http.Error(w, "stack not found", http.StatusNotFound)
        return
    }

    // Insert instance
    var instanceID int
    err = h.db.QueryRow(`
        INSERT INTO service_instances (stack_id, instance_id, template_service_id, description)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `, stackID, req.InstanceID, req.TemplateServiceID, req.Description).Scan(&instanceID)

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]int{"id": instanceID})
}
```

### Loading Instance with Template and Overrides

```go
// Source: api/internal/api/handlers/service.go loadServiceRelations pattern
func (h *InstanceHandler) Get(w http.ResponseWriter, r *http.Request) {
    stackName := chi.URLParam(r, "stack")
    instanceName := chi.URLParam(r, "instance")

    // Load instance metadata
    var instance models.ServiceInstance
    var stackID int
    err := h.db.QueryRow(`
        SELECT si.id, si.stack_id, si.instance_id, si.template_service_id,
               si.description, si.enabled, si.created_at, si.updated_at
        FROM service_instances si
        JOIN stacks s ON s.id = si.stack_id
        WHERE s.name = $1 AND si.instance_id = $2 AND s.deleted_at IS NULL
    `, stackName, instanceName).Scan(
        &instance.ID, &instance.StackID, &instance.InstanceID,
        &instance.TemplateServiceID, &instance.Description, &instance.Enabled,
        &instance.CreatedAt, &instance.UpdatedAt,
    )

    if err == sql.ErrNoRows {
        http.Error(w, "instance not found", http.StatusNotFound)
        return
    }

    // Load template service (reuse existing loadServiceByName)
    template, err := h.loadServiceByID(instance.TemplateServiceID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Load overrides (follow service ports pattern)
    h.loadInstanceOverrides(&instance)

    response := struct {
        Instance models.ServiceInstance `json:"instance"`
        Template models.Service         `json:"template"`
    }{
        Instance: instance,
        Template: *template,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func (h *InstanceHandler) loadInstanceOverrides(instance *models.ServiceInstance) {
    // Ports
    rows, _ := h.db.Query(`
        SELECT host_ip, host_port, container_port, protocol
        FROM instance_port_overrides WHERE instance_id = $1
    `, instance.ID)
    if rows != nil {
        defer rows.Close()
        for rows.Next() {
            var p models.ServicePort
            rows.Scan(&p.HostIP, &p.HostPort, &p.ContainerPort, &p.Protocol)
            instance.PortOverrides = append(instance.PortOverrides, p)
        }
    }

    // Repeat for volumes, env_vars, labels, domains, healthcheck, config_files
}
```

### Effective Config Endpoint

```go
// New endpoint: GET /stacks/{stack}/instances/{instance}/effective-config
func (h *InstanceHandler) EffectiveConfig(w http.ResponseWriter, r *http.Request) {
    stackName := chi.URLParam(r, "stack")
    instanceName := chi.URLParam(r, "instance")

    // Load instance + template + overrides (reuse Get logic)
    instance, template := h.loadInstanceWithTemplate(stackName, instanceName)

    // Merge logic: overrides win
    effective := &models.EffectiveConfig{
        ImageName:     template.ImageName,
        ImageTag:      template.ImageTag,
        RestartPolicy: template.RestartPolicy,
    }

    // Ports: use overrides if present, else template
    if len(instance.PortOverrides) > 0 {
        effective.Ports = instance.PortOverrides
    } else {
        effective.Ports = template.Ports
    }

    // EnvVars: merge by key (override wins on collision)
    envMap := make(map[string]models.ServiceEnvVar)
    for _, e := range template.EnvVars {
        envMap[e.Key] = e
    }
    for _, e := range instance.EnvVarOverrides {
        envMap[e.Key] = e  // Override wins
    }
    for _, e := range envMap {
        effective.EnvVars = append(effective.EnvVars, e)
    }

    // Healthcheck: override OR template (not merge)
    if instance.HealthcheckOverride != nil {
        effective.Healthcheck = instance.HealthcheckOverride
    } else {
        effective.Healthcheck = template.Healthcheck
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(effective)
}
```

### React Instance Detail Page (Following Service Detail Pattern)

```tsx
// Source: dashboard/src/routes/services/$name.tsx pattern
import { useParams } from '@tanstack/react-router'
import { useQuery, useMutation } from '@tanstack/react-query'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'

function InstanceDetailPage() {
  const { stack, instance } = useParams()

  const { data, isLoading } = useQuery({
    queryKey: ['instance', stack, instance],
    queryFn: () => api.get(`/stacks/${stack}/instances/${instance}`),
  })

  if (isLoading) {
    return <div>Loading...</div>
  }

  const { instance: inst, template } = data

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">{inst.instance_id}</h1>
          <p className="text-muted-foreground">
            Template: {template.name}
          </p>
        </div>
      </div>

      <Tabs defaultValue="info">
        <TabsList>
          <TabsTrigger value="info">Info</TabsTrigger>
          <TabsTrigger value="env">Environment</TabsTrigger>
          <TabsTrigger value="effective">Effective Config</TabsTrigger>
        </TabsList>

        <TabsContent value="info">
          <InstancePortOverrides
            template={template.ports}
            overrides={inst.port_overrides}
            onSave={(ports) => updatePorts.mutate(ports)}
          />
          {/* Similar components for volumes, healthcheck, labels, domains */}
        </TabsContent>

        <TabsContent value="env">
          <InstanceEnvOverrides
            template={template.env_vars}
            overrides={inst.env_var_overrides}
            onSave={(envVars) => updateEnvVars.mutate(envVars)}
          />
        </TabsContent>

        <TabsContent value="effective">
          <EffectiveConfigPreview stack={stack} instance={instance} />
        </TabsContent>
      </Tabs>
    </div>
  )
}
```

### Override Editor Component (Placeholder Pattern)

```tsx
// Reusable pattern for showing template value as placeholder
function PortOverrideEditor({ templatePorts, overridePorts, onChange }) {
  const [localOverrides, setLocalOverrides] = useState(overridePorts ?? [])
  const hasOverrides = localOverrides.length > 0

  const handleSave = () => {
    onChange(localOverrides)
  }

  const handleReset = () => {
    setLocalOverrides([])
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center justify-between">
          Ports
          {hasOverrides && <Badge>Overridden</Badge>}
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Show template ports as reference */}
        <div className="text-sm text-muted-foreground">
          Template ports: {templatePorts.map(p => `${p.host_port}:${p.container_port}`).join(', ')}
        </div>

        {/* Override editor */}
        <div className={cn("space-y-2", hasOverrides && "border-l-2 border-l-blue-500 pl-4")}>
          {(hasOverrides ? localOverrides : templatePorts).map((port, idx) => (
            <div key={idx} className="grid grid-cols-3 gap-2">
              <Input
                placeholder="127.0.0.1"
                value={port.host_ip}
                onChange={(e) => {
                  const updated = [...localOverrides]
                  updated[idx] = { ...port, host_ip: e.target.value }
                  setLocalOverrides(updated)
                }}
              />
              <Input
                placeholder={hasOverrides ? '' : port.host_port.toString()}
                value={hasOverrides ? port.host_port : ''}
                onChange={(e) => {
                  const updated = [...localOverrides]
                  updated[idx] = { ...port, host_port: parseInt(e.target.value) }
                  setLocalOverrides(updated)
                }}
              />
              <Input
                placeholder={hasOverrides ? '' : port.container_port.toString()}
                value={hasOverrides ? port.container_port : ''}
                onChange={(e) => {
                  const updated = [...localOverrides]
                  updated[idx] = { ...port, container_port: parseInt(e.target.value) }
                  setLocalOverrides(updated)
                }}
              />
            </div>
          ))}
        </div>

        <div className="flex gap-2">
          <Button onClick={handleSave}>Save</Button>
          {hasOverrides && (
            <Button variant="outline" onClick={handleReset}>
              Reset to Template
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| EAV tables for flexible config | Separate tables per resource + JSONB for sparse data | ~2020 with JSONB maturity | Better query performance, type safety, simpler queries |
| Deep merge libraries | Explicit merge logic per resource type | Ongoing (2024-2026) | Clearer semantics, fewer surprises, easier debugging |
| defaultValue in React forms | Controlled components with useState + placeholder | React 16.8+ (hooks) | More predictable state, better control flow |
| Uncontrolled forms | React Hook Form with controlled pattern | ~2021 with RHF v7 | Better validation, easier state management |
| Auto-save forms | Explicit save button | UX trend 2022+ | Clearer intent, prevents accidental edits |

**Deprecated/outdated:**
- EAV models in Postgres: use separate tables or JSONB
- Form libraries that mutate DOM (jQuery plugins): use controlled React components
- Global state for forms (Redux): use local state + server state library (TanStack Query)
- Manual SQL string building: use parameterized queries (already done in codebase)

## Open Questions

Things that couldn't be fully resolved:

1. **Override count badge performance**
   - What we know: Need to show override count per instance on stack detail list
   - What's unclear: COUNT() subquery per instance could be slow with many instances
   - Recommendation: Start with JOIN + COUNT aggregate (single query for all instances), optimize later if needed

2. **Template service deletion UX**
   - What we know: Need to prevent deleting templates with active instances
   - What's unclear: Should UI show instance list before deletion attempt, or error after?
   - Recommendation: Add `/services/{name}/instances/count` endpoint, show warning in service delete dialog if count > 0

3. **Effective config caching**
   - What we know: Effective config is computed on-demand (template + overrides)
   - What's unclear: Should we cache computed config, or always compute fresh?
   - Recommendation: Compute fresh for Phase 3 (simpler), cache later if performance issue (Phase 6 when containers actually start)

## Sources

### Primary (HIGH confidence)
- [DevArch codebase](file:///home/priz/projects/devarch) - Existing patterns in service.go, stack.go, routes.go, models.go, migrations
- [PostgreSQL Official Docs - Constraints](https://www.postgresql.org/docs/current/ddl-constraints.html) - Foreign key CASCADE behavior
- [React Official Docs - Input](https://react.dev/reference/react-dom/components/input) - Controlled component patterns

### Secondary (MEDIUM confidence)
- [CYBERTEC PostgreSQL - EAV Design](https://www.cybertec-postgresql.com/en/entity-attribute-value-eav-design-in-postgresql-dont-do-it/) - Why not to use EAV
- [Raz Samuel - PostgreSQL JSONB vs EAV](https://www.razsamuel.com/postgresql-jsonb-vs-eav-dynamic-data/) - Storage comparison showing 3x size difference
- [Medium - JSONB PostgreSQL's Secret Weapon](https://medium.com/@richardhightower/jsonb-postgresqls-secret-weapon-for-flexible-data-modeling-cf2f5087168f) - JSONB use cases
- [Heap - When to Avoid JSONB](https://www.heap.io/blog/when-to-avoid-jsonb-in-a-postgresql-schema) - JSONB limitations for analytics
- [github.com/imdario/mergo](https://pkg.go.dev/github.com/imdario/mergo) - Go struct merge library
- [shadcn/ui Form Docs](https://ui.shadcn.com/docs/components/form) - Form patterns with React Hook Form
- [React Hook Form - reset](https://www.react-hook-form.com/api/useform/reset/) - Form reset API
- [Docker Compose - Merge](https://docs.docker.com/compose/how-tos/multiple-compose-files/merge/) - Config merge precedence patterns
- [Helm Override Nested Values](https://oneuptime.com/blog/post/2026-01-17-helm-override-nested-values/view) - Lists replace, deep merge preserves, precedence order

### Tertiary (LOW confidence)
- [Builder.io - React Component Libraries 2026](https://www.builder.io/blog/react-component-libraries-2026) - UI library trends
- [DEV Community - Controlled vs Uncontrolled](https://dev.to/fazal_mansuri_/controlled-vs-uncontrolled-components-in-react-a-deep-dive-1hhb) - React form patterns

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Verified with existing package.json and go.mod, all libraries already in use
- Architecture: HIGH - Patterns extracted directly from existing codebase (service.go, stack.go, models.go)
- Pitfalls: HIGH - Based on PostgreSQL official docs + existing soft delete pattern in codebase
- Don't Hand-Roll: HIGH - Analyzed existing helpers (container.ContainerName, models.NullableJSON, migrate runner)

**Research date:** 2026-02-03
**Valid until:** 60 days (stable domain, unlikely to change significantly)

**Codebase analysis scope:**
- api/internal/api/handlers/service.go (1455 lines) - Service CRUD patterns
- api/internal/api/handlers/stack.go (857 lines) - Stack CRUD patterns
- api/pkg/models/models.go (365 lines) - Data model patterns
- api/migrations/013_stacks_instances.up.sql - Existing schema
- dashboard/src/routes/services/$name.tsx (300 lines) - Detail page pattern
- dashboard/src/routes/stacks/index.tsx (200 lines) - List page pattern
- dashboard/src/types/api.ts (347 lines) - Type definitions
