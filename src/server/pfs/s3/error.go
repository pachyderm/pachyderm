package s3

import (
	"net/http"

	"github.com/pachyderm/pachyderm/src/server/pfs"
	"github.com/pachyderm/s3server"
)

func enterpriseDisabledError(r *http.Request) *s3server.Error {
	return s3server.NewError(r, http.StatusForbidden, "EnterpriseDisabled", "Enterprise mode must be enabled to use the s3gateway.")
}

func invalidDelimiterError(r *http.Request) *s3server.Error {
	return s3server.NewError(r, http.StatusBadRequest, "InvalidDelimiter", "The delimiter you specified is invalid. It must be '' or '/'.")
}

func invalidFilePathError(r *http.Request) *s3server.Error {
	return s3server.NewError(r, http.StatusBadRequest, "InvalidFilePath", "Invalid file path")
}

func maybeNotFoundError(r *http.Request, err error) *s3server.Error {
	if pfs.IsRepoNotFoundErr(err) || pfs.IsBranchNotFoundErr(err) {
		return s3server.NoSuchBucketError(r)
	} else if pfs.IsFileNotFoundErr(err) {
		return s3server.NoSuchKeyError(r)
	}
	return s3server.InternalError(r, err)
}
