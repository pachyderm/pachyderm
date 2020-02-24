package s3

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/gogo/protobuf/types"
	"github.com/gorilla/mux"
	"github.com/pachyderm/pachyderm/src/client"
	pfsClient "github.com/pachyderm/pachyderm/src/client/pfs"
	pfsServer "github.com/pachyderm/pachyderm/src/server/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"github.com/pachyderm/s2"
)

var multipartChunkPathMatcher = regexp.MustCompile(`([^/]+)/([^/]+)/(.+)/([^/]+)/(\d+)`)
var multipartKeepPathMatcher = regexp.MustCompile(`([^/]+)/([^/]+)/(.+)/([^/]+)/\.keep`)

func multipartChunkArgs(path string) (repo string, branch string, key string, uploadID string, partNumber int, err error) {
	match := multipartChunkPathMatcher.FindStringSubmatch(path)

	if len(match) == 0 {
		err = errors.New("invalid file path found in multipath bucket")
		return
	}

	repo = match[1]
	branch = match[2]
	key = match[3]
	uploadID = match[4]
	partNumber, err = strconv.Atoi(match[5])
	if err != nil {
		err = fmt.Errorf("invalid file path found in multipath bucket: %s", err)
		return
	}
	return
}

func multipartKeepArgs(path string) (repo string, branch string, key string, uploadID string, err error) {
	match := multipartKeepPathMatcher.FindStringSubmatch(path)

	if len(match) == 0 {
		err = errors.New("invalid file path found in multipath bucket")
		return
	}

	repo = match[1]
	branch = match[2]
	key = match[3]
	uploadID = match[4]
	return
}

func parentDirPath(repo, branch, key, uploadID string) string {
	return path.Join(repo, branch, key, uploadID)
}

func chunkPath(repo, branch, key, uploadID string, partNumber int) string {
	return path.Join(parentDirPath(repo, branch, key, uploadID), strconv.Itoa(partNumber))
}

func keepPath(repo, branch, key, uploadID string) string {
	return path.Join(parentDirPath(repo, branch, key, uploadID), ".keep")
}

func (c *controller) ensureRepo(pc *client.APIClient) error {
	_, err := pc.InspectBranch(c.repo, "master")
	if err != nil {
		err = pc.UpdateRepo(c.repo)
		if err != nil {
			return err
		}

		err = pc.CreateBranch(c.repo, "master", "", nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *controller) ListMultipart(r *http.Request, bucketName, keyMarker, uploadIDMarker string, maxUploads int) (*s2.ListMultipartResult, error) {
	vars := mux.Vars(r)
	pc, err := c.pachClient(vars["authAccessKey"])
	if err != nil {
		return nil, err
	}

	if err = c.ensureRepo(pc); err != nil {
		return nil, err
	}

	bucket, err := c.driver.Bucket(pc, r, bucketName)
	if err != nil {
		return nil, err
	}

	result := s2.ListMultipartResult{
		Uploads: []s2.Upload{},
	}

	globPattern := path.Join(bucket.Repo, bucket.Commit, "*", "*", ".keep")
	err = pc.GlobFileF(c.repo, "master", globPattern, func(fileInfo *pfsClient.FileInfo) error {
		_, _, key, uploadID, err := multipartKeepArgs(fileInfo.File.Path)
		if err != nil {
			return nil
		}

		if key <= keyMarker || uploadID <= uploadIDMarker {
			return nil
		}

		if len(result.Uploads) >= maxUploads {
			if maxUploads > 0 {
				result.IsTruncated = true
			}
			return errutil.ErrBreak
		}

		timestamp, err := types.TimestampFromProto(fileInfo.Committed)
		if err != nil {
			return err
		}

		result.Uploads = append(result.Uploads, s2.Upload{
			Key:          key,
			UploadID:     uploadID,
			Initiator:    defaultUser,
			StorageClass: globalStorageClass,
			Initiated:    timestamp,
		})

		return nil
	})

	return &result, err
}

func (c *controller) InitMultipart(r *http.Request, bucketName, key string) (string, error) {
	vars := mux.Vars(r)
	pc, err := c.pachClient(vars["authAccessKey"])
	if err != nil {
		return "", err
	}

	if err = c.ensureRepo(pc); err != nil {
		return "", err
	}

	bucket, err := c.driver.Bucket(pc, r, bucketName)
	if err != nil {
		return "", err
	}

	uploadID := uuid.NewWithoutDashes()

	_, err = pc.PutFileOverwrite(c.repo, "master", keepPath(bucket.Repo, bucket.Commit, key, uploadID), strings.NewReader(""), 0)
	if err != nil {
		return "", err
	}

	return uploadID, nil
}

func (c *controller) AbortMultipart(r *http.Request, bucketName, key, uploadID string) error {
	vars := mux.Vars(r)
	pc, err := c.pachClient(vars["authAccessKey"])
	if err != nil {
		return err
	}

	if err = c.ensureRepo(pc); err != nil {
		return err
	}

	bucket, err := c.driver.Bucket(pc, r, bucketName)
	if err != nil {
		return err
	}

	_, err = pc.InspectFile(c.repo, "master", keepPath(bucket.Repo, bucket.Commit, key, uploadID))
	if err != nil {
		return s2.NoSuchUploadError(r)
	}

	err = pc.DeleteFile(c.repo, "master", parentDirPath(bucket.Repo, bucket.Commit, key, uploadID))
	if err != nil {
		return s2.InternalError(r, err)
	}

	return nil
}

func (c *controller) CompleteMultipart(r *http.Request, bucketName, key, uploadID string, parts []s2.Part) (*s2.CompleteMultipartResult, error) {
	vars := mux.Vars(r)
	pc, err := c.pachClient(vars["authAccessKey"])
	if err != nil {
		return nil, err
	}

	if err = c.ensureRepo(pc); err != nil {
		return nil, err
	}

	bucket, err := c.driver.Bucket(pc, r, bucketName)
	if err != nil {
		return nil, err
	}

	_, err = pc.InspectFile(c.repo, "master", keepPath(bucket.Repo, bucket.Commit, key, uploadID))
	if err != nil {
		if pfsServer.IsFileNotFoundErr(err) {
			return nil, s2.NoSuchUploadError(r)
		}
		return nil, err
	}

	// check if the destination file already exists, and if so, delete it
	_, err = pc.InspectFile(bucket.Repo, bucket.Commit, key)
	if err != nil && !pfsServer.IsFileNotFoundErr(err) && !pfsServer.IsNoHeadErr(err) {
		return nil, err
	} else if err == nil {
		err = pc.DeleteFile(bucket.Repo, bucket.Commit, key)
		if err != nil {
			if errutil.IsWriteToOutputBranchError(err) {
				return nil, writeToOutputBranchError(r)
			}
			return nil, err
		}
	}

	for i, part := range parts {
		srcPath := chunkPath(bucket.Repo, bucket.Commit, key, uploadID, part.PartNumber)

		fileInfo, err := pc.InspectFile(c.repo, "master", srcPath)
		if err != nil {
			if pfsServer.IsFileNotFoundErr(err) {
				return nil, s2.NoSuchUploadError(r)
			}
			return nil, err
		}

		// Only verify the ETag when it's of the same length as PFS file
		// hashes. This is because s3 clients will generally use md5 for
		// ETags, and would otherwise fail.
		expectedETag := fmt.Sprintf("%x", fileInfo.Hash)
		if len(part.ETag) == len(expectedETag) && part.ETag != expectedETag {
			return nil, s2.InvalidPartError(r)
		}

		if i < len(parts)-1 && fileInfo.SizeBytes < 5*1024*1024 {
			// each part, except for the last, is expected to be at least 5mb
			// in s3
			return nil, s2.EntityTooSmallError(r)
		}

		err = pc.CopyFile(c.repo, "master", srcPath, bucket.Repo, bucket.Commit, key, false)
		if err != nil {
			if errutil.IsWriteToOutputBranchError(err) {
				return nil, writeToOutputBranchError(r)
			}
			return nil, err
		}
	}

	err = pc.DeleteFile(c.repo, "master", parentDirPath(bucket.Repo, bucket.Commit, key, uploadID))
	if err != nil {
		return nil, err
	}

	fileInfo, err := pc.InspectFile(bucket.Repo, bucket.Commit, key)
	if err != nil && !pfsServer.IsOutputCommitNotFinishedErr(err) {
		return nil, err
	}

	result := s2.CompleteMultipartResult{Location: globalLocation}
	if fileInfo != nil {
		result.ETag = fmt.Sprintf("%x", fileInfo.Hash)
		result.Version = fileInfo.File.Commit.ID
	}

	return &result, nil
}

func (c *controller) ListMultipartChunks(r *http.Request, bucketName, key, uploadID string, partNumberMarker, maxParts int) (*s2.ListMultipartChunksResult, error) {
	vars := mux.Vars(r)
	pc, err := c.pachClient(vars["authAccessKey"])
	if err != nil {
		return nil, err
	}

	if err = c.ensureRepo(pc); err != nil {
		return nil, err
	}

	bucket, err := c.driver.Bucket(pc, r, bucketName)
	if err != nil {
		return nil, err
	}

	result := s2.ListMultipartChunksResult{
		Initiator:    &defaultUser,
		Owner:        &defaultUser,
		StorageClass: globalStorageClass,
		Parts:        []s2.Part{},
	}

	globPattern := path.Join(parentDirPath(bucket.Repo, bucket.Commit, key, uploadID), "*")
	err = pc.GlobFileF(c.repo, "master", globPattern, func(fileInfo *pfsClient.FileInfo) error {
		_, _, _, _, partNumber, err := multipartChunkArgs(fileInfo.File.Path)
		if err != nil {
			return nil
		}

		if partNumber <= partNumberMarker {
			return nil
		}

		if len(result.Parts) >= maxParts {
			if maxParts > 0 {
				result.IsTruncated = true
			}
			return errutil.ErrBreak
		}

		result.Parts = append(result.Parts, s2.Part{
			PartNumber: partNumber,
			ETag:       fmt.Sprintf("%x", fileInfo.Hash),
		})

		return nil
	})

	return &result, err
}

func (c *controller) UploadMultipartChunk(r *http.Request, bucketName, key, uploadID string, partNumber int, reader io.Reader) (string, error) {
	vars := mux.Vars(r)
	pc, err := c.pachClient(vars["authAccessKey"])
	if err != nil {
		return "", err
	}

	if err = c.ensureRepo(pc); err != nil {
		return "", err
	}

	bucket, err := c.driver.Bucket(pc, r, bucketName)
	if err != nil {
		return "", err
	}

	_, err = pc.InspectFile(c.repo, "master", keepPath(bucket.Repo, bucket.Commit, key, uploadID))
	if err != nil {
		if pfsServer.IsFileNotFoundErr(err) {
			return "", s2.NoSuchUploadError(r)
		}
		return "", err
	}

	path := chunkPath(bucket.Repo, bucket.Commit, key, uploadID, partNumber)
	_, err = pc.PutFileOverwrite(c.repo, "master", path, reader, 0)
	if err != nil {
		return "", err
	}

	fileInfo, err := pc.InspectFile(c.repo, "master", path)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", fileInfo.Hash), nil
}

func (c *controller) DeleteMultipartChunk(r *http.Request, bucketName, key, uploadID string, partNumber int) error {
	vars := mux.Vars(r)
	pc, err := c.pachClient(vars["authAccessKey"])
	if err != nil {
		return err
	}

	if err = c.ensureRepo(pc); err != nil {
		return err
	}

	bucket, err := c.driver.Bucket(pc, r, bucketName)
	if err != nil {
		return err
	}

	_, err = pc.InspectFile(c.repo, "master", keepPath(bucket.Repo, bucket.Commit, key, uploadID))
	if err != nil {
		return s2.NoSuchUploadError(r)
	}

	path := chunkPath(bucket.Repo, bucket.Commit, key, uploadID, partNumber)
	return pc.DeleteFile(c.repo, "master", path)
}
