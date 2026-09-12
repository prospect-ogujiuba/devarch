# REC-002: Laravel API plus TanStack Start recipe

- Status: planned
- Plan revision: 2
- Spec: r2
- Dependencies: `REC-001`

## Goal and observable behavior

Ship the first multi-app recipe as separately routable `apps/showcase-laravel-tanstack-api` and `apps/showcase-laravel-tanstack-web`, using Laravel API/Sanctum and TanStack Start/Router/Query with coordinated local URLs.

## Scope

Add a current TanStack Start JavaScript definition through the existing JavaScript bootstrap model with support tier, compatibility probe, and resolved-version evidence; define recipe nodes/edge; configure frontend API URL and Laravel CORS/Sanctum stateful domains through reviewed allowlisted adapter-owned non-secret outputs; add start/stop/status/resume/verify and documentation.

## Non-goals

No monorepo, SSR authentication guarantee beyond documented Sanctum flow, external identity provider, deployment topology, or generic code templating.

## Expected files

TanStack JavaScript framework/profile definition and tests; one recipe definition; adapter-owned configuration handlers; recipe integration tests/docs.

## Acceptance criteria

- AC-10 passes.
- Both apps are ignored, independently routable, reproducibly named, and reported as one recipe.
- Browser/API origin configuration is explicit and contains no credential.
- Verification proves frontend HTTPS, API health/JSON response, allowed local CORS origin, and recovery after injected second-node failure.

## Verification

Red offline definition/edge/config fixture and hostile-output rejection; scaffold dry-run; injected failure/atomic recovery/resume. Then opt-in uniquely named bounded real creation and HTTPS/API smoke with prerequisites, preserved logs, and cleanup instructions.

## Rollback

Remove definition/config handlers; generated apps remain separable and recoverable.
