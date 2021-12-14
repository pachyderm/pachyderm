package s3

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/client"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	pfsServer "github.com/pachyderm/pachyderm/v2/src/server/pfs"
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

	if version != "" && !bucketCaps.historicVersions {
		return nil, s2.NotImplementedError(r)
	}

	if result, err := c.driver.getObject(pc, bucket, file, version); err != nil {
		return nil, maybeNotFoundError(r, err)
	} else {
		return result, nil
	}
}

func (pfs *pachFS) getObject(pc *client.APIClient, bucket *Bucket, file, version string) (*s2.GetObjectResult, error) {
	commitID := bucket.Commit.ID
	if version != "" {
		commitID = version
	}

	fileInfo, err := pc.InspectFile(bucket.Commit, file)
	if err != nil {
		return nil, err
	}

	modTime, err := types.TimestampFromProto(fileInfo.Committed)
	if err != nil {
		return nil, err
	}

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

func (lfs *localFS) getObject(pc *client.APIClient, bucket *Bucket, path, version string) (*s2.GetObjectResult, error) {
	fullPath := filepath.Join(bucket.Path, path)
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	return &s2.GetObjectResult{
		ModTime: info.ModTime(),
		Content: file,
	}, nil
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

	if version, err := c.driver.copyObject(pc, srcBucket, destBucket, srcFile, destFile); err != nil {
		if errutil.IsWriteToOutputBranchError(err) {
			return "", writeToOutputBranchError(r)
		} else if errutil.IsNotADirectoryError(err) {
			return "", invalidFileParentError(r)
		} else if errutil.IsInvalidPathError(err) {
			return "", invalidFilePathError(r)
		}
		return "", err
	} else {
		return version, nil
	}
}

func (pfs *pachFS) copyObject(pc *client.APIClient, srcBucket, destBucket *Bucket, srcFile, destFile string) (string, error) {
	if err := pc.CopyFile(destBucket.Commit, destFile, srcBucket.Commit, srcFile); err != nil {
		return "", err
	}

	fileInfo, err := pc.InspectFile(destBucket.Commit, destFile)
	if err != nil && !pfsServer.IsOutputCommitNotFinishedErr(err) {
		return "", err
	}
	var version string
	if fileInfo != nil {
		version = fileInfo.File.Commit.ID
	}

	return version, nil
}

func (lfs *localFS) copyObject(pc *client.APIClient, srcBucket, destBucket *Bucket, srcFile, destFile string) (string, error) {
	src, err := os.Open(filepath.Join(srcBucket.Path, srcFile))
	if err != nil {
		return "", err
	}
	defer src.Close()
	dst, err := os.Create(filepath.Join(destBucket.Path, destFile))
	if err != nil {
		return "", err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return "", err
	}

	return "", nil
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

	if result, err := c.driver.putObject(pc, bucket, file, reader); err != nil {
		if errutil.IsWriteToOutputBranchError(err) {
			return nil, writeToOutputBranchError(r)
		} else if errutil.IsNotADirectoryError(err) {
			return nil, invalidFileParentError(r)
		} else if errutil.IsInvalidPathError(err) {
			return nil, invalidFilePathError(r)
		}
		return nil, err
	} else {
		return result, nil
	}
}

func (pfs *pachFS) putObject(pc *client.APIClient, bucket *Bucket, path string, reader io.Reader) (*s2.PutObjectResult, error) {
	bucketCommit := bucket.Commit
	if err := pc.PutFile(bucketCommit, path, reader); err != nil {
		return nil, err
	}

	fileInfo, err := pc.InspectFile(bucketCommit, path)
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

func (lfs *localFS) putObject(pc *client.APIClient, bucket *Bucket, path string, reader io.Reader) (*s2.PutObjectResult, error) {
	file, err := os.Create(filepath.Join(bucket.Path, path))
	if err != nil {
		return nil, err
	}
	defer file.Close()
	_, err = io.Copy(file, reader)
	if err != nil {
		return nil, err
	}
	return &s2.PutObjectResult{}, nil
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

	if err = c.driver.deleteObject(pc, bucket, file); err != nil {
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

func (pfs *pachFS) deleteObject(pc *client.APIClient, bucket *Bucket, file string) error {
	return pc.DeleteFile(bucket.Commit, file)
}

func (lfs *localFS) deleteObject(pc *client.APIClient, bucket *Bucket, file string) error {
	err := os.Remove(filepath.Join(bucket.Path, file))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}
