package domain

// FieldError describes one invalid field in a submitted scenario.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors contains all field-level problems found in a scenario.
type ValidationErrors []FieldError

// Add records one field-level validation problem.
func (validationErrors *ValidationErrors) Add(field, message string) {
	*validationErrors = append(*validationErrors, FieldError{
		Field:   field,
		Message: message,
	})
}

// Empty reports whether validation found no problems.
func (validationErrors ValidationErrors) Empty() bool {
	return len(validationErrors) == 0
}

// Error allows ValidationErrors to satisfy Go's error interface.
func (validationErrors ValidationErrors) Error() string {
	if validationErrors.Empty() {
		return ""
	}

	return "scenario validation failed"
}
