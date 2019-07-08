package s3

import (
	"io"
	"net/http"
	"strings"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/s3server"
	"github.com/sirupsen/logrus"
)

type objectController struct {
	pc     *client.APIClient
	logger *logrus.Entry
}

func (c objectController) Get(r *http.Request, bucket, file string, result *s3server.GetObjectResult) *s3server.Error {
	repo, branch, s3Err := bucketArgs(r, bucket)
	if s3Err != nil {
		return s3Err
	}

	branchInfo, err := c.pc.InspectBranch(repo, branch)
	if err != nil {
		return maybeNotFoundError(r, err)
	}
	if branchInfo.Head == nil {
		return s3server.NoSuchKeyError(r)
	}
	if strings.HasSuffix(file, "/") {
		return invalidFilePathError(r)
	}

	fileInfo, err := c.pc.InspectFile(branchInfo.Branch.Repo.Name, branchInfo.Head.ID, file)
	if err != nil {
		return maybeNotFoundError(r, err)
	}

	timestamp, err := types.TimestampFromProto(fileInfo.Committed)
	if err != nil {
		return s3server.InternalError(r, err)
	}

	reader, err := c.pc.GetFileReadSeeker(branchInfo.Branch.Repo.Name, branchInfo.Head.ID, file)
	if err != nil {
		return s3server.InternalError(r, err)
	}

	result.Name = file
	result.Hash = fileInfo.Hash
	result.ModTime = timestamp
	result.Content = reader
	return nil
}

func (c objectController) Put(r *http.Request, bucket, file string, reader io.Reader) *s3server.Error {
	repo, branch, s3Err := bucketArgs(r, bucket)
	if s3Err != nil {
		return s3Err
	}

	branchInfo, err := c.pc.InspectBranch(repo, branch)
	if err != nil {
		return maybeNotFoundError(r, err)
	}
	if strings.HasSuffix(file, "/") {
		return invalidFilePathError(r)
	}

	_, err = c.pc.PutFileOverwrite(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, file, reader, 0)
	if err != nil {
		return s3server.InternalError(r, err)
	}

	return nil
}

func (c objectController) Del(r *http.Request, bucket, key string) *s3server.Error {
	repo, branch, s3Err := bucketArgs(r, bucket)
	if s3Err != nil {
		return s3Err
	}

	branchInfo, err := c.pc.InspectBranch(repo, branch)
	if err != nil {
		return maybeNotFoundError(r, err)
	}
	if branchInfo.Head == nil {
		return s3server.NoSuchKeyError(r)
	}
	if strings.HasSuffix(key, "/") {
		return invalidFilePathError(r)
	}

	if err := c.pc.DeleteFile(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, key); err != nil {
		return maybeNotFoundError(r, err)
	}

	return nil
}
