package s3

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gogo/protobuf/types"
	"github.com/gorilla/mux"
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

	if strings.HasSuffix(file, "/") {
		return nil, invalidFilePathError(r)
	}

	ref, err := c.driver.DereferenceBucket(pc, r, bucket)
	if err != nil {
		return nil, err
	}

	if c.driver.CanGetHistoricObject() && version != "" {
		commitInfo, err := pc.InspectCommit(ref.repo, version)
		if err != nil {
			return nil, maybeNotFoundError(r, err)
		}
		if commitInfo.Branch.Name != ref.commit {
			return nil, s2.NoSuchVersionError(r)
		}
		ref.commit = commitInfo.Commit.ID
	}

	fileInfo, err := pc.InspectFile(ref.repo, ref.commit, file)
	if err != nil {
		return nil, maybeNotFoundError(r, err)
	}

	modTime, err := types.TimestampFromProto(fileInfo.Committed)
	if err != nil {
		return nil, err
	}

	content, err := pc.GetFileReadSeeker(ref.repo, ref.commit, file)
	if err != nil {
		return nil, err
	}

	result := s2.GetObjectResult{
		ModTime:      modTime,
		Content:      content,
		ETag:         fmt.Sprintf("%x", fileInfo.Hash),
		Version:      ref.commit,
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

	if strings.HasSuffix(file, "/") {
		return nil, invalidFilePathError(r)
	}

	ref, err := c.driver.DereferenceBucket(pc, r, bucket)
	if err != nil {
		return nil, err
	}

	_, err = pc.PutFileOverwrite(ref.repo, ref.commit, file, reader, 0)
	if err != nil {
		if errutil.IsWriteToOutputBranchError(err) {
			return nil, writeToOutputBranchError(r)
		}
		return nil, err
	}

	fileInfo, err := pc.InspectFile(ref.repo, ref.commit, file)
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

	if strings.HasSuffix(file, "/") {
		return nil, invalidFilePathError(r)
	}
	if version != "" {
		return nil, s2.NotImplementedError(r)
	}

	ref, err := c.driver.DereferenceBucket(pc, r, bucket)
	if err != nil {
		return nil, err
	}

	if err = pc.DeleteFile(ref.repo, ref.commit, file); err != nil {
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
