package validation

import "regexp"

var (
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

type MapErrors map[string]string

// Validator struct which contains a map of validation errors.
type Validator struct {
	Errors MapErrors
}

// New constructor which creates a new validator instance with an empty errors map
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// Valid returns true if the errors map doesn't contain any entries
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// AddError adds an error message to the map
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// Check adds error message to the map onl if a validation check is not ok
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// PermittedValue generic function which returns true if a specific value is in a list
func PermittedValue[T comparable](value T, permittedValue ...T) bool {
	for i := range permittedValue {
		if value == permittedValue[i] {
			return true
		}
	}
	return false
}

// Matches function returns true if a string value matches a specific regexp pattern.
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// Unique generic function which returns true if all values in a slice are unique.
func Unique[T comparable](values []T) bool {
	uniqueValues := make(map[T]bool)

	for _, val := range values {
		uniqueValues[val] = true
	}

	return len(values) == len(uniqueValues)
}
