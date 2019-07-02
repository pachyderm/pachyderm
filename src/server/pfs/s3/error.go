package s3

import (
	"net/http"

	"github.com/pachyderm/pachyderm/src/server/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/s3server"
)

func enterpriseDisabledError(r *http.Request) *s3server.S3Error {
	return s3server.NewS3Error(r, http.StatusForbidden, "EnterpriseDisabled", "Enterprise mode must be enabled to use the s3gateway.")
}

func invalidDelimiterError(r *http.Request) *s3server.S3Error {
	return s3server.NewS3Error(r, http.StatusBadRequest, "InvalidDelimiter", "The delimiter you specified is invalid. It must be '' or '/'.")
}

func invalidFilePathError(r *http.Request) *s3server.S3Error {
	return s3server.NewS3Error(r, http.StatusBadRequest, "InvalidFilePath", "Invalid file path")
}

func maybeNotFoundError(r *http.Request, err error) *s3server.S3Error {
	if pfs.IsRepoNotFoundErr(err) || pfs.IsBranchNotFoundErr(err) {
		return s3server.NoSuchBucketError(r)
	} else if pfs.IsFileNotFoundErr(err) {
		return s3server.NoSuchKeyError(r)
	}
	return s3server.InternalError(r, err)
}
