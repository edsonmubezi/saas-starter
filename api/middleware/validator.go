package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

var validate = validator.New()

func ParseAndValidateBody[T any](r *http.Request) (*T, []FieldError, error) {
	// Create a new instance of T
	var obj T

	// Decode JSON into T
	if err := json.NewDecoder(r.Body).Decode(&obj); err != nil {
		return nil, []FieldError{{Field: "body", Message: "Invalid JSON format"}}, nil
	}

	// Validate
	if err := validate.Struct(obj); err != nil {
		var errs []FieldError
		for _, fe := range err.(validator.ValidationErrors) {
			field := toJSONName(fe.Field())
			message := getErrorMessage(field, fe.Tag(), fe.Param())
			errs = append(errs, FieldError{Field: field, Message: message})
		}
		return nil, errs, nil
	}

	return &obj, nil, nil
}

func getErrorMessage(field, tag, param string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, param)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "gte":
		return fmt.Sprintf("%s must be at least %s", field, param)
	case "lte":
		return fmt.Sprintf("%s must be no more than %s", field, param)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// toJSONName converts a struct field name to snake_case for JSON consistency
// e.g., "FullName" → "full_name", "RoleID" → "role_id", "OrganizationID" → "organization_id"
func toJSONName(field string) string {
	if len(field) == 0 {
		return field
	}

	var result []byte
	for i, r := range field {
		// If uppercase, add underscore before it (except at start)
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				// Check if previous char is lowercase or if next char exists and is lowercase
				// This handles "ID" → "id" not "i_d"
				prevIsLower := field[i-1] >= 'a' && field[i-1] <= 'z'
				nextIsLower := i+1 < len(field) && field[i+1] >= 'a' && field[i+1] <= 'z'
				if prevIsLower || nextIsLower {
					result = append(result, '_')
				}
			}
			result = append(result, byte(r)|0x20) // to lowercase
		} else {
			result = append(result, byte(r))
		}
	}
	return string(result)
}
