# TimeTrap demonstration script

## Demonstration objective

A judge should understand within a few seconds:

- What authorization state became stale
- When the source was revoked
- When the policy violation began
- When the stale copy expired
- How long the violation lasted
- How the model can be corrected

## 30-second pitch

Revoking access in one database does not mean every cache, session or token immediately stops trusting it. TimeTrap lets backend developers model those timed authorization copies and define rules such as “premium access must end within five minutes of cancellation.” It then returns a deterministic counterexample showing exactly when that rule fails. In this example, a subscription is revoked at minute 10, but its cached entitlement remains valid until minute 60. TimeTrap makes that invisible timing bug visible before it reaches production.

## Frozen demonstration model

| Setting | Value |
| --- | --- |
| Scenario | Subscription cancellation leak |
| Horizon | 90 minutes |
| Source | Premium subscription |
| Source issued | Minute 0 |
| Source revoked | Minute 10 |
| Dependent | Premium entitlement cache |
| Dependent issued | Minute 0 |
| Dependent expires | Minute 60 |
| Rule | Access must end within 5 minutes of source revocation |

## Expected unsafe result

- Source revocation: minute 10
- Allowed grace: 5 minutes
- Policy deadline: minute 15
- Cache expiration: minute 60
- Stale-access interval: `[10, 60)`
- Total stale access: 50 minutes
- Policy-violation interval: `[15, 60)`
- Policy-violation duration: 45 minutes
- Verdict: unsafe

The spoken presentation must not call `[15, 60)` a 50-minute violation.

## Corrected model

Apply active invalidation:

```text
Premium entitlement cache: revoke at minute 10