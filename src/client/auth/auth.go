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
