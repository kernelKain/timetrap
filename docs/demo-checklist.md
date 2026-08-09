# TimeTrap demo verification

- Date: 2026-08-09
- Starting commit: `8957bf39409b31d86d757f74e44cf506d0859a2a`
- Working tree: Intentional unrelated staged file `stagedn'` present and preserved; otherwise clean before Phase 8.
- Go version: Go 1.25.0 module toolchain (host launcher Go 1.22.2)
- Node version: v24.19.0
- npm version: 11.17.0
- PostgreSQL version: BLOCKED — no guarded disposable connection was available for inspection.
- Environment: Zerops ZCP workspace, Linux 7.0.14-zabbly+ x86_64
- Test database: BLOCKED — `TEST_DATABASE_URL` was not available; no shared database was used.
- Tester: Codex
- Verification method: Automated unit/static checks plus source-contract review. Persisted browser and database checks are explicitly blocked below.

Status values are PASS, FAIL, BLOCKED, SKIPPED, or MANUAL. A PASS means the check was executed successfully in this verification run.

## Automated verification

| Check | Status | Evidence |
| --- | --- | --- |
| `go test ./...` | PASS | All backend packages passed with an isolated temporary Go build cache. |
| `go test -count=20 ./internal/analyzer` | PASS | Twenty repeated package runs passed. |
| `go vet ./...` | PASS | No findings. |
| `go test -race ./...` | BLOCKED | The race-enabled build produced no result after several minutes in the constrained workspace and was interrupted; ordinary tests passed. |
| PostgreSQL integration test | BLOCKED | `TEST_DATABASE_URL` was unavailable; the build-tagged test safely skips and requires exactly `timetrap_test`. |
| `npm run test` | PASS | Five frontend test files passed, including Phase 8 validation paths. |
| `npm run lint` | PASS | ESLint completed without findings. |
| `npm run build` | PASS | TypeScript and Vite production build completed. |

## Canonical unsafe result

| Check | Status | Evidence |
| --- | --- | --- |
| Horizon 90; source issue 0/revoke 10; cache issue 0/expire 60; grace 5 | PASS | Fresh explicit Go fixture and exact regression assertions. |
| Deadline minute 15 | PASS | Exact analyzer assertion. |
| Total stale exposure `[10,60)` = 50 minutes | PASS | Exact structured-evidence assertion. |
| Policy violation `[15,60)` = 45 minutes | PASS | Exact half-open interval and duration assertions. |
| Unsafe with deterministic earliest violation | PASS | Full result assertions and twenty-repeat suite. |

## Corrected result

| Check | Status | Evidence |
| --- | --- | --- |
| Cache revoke at minute 10; expiry at minute 60 preserved | PASS | Fresh corrected fixture and frontend remediation test. |
| Analyzer returns safe, no violations, no earliest violation | PASS | Exact Go regression repeated in-process twenty times. |
| New persisted scenario and analysis IDs | BLOCKED | Requires the unavailable disposable PostgreSQL database. |
| Original unsafe persisted result remains unchanged | BLOCKED | Store test covers snapshot semantics, but a fresh Phase 8 database run was unavailable. |

## Persistence and direct links

| Check | Status | Reason |
| --- | --- | --- |
| Create → analyze → retrieve with PostgreSQL | BLOCKED | No guarded disposable database connection. |
| Pool/API restart persistence | BLOCKED | No guarded disposable database connection. |
| Historical snapshot after scenario update | BLOCKED | Existing build-tagged integration assertion was inspected but not executed. |
| Result refresh, direct link, and fresh context | BLOCKED | Requires a persisted local result on the disposable database. |
| Unknown analysis UUID state | MANUAL | Handler 404 is automated; browser presentation was not exercised in this environment. |

## Failure behavior

| Check | Status | Evidence or reason |
| --- | --- | --- |
| Invalid/trailing/unknown-field JSON | PASS | Handler tests assert `400 INVALID_JSON`, request IDs, JSON content type, and sanitized messages. |
| Invalid model field errors | PASS | Domain and handler tests assert `422 VALIDATION_FAILED` with stable field paths. |
| Missing scenario/analysis | PASS | Fake-repository handlers assert stable 404 codes. |
| Storage and analyzer failures | PASS | Handler tests assert stable safe envelopes without internal details. |
| Health unavailable/timeout safety | PASS | Injected health-check tests cover degraded responses, deadlines, and sensitive-error redaction. |
| API unavailable in browser | BLOCKED | No disposable end-to-end browser environment was available. |
| Database unavailable in browser | BLOCKED | Shared Zerops PostgreSQL was deliberately not stopped or reconfigured. |
| Slow network and duplicate submission | MANUAL | Source guards and unit behavior were inspected; browser network evidence was unavailable. |
| Clipboard denial | MANUAL | Fallback implementation was inspected; permission denial was not browser-exercised. |

## Responsive and accessibility

| Check | Status | Reason |
| --- | --- | --- |
| Desktop, tablet, 390×844, and 320×568 | MANUAL | Requires viewport-capable browser QA. |
| Keyboard-only workflow | MANUAL | Requires interactive browser QA. |
| 200% zoom | MANUAL | Requires interactive browser QA. |
| Reduced motion | PASS | Static CSS verification confirms the reduced-motion override remains present. |
| Status not conveyed only by color | PASS | Result source uses text and icons/shapes with status color. |

## Optional clean-clone verification

SKIPPED — dependency reinstallation and a fresh disposable PostgreSQL database were not available; the working repository was not destructively cleaned.
