package utilis

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// FormatValidationError converts Gin/validator errors into a frontend-friendly
// map of { fieldName: humanReadableMessage }.
func FormatValidationError(err error) map[string]string {
	details := make(map[string]string)

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		details["error"] = err.Error()
		return details
	}

	for _, fe := range validationErrors {
		field := toCamelCase(fe.Field())
		details[field] = messageFor(fe)
	}

	return details
}

// FirstValidationMessage returns the first human-readable validation error
// message for a bind error, suitable for use as a flat "message" field.
func FirstValidationMessage(err error) string {
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok || len(validationErrors) == 0 {
		return "Validation failed"
	}
	return messageFor(validationErrors[0])
}

func messageFor(fe validator.FieldError) string {
	field := toCamelCase(fe.Field())

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return "Invalid email format"
	case "min":
		if fe.Kind().String() == "string" {
			return fmt.Sprintf("%s must be at least %s characters", field, fe.Param())
		}
		return fmt.Sprintf("%s must be at least %s", field, fe.Param())
	case "max":
		if fe.Kind().String() == "string" {
			return fmt.Sprintf("%s must be at most %s characters", field, fe.Param())
		}
		return fmt.Sprintf("%s must be at most %s", field, fe.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, fe.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, fe.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, fe.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, fe.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, fe.Param())
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

func toCamelCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}
