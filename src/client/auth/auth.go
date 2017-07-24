package auth

import (
	"fmt"
	"strings"
)

const (
	// ContextTokenKey is the key of the auth token in an
	// authenticated context
	ContextTokenKey = "authn-token"
)

// ParseScope parses the string 's' to a scope (for example, parsing a command-
// line argument.
func ParseScope(s string) (Scope, error) {
	for name, value := range Scope_value {
		if strings.EqualFold(s, name) {
			return Scope(value), nil
		}
	}
	return Scope_NONE, fmt.Errorf("unrecognized scope: %s", s)
}

// NotActivatedError is returned by an Auth API if the Auth service
// has not been activated.
type NotActivatedError struct{}

const notActivatedErrorMsg = "the auth service is not activated"

func (e NotActivatedError) Error() string {
	return notActivatedErrorMsg
}

// IsNotActivatedError checks if an error is a NotActivatedError
func IsNotActivatedError(e error) bool {
	return strings.Contains(e.Error(), notActivatedErrorMsg)
}
