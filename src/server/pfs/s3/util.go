package s3

import (
	"net/http"

	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"github.com/pachyderm/s2"
)

// The S3 storage class that all PFS content will be reported to be stored in
const globalStorageClass = "STANDARD"

// The S3 location served back
const globalLocation = "PACHYDERM"

// The S3 user associated with all PFS content
var defaultUser = s2.User{ID: "00000000000000000000000000000000", DisplayName: "pachyderm"}

// isCommit checks whether a given value is a commit ID
func isCommit(value string) bool {
	return uuid.IsUUIDWithoutDashes(value)
}

func branchArg(r *http.Request) string {
	value := r.FormValue("branch")
	if value == "" {
		return "master"
	}
	return value
}
