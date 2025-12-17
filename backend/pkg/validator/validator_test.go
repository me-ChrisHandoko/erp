package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test struct for validation
type TestPasswordStruct struct {
	Password string `validate:"required,min=8,password_strength"`
}

type TestPhoneStruct struct {
	Phone string `validate:"required,phone_number"`
}

func TestPasswordStrengthValidator_Valid(t *testing.T) {
	validator := New()

	validPasswords := []string{
		"Password1",           // Min requirements
		"P@ssw0rd",            // With special char
		"MyP@ssw0rd123",       // Complex
		"Aa1bcdefgh",          // Long with requirements
		"Test1234Pass",        // Mixed case with numbers
		"ComplexP@ss123",      // All character types
	}

	for _, password := range validPasswords {
		test := TestPasswordStruct{Password: password}
		err := validator.ValidateStruct(test)
		assert.NoError(t, err, "Password '%s' should be valid", password)
	}
}

func TestPasswordStrengthValidator_Invalid(t *testing.T) {
	validator := New()

	invalidPasswords := []struct {
		password string
		reason   string
	}{
		{"short1A", "too short (less than 8 chars)"},
		{"password", "no uppercase or digit"},
		{"PASSWORD", "no lowercase or digit"},
		{"12345678", "no letters"},
		{"Password", "no digit"},
		{"password1", "no uppercase"},
		{"PASSWORD1", "no lowercase"},
		{"", "empty"},
		{"abc", "too short and weak"},
	}

	for _, tc := range invalidPasswords {
		test := TestPasswordStruct{Password: tc.password}
		err := validator.ValidateStruct(test)
		assert.Error(t, err, "Password '%s' should be invalid (%s)", tc.password, tc.reason)
	}
}

func TestPhoneNumberValidator_Valid(t *testing.T) {
	validator := New()

	validPhones := []string{
		"081234567890",      // 12 digits with 08 (standard mobile)
		"08123456789",       // 11 digits with 08
		"0812345678",        // 10 digits with 08 (minimum)
		"+6281234567890",    // With +62 country code
		"6281234567890",     // With 62 (without +)
		"08987654321",       // Different operator prefix
	}

	for _, phone := range validPhones {
		test := TestPhoneStruct{Phone: phone}
		err := validator.ValidateStruct(test)
		assert.NoError(t, err, "Phone '%s' should be valid", phone)
	}
}

func TestPhoneNumberValidator_Invalid(t *testing.T) {
	validator := New()

	invalidPhones := []struct {
		phone  string
		reason string
	}{
		{"0712345678", "doesn't start with 08"},
		{"1234567890", "doesn't start with 0, 62, or +62"},
		{"08123", "too short (less than 10 digits)"},
		{"0812345678901234", "too long (more than 13 digits)"},
		{"+6371234567890", "country code wrong (not 62)"},
		{"081-234-5678", "contains dashes"},
		{"0812 3456 7890", "contains spaces"},
		{"08abc1234567", "contains letters"},
		{"", "empty"},
	}

	for _, tc := range invalidPhones {
		test := TestPhoneStruct{Phone: tc.phone}
		err := validator.ValidateStruct(test)
		assert.Error(t, err, "Phone '%s' should be invalid (%s)", tc.phone, tc.reason)
	}
}

func TestPhoneNumberValidator_Empty(t *testing.T) {
	// Test with omitempty - empty should be valid
	type OptionalPhoneStruct struct {
		Phone string `validate:"omitempty,phone_number"`
	}

	validator := New()
	test := OptionalPhoneStruct{Phone: ""}
	err := validator.ValidateStruct(test)
	assert.NoError(t, err, "Empty phone with omitempty should be valid")
}

func TestValidator_MultipleErrors(t *testing.T) {
	type ComplexStruct struct {
		Email    string `validate:"required,email"`
		Password string `validate:"required,min=8,password_strength"`
		Phone    string `validate:"omitempty,phone_number"`
		Age      int    `validate:"required,min=18,max=100"`
	}

	validator := New()

	// Test with multiple validation errors
	test := ComplexStruct{
		Email:    "invalid-email",
		Password: "weak",
		Phone:    "123",
		Age:      15,
	}

	err := validator.ValidateStruct(test)
	assert.Error(t, err)

	// Check that error message contains field names
	errMsg := err.Error()
	assert.Contains(t, errMsg, "email")
	assert.Contains(t, errMsg, "password")
}

func TestValidator_AllValid(t *testing.T) {
	type ComplexStruct struct {
		Email    string `validate:"required,email,max=255"`
		Password string `validate:"required,min=8,max=72,password_strength"`
		Phone    string `validate:"omitempty,phone_number"`
		FullName string `validate:"required,min=2,max=255"`
	}

	validator := New()

	test := ComplexStruct{
		Email:    "user@example.com",
		Password: "SecurePass123",
		Phone:    "081234567890",
		FullName: "John Doe",
	}

	err := validator.ValidateStruct(test)
	assert.NoError(t, err)
}

func TestFormatErrorMessage_Required(t *testing.T) {
	validator := New()

	type TestStruct struct {
		Name string `validate:"required"`
	}

	test := TestStruct{Name: ""}
	err := validator.ValidateStruct(test)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

func TestFormatErrorMessage_Email(t *testing.T) {
	validator := New()

	type TestStruct struct {
		Email string `validate:"required,email"`
	}

	test := TestStruct{Email: "not-an-email"}
	err := validator.ValidateStruct(test)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "valid email")
}

func TestFormatErrorMessage_MinMax(t *testing.T) {
	validator := New()

	type TestStruct struct {
		ShortString string `validate:"min=5"`
		LongString  string `validate:"max=10"`
		SmallNumber int    `validate:"min=10"`
		BigNumber   int    `validate:"max=100"`
	}

	// Test min string
	test1 := TestStruct{ShortString: "abc"}
	err := validator.ValidateStruct(test1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least 5 characters")

	// Test max string
	test2 := TestStruct{LongString: "this is a very long string"}
	err = validator.ValidateStruct(test2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not exceed 10 characters")
}

func TestGetJSONFieldName(t *testing.T) {
	// Test camelCase conversion
	type TestStruct struct {
		FullName string `validate:"required"`
		Email    string `validate:"required"`
	}

	validator := New()
	test := TestStruct{}
	err := validator.ValidateStruct(test)

	assert.Error(t, err)
	// Validation error should contain field names
	assert.Contains(t, err.Error(), "required")
}

func TestPasswordStrength_EdgeCases(t *testing.T) {
	validator := New()

	// Exactly 8 characters with all requirements
	test1 := TestPasswordStruct{Password: "Passw0rd"}
	err := validator.ValidateStruct(test1)
	assert.NoError(t, err)

	// Very long password
	test2 := TestPasswordStruct{Password: "ThisIsAVeryLongPasswordWith1UppercaseAndDigits"}
	err = validator.ValidateStruct(test2)
	assert.NoError(t, err)

	// Unicode characters (should still validate requirements)
	test3 := TestPasswordStruct{Password: "Pāssw0rd"}
	err = validator.ValidateStruct(test3)
	assert.NoError(t, err)
}

func TestPhoneNumber_EdgeCases(t *testing.T) {
	validator := New()

	// Minimum valid length (10 digits)
	test1 := TestPhoneStruct{Phone: "0812345678"}
	err := validator.ValidateStruct(test1)
	assert.NoError(t, err, "10-digit phone should be valid")

	// With country code +62
	test2 := TestPhoneStruct{Phone: "+6281234567890"}
	err = validator.ValidateStruct(test2)
	assert.NoError(t, err, "+62 prefix should be valid")

	// With country code 62 (no plus)
	test3 := TestPhoneStruct{Phone: "6281234567890"}
	err = validator.ValidateStruct(test3)
	assert.NoError(t, err, "62 prefix (no plus) should be valid")
}
