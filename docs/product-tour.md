# TimeTrap product tour

This walkthrough verifies whether cached premium access ends soon enough after its subscription authority is revoked.

## Open the canonical model

Open [TimeTrap](https://web-2b33.prg1.zerops.app) and select **Try subscription cancellation**. The builder loads a fresh editable copy containing:

- a 90-minute horizon;
- `subscription`: issue at minute 0, revoke at minute 10;
- `premium-cache`: issue at minute 0, expire at minute 60;
- `revoked_access_grace`: the cache must end within five minutes of subscription revocation.

The builder timeline is an orientation preview, not a verdict. The Go API remains authoritative.

## Run the analysis

Select **Save and analyze**. TimeTrap first persists the scenario, then runs and stores an analysis for that scenario. The browser opens a durable `/results/{analysisId}` route.

The [persisted unsafe example](https://web-2b33.prg1.zerops.app/results/d28ecf3d-e28a-4538-a74c-d88d791d57fe) shows:

- source revocation at minute 10;
- grace deadline at minute 15;
- dependent expiration at minute 60;
- total stale exposure `[10,60)` — 50 minutes;
- policy violation `[15,60)` — 45 minutes.

The first five stale minutes are permitted by policy. Therefore, 50 minutes is the exposure duration, while 45 minutes is the violation duration.

## Read the result

The verdict and interval appear first. The shared-scale timeline aligns source and dependent events. A hatched overlay marks only the server-reported violation. The evidence section lists the objects, invariant, relevant event boundaries, and exact diagnostic values.

## Apply the correction

The structured remediation proposes adding a cache revoke event at minute 10. **Apply fix and rerun** creates a corrected scenario copy; it does not edit the original historical model.

The [persisted corrected example](https://web-2b33.prg1.zerops.app/results/1998c072-7095-4d5f-a45f-7a520a3d62c8) is safe within the same 90-minute horizon. It retains the later expiry event while active revocation ends cache validity immediately. Refreshing either result URL loads its stored server snapshot directly.
