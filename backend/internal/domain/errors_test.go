package domain

import "testing"

func TestValidationErrorsEmpty(t *testing.T) {
	var validationErrors ValidationErrors

	if !validationErrors.Empty() {
		t.Fatal("expected new ValidationErrors to be empty")
	}

	validationErrors.Add("name", "is required")

	if validationErrors.Empty() {
		t.Fatal("expected ValidationErrors to be non-empty after Add")
	}
}

func TestValidationErrorsAddPreservesOrder(t *testing.T) {
	var validationErrors ValidationErrors

	validationErrors.Add("name", "is required")
	validationErrors.Add("objects[0].kind", "is not a supported object kind")
	validationErrors.Add(
		"invariants[0].sourceObjectId",
		"must reference an existing object",
	)

	expected := []FieldError{
		{
			Field:   "name",
			Message: "is required",
		},
		{
			Field:   "objects[0].kind",
			Message: "is not a supported object kind",
		},
		{
			Field:   "invariants[0].sourceObjectId",
			Message: "must reference an existing object",
		},
	}

	if len(validationErrors) != len(expected) {
		t.Fatalf(
			"expected %d validation errors, got %d",
			len(expected),
			len(validationErrors),
		)
	}

	for index, expectedError := range expected {
		actualError := validationErrors[index]

		if actualError != expectedError {
			t.Errorf(
				"error %d: expected %#v, got %#v",
				index,
				expectedError,
				actualError,
			)
		}
	}
}

func TestValidationErrorsError(t *testing.T) {
	t.Run("empty collection", func(t *testing.T) {
		var validationErrors ValidationErrors

		if message := validationErrors.Error(); message != "" {
			t.Fatalf("expected an empty message, got %q", message)
		}
	})

	t.Run("non-empty collection", func(t *testing.T) {
		validationErrors := ValidationErrors{
			{
				Field:   "horizonMinutes",
				Message: "must be between 1 and 10080",
			},
		}

		const expected = "scenario validation failed"

		if message := validationErrors.Error(); message != expected {
			t.Fatalf("expected %q, got %q", expected, message)
		}
	})
}
