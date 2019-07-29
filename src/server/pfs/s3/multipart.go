package s3

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	pfsClient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"github.com/pachyderm/s2"
	"github.com/sirupsen/logrus"
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
	return fmt.Sprintf("%s/%s/%s/%s", repo, branch, key, uploadID)
}

func chunkPath(repo, branch, key, uploadID string, partNumber int) string {
	return fmt.Sprintf("%s/%d", parentDirPath(repo, branch, key, uploadID), partNumber)
}

func keepPath(repo, branch, key, uploadID string) string {
	return fmt.Sprintf("%s/.keep", parentDirPath(repo, branch, key, uploadID))
}

type multipartController struct {
	pc     *client.APIClient
	logger *logrus.Entry

	// Name of the PFS repo holding multipart content
	repo string

	// the maximum number of allowed parts that can be associated with any
	// given file
	maxAllowedParts int
}

func newMultipartController(pc *client.APIClient, logger *logrus.Entry, repo string, maxAllowedParts int) *multipartController {
	return &multipartController{
		pc:              pc,
		logger:          logger,
		repo:            repo,
		maxAllowedParts: maxAllowedParts,
	}
}

func (c *multipartController) ensureRepo() error {
	_, err := c.pc.InspectBranch(c.repo, "master")
	if err != nil {
		err = c.pc.CreateRepo(c.repo)
		if err != nil && !strings.Contains(err.Error(), "as it already exists") {
			return err
		}

		err = c.pc.CreateBranch(c.repo, "master", "", nil)
		if err != nil && !strings.Contains(err.Error(), "as it already exists") {
			return err
		}
	}

	return nil
}

func (c *multipartController) ListMultipart(r *http.Request, repo, keyMarker, uploadIDMarker string, maxUploads int) (isTruncated bool, uploads []s2.Upload, err error) {
	branch := branchArg(r)
	_, err = c.pc.InspectBranch(repo, branch)
	if err != nil {
		err = maybeNotFoundError(r, err)
		return
	}
	if err = c.ensureRepo(); err != nil {
		return
	}

	err = c.pc.GlobFileF(c.repo, "master", fmt.Sprintf("%s/%s/*/*/.keep", repo, branch), func(fileInfo *pfsClient.FileInfo) error {
		_, _, key, uploadID, err := multipartKeepArgs(fileInfo.File.Path)
		if err != nil {
			return nil
		}

		if key <= keyMarker || uploadID <= uploadIDMarker {
			return nil
		}

		if len(uploads) >= maxUploads {
			if maxUploads > 0 {
				isTruncated = true
			}
			return errutil.ErrBreak
		}

		timestamp, err := types.TimestampFromProto(fileInfo.Committed)
		if err != nil {
			return err
		}

		uploads = append(uploads, s2.Upload{
			Key:          key,
			UploadID:     uploadID,
			Initiator:    defaultUser,
			StorageClass: globalStorageClass,
			Initiated:    timestamp,
		})

		return nil
	})

	return
}

func (c *multipartController) InitMultipart(r *http.Request, repo, key string) (uploadID string, err error) {
	branch := branchArg(r)
	_, err = c.pc.InspectBranch(repo, branch)
	if err != nil {
		err = maybeNotFoundError(r, err)
		return
	}
	if err = c.ensureRepo(); err != nil {
		return "", err
	}

	uploadID = uuid.NewWithoutDashes()

	path := fmt.Sprintf("%s/.keep", parentDirPath(repo, branch, key, uploadID))
	_, err = c.pc.PutFileOverwrite(c.repo, "master", path, strings.NewReader(""), 0)
	if err != nil {
		return
	}
	return
}

func (c *multipartController) AbortMultipart(r *http.Request, repo, key, uploadID string) error {
	branch := branchArg(r)
	_, err := c.pc.InspectBranch(repo, branch)
	if err != nil {
		return maybeNotFoundError(r, err)
	}
	if err = c.ensureRepo(); err != nil {
		return err
	}

	_, err = c.pc.InspectFile(c.repo, "master", keepPath(repo, branch, key, uploadID))
	if err != nil {
		return s2.NoSuchUploadError(r)
	}

	err = c.pc.DeleteFile(c.repo, "master", parentDirPath(repo, branch, key, uploadID))
	if err != nil {
		return s2.InternalError(r, err)
	}

	return nil
}

func (c *multipartController) CompleteMultipart(r *http.Request, repo, key, uploadID string, parts []s2.Part) (location, etag string, err error) {
	branch := branchArg(r)
	_, err = c.pc.InspectBranch(repo, branch)
	if err != nil {
		err = maybeNotFoundError(r, err)
		return
	}
	if err = c.ensureRepo(); err != nil {
		return
	}

	_, err = c.pc.InspectFile(c.repo, "master", keepPath(repo, branch, key, uploadID))
	if err != nil {
		err = s2.NoSuchUploadError(r)
		return
	}

	for i, part := range parts {
		srcPath := chunkPath(repo, branch, key, uploadID, part.PartNumber)

		var fileInfo *pfsClient.FileInfo
		fileInfo, err = c.pc.InspectFile(c.repo, "master", srcPath)
		if err != nil {
			err = s2.InvalidPartError(r)
			return
		}

		// Only verify the ETag when it's of the same length as PFS file
		// hashes. This is because s3 clients will generally use md5 for
		// ETags, and would otherwise fail.
		expectedETag := fmt.Sprintf("%x", fileInfo.Hash)
		if len(part.ETag) == len(expectedETag) && part.ETag != expectedETag {
			err = s2.InvalidPartError(r)
			return
		}

		if i < len(parts)-1 && fileInfo.SizeBytes < 5*1024*1024 {
			// each part, except for the last, is expected to be at least 5mb
			// in s3
			err = s2.EntityTooSmallError(r)
			return
		}

		err = c.pc.CopyFile(c.repo, "master", srcPath, repo, branch, key, false)
		if err != nil {
			err = s2.InternalError(r, err)
			return
		}
	}

	// TODO: verify that this works
	err = c.pc.DeleteFile(c.repo, "master", parentDirPath(repo, branch, key, uploadID))
	if err != nil {
		return
	}

	fileInfo, err := c.pc.InspectFile(repo, branch, key)
	if err != nil {
		return
	}

	location = globalLocation
	etag = fmt.Sprintf("%x", fileInfo.Hash)
	return
}

func (c *multipartController) ListMultipartChunks(r *http.Request, repo, key, uploadID string, partNumberMarker, maxParts int) (initiator, owner *s2.User, storageClass string, isTruncated bool, parts []s2.Part, err error) {
	branch := branchArg(r)
	_, err = c.pc.InspectBranch(repo, branch)
	if err != nil {
		err = maybeNotFoundError(r, err)
		return
	}
	if err = c.ensureRepo(); err != nil {
		return
	}

	err = c.pc.GlobFileF(c.repo, "master", fmt.Sprintf("%s/%s/%s/%s/*", repo, branch, key, uploadID), func(fileInfo *pfsClient.FileInfo) error {
		_, _, _, _, partNumber, err := multipartChunkArgs(fileInfo.File.Path)
		if err != nil {
			return nil
		}

		if partNumber <= partNumberMarker {
			return nil
		}

		if len(parts) >= maxParts {
			if maxParts > 0 {
				isTruncated = true
			}
			return errutil.ErrBreak
		}

		parts = append(parts, s2.Part{
			PartNumber: partNumber,
			ETag:       fmt.Sprintf("%x", fileInfo.Hash),
		})

		return nil
	})
	if err != nil {
		return
	}

	initiator = &defaultUser
	owner = &defaultUser
	storageClass = globalStorageClass
	return
}

func (c *multipartController) UploadMultipartChunk(r *http.Request, repo, key, uploadID string, partNumber int, reader io.Reader) (etag string, err error) {
	branch := branchArg(r)
	_, err = c.pc.InspectBranch(repo, branch)
	if err != nil {
		err = maybeNotFoundError(r, err)
		return
	}
	if err = c.ensureRepo(); err != nil {
		return "", err
	}

	_, err = c.pc.InspectFile(c.repo, "master", keepPath(repo, branch, key, uploadID))
	if err != nil {
		err = s2.NoSuchUploadError(r)
		return
	}

	path := chunkPath(repo, branch, key, uploadID, partNumber)
	_, err = c.pc.PutFileOverwrite(c.repo, "master", path, reader, 0)
	if err != nil {
		return
	}

	fileInfo, err := c.pc.InspectFile(c.repo, "master", path)
	if err != nil {
		return
	}

	etag = fmt.Sprintf("%x", fileInfo.Hash)
	return
}

func (c *multipartController) DeleteMultipartChunk(r *http.Request, repo, key, uploadID string, partNumber int) error {
	branch := branchArg(r)
	_, err := c.pc.InspectBranch(repo, branch)
	if err != nil {
		return maybeNotFoundError(r, err)
	}
	if err = c.ensureRepo(); err != nil {
		return err
	}

	_, err = c.pc.InspectFile(c.repo, "master", keepPath(repo, branch, key, uploadID))
	if err != nil {
		return s2.NoSuchUploadError(r)
	}

	path := chunkPath(repo, branch, key, uploadID, partNumber)
	return c.pc.DeleteFile(c.repo, "master", path)
}
