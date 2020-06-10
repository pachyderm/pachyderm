package s3

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gogo/protobuf/types"
	pfsServer "github.com/pachyderm/pachyderm/src/server/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"
	"github.com/pachyderm/s2"
)

func (c *controller) GetObject(r *http.Request, bucketName, file, version string) (*s2.GetObjectResult, error) {
	c.logger.Debugf("GetObject: bucketName=%+v, file=%+v, version=%+v", bucketName, file, version)

	pc, err := c.requestClient(r)
	if err != nil {
		return nil, err
	}

	if strings.HasSuffix(file, "/") {
		return nil, invalidFilePathError(r)
	}

	bucket, err := c.driver.bucket(pc, r, bucketName)
	if err != nil {
		return nil, err
	}
	bucketCaps, err := c.driver.bucketCapabilities(pc, r, bucket)
	if err != nil {
		return nil, err
	}
	if !bucketCaps.readable {
		return nil, s2.NoSuchKeyError(r)
	}

	if bucketCaps.historicVersions && version != "" {
		commitInfo, err := pc.InspectCommit(bucket.Repo, version)
		if err != nil {
			return nil, maybeNotFoundError(r, err)
		}
		if commitInfo.Branch.Name != bucket.Commit {
			return nil, s2.NoSuchVersionError(r)
		}
		bucket.Commit = commitInfo.Commit.ID
	}

	fileInfo, err := pc.InspectFile(bucket.Repo, bucket.Commit, file)
	if err != nil {
		return nil, maybeNotFoundError(r, err)
	}

	modTime, err := types.TimestampFromProto(fileInfo.Committed)
	if err != nil {
		return nil, err
	}

	content, err := pc.GetFileReadSeeker(bucket.Repo, bucket.Commit, file)
	if err != nil {
		return nil, err
	}

	result := s2.GetObjectResult{
		ModTime:      modTime,
		Content:      content,
		ETag:         fmt.Sprintf("%x", fileInfo.Hash),
		Version:      bucket.Commit,
		DeleteMarker: false,
	}

	return &result, nil
}

func (c *controller) CopyObject(r *http.Request, srcBucketName, srcFile string, srcObj *s2.GetObjectResult, destBucketName, destFile string) (string, error) {
	c.logger.Tracef("CopyObject: srcBucketName=%+v, srcFile=%+v, srcObj=%+v, destBucketName=%+v, destFile=%+v", srcBucketName, srcFile, srcObj, destBucketName, destFile)

	pc, err := c.requestClient(r)
	if err != nil {
		return "", err
	}

	if strings.HasSuffix(destFile, "/") {
		return "", invalidFilePathError(r)
	}

	srcBucket, err := c.driver.bucket(pc, r, srcBucketName)
	if err != nil {
		return "", err
	}
	// srcBucket capabilities were already verified, since s2 will call
	// `GetObject` under the hood before calling `CopyObject`

	destBucket, err := c.driver.bucket(pc, r, destBucketName)
	if err != nil {
		return "", err
	}
	destBucketCaps, err := c.driver.bucketCapabilities(pc, r, destBucket)
	if err != nil {
		return "", err
	}
	if !destBucketCaps.writable {
		return "", s2.NotImplementedError(r)
	}

	if err = pc.CopyFile(srcBucket.Repo, srcBucket.Commit, srcFile, destBucket.Repo, destBucket.Commit, destFile, true); err != nil {
		if errutil.IsWriteToOutputBranchError(err) {
			return "", writeToOutputBranchError(r)
		} else if errutil.IsNotADirectoryError(err) {
			return "", invalidFileParentError(r)
		} else if errutil.IsInvalidPathError(err) {
			return "", invalidFilePathError(r)
		}
		return "", err
	}

	fileInfo, err := pc.InspectFile(destBucket.Repo, destBucket.Commit, destFile)
	if err != nil && !pfsServer.IsOutputCommitNotFinishedErr(err) {
		return "", err
	}
	var version string
	if fileInfo != nil {
		version = fileInfo.File.Commit.ID
	}

	return version, nil
}

func (c *controller) PutObject(r *http.Request, bucketName, file string, reader io.Reader) (*s2.PutObjectResult, error) {
	c.logger.Debugf("PutObject: bucketName=%+v, file=%+v", bucketName, file)

	pc, err := c.requestClient(r)
	if err != nil {
		return nil, err
	}

	if strings.HasSuffix(file, "/") {
		return nil, invalidFilePathError(r)
	}

	bucket, err := c.driver.bucket(pc, r, bucketName)
	if err != nil {
		return nil, err
	}
	bucketCaps, err := c.driver.bucketCapabilities(pc, r, bucket)
	if err != nil {
		return nil, err
	}
	if !bucketCaps.writable {
		return nil, s2.NotImplementedError(r)
	}

	_, err = pc.PutFileOverwrite(bucket.Repo, bucket.Commit, file, reader, 0)
	if err != nil {
		if errutil.IsWriteToOutputBranchError(err) {
			return nil, writeToOutputBranchError(r)
		} else if errutil.IsNotADirectoryError(err) {
			return nil, invalidFileParentError(r)
		} else if errutil.IsInvalidPathError(err) {
			return nil, invalidFilePathError(r)
		}
		return nil, err
	}

	fileInfo, err := pc.InspectFile(bucket.Repo, bucket.Commit, file)
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

func (c *controller) DeleteObject(r *http.Request, bucketName, file, version string) (*s2.DeleteObjectResult, error) {
	c.logger.Debugf("DeleteObject: bucketName=%+v, file=%+v, version=%+v", bucketName, file, version)

	pc, err := c.requestClient(r)
	if err != nil {
		return nil, err
	}

	if strings.HasSuffix(file, "/") {
		return nil, invalidFilePathError(r)
	}
	if version != "" {
		return nil, s2.NotImplementedError(r)
	}

	bucket, err := c.driver.bucket(pc, r, bucketName)
	if err != nil {
		return nil, err
	}
	bucketCaps, err := c.driver.bucketCapabilities(pc, r, bucket)
	if err != nil {
		return nil, err
	}
	if !bucketCaps.writable {
		return nil, s2.NotImplementedError(r)
	}

	if err = pc.DeleteFile(bucket.Repo, bucket.Commit, file); err != nil {
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
