package s3

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gogo/protobuf/types"
	"github.com/gorilla/mux"
	pfsClient "github.com/pachyderm/pachyderm/src/client/pfs"
	pfsServer "github.com/pachyderm/pachyderm/src/server/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"
	"github.com/pachyderm/s2"
)

func (c *controller) GetObject(r *http.Request, bucket, file, version string) (*s2.GetObjectResult, error) {
	vars := mux.Vars(r)
	pc, err := c.pachClient(vars["authAccessKey"])
	if err != nil {
		return nil, err
	}
	repo, branch, err := bucketArgs(r, bucket)
	if err != nil {
		return nil, err
	}

	branchInfo, err := pc.InspectBranch(repo, branch)
	if err != nil {
		return nil, maybeNotFoundError(r, err)
	}
	if branchInfo.Head == nil {
		return nil, s2.NoSuchKeyError(r)
	}
	if strings.HasSuffix(file, "/") {
		return nil, invalidFilePathError(r)
	}

	var commitInfo *pfsClient.CommitInfo
	commitID := branch
	if version != "" {
		commitInfo, err = pc.InspectCommit(repo, version)
		if err != nil {
			return nil, maybeNotFoundError(r, err)
		}
		if commitInfo.Branch.Name != branch {
			return nil, s2.NoSuchVersionError(r)
		}
		commitID = commitInfo.Commit.ID
	}

	fileInfo, err := pc.InspectFile(branchInfo.Branch.Repo.Name, commitID, file)
	if err != nil {
		return nil, maybeNotFoundError(r, err)
	}

	modTime, err := types.TimestampFromProto(fileInfo.Committed)
	if err != nil {
		return nil, err
	}

	content, err := pc.GetFileReadSeeker(branchInfo.Branch.Repo.Name, commitID, file)
	if err != nil {
		return nil, err
	}

	result := s2.GetObjectResult{
		ModTime:      modTime,
		Content:      content,
		ETag:         fmt.Sprintf("%x", fileInfo.Hash),
		Version:      commitID,
		DeleteMarker: false,
	}

	return &result, nil
}

func (c *controller) PutObject(r *http.Request, bucket, file string, reader io.Reader) (*s2.PutObjectResult, error) {
	vars := mux.Vars(r)
	pc, err := c.pachClient(vars["authAccessKey"])
	if err != nil {
		return nil, err
	}
	repo, branch, err := bucketArgs(r, bucket)
	if err != nil {
		return nil, err
	}

	branchInfo, err := pc.InspectBranch(repo, branch)
	if err != nil {
		return nil, maybeNotFoundError(r, err)
	}
	if strings.HasSuffix(file, "/") {
		return nil, invalidFilePathError(r)
	}

	_, err = pc.PutFileOverwrite(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, file, reader, 0)
	if err != nil {
		if errutil.IsWriteToOutputBranchError(err) {
			return nil, writeToOutputBranchError(r)
		}
		return nil, err
	}

	fileInfo, err := pc.InspectFile(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, file)
	if err != nil && !pfsServer.IsOutputCommitNotFinishedErr(err) {
		return nil, err
	}

	result := s2.PutObjectResult{}
	if fileInfo != nil {
		result.ETag = fmt.Sprintf("%x", fileInfo.Hash)
		result.Version = fileInfo.File.Commit.ID
	}

	return &result, nil
}

func (c *controller) DeleteObject(r *http.Request, bucket, file, version string) (*s2.DeleteObjectResult, error) {
	vars := mux.Vars(r)
	pc, err := c.pachClient(vars["authAccessKey"])
	if err != nil {
		return nil, err
	}
	repo, branch, err := bucketArgs(r, bucket)
	if err != nil {
		return nil, err
	}

	branchInfo, err := pc.InspectBranch(repo, branch)
	if err != nil {
		return nil, maybeNotFoundError(r, err)
	}
	if branchInfo.Head == nil {
		return nil, s2.NoSuchKeyError(r)
	}
	if strings.HasSuffix(file, "/") {
		return nil, invalidFilePathError(r)
	}
	if version != "" {
		return nil, s2.NotImplementedError(r)
	}

	if err = pc.DeleteFile(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, file); err != nil {
		if errutil.IsWriteToOutputBranchError(err) {
			return nil, writeToOutputBranchError(r)
		}
		return nil, maybeNotFoundError(r, err)
	}

	result := s2.DeleteObjectResult{
		Version:      "",
		DeleteMarker: false,
	}

	return &result, nil
}
