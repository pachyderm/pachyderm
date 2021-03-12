package identity

import (
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// ErrInvalidID is returned if the client or connector ID does not exist
	ErrInvalidID = status.Error(codes.Internal, "ID does not exist")

	// ErrAlreadyExists is returned if the clientor connector ID already exists
	ErrAlreadyExists = status.Error(codes.Internal, "ID already exists")
)

// IsErrInvalidID checks if an error is a ErrAlreadyActivated
func IsErrInvalidID(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), status.Convert(ErrInvalidID).Message())
}

// IsErrAlreadyExists checks if an error is a ErrAlreadyExists
func IsErrAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), status.Convert(ErrAlreadyExists).Message())
}
