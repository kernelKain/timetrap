import type { ValidationIssue } from '../../types/forms'
export function FieldError({ id, errors }: { id: string; errors: ValidationIssue[] }) {
  if (!errors.length) return null
  return <p className="field-error" id={id} role="alert">{errors[0].message}</p>
}
