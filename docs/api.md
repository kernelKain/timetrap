# TimeTrap API

The public API base is `https://api-2b33-8080.prg1.zerops.app/api/v1`. JSON requests reject unknown fields.

| Method | Path | Success | Purpose |
| --- | --- | --- | --- |
| `GET` | `/health` | `200` | API and database readiness |
| `POST` | `/scenarios/validate` | `200` | Validate without persistence |
| `POST` | `/scenarios` | `201` | Validate and persist a scenario |
| `GET` | `/scenarios/{scenarioId}` | `200` | Retrieve a scenario |
| `PUT` | `/scenarios/{scenarioId}` | `200` | Replace a scenario definition |
| `POST` | `/scenarios/{scenarioId}/analyses` | `201` | Analyze and persist a snapshot |
| `GET` | `/analyses/{analysisId}` | `200` | Retrieve a persisted analysis |

## Scenario request

```json
{
  "name": "Subscription cancellation",
  "horizonMinutes": 90,
  "objects": [
    {"clientId":"subscription","name":"Subscription","kind":"subscription","events":[{"type":"issue","atMinute":0},{"type":"revoke","atMinute":10}]},
    {"clientId":"premium-cache","name":"Premium cache","kind":"cached_entitlement","events":[{"type":"issue","atMinute":0},{"type":"expire","atMinute":60}]}
  ],
  "invariants": [
    {"type":"revoked_access_grace","sourceObjectId":"subscription","dependentObjectId":"premium-cache","graceMinutes":5}
  ]
}
```

Persisted-resource responses wrap the resource and include `requestId`, for example `{ "scenario": { ... }, "requestId": "..." }` or `{ "analysis": { ... }, "requestId": "..." }`. An analysis includes its UUID, scenario UUID, immutable scenario snapshot, result, engine version, and timestamp.

## Errors

Errors use a stable envelope:

```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "The scenario is invalid.",
    "fields": [{"field":"objects[0].events[1].atMinute","message":"must be within the scenario horizon"}]
  },
  "requestId": "..."
}
```

Common statuses are `400` invalid JSON, `404` missing resource, `405` wrong method, `422` invalid domain model, `500` internal/storage failure, and `503` unavailable database health. Responses never expose raw internal or database errors.
