export function TechnicalBadge({ children, tone = 'neutral' }: { children: React.ReactNode; tone?: 'neutral' | 'cyan' | 'warning' | 'danger' | 'safe' }) {
  return <span className={`technical-badge ${tone}`}>{children}</span>
}
