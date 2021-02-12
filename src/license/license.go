package license

import (
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// ErrDuplicateClusterID is thrown when a cluster is registered but the ID already exists
	ErrDuplicateClusterID = status.Error(codes.Unimplemented, "cluster ID is not unique")

	// ErrInvalidIDOrSecret is thrown when the provided cluster ID or secret is not valid
	ErrInvalidIDOrSecret = status.Error(codes.Unimplemented, "cluster ID or secret is not valid")

	// ErrNotActivated is thrown when a cluster does not have an enterprise key activated
	ErrNotActivated = status.Error(codes.Unimplemented, "cluster does not have enterprise features enabled")
)

// IsErrDuplicateClusterID checks if an error is an ErrDuplicateClusterID
func IsErrDuplicateClusterID(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), status.Convert(ErrDuplicateClusterID).Message())
}

// IsErrInvalidIDOrSecret checks if an error is an ErrInvalidIDOrSecret
func IsErrInvalidIDOrSecret(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), status.Convert(ErrInvalidIDOrSecret).Message())
}

// IsErrNotActivated checks if an error is an ErrNotActivated
func IsErrNotActivated(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), status.Convert(ErrNotActivated).Message())
}
