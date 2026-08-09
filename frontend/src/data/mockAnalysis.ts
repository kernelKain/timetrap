export const canonicalMockAnalysis = {
  verdict: 'Unsafe', sourceRevocation: 10, deadline: 15, expiry: 60,
  staleInterval: '[10,60)', staleDuration: 50, violationInterval: '[15,60)', violationDuration: 45,
  remediation: 'Invalidate the cache when the subscription is revoked.',
} as const
