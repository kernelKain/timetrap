# TimeTrap

**Find the minute trust outlives truth.**

TimeTrap is a design-time verifier that finds intervals where cached entitlements, sessions, tokens, or other authorization copies remain valid after their source authorization has been revoked.

> Project status: Phase 1 — runnable React frontend and Go API shells.

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

TimeTrap will not include a general-purpose temporal-logic language during the hackathon.

## Planned architecture

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

The PostgreSQL integration and temporal analyzer are intentionally deferred to later phases.

## Current implementation

Phase 1 provides:

- A React, TypeScript, Vite, and Tailwind frontend
- A standard-library Go HTTP API
- `GET /api/v1/health`
- Environment-controlled API URL and CORS origin
- Healthy, checking, and unavailable frontend states
- Structured API request and lifecycle logs
- Graceful API shutdown

No database, analyzer, authentication, scenario builder, or deployment logic is implemented yet.

## Requirements

- Node.js 20.19+ or 22.12+
- npm
- Go 1.22+

## Local development

Install frontend dependencies:

```bash
npm --prefix frontend install
```

Start the Go API:

```bash
go -C backend run ./cmd/api
```

The API defaults to `http://localhost:8080`.

Start the frontend in another terminal:

```bash
npm --prefix frontend run dev
```

Open `http://localhost:5173`. The page should change from `Checking API…` to `API healthy`.

Test the API directly:

```bash
curl http://localhost:8080/api/v1/health
```

Copy the safe example values from `.env.example` when environment customization is needed. Never commit `.env`, `.env.local`, `.mcp.json`, `.zcp/`, access tokens, or credentials.

## Remote ZCP development

The ZCP workspace already provides Browser VS Code, Codex, and project-scoped ZCP access.

If port `8080` is occupied, run the API on another port:

```bash
PORT=8082 go -C backend run ./cmd/api
```

Remote code-server development can use an ignored `frontend/.env.development.local` file to configure its `/absproxy/5173/` path, allowed host, and API proxy target. Workspace-specific hostnames must not be committed.

## Verification

Backend:

```bash
gofmt -w backend
go -C backend test ./...
go -C backend vet ./...
```

Frontend:

```bash
npm --prefix frontend run lint
npm --prefix frontend run build
```

## AI assistance

Codex assisted with Phase 0 planning, Phase 1 scaffolding, implementation guidance, debugging, and documentation. Zerops Control Plane was used to inspect the project and adopt the existing development and staging services. All commands and acceptance checks were manually run and reviewed.                         