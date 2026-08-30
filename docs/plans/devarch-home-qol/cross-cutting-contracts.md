# Cross-cutting implementation contracts

Created: 2026-08-29
Status: Planned prerequisite
Applies to: all DevArch Home quality-of-life plans

## Purpose

Resolve shared architecture, testing, security, performance, and rollout decisions once so feature phases do not invent incompatible approaches.

## 1. Incremental module boundaries

Do not continue growing `scripts/dashboard/static/dashboard.js` as one feature file and do not perform a speculative big-bang rewrite. Extract a boundary immediately before the first feature needs it:

- `scripts/dashboard/static/core/identity.mjs` — stable app/service/container identities and validation.
- `scripts/dashboard/static/core/preferences.mjs` — versioned favorites/recent storage.
- `scripts/dashboard/static/core/view-state.mjs` — collection filtering, sorting, and URL serialization.
- `scripts/dashboard/static/core/actions.mjs` — safe navigate/open/copy action descriptors and ranking inputs.
- `scripts/dashboard/static/core/router.mjs` — route parsing, history, and canonicalization.
- `scripts/dashboard/static/dashboard.js` — DOM composition and event wiring only; load it as an ES module once imports begin.

Backend collectors may remain in `server.py` until Git or relationship work begins. Before two collectors are added, extract pure inventory normalization/resolution into `scripts/dashboard/inventory.py`; keep HTTP serving in `server.py`. This is a refactor boundary, not a new service.

Each extraction must preserve behavior in a dedicated commit/slice and pass existing tests before feature behavior is added.

## 2. Test architecture

Use production-dependency-free test layers:

- Python: existing `unittest` suite for server routes, collectors, API serialization, timeouts, and Unix-socket behavior.
- JavaScript pure logic: Node's built-in `node:test` against `.mjs` modules; add `npm test` without introducing a browser framework solely for pure functions.
- Browser behavior: a small Playwright test suite is justified for navigation, history, dialog focus, mobile menu/filter disclosure, clipboard fallback, and no-polling assertions. Keep Playwright as a development dependency only and pin its version. If browser installation is unavailable locally, CI or an explicitly prepared verification environment must run these tests before a feature is marked complete; manual checks alone do not close browser-critical phases.
- HTTPS smoke: run against the installed user service and `https://devarch.test`; local certificate setup is a documented prerequisite.

Every phase identifies which layer proves each acceptance criterion. Tests must use temporary repositories/directories and inert subprocess doubles; they must never mutate live containers or app workspaces.

## 3. Security and data handling

- Render inventory, URL, and storage strings with DOM `textContent`/attribute APIs; never use `innerHTML` for dynamic data.
- Add/retain a restrictive response policy compatible with the static app: at minimum `default-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'`. Explicitly justify any additional source. No CDN scripts, inline executable script, or remote font dependency.
- Same-origin navigation accepts only validated dashboard paths. External web opens accept parsed `http:`/`https:` URLs and use `noopener,noreferrer`. The existing editor handoff is a separate non-palette action restricted to an encoded `vscode://file` workspace path; no arbitrary custom scheme is accepted.
- Clipboard payloads come only from typed action/command builders. Clipboard failure must expose selectable text without executing anything.
- Query strings are browser-history/shareable data. Never place absolute paths, container IDs, localStorage contents, Git data, or commands in URLs. Document that search text in `q` becomes part of a copied/history URL.
- localStorage contains identities, action kinds, and timestamps only—never inventory records, paths, URLs, commands, or labels.
- API additions are allowlisted and serialized through explicit response builders. Raw subprocess output, stderr, Git config, label maps, environment variables, and argv/operator arrays never cross the boundary. Plan 07 may serialize only the backend-rendered copy string plus bounded descriptor metadata; its internal argv segments remain server-side and no execution endpoint accepts the rendered string.
- The existing Host allowlist and localhost Nginx publication remain required.

## 4. Inventory and refresh consistency

- At most one inventory request may be active. Disable refresh while active and use an `AbortController` or request generation token so a stale response cannot overwrite newer state.
- Feature-local interactions never fetch inventory.
- If refresh fails, preserve the last successful inventory and show a non-blocking stale/error state; do not blank favorites, filters, relationships, or detail pages.
- Add top-level `schemaVersion: 1` and a bounded `capabilities` object before optional collectors land. During migration the client treats an absent version as legacy version 0. Additive fields remain optional to the client for one compatibility slice so static assets tolerate an older service during restart/deployment; incompatible changes require a version increment and explicit dual-read migration.
- One failed optional collector produces a stable bounded code in `capabilities` (for example `git: unavailable|partial|ready`) and optional per-entity codes, not failure of the base inventory. Human-readable raw stderr is never the capability contract.

## 5. Performance budgets

Measure on a warm local machine with 50 apps, 250 services, and 100 containers unless a phase specifies a larger synthetic set:

- Base inventory excluding Git: target ≤500ms, hard acceptance ≤1s.
- Inventory including Git: target ≤2s; hard total deadline ≤5s, after which incomplete Git entries return timeout capability states.
- Serialized inventory: target <1MB, hard cap 2MB; optional fields and relationship counts are bounded.
- Search/filter/sort rerender for 1,000 items: ≤100ms at p95 across 20 runs; pure transformation ≤50ms.
- Command palette ranking for 5,000 actions: ≤50ms at p95 across 20 runs.
- localStorage document: hard cap 10KB.

Record hardware/runtime, fixture size, median, and p95. “No noticeable delay” is not sufficient evidence.

## 6. Accessibility and responsive baseline

Target the current and previous major releases of Chromium, Firefox, and Safari on desktop plus current mobile Safari/Chrome. Use standards available across that baseline; do not carry an untested legacy fallback.

- Interactive targets are at least 44×44 CSS pixels on touch layouts.
- Visible focus, logical tab order, programmatic names, pressed/expanded/selected state, and Escape dismissal are mandatory.
- Dynamic counts use restrained live-region announcements; typing must not produce excessive speech.
- Honor `prefers-reduced-motion`; functionality cannot depend on animation.
- Verify 320, 375, 768, 1024, and ≥1280px widths, 200% zoom, keyboard-only use, and one screen-reader pass for dialog/filter/pinning additions.
- Focus restoration after dialogs/disclosures and focus retention after rerender are explicit browser tests.

## 7. Rollout and compatibility

- Implement one subphase at a time; do not combine multiple feature plans in one review slice.
- Preserve existing routes, API fields, manual refresh, and no-polling behavior throughout.
- Additive server fields land before dependent UI and remain optional for one slice.
- Client-only features require no runtime feature-flag system. Rollback is the previous static assets plus tolerant optional fields. Backend collectors need a single configuration/capability switch only when their phase identifies an operational rollback need.
- A canonical plan status changes only after its verification evidence exists. Record deviations and contract changes in these canonical docs.

## 8. Expected file ownership

Use these boundaries to keep phases surgical:

- Plan 01: `scripts/dashboard/static/core/identity.mjs`, `scripts/dashboard/static/core/preferences.mjs`, dashboard favorite/recent renderers, JavaScript tests, CSS/HTML as needed.
- Plans 02–03: `scripts/dashboard/static/core/view-state.mjs`, `scripts/dashboard/static/core/router.mjs`, collection controls, JavaScript/browser tests.
- Plan 04: `scripts/dashboard/static/core/actions.mjs`, palette markup/controller/styles, JavaScript/browser tests.
- Plan 05: Git collector module under `scripts/dashboard/`, inventory assembler integration, Python tests, app UI.
- Plan 06: Podman normalization/relationship resolver, inventory assembler integration, Python tests, relationship UI and the deferred runtime facet.
- Plan 07: backend command descriptor/quoting builder, action registry extension, command UI, Python/JavaScript/browser tests.

`scripts/dashboard/static/dashboard.css` is generated by Tailwind and updated whenever classes change. `scripts/dashboard/README.md` changes only in the verification/documentation phase for the feature. Avoid unrelated root README changes unless launch behavior changes.

## Cross-cutting definition of done

- Module/test boundary required by the feature exists and is covered before behavior expands.
- Security/data rules have explicit negative tests.
- Performance evidence uses the shared fixture and percentile format.
- Browser-critical behavior has automated Playwright evidence in an environment with a browser installed.
- No phase adds lifecycle execution, hidden mutation, remote access, or recurring inventory requests.
