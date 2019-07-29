package s3

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	pfsClient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/s2"
	"github.com/sirupsen/logrus"
)

type objectController struct {
	pc     *client.APIClient
	logger *logrus.Entry
}

func newObjectController(pc *client.APIClient, logger *logrus.Entry) *objectController {
	c := objectController{
		pc:     pc,
		logger: logger,
	}

	return &c
}

func (c *objectController) GetObject(r *http.Request, repo, file, version string) (etag, fetchedVersion string, modTime time.Time, content io.ReadSeeker, err error) {
	if version == "" {
		version = "master"
	}
	if isCommit(version) {
		_, err = c.pc.InspectCommit(repo, version)
		if err != nil {
			err = maybeNotFoundError(r, err)
			return
		}
	} else {
		var branchInfo *pfsClient.BranchInfo
		branchInfo, err = c.pc.InspectBranch(repo, version)
		if err != nil {
			err = maybeNotFoundError(r, err)
			return
		}
		if branchInfo.Head == nil {
			err = s2.NoSuchKeyError(r)
			return
		}
	}

	if strings.HasSuffix(file, "/") {
		err = invalidFilePathError(r)
		return
	}

	fileInfo, err := c.pc.InspectFile(repo, version, file)
	if err != nil {
		err = maybeNotFoundError(r, err)
		return
	}

	modTime, err = types.TimestampFromProto(fileInfo.Committed)
	if err != nil {
		return
	}

	content, err = c.pc.GetFileReadSeeker(repo, version, file)
	if err != nil {
		return
	}

	etag = fmt.Sprintf("%x", fileInfo.Hash)
	return
}

func (c *objectController) PutObject(r *http.Request, repo, file string, reader io.Reader) (etag, createdVersion string, err error) {
	branch := branchArg(r)
	_, err = c.pc.InspectBranch(repo, branch)
	if err != nil {
		err = maybeNotFoundError(r, err)
		return
	}
	if strings.HasSuffix(file, "/") {
		err = invalidFilePathError(r)
		return
	}

	_, err = c.pc.PutFileOverwrite(repo, branch, file, reader, 0)
	if err != nil {
		return
	}

	fileInfo, err := c.pc.InspectFile(repo, branch, file)
	if err != nil {
		return
	}

	etag = fmt.Sprintf("%x", fileInfo.Hash)
	return
}

func (c *objectController) DeleteObject(r *http.Request, repo, file, version string) (removedVersion string, err error) {
	// It's only possible to delete objects when no version is specified (i.e.
	// it runs on the HEAD of master), or when the specified version is a
	// branch (i.e. it runs on the HEAD of the specified branch.)
	if isCommit(version) {
		return "", illegalVersioningConfigurationError(r)
	}
	if version == "" {
		version = "master"
	}

	branchInfo, err := c.pc.InspectBranch(repo, version)
	if err != nil {
		return "", maybeNotFoundError(r, err)
	}
	if branchInfo.Head == nil {
		return "", s2.NoSuchKeyError(r)
	}
	if strings.HasSuffix(file, "/") {
		return "", invalidFilePathError(r)
	}

	if err := c.pc.DeleteFile(repo, version, file); err != nil {
		return "", maybeNotFoundError(r, err)
	}

	return version, nil
}
