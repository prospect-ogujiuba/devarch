# Codebase Concerns

**Analysis Date:** 2026-01-07

## Tech Debt

**Monolithic Shell Scripts:**
- Issue: `scripts/service-manager.sh` is 1,787 lines in single file
- Files: `scripts/service-manager.sh`
- Why: Grew organically as features added
- Impact: Hard to test, difficult to maintain, complex to debug
- Fix approach: Extract into modules (cli.sh, bulk-ops.sh, cleanup.sh)

**Duplicate Validation Logic:**
- Issue: `validateServiceName()` and `validateCategoryName()` duplicated
- Files: `scripts/config.sh` lines 42-58, `config/devarch/api/lib/shell.php` lines 42-58
- Why: PHP and shell need same validation
- Impact: Rules can diverge between systems
- Fix approach: Single source of truth (shell generates PHP validation or shared config)

**No API Authentication:**
- Issue: Dashboard API has no authentication
- Files: `config/devarch/api/public/index.php`
- Why: Designed as local development tool
- Impact: Cannot safely expose to network
- Fix approach: Add API key or session auth if network exposure needed

## Known Bugs

**No critical bugs detected during analysis.**

## Security Considerations

**Exposed Credentials in .env:**
- Files: `.env` (lines 11, 23, 31-32, 42, 45)
- Risk: Hardcoded passwords, GitHub token in plain text
- Current mitigation: File is gitignored
- Recommendations: Use secrets manager, rotate exposed tokens, never commit .env

**Unrestricted CORS:**
- Files: `config/devarch/api/public/index.php` line 8
- Risk: `Access-Control-Allow-Origin: *` allows any domain
- Current mitigation: None
- Recommendations: Restrict to specific origins or localhost only

**Unvalidated GET Parameters:**
- Files: `config/devarch/api/endpoints/apps.php` lines 17, 24, 36-37
- Files: `config/devarch/api/endpoints/containers.php` lines 17, 39, 51-52
- Files: `config/devarch/api/endpoints/logs.php` lines 31-33
- Risk: `$_GET` parameters used without sanitization
- Current mitigation: Some regex validation on container names
- Recommendations: Validate all input, whitelist allowed values

**Shell Command Invocation:**
- Files: `config/devarch/api/lib/shell.php` lines 176, 183
- Risk: `shell_exec()` for `whoami` and `id -u` without error handling
- Current mitigation: None
- Recommendations: Add error checking, use safer alternatives

## Performance Bottlenecks

**No Pagination on Container Listing:**
- Files: `config/devarch/api/endpoints/containers.php`
- Problem: Returns all containers in single response
- Measurement: Not measured, degrades with 50+ services
- Cause: All containers loaded, filtered in memory
- Improvement path: Add limit/offset pagination, server-side filtering

**Repeated Podman Queries:**
- Files: `config/devarch/api/lib/containers.php` lines 59-80
- Problem: Static cache only, no persistent caching
- Measurement: Each request re-queries if cache expired (30s TTL)
- Cause: No Redis/persistent cache wired
- Improvement path: Add Redis caching layer

## Fragile Areas

**WordPress Workflow Script:**
- Files: `scripts/wordpress/wp-workflow.sh`
- Why fragile: No `set -e`, errors suppressed with `|| true`
- Common failures: Silent failures on plugin install, database operations
- Safe modification: Add proper error handling before changes
- Test coverage: No automated tests

**Service Startup Order:**
- Files: `scripts/config.sh` lines 93-117
- Why fragile: Order defined in array, no dependency resolution
- Common failures: Services fail if dependencies not running
- Safe modification: Document dependencies in comments
- Test coverage: No automated tests

## Scaling Limits

**Docker Compose Per-Service:**
- Current capacity: ~50 concurrent services tested
- Limit: Unknown, potentially compose orchestration overhead
- Symptoms at limit: Slow startup, resource contention
- Scaling path: Consider Kubernetes for larger deployments

## Dependencies at Risk

**react-hot-toast (if used):**
- Risk: Unmaintained packages in WordPress ecosystem
- Impact: Security vulnerabilities, React compatibility
- Migration plan: Audit and update dependencies regularly

## Missing Critical Features

**No Test Coverage for Dashboard:**
- Problem: Zero React tests
- Current workaround: Manual testing
- Blocks: Cannot verify UI regressions
- Implementation complexity: Medium (add React Testing Library)

**No Integration Tests for Service Lifecycle:**
- Problem: No automated tests for service start/stop/restart
- Current workaround: Manual verification
- Blocks: Cannot verify orchestration correctness
- Implementation complexity: Medium (mock Podman, test scripts)

## Test Coverage Gaps

**Dashboard (React):**
- What's not tested: All components, hooks, utilities
- Risk: UI regressions undetected
- Priority: Medium
- Difficulty: Medium (requires React Testing Library setup)

**PHP API Endpoints:**
- What's not tested: All 16 endpoints in `config/devarch/api/endpoints/`
- Risk: API regressions, security issues
- Priority: High
- Difficulty: Medium (add PHPUnit, mock Podman)

**Shell Scripts:**
- What's not tested: `service-manager.sh`, `config.sh`
- Risk: Orchestration bugs, silent failures
- Priority: High
- Difficulty: Medium (use bats or shunit2)

**Service Backup/Restore:**
- What's not tested: `wp-workflow.sh` backup operations
- Risk: Data loss on failed restore
- Priority: High
- Difficulty: Medium (requires test environment)

## Documentation Gaps

**No API Documentation:**
- Files: `config/devarch/api/`
- Problem: No README, no endpoint documentation
- Risk: Developers must read code to understand API
- Recommendation: Add OpenAPI spec or markdown docs

**Service Category Dependencies Undocumented:**
- Files: `scripts/config.sh` lines 50-76
- Problem: 22 categories with 80+ services, no dependency docs
- Risk: Incorrect startup order, service failures
- Recommendation: Document which services depend on which

---

*Concerns audit: 2026-01-07*
*Update as issues are fixed or new ones discovered*
