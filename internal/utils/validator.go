package utils

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidateStruct validates v and returns first error message suitable for API.
func ValidateStruct(v *validator.Validate, s interface{}) error {
	if v == nil {
		v = validator.New()
	}
	err := v.Struct(s)
	if err == nil {
		return nil
	}
	var msg []string
	for _, e := range err.(validator.ValidationErrors) {
		msg = append(msg, fmt.Sprintf("%s: %s", e.Field(), e.Tag()))
	}
	return fmt.Errorf("%s", strings.Join(msg, "; "))
}

// NewValidator returns a validator instance (can add custom rules here).
func NewValidator() *validator.Validate {
	return validator.New()
}
