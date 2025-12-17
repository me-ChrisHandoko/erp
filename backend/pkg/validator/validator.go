package validator

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"

	"backend/pkg/errors"
)

var (
	// phoneRegex validates Indonesian phone numbers (starts with 08 or +62)
	phoneRegex = regexp.MustCompile(`^(\+62|62|0)8[0-9]{8,11}$`)

	// passwordRegex checks for at least one uppercase, one lowercase, one digit
	passwordRegex = regexp.MustCompile(`^.*[a-z].*[A-Z].*[0-9]|[A-Z].*[a-z].*[0-9]|[0-9].*[a-z].*[A-Z]|[0-9].*[A-Z].*[a-z]|[a-z].*[0-9].*[A-Z]|[A-Z].*[0-9].*[a-z].*$`)
)

// Validator wraps go-playground/validator with custom validators
type Validator struct {
	validate *validator.Validate
}

// New creates a new Validator instance with custom validators
func New() *Validator {
	v := validator.New()

	// Register custom validators
	v.RegisterValidation("password_strength", validatePasswordStrength)
	v.RegisterValidation("phone_number", validatePhoneNumber)

	return &Validator{
		validate: v,
	}
}

// ValidateStruct validates a struct and returns formatted errors
func (v *Validator) ValidateStruct(s interface{}) error {
	err := v.validate.Struct(s)
	if err == nil {
		return nil
	}

	// Convert validation errors to our custom error format
	validationErrors := make([]errors.ValidationError, 0)

	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, fieldErr := range validationErrs {
			validationErrors = append(validationErrors, errors.ValidationError{
				Field:   getJSONFieldName(fieldErr),
				Message: formatErrorMessage(fieldErr),
			})
		}
	}

	return errors.NewValidationError(validationErrors)
}

// validatePasswordStrength checks password meets security requirements
// Requirements:
// - At least 8 characters (enforced by min tag)
// - Contains at least one uppercase letter
// - Contains at least one lowercase letter
// - Contains at least one digit
// - Optionally contains special characters for extra security
func validatePasswordStrength(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}

	var (
		hasUpper bool
		hasLower bool
		hasDigit bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		}
	}

	// Require at least uppercase, lowercase, and digit
	// Special characters are recommended but not required
	return hasUpper && hasLower && hasDigit
}

// validatePhoneNumber validates Indonesian phone numbers
// Accepts formats:
// - 08xxxxxxxxxx (10-13 digits)
// - +628xxxxxxxxxx
// - 628xxxxxxxxxx
func validatePhoneNumber(fl validator.FieldLevel) bool {
	phone := fl.Field().String()

	// Empty phone is valid (handled by required/omitempty tags)
	if phone == "" {
		return true
	}

	return phoneRegex.MatchString(phone)
}

// getJSONFieldName extracts the JSON field name from validator.FieldError
func getJSONFieldName(fe validator.FieldError) string {
	// Get the field name from the struct tag
	field := fe.Field()

	// Convert to camelCase for JSON consistency
	if len(field) > 0 {
		return strings.ToLower(field[:1]) + field[1:]
	}

	return field
}

// formatErrorMessage generates user-friendly error messages
func formatErrorMessage(fe validator.FieldError) string {
	field := getJSONFieldName(fe)

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		if fe.Type().String() == "string" {
			return fmt.Sprintf("%s must be at least %s characters long", field, fe.Param())
		}
		return fmt.Sprintf("%s must be at least %s", field, fe.Param())
	case "max":
		if fe.Type().String() == "string" {
			return fmt.Sprintf("%s must not exceed %s characters", field, fe.Param())
		}
		return fmt.Sprintf("%s must not exceed %s", field, fe.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, fe.Param())
	case "password_strength":
		return fmt.Sprintf("%s must contain at least one uppercase letter, one lowercase letter, and one digit", field)
	case "phone_number":
		return fmt.Sprintf("%s must be a valid Indonesian phone number (format: 08xxxxxxxxxx or +628xxxxxxxxxx)", field)
	case "nefield":
		return fmt.Sprintf("%s must be different from %s", field, fe.Param())
	case "eqfield":
		return fmt.Sprintf("%s must match %s", field, fe.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, fe.Param())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "uri":
		return fmt.Sprintf("%s must be a valid URI", field)
	case "alpha":
		return fmt.Sprintf("%s must contain only alphabetic characters", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only alphanumeric characters", field)
	case "numeric":
		return fmt.Sprintf("%s must be a valid numeric value", field)
	case "number":
		return fmt.Sprintf("%s must be a valid number", field)
	case "hexadecimal":
		return fmt.Sprintf("%s must be a valid hexadecimal value", field)
	case "hexcolor":
		return fmt.Sprintf("%s must be a valid hex color", field)
	case "rgb":
		return fmt.Sprintf("%s must be a valid RGB color", field)
	case "rgba":
		return fmt.Sprintf("%s must be a valid RGBA color", field)
	case "hsl":
		return fmt.Sprintf("%s must be a valid HSL color", field)
	case "hsla":
		return fmt.Sprintf("%s must be a valid HSLA color", field)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "uuid3":
		return fmt.Sprintf("%s must be a valid UUID v3", field)
	case "uuid4":
		return fmt.Sprintf("%s must be a valid UUID v4", field)
	case "uuid5":
		return fmt.Sprintf("%s must be a valid UUID v5", field)
	case "ascii":
		return fmt.Sprintf("%s must contain only ASCII characters", field)
	case "printascii":
		return fmt.Sprintf("%s must contain only printable ASCII characters", field)
	case "base64":
		return fmt.Sprintf("%s must be a valid Base64 string", field)
	case "btc_addr":
		return fmt.Sprintf("%s must be a valid Bitcoin address", field)
	case "btc_addr_bech32":
		return fmt.Sprintf("%s must be a valid Bech32 Bitcoin address", field)
	case "eth_addr":
		return fmt.Sprintf("%s must be a valid Ethereum address", field)
	case "ip":
		return fmt.Sprintf("%s must be a valid IP address", field)
	case "ipv4":
		return fmt.Sprintf("%s must be a valid IPv4 address", field)
	case "ipv6":
		return fmt.Sprintf("%s must be a valid IPv6 address", field)
	case "cidr":
		return fmt.Sprintf("%s must be a valid CIDR notation", field)
	case "cidrv4":
		return fmt.Sprintf("%s must be a valid CIDR v4 notation", field)
	case "cidrv6":
		return fmt.Sprintf("%s must be a valid CIDR v6 notation", field)
	case "tcp_addr":
		return fmt.Sprintf("%s must be a valid TCP address", field)
	case "tcp4_addr":
		return fmt.Sprintf("%s must be a valid TCPv4 address", field)
	case "tcp6_addr":
		return fmt.Sprintf("%s must be a valid TCPv6 address", field)
	case "udp_addr":
		return fmt.Sprintf("%s must be a valid UDP address", field)
	case "udp4_addr":
		return fmt.Sprintf("%s must be a valid UDPv4 address", field)
	case "udp6_addr":
		return fmt.Sprintf("%s must be a valid UDPv6 address", field)
	case "ip_addr":
		return fmt.Sprintf("%s must be a valid IP address", field)
	case "unix_addr":
		return fmt.Sprintf("%s must be a valid Unix domain socket address", field)
	case "mac":
		return fmt.Sprintf("%s must be a valid MAC address", field)
	case "hostname":
		return fmt.Sprintf("%s must be a valid hostname", field)
	case "hostname_rfc1123":
		return fmt.Sprintf("%s must be a valid hostname (RFC 1123)", field)
	case "fqdn":
		return fmt.Sprintf("%s must be a valid FQDN", field)
	case "latitude":
		return fmt.Sprintf("%s must be a valid latitude", field)
	case "longitude":
		return fmt.Sprintf("%s must be a valid longitude", field)
	default:
		return fmt.Sprintf("%s failed validation (%s)", field, fe.Tag())
	}
}
