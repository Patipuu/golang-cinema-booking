package helpers

import (
    "regexp"
    "strings"
)

// Validator validates input data
type Validator struct {
    Errors map[string]string
}

func NewValidator() *Validator {
    return &Validator{
        Errors: make(map[string]string),
    }
}

func (v *Validator) IsValid() bool {
    return len(v.Errors) == 0
}

func (v *Validator) AddError(key, message string) {
    if _, exists := v.Errors[key]; !exists {
        v.Errors[key] = message
    }
}

// Required kiểm tra field không empty
func (v *Validator) Required(field, value string) {
    if strings.TrimSpace(value) == "" {
        v.AddError(field, "không được để trống")
    }
}

// MinLength kiểm tra độ dài tối thiểu
func (v *Validator) MinLength(field, value string, min int) {
    if len(strings.TrimSpace(value)) < min {
        v.AddError(field, "phải có ít nhất "+string(rune(min))+" ký tự")
    }
}

// MaxLength kiểm tra độ dài tối đa
func (v *Validator) MaxLength(field, value string, max int) {
    if len(strings.TrimSpace(value)) > max {
        v.AddError(field, "không được vượt quá "+string(rune(max))+" ký tự")
    }
}

// Email kiểm tra định dạng email
func (v *Validator) Email(field, value string) {
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    if !emailRegex.MatchString(value) {
        v.AddError(field, "email không hợp lệ")
    }
}

// Phone kiểm tra số điện thoại
func (v *Validator) Phone(field, value string) {
    phoneRegex := regexp.MustCompile(`^[0-9]{10,11}$`)
    if !phoneRegex.MatchString(value) {
        v.AddError(field, "số điện thoại không hợp lệ")
    }
}

// PositiveNumber kiểm tra số dương
func (v *Validator) PositiveNumber(field string, value float64) {
    if value <= 0 {
        v.AddError(field, "phải lớn hơn 0")
    }
}

// InList kiểm tra value có trong list không
func (v *Validator) InList(field, value string, list []string) {
    for _, item := range list {
        if item == value {
            return
        }
    }
    v.AddError(field, "giá trị không hợp lệ")
}