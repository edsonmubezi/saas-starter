package validator

import (
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
)

var (
	Validate *validator.Validate
	Trans    ut.Translator
)

func Init() error {
	enLocale := en.New()
	uni := ut.New(enLocale, enLocale)
	translator, _ := uni.GetTranslator("en")

	// ✅ enable required on struct fields (so custom types like DateOnly work with `required`)
	Validate = validator.New(validator.WithRequiredStructEnabled())
	Trans = translator

	// ✅ use JSON tag names in error messages
	Validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		tag := fld.Tag.Get("json")
		if tag == "-" {
			return ""
		}
		if idx := strings.Index(tag, ","); idx != -1 {
			return tag[:idx]
		}
		return tag
	})

	// default EN translations
	if err := enTranslations.RegisterDefaultTranslations(Validate, Trans); err != nil {
		return err
	}

	// Register custom messages (your existing function)
	registerCustomMessages()

	return nil
}

// Helper that returns translated messages
func ValidateStruct(s interface{}) []string {
	var errors []string
	if err := Validate.Struct(s); err != nil {
		for _, verr := range err.(validator.ValidationErrors) {
			errors = append(errors, verr.Translate(Trans))
		}
	}
	return errors
}
