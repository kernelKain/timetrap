# Local development

## Prerequisites

- Go 1.25
- Node.js 20.19+ or 22.12+
- npm
- PostgreSQL 16
- Goose v3

## Environment

Copy the example file to an ignored local file:

```bash
cp .env.example .env
```

Set `DATABASE_URL` for a local database, `CORS_ALLOWED_ORIGINS=http://localhost:5173`, and `VITE_API_BASE_URL=http://localhost:8080`. Never commit the populated file.

## Database and migrations

Create a local database, then run Goose with connection details supplied only through process environment variables:

```bash
(
  set +x
  set -a; source .env; set +a
  export GOOSE_DRIVER=postgres
  export GOOSE_DBSTRING="$DATABASE_URL"
  goose -dir backend/migrations status
  goose -dir backend/migrations up
)
```

Use only forward migrations on persistent databases. Integration tests require `TEST_DATABASE_URL` whose database name clearly contains `test`.

## Start the application

```bash
go -C backend run ./cmd/api
npm --prefix frontend ci
npm --prefix frontend run dev
```

The defaults are API `http://localhost:8080` and frontend `http://localhost:5173`.

## Verification

```bash
go -C backend test ./...
go -C backend vet ./...
npm --prefix frontend run test
npm --prefix frontend run lint
npm --prefix frontend run build
```

## Troubleshooting

- Health unavailable: check PostgreSQL connectivity and `DATABASE_URL`; do not print the value.
- Browser CORS failure: make `CORS_ALLOWED_ORIGINS` exactly match the frontend origin without a trailing slash.
- Frontend uses the wrong API: `VITE_API_BASE_URL` is a build-time variable; rebuild after changing it.
- Direct result route fails under a custom server: configure SPA fallback to `index.html`.
