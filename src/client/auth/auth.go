package auth

import "strings"

// NotActivatedError is returned by an Auth API if the Auth service
// has not been activated.
type NotActivatedError struct{}

const notActivatedErrorMsg = "the auth service is not activated"

func (e NotActivatedError) Error() string {
	return notActivatedErrorMsg
}

func IsNotActivatedError(e error) bool {
	return strings.Contains(e.Error(), notActivatedErrorMsg)
}
