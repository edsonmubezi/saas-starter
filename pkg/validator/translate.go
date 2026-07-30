package validator

import (
	"regexp"
	"strings"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

func registerCustomMessages() {
	translations := []struct {
		tag         string
		translation string
		paramFunc   func(fe validator.FieldError) []string
	}{
		{"required", "{0} is required", func(fe validator.FieldError) []string {
			return []string{humanizeFieldName(fe.Field())}
		}},
		{"min", "{0} must be at least {1} characters", func(fe validator.FieldError) []string {
			return []string{humanizeFieldName(fe.Field()), fe.Param()}
		}},
		{"max", "{0} must be at most {1} characters", func(fe validator.FieldError) []string {
			return []string{humanizeFieldName(fe.Field()), fe.Param()}
		}},
		{"email", "{0} must be a valid email address", func(fe validator.FieldError) []string {
			return []string{humanizeFieldName(fe.Field())}
		}},
		{"gte", "{0} must be greater than or equal to {1}", func(fe validator.FieldError) []string {
			return []string{humanizeFieldName(fe.Field()), fe.Param()}
		}},
		{"lte", "{0} must be less than or equal to {1}", func(fe validator.FieldError) []string {
			return []string{humanizeFieldName(fe.Field()), fe.Param()}
		}},
		{"gt", "{0} must be greater than {1}", func(fe validator.FieldError) []string {
			return []string{humanizeFieldName(fe.Field()), fe.Param()}
		}},
		{"lt", "{0} must be less than {1}", func(fe validator.FieldError) []string {
			return []string{humanizeFieldName(fe.Field()), fe.Param()}
		}},
		{"len", "{0} must be exactly {1} characters long", func(fe validator.FieldError) []string {
			return []string{humanizeFieldName(fe.Field()), fe.Param()}
		}},
		{"oneof", "{0} must be one of [{1}]", func(fe validator.FieldError) []string {
			return []string{humanizeFieldName(fe.Field()), fe.Param()}
		}},
		{"url", "{0} must be a valid URL", func(fe validator.FieldError) []string {
			return []string{humanizeFieldName(fe.Field())}
		}},
		{"uuid", "{0} must be a valid UUID", func(fe validator.FieldError) []string {
			return []string{humanizeFieldName(fe.Field())}
		}},
		{"numeric", "{0} must be a number", func(fe validator.FieldError) []string {
			return []string{humanizeFieldName(fe.Field())}
		}},
	}

	for _, tr := range translations {
		_ = Validate.RegisterTranslation(tr.tag, Trans,
			func(ut ut.Translator) error {
				return ut.Add(tr.tag, tr.translation, true)
			},
			func(ut ut.Translator, fe validator.FieldError) string {
				t, _ := ut.T(tr.tag, tr.paramFunc(fe)...)
				return t
			},
		)
	}
}

func humanizeFieldName(name string) string {
	// Split camelCase words
	re := regexp.MustCompile("([a-z])([A-Z])")
	spaced := re.ReplaceAllString(name, "${1} ${2}")

	// Capitalize first letter
	return strings.Title(spaced)
}
