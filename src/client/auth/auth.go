package auth

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc/metadata"
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

// In2Out converts an incoming context containing auth information into an
// outgoing context containing auth information, stripping other keys (e.g.
// for metrics) in the process. If the incoming context doesn't have any auth
// information, then the returned context won't either.
func In2Out(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}
	mdOut := make(metadata.MD)
	if value, ok := md[ContextTokenKey]; ok {
		mdOut[ContextTokenKey] = value
	}
	return metadata.NewOutgoingContext(ctx, mdOut)
}

// NotActivatedError is returned by an Auth API if the Auth service
// has not been activated.
type NotActivatedError struct{}

// This error message string is matched in the UI. If edited,
// it also needs to be updated in the UI code
const notActivatedErrorMsg = "the auth service is not activated"

func (e NotActivatedError) Error() string {
	return notActivatedErrorMsg
}

// IsNotActivatedError checks if an error is a NotActivatedError
func IsNotActivatedError(err error) bool {
	if err == nil {
		return false
	}
	// TODO(msteffen) This is unstructured because we have no way to propagate
	// structured errors across GRPC boundaries. Fix
	return strings.Contains(err.Error(), notActivatedErrorMsg)
}

// NotAuthorizedError is returned if the user is not authorized to perform
// a certain operation on the repo 'Repo' (to do so, they would need to have the
// authorization scope in 'Required').
type NotAuthorizedError struct {
	Repo     string
	Required Scope
}

// This error message string is matched in the UI. If edited,
// it also needs to be updated in the UI code
const notAuthorizedErrorMsg = "not authorized to perform this operation"

func (e *NotAuthorizedError) Error() string {
	msg := notAuthorizedErrorMsg
	if e.Repo != "" {
		msg += " on the repo " + e.Repo
	}
	if e.Required != Scope_NONE {
		msg += ", must have at least " + e.Required.String() + " access"
	}
	return msg
}

// IsNotAuthorizedError checks if an error is a NotAuthorizedError
func IsNotAuthorizedError(err error) bool {
	if err == nil {
		return false
	}
	// TODO(msteffen) This is unstructured because we have no way to propagate
	// structured errors across GRPC boundaries. Fix
	return strings.HasPrefix(err.Error(), notAuthorizedErrorMsg)
}

// This error message string is matched in the UI. If edited,
// it also needs to be updated in the UI code
const notSignedInErrMsg = "auth token not found in context (user may not be signed in)"

// NotSignedInError indicates that the caller isn't signed in
type NotSignedInError struct{}

func (e NotSignedInError) Error() string {
	return notSignedInErrMsg
}

// IsNotSignedInError returns true if 'err' is a NotSignedInError
func IsNotSignedInError(err error) bool {
	// TODO(msteffen) This is unstructured because we have no way to propagate
	// structured errors across GRPC boundaries. Fix
	return strings.Contains(err.Error(), notSignedInErrMsg)
}

// InvalidPrincipalError indicates that a an argument to e.g. GetScope,
// SetScope, or SetACL is invalid
type InvalidPrincipalError struct {
	Principal string
}

func (e *InvalidPrincipalError) Error() string {
	return fmt.Sprintf("invalid principal \"%s\"; must start with \"robot:\" or have no \":\"", e.Principal)
}

// IsInvalidPrincipalError returns true if 'err' is an InvalidPrincipalError
func IsInvalidPrincipalError(err error) bool {
	if err == nil {
		return false
	}
	return strings.HasPrefix(err.Error(), "invalid principal \"") &&
		strings.HasSuffix(err.Error(), "\"; must start with \"robot:\" or have no \":\"")
}

const badTokenErrorMsg = "provided auth token is corrupted or has expired (try logging in again)"

// BadTokenError is returned by the Auth API if the caller's token is corruped
// or has expired.
type BadTokenError struct{}

func (e BadTokenError) Error() string {
	return badTokenErrorMsg
}

// IsBadTokenError returns true if 'err' is a BadTokenError
func IsBadTokenError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), badTokenErrorMsg)
}
