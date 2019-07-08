package s3

import (
	"net/http"
	"strings"

	"github.com/pachyderm/s2"
)

// The S3 storage class that all PFS content will be reported to be stored in
const storageClass = "STANDARD"

// The S3 user associated with all PFS content
var defaultUser = s2.User{ID: "00000000000000000000000000000000", DisplayName: "pachyderm"}

func bucketArgs(r *http.Request, bucket string) (string, string, *s2.Error) {
	parts := strings.SplitN(bucket, ".", 2)
	if len(parts) != 2 || len(parts[0]) == 0 || len(parts[1]) == 0 {
		return "", "", s2.InvalidBucketNameError(r)
	}
	return parts[1], parts[0], nil
}
