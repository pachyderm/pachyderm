package s3

import (
	"net/http"

	"github.com/pachyderm/pachyderm/src/server/pfs"
	"github.com/pachyderm/s2"
)

func enterpriseDisabledError(r *http.Request) *s2.Error {
	return s2.NewError(r, http.StatusForbidden, "EnterpriseDisabled", "Enterprise mode must be enabled to use the s3gateway.")
}

func invalidDelimiterError(r *http.Request) *s2.Error {
	return s2.NewError(r, http.StatusBadRequest, "InvalidDelimiter", "The delimiter you specified is invalid. It must be '' or '/'.")
}

func invalidFilePathError(r *http.Request) *s2.Error {
	return s2.NewError(r, http.StatusBadRequest, "InvalidFilePath", "Invalid file path")
}

func writeToOutputBranchError(r *http.Request) *s2.Error {
	return s2.NewError(r, http.StatusBadRequest, "WriteToOutputBranch", "You cannot write to an output branch")
}

func maybeNotFoundError(r *http.Request, err error) *s2.Error {
	if pfs.IsRepoNotFoundErr(err) || pfs.IsBranchNotFoundErr(err) {
		return s2.NoSuchBucketError(r)
	} else if pfs.IsFileNotFoundErr(err) {
		return s2.NoSuchKeyError(r)
	}
	return s2.InternalError(r, err)
}
