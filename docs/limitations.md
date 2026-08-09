# Limitations

- **Finite horizon:** a safe verdict applies only through the configured simulation horizon.
- **Minute resolution:** events occur at integer minutes; clock skew, sub-minute latency, and distributed races are not modeled.
- **Simplified state:** each timed object is reduced to valid or invalid. Claims, scopes, audiences, partial permissions, and probabilistic behavior are outside the model.
- **Fixed vocabulary:** supported events are `issue`, `refresh`, `revoke`, and `expire`. Supported invariants are revoked-access grace, dependent-not-outlive-source, and maximum validity.
- **Model quality:** an incomplete or inaccurate scenario can produce a technically correct but irrelevant result.
- **Unknown state:** malformed references, contradictory events, or invalid models are rejected rather than assumed safe.
- **No live inspection:** TimeTrap analyzes submitted designs; it does not observe traffic or automatically verify a deployed application.
- **No production mutation:** remediation changes only a corrected scenario copy inside TimeTrap.
- **Shareable UUIDs:** result URLs are public bearer-style links. There are no authentication, ownership, privacy, team, or access-control features.
- **Single-product workflow:** collaboration, organization management, and large-scale batch analysis are not currently implemented.
