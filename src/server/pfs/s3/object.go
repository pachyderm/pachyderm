package s3

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/gorilla/mux"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/s2"
	"github.com/sirupsen/logrus"
)

type objectController struct {
	pc     *client.APIClient
	logger *logrus.Entry
}

func (c objectController) GetObject(r *http.Request, bucket, file string) (etag string, modTime time.Time, content io.ReadSeeker, err error) {
	vars := mux.Vars(r)
	pc, err := pachClient(vars["authAccessKey"])
	if err != nil {
		return
	}
	repo, branch, err := bucketArgs(r, bucket)
	if err != nil {
		return
	}

	branchInfo, err := pc.InspectBranch(repo, branch)
	if err != nil {
		err = maybeNotFoundError(r, err)
		return
	}
	if branchInfo.Head == nil {
		err = s2.NoSuchKeyError(r)
		return
	}
	if strings.HasSuffix(file, "/") {
		invalidFilePathError(r)
		return
	}

	fileInfo, err := pc.InspectFile(branchInfo.Branch.Repo.Name, branchInfo.Head.ID, file)
	if err != nil {
		err = maybeNotFoundError(r, err)
		return
	}

	modTime, err = types.TimestampFromProto(fileInfo.Committed)
	if err != nil {
		return
	}

	content, err = pc.GetFileReadSeeker(branchInfo.Branch.Repo.Name, branchInfo.Head.ID, file)
	if err != nil {
		return
	}

	etag = fmt.Sprintf("%x", fileInfo.Hash)
	return
}

func (c objectController) PutObject(r *http.Request, bucket, file string, reader io.Reader) (etag string, err error) {
	vars := mux.Vars(r)
	pc, err := pachClient(vars["authAccessKey"])
	if err != nil {
		return
	}
	repo, branch, err := bucketArgs(r, bucket)
	if err != nil {
		return
	}

	branchInfo, err := pc.InspectBranch(repo, branch)
	if err != nil {
		err = maybeNotFoundError(r, err)
		return
	}
	if strings.HasSuffix(file, "/") {
		err = invalidFilePathError(r)
		return
	}

	_, err = pc.PutFileOverwrite(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, file, reader, 0)
	if err != nil {
		return
	}

	fileInfo, err := pc.InspectFile(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, file)
	if err != nil {
		return
	}

	etag = fmt.Sprintf("%x", fileInfo.Hash)
	return
}

func (c objectController) DeleteObject(r *http.Request, bucket, file string) error {
	vars := mux.Vars(r)
	pc, err := pachClient(vars["authAccessKey"])
	if err != nil {
		return err
	}
	repo, branch, err := bucketArgs(r, bucket)
	if err != nil {
		return err
	}

	branchInfo, err := pc.InspectBranch(repo, branch)
	if err != nil {
		return maybeNotFoundError(r, err)
	}
	if branchInfo.Head == nil {
		return s2.NoSuchKeyError(r)
	}
	if strings.HasSuffix(file, "/") {
		return invalidFilePathError(r)
	}

	if err := pc.DeleteFile(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, file); err != nil {
		return maybeNotFoundError(r, err)
	}

	return nil
}
