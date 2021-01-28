package license

import (
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// ErrDuplicateClusterID is thrown when a cluster is registered but the ID already exists
	ErrDuplicateClusterID = status.Error(codes.Unimplemented, "cluster ID is not unique")
)

// IsErrDuplicateClusterID checks if an error is a IsErrDuplicateClusterID
func IsErrDuplicateClusterID(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), status.Convert(ErrDuplicateClusterID).Message())
}
