package s3

import "github.com/pachyderm/s2"

// The S3 storage class that all PFS content will be reported to be stored in
const globalStorageClass = "STANDARD"

// The S3 location served back
const globalLocation = "PACHYDERM"

// The S3 user associated with all PFS content
var defaultUser = s2.User{ID: "00000000000000000000000000000000", DisplayName: "pachyderm"}
