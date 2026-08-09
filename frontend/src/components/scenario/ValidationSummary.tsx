import { AlertCircle } from 'lucide-react'
import type { ValidationIssue } from '../../types/forms'
export function ValidationSummary({ errors }: { errors: ValidationIssue[] }) {
  if (!errors.length) return null
  return <section className="validation-summary" role="alert" aria-labelledby="validation-heading"><AlertCircle aria-hidden="true" /><div><h2 id="validation-heading">Resolve {errors.length} validation {errors.length === 1 ? 'issue' : 'issues'}</h2><ul>{errors.map((error, index) => <li key={`${error.field}-${index}`}><code>{error.field}</code> — {error.message}</li>)}</ul></div></section>
}
