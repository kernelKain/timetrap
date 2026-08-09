# Testing and verification

## Commands

```bash
go -C backend test ./...
go -C backend test -count=20 ./internal/analyzer
go -C backend vet ./...
npm --prefix frontend run test
npm --prefix frontend run lint
npm --prefix frontend run build
```

PostgreSQL integration coverage is build-tagged and runs only with a guarded `TEST_DATABASE_URL` whose database name contains `test`.

## Analyzer coverage

Tests cover the canonical unsafe and corrected-safe scenarios, event ordering, minute zero, horizon boundaries, zero grace, simultaneous expiry, ongoing violations, duplicate boundaries, missing references, refresh behavior, earliest-violation ordering, input immutability, normalized timelines, and structured remediation.

The canonical assertions are exact:

- source revoke `10`;
- grace deadline `15`;
- dependent expiry `60`;
- stale exposure `[10,60)` and `50` minutes;
- violation `[15,60)` and `45` minutes;
- unsafe original;
- safe corrected copy with no violations or earliest violation.

Twenty repeated analyzer runs compare deterministic results and ordering.

## HTTP and persistence coverage

Handler tests cover strict decoding, unknown fields, field-specific validation, request IDs, not-found resources, repository failures, analyzer failures, panic recovery, health degradation, CORS, and sensitive-error redaction. PostgreSQL integration tests cover create, retrieve, analyze, pool restart, foreign keys, and immutable historical snapshots.

## Frontend coverage

Native frontend tests cover request mapping, deterministic fingerprints, validation paths, timeline positioning, remediation immutability and conflicts, typed API success/errors, non-JSON failures, network errors, and aborts. Lint, TypeScript compilation, and the production Vite build run before deployment.

## Production smoke checks

Production verification includes API health, canonical unsafe evidence, corrected safe evidence, direct URL refresh, fresh browser context, unknown UUID behavior, browser console errors, exact CORS origin, and persistence after an API-only restart.
