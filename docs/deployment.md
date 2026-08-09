# Zerops deployment

## Service topology

| Service | Runtime | Public | Purpose |
| --- | --- | --- | --- |
| `web` | Zerops Static | Yes | React production bundle |
| `api` | Go build, Alpine runtime | Yes | HTTP API and analyzer orchestration |
| `db` | Managed PostgreSQL 16 | No | Scenarios and immutable analyses |
| `zcp` | Control workspace | No | Development and deployment operations |

## Build and runtime configuration

The root [`zerops.yaml`](../zerops.yaml) defines separate `web` and `api` setups. The web build uses `npm ci`, tests, lint, and Vite build; only `frontend/dist` is deployed to Static. `VITE_API_BASE_URL` is injected during the build because Vite embeds it in the bundle.

The API build runs Go tests and vet, then produces a static Linux binary. Runtime configuration includes `PORT`, `APP_ENV`, `DATABASE_URL`, `CORS_ALLOWED_ORIGINS`, `ENGINE_VERSION`, and `LOG_LEVEL`. Database credentials remain Zerops references and are never stored in Git.

## Deployment order

1. Verify backend and frontend locally.
2. Deploy the API and wait for `/api/v1/health`.
3. Check Goose status and apply forward migrations explicitly from a private-network context.
4. Verify scenario creation, analysis, retrieval, 404, and 422 behavior.
5. Build and deploy the web service with the final API origin.
6. Verify direct SPA routes, CORS, browser console, and persisted workflow.

## Networking and CORS

Only the HTTPS web and API origins are public. The API allowlist contains the exact web origin; wildcard and credentialed CORS are disabled. PostgreSQL has no public endpoint.

## Health, logs, and persistence

The API health probe checks HTTP and database readiness. Structured access logs record method, path, status, duration, and request ID without connection details. Persistence verification retrieves an analysis, restarts only the API process, and retrieves the identical UUID and response afterward.

## Troubleshooting

- Build failure: inspect the failed build stage before redeploying.
- API starts but health fails: inspect database reference and query timeout configuration without printing values.
- Web requests localhost: correct the web build variable and rebuild; runtime-only changes cannot alter a Vite bundle.
- Direct route returns 404: verify Static SPA fallback and deployed artifact root.
- CORS rejection: compare the exact requesting origin with the configured allowlist.
