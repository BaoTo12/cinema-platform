package validator

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator wraps go-playground validator
type Validator struct {
	validate *validator.Validate
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   any    `json:"value,omitempty"`
	Message string `json:"message"`
}

// New creates a new validator instance
func New() *Validator {
	v := validator.New()

	// Register custom tag name function to use json tag names
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// Register custom validations
	registerCustomValidations(v)

	return &Validator{validate: v}
}

// registerCustomValidations registers custom validation functions
func registerCustomValidations(v *validator.Validate) {
	// Phone number validation
	v.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
		phone := fl.Field().String()
		if phone == "" {
			return true // Optional field
		}
		// Basic phone validation: allows + at start, then digits, spaces, dashes
		re := regexp.MustCompile(`^\+?[0-9\s\-]{10,15}$`)
		return re.MatchString(phone)
	})

	// Password strength validation
	v.RegisterValidation("password", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()
		if len(password) < 8 {
			return false
		}
		// At least one uppercase, one lowercase, one digit
		hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
		hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
		hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
		return hasUpper && hasLower && hasDigit
	})

	// Slug validation
	v.RegisterValidation("slug", func(fl validator.FieldLevel) bool {
		slug := fl.Field().String()
		re := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
		return re.MatchString(slug)
	})
}

// Validate validates a struct and returns validation errors
func (v *Validator) Validate(i interface{}) []ValidationError {
	err := v.validate.Struct(i)
	if err == nil {
		return nil
	}

	var errors []ValidationError
	for _, err := range err.(validator.ValidationErrors) {
		errors = append(errors, ValidationError{
			Field:   err.Field(),
			Tag:     err.Tag(),
			Value:   err.Value(),
			Message: getErrorMessage(err),
		})
	}

	return errors
}

// ValidateField validates a single field
func (v *Validator) ValidateField(field interface{}, tag string) error {
	return v.validate.Var(field, tag)
}

// getErrorMessage returns a human-readable error message
func getErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return err.Field() + " is required"
	case "email":
		return "Invalid email format"
	case "min":
		if err.Type().Kind() == reflect.String {
			return err.Field() + " must be at least " + err.Param() + " characters"
		}
		return err.Field() + " must be at least " + err.Param()
	case "max":
		if err.Type().Kind() == reflect.String {
			return err.Field() + " must be at most " + err.Param() + " characters"
		}
		return err.Field() + " must be at most " + err.Param()
	case "len":
		return err.Field() + " must be exactly " + err.Param() + " characters"
	case "uuid":
		return "Invalid UUID format"
	case "oneof":
		return err.Field() + " must be one of: " + err.Param()
	case "password":
		return "Password must be at least 8 characters and contain uppercase, lowercase, and digit"
	case "phone":
		return "Invalid phone number format"
	case "url":
		return "Invalid URL format"
	case "gte":
		return err.Field() + " must be greater than or equal to " + err.Param()
	case "lte":
		return err.Field() + " must be less than or equal to " + err.Param()
	case "gt":
		return err.Field() + " must be greater than " + err.Param()
	case "lt":
		return err.Field() + " must be less than " + err.Param()
	case "eqfield":
		return err.Field() + " must match " + err.Param()
	default:
		return "Invalid value for " + err.Field()
	}
}
