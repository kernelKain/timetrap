# TimeTrap

**Find the minute trust outlives truth.**

TimeTrap is a design-time verifier that finds intervals where cached entitlements, sessions, tokens, or other authorization copies remain valid after their source authorization has been revoked.

> Project status: Phase 0 — MVP scope and demonstration design.

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