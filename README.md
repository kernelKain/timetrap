# TimeTrap

**Find the minute trust outlives truth.**

TimeTrap is a design-time verifier that finds intervals where cached entitlements, sessions, tokens, or other authorization copies remain valid after their source authorization has been revoked.

> Project status: Phase 4 — deterministic analysis, persistent scenarios, and immutable analysis reports.

## The problem

Modern authorization state is rarely stored in only one place.

A subscription might be stored in PostgreSQL while access is also represented by:

- A cached premium entitlement
- An application session
- An access or refresh token
- A browser or CDN cache
- A third-party identity provider
- A delayed background synchronization process

These copies expire and refresh independently. Revoking the source record does not necessarily invalidate every previously issued copy.

Each component can appear correctly configured while the complete system still permits access for too long.

## What TimeTrap does

A user describes a bounded authorization model containing:

- Timed objects
- State-changing events
- A simulation horizon
- One or more predefined temporal invariants

TimeTrap deterministically evaluates the relevant boundary moments and returns:

- Whether the submitted model is safe or unsafe
- The earliest violation
- The exact violation interval
- The violation duration
- The objects and events responsible
- The invariant that failed
- A deterministic configuration-level remediation

The same input always produces the same result.

## Canonical demonstration

The primary demonstration models subscription cancellation:

| Item | Configuration |
| --- | --- |
| Simulation horizon | 90 minutes |
| Subscription | Issued at minute 0 and revoked at minute 10 |
| Cached entitlement | Issued at minute 0 and expires at minute 60 |
| Policy | Cached access must end within 5 minutes of revocation |
| Policy deadline | Minute 15 |
| Stale-access interval | `[10, 60)` |
| Total stale access | 50 minutes |
| Policy-violation interval | `[15, 60)` |
| Policy-violation duration | 45 minutes |

The five minutes between minute 10 and minute 15 are permitted by the configured grace period. Therefore, the policy violation lasts 45 minutes even though the cached authorization remains stale for 50 minutes in total.

The corrected scenario actively revokes the cached entitlement when the subscription is revoked. Running the same analysis then produces a safe result.

## MVP

The hackathon MVP will allow users to:

1. Load a built-in demonstration template.
2. Create or edit a bounded authorization timeline.
3. Add supported objects and events.
4. Select a predefined invariant.
5. Run deterministic analysis.
6. View the earliest violation on a multi-lane timeline.
7. Inspect evidence explaining the violation.
8. Apply a suggested configuration-level correction.
9. Rerun the analysis and verify the corrected model.
10. Open and share a persistent analysis URL.

## Supported model

### Objects

- Subscription
- Cached entitlement
- Session
- Token
- Invitation
- Credential

### Events

- Issue
- Refresh
- Revoke
- Expire

### Invariants

1. Revoked access must end within a configured grace period.
2. A dependent object must not outlive its source.
3. An object must not remain valid longer than a configured maximum.

TimeTrap does not include a general-purpose temporal-logic language during the hackathon.

## Architecture

```text
User browser
    |
    | HTTPS and JSON
    v
React + TypeScript frontend on Zerops
    |
    v
Stateless Go API on Zerops
    |                       |
    v                       v
Pure temporal analyzer   Managed PostgreSQL
                         over a private network
```

The analyzer is a pure Go package with no HTTP or database dependency.

PostgreSQL stores complete scenario definitions and immutable analysis snapshots. HTTP handlers depend only on repository interfaces and do not know about pgx or PostgreSQL implementation details.

## Current implementation

Phase 4 provides:

- A React, TypeScript, Vite, and Tailwind frontend shell
- A standard-library Go HTTP API
- Strict JSON decoding and bounded domain validation
- Request IDs and structured logs
- Panic recovery and CORS middleware
- Graceful HTTP and database shutdown
- A pure deterministic temporal analyzer
- Earliest half-open violation intervals
- Deterministic remediation suggestions
- PostgreSQL-backed scenario creation, retrieval, and updates
- Persistent analysis reports with immutable scenario snapshots
- Explicit Goose migrations
- Database-aware health checks
- Isolated handler tests using fake repositories
- Build-tagged PostgreSQL integration tests

Stored analysis reports use the scenario snapshot captured when analysis ran. Updating the scenario later does not alter an earlier report.

## API

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/api/v1/health` | Report API and database readiness |
| `POST` | `/api/v1/scenarios/validate` | Validate without storing a scenario |
| `POST` | `/api/v1/scenarios` | Validate and persist a scenario |
| `GET` | `/api/v1/scenarios/{scenarioId}` | Retrieve a stored scenario |
| `PUT` | `/api/v1/scenarios/{scenarioId}` | Replace a stored scenario definition |
| `POST` | `/api/v1/scenarios/{scenarioId}/analyses` | Analyze and persist a scenario snapshot |
| `GET` | `/api/v1/analyses/{analysisId}` | Retrieve a persistent analysis report |

### Error contract

API errors use a stable JSON envelope containing:

- A machine-readable error code
- A safe client-facing message
- Optional validation fields
- A request ID

Internal analyzer, PostgreSQL, hostname, SQL, and credential details are never returned to clients.

## Persistence

Scenarios are stored as JSONB alongside relational metadata.

Analysis records store:

- Analysis UUID
- Original scenario UUID
- Immutable scenario snapshot
- Safe or unsafe verdict
- Complete analysis result
- Analyzer engine version
- Creation timestamp

The snapshot ensures that an old analysis remains reproducible after its scenario is updated.

PostgreSQL migrations are explicit. The API never runs migrations automatically.

## Requirements

- Node.js 20.19+ or 22.12+
- npm
- Go 1.25+
- PostgreSQL 16
- Goose v3 as a separate migration CLI

Install Goose:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Ensure the Go binary directory is available:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

## Local development

Install frontend dependencies:

```bash
npm --prefix frontend install
```

Verify that local environment files are ignored:

```bash
git check-ignore .env
```

Expected:

```text
.env
```

Create an ignored local environment file:

```bash
cp .env.example .env
```

Replace the example database URL only inside the ignored `.env` file. Never commit a real connection string.

The backend configuration includes:

- `DATABASE_URL`
- `DATABASE_MAX_CONNS`
- `DATABASE_MIN_CONNS`
- `DATABASE_CONNECT_TIMEOUT`
- `DATABASE_QUERY_TIMEOUT`

### Apply migrations

Load the ignored environment inside a temporary subshell and apply migrations explicitly:

```bash
(
  set +x
  set -a
  source .env
  set +a

  export GOOSE_DRIVER=postgres
  export GOOSE_DBSTRING="$DATABASE_URL"

  goose -dir backend/migrations status
  goose -dir backend/migrations up
  goose -dir backend/migrations status
)
```

The environment variables disappear when the subshell exits.

Use `goose down` only with a disposable development or test database and only when intentionally destroying its tables.

### Start the API

Load the ignored local configuration and start the API:

```bash
(
  set +x
  set -a
  source .env
  set +a

  go -C backend run ./cmd/api
)
```

The development API uses:

```text
http://localhost:8082
```

Test database readiness:

```bash
curl http://localhost:8082/api/v1/health
```

A healthy response reports:

```json
{
  "status": "ok",
  "service": "timetrap-api",
  "database": "connected",
  "timestamp": "..."
}
```

If PostgreSQL is unavailable, health returns HTTP `503` with:

```json
{
  "status": "degraded",
  "service": "timetrap-api",
  "database": "unavailable",
  "timestamp": "..."
}
```

Database errors, credentials, and hostnames are never included.

### Start the frontend

In another terminal:

```bash
npm --prefix frontend run dev
```

Open:

```text
http://localhost:5173
```

The frontend reads its API URL from `VITE_API_BASE_URL`.

Never commit `.env`, `.env.local`, `.mcp.json`, `.zcp/`, access tokens, database credentials, or resolved Zerops references.

## Remote ZCP development

The ZCP workspace provides:

- Browser VS Code
- Codex
- Project-scoped ZCP access
- Private connectivity to managed PostgreSQL
- Development and staging Go services

Database credentials must be supplied through unresolved Zerops service references and temporary process environments. Do not print or persist resolved values.

The backend requires Go 1.25. Zerops `appdev` and `appstage` must use `alpine/golang@latest` before this module is deployed. Do not deploy it to the older `go@1.22` runtime.

Runtime upgrades and deployments require explicit approval.

PostgreSQL must remain private and must not receive a public subdomain.

Runtime-local files and process state are not persistent. PostgreSQL is the persistence boundary.

## Verification

### Backend

Format and run unit tests:

```bash
go -C backend fmt ./...
go -C backend test ./...
go -C backend vet ./...
```

Check patch formatting:

```bash
git diff --check
```

### PostgreSQL integration test

PostgreSQL integration tests are build-tagged and must use an isolated database named exactly `timetrap_test`.

Apply migrations explicitly to the isolated test database before running the test.

Supply `TEST_DATABASE_URL` securely through the process environment, then run:

```bash
go -C backend test \
  -tags=integration \
  ./internal/store/postgres \
  -run '^TestPostgresPersistenceIntegration$' \
  -count=1 \
  -v
```

The integration test:

- Skips when `TEST_DATABASE_URL` is absent
- Refuses to run unless the parsed database name is exactly `timetrap_test`
- Creates uniquely identified records
- Deletes only the scenario created by that test
- Uses the foreign key’s `ON DELETE CASCADE` behavior for its analysis
- Verifies persistence after closing and reopening the pool

Never run integration tests against production or the shared development database.

### Frontend

```bash
npm --prefix frontend run lint
npm --prefix frontend run build
```

## AI assistance

Codex assisted with phased planning, implementation guidance, debugging, test design, security review, and documentation.

Zerops Control Plane was used for read-only project discovery, approval-gated database creation, explicit migrations, and isolated integration-test execution.

Commands, infrastructure mutations, and acceptance checks were manually approved and reviewed.
