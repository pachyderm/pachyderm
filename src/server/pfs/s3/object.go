package s3

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
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

func (c *objectController) GetObject(r *http.Request, bucket, file string, result *s2.GetObjectResult) error {
	repo, branch, err := bucketArgs(r, bucket)
	if err != nil {
		return err
	}

	branchInfo, err := c.pc.InspectBranch(repo, branch)
	if err != nil {
		return maybeNotFoundError(r, err)
	}
	if branchInfo.Head == nil {
		return s2.NoSuchKeyError(r)
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
		return s2.InternalError(r, err)
	}

	reader, err := c.pc.GetFileReadSeeker(branchInfo.Branch.Repo.Name, branchInfo.Head.ID, file)
	if err != nil {
		return s2.InternalError(r, err)
	}

	result.Name = file
	result.ETag = fmt.Sprintf("%x", fileInfo.Hash)
	result.ModTime = timestamp
	result.Content = reader
	return nil
}

func (c *objectController) PutObject(r *http.Request, bucket, file string, reader io.Reader) error {
	repo, branch, err := bucketArgs(r, bucket)
	if err != nil {
		return err
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
		return s2.InternalError(r, err)
	}

	return nil
}

func (c *objectController) DeleteObject(r *http.Request, bucket, key string) error {
	repo, branch, err := bucketArgs(r, bucket)
	if err != nil {
		return err
	}

	branchInfo, err := c.pc.InspectBranch(repo, branch)
	if err != nil {
		return maybeNotFoundError(r, err)
	}
	if branchInfo.Head == nil {
		return s2.NoSuchKeyError(r)
	}
	if strings.HasSuffix(key, "/") {
		return invalidFilePathError(r)
	}

	if err := c.pc.DeleteFile(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, key); err != nil {
		return maybeNotFoundError(r, err)
	}

	return nil
}
