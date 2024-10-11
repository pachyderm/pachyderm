package s3

// TODO: the s2 library checks the type of the error to decide how to handle it,
// which doesn't work properly with wrapped errors

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	pfsServer "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	"github.com/pachyderm/s2"
	"go.uber.org/zap"
)

func (c *controller) GetObject(r *http.Request, bucketName, file, version string) (*s2.GetObjectResult, error) {
	defer log.Span(r.Context(), "GetObject", zap.String("bucketName", bucketName), zap.String("file", file), zap.String("version", version))()

	pc := c.requestClient(r)
	file = strings.TrimSuffix(file, "/")

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

	commitID := bucket.Commit.Id
	if version != "" {
		if !bucketCaps.historicVersions {
			return nil, s2.NotImplementedError(r)
		}
		commitID = version
	}

	// We use listFileResult[0] rather than InspectFile result since InspectFile
	// on a path that has both a file and a directory in it returns the
	// directory. However, ListFile will show it as a file, if it exists.
	var firstFile *pfs.FileInfo
	err = pc.ListFile(bucket.Commit, file, func(fi *pfs.FileInfo) (retErr error) {
		if firstFile == nil {
			firstFile = fi
		}
		return errutil.ErrBreak
	})
	if err != nil {
		return nil, maybeNotFoundError(r, err)
	}
	if firstFile == nil {
		// we never set it, probably zero results
		return nil, s2.NoSuchKeyError(r)
	}
	fileInfo := firstFile

	// the exact object named does not exist, but perhaps is a "directory".
	// "directories" do not actually exist, and certainly cannot be read.
	// ("seeker can't seek")
	if fileInfo.File.Path[1:] != file {
		return nil, s2.NoSuchKeyError(r)
	}

	modTime := fileInfo.Committed.AsTime()

	content, err := pc.GetFileReadSeeker(bucket.Commit, file)
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

func (c *controller) CopyObject(r *http.Request, srcBucketName, srcFile string, srcObj *s2.GetObjectResult, destBucketName, destFile string) (string, error) {
	defer log.Span(r.Context(), "CopyObject", zap.String("srcBucketName", srcBucketName), zap.String("srcFile", srcFile), zap.Any("srcObj", srcObj), zap.String("destBucketName", destBucketName), zap.String("destFile", destFile))()

	pc := c.requestClient(r)
	destFile = strings.TrimSuffix(destFile, "/")

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

	if err = pc.CopyFile(destBucket.Commit, destFile, srcBucket.Commit, srcFile); err != nil {
		if errutil.IsWriteToOutputBranchError(err) {
			return "", writeToOutputBranchError(r)
		} else if errutil.IsNotADirectoryError(err) {
			return "", invalidFileParentError(r)
		} else if errutil.IsInvalidPathError(err) {
			return "", invalidFilePathError(r)
		}
		return "", err
	}

	fileInfo, err := pc.InspectFile(destBucket.Commit, destFile)
	if err != nil && !pfsServer.IsOutputCommitNotFinishedErr(err) {
		return "", err
	}
	var version string
	if fileInfo != nil {
		version = fileInfo.File.Commit.Id
	}

	return version, nil
}

func (c *controller) PutObject(r *http.Request, bucketName, file string, reader io.Reader) (*s2.PutObjectResult, error) {
	defer log.Span(r.Context(), "PutObject", zap.String("bucketName", bucketName), zap.String("file", file))()

	pc := c.requestClient(r)
	file = strings.TrimSuffix(file, "/")

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

	bucketCommit := bucket.Commit
	if err := pc.PutFile(bucketCommit, file, reader); err != nil {
		if errutil.IsWriteToOutputBranchError(err) {
			return nil, writeToOutputBranchError(r)
		} else if errutil.IsNotADirectoryError(err) {
			return nil, invalidFileParentError(r)
		} else if errutil.IsInvalidPathError(err) {
			return nil, invalidFilePathError(r)
		}
		return nil, err
	}

	fileInfo, err := pc.InspectFile(bucketCommit, file)
	if err != nil && !pfsServer.IsOutputCommitNotFinishedErr(err) {
		return nil, err
	}

	result := s2.PutObjectResult{}
	if fileInfo != nil {
		result.ETag = fmt.Sprintf("%x", fileInfo.Hash)
		result.Version = fileInfo.File.Commit.Id
	}

	return &result, nil
}

func (c *controller) DeleteObject(r *http.Request, bucketName, file, version string) (*s2.DeleteObjectResult, error) {
	defer log.Span(r.Context(), "DeleteObject", zap.String("bucketName", bucketName), zap.String("file", file), zap.String("version", version))()

	pc := c.requestClient(r)
	file = strings.TrimSuffix(file, "/")
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

	if err = pc.DeleteFile(bucket.Commit, file); err != nil {
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
