//nolint:wrapcheck
// TODO: the s2 library checks the type of the error to decide how to handle it,
// which doesn't work properly with wrapped errors
package s3

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gogo/protobuf/types"
	glob "github.com/pachyderm/ohmyglob"
	"github.com/pachyderm/pachyderm/v2/src/internal/ancestry"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	pfsClient "github.com/pachyderm/pachyderm/v2/src/pfs"
	pfsServer "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	"github.com/pachyderm/s2"
)

func (c *controller) newContents(fileInfo *pfsClient.FileInfo) (s2.Contents, error) {
	t, err := types.TimestampFromProto(fileInfo.Committed)
	if err != nil {
		c.logger.Debugf("Warning: using nil timestamp (file probably in open commit), on error %s", err)
	}

	return s2.Contents{
		Key:          fileInfo.File.Path,
		LastModified: t,
		ETag:         fmt.Sprintf("%x", fileInfo.Hash),
		Size:         uint64(fileInfo.SizeBytes),
		StorageClass: globalStorageClass,
		Owner:        defaultUser,
	}, nil
}

func (c *controller) GetLocation(r *http.Request, bucketName string) (string, error) {
	c.logger.Debugf("GetLocation: %+v", bucketName)

	pc := c.requestClient(r)
	bucket, err := c.driver.bucket(pc, r, bucketName)
	if err != nil {
		return "", err
	}
	_, err = c.driver.bucketCapabilities(pc, r, bucket)
	if err != nil {
		return "", err
	}

	return globalLocation, nil
}

func (c *controller) ListObjects(r *http.Request, bucketName, prefix, marker, delimiter string, maxKeys int) (*s2.ListObjectsResult, error) {
	c.logger.Debugf("ListObjects: bucketName=%+v, prefix=%+v, marker=%+v, delimiter=%+v, maxKeys=%+v", bucketName, prefix, marker, delimiter, maxKeys)

	// Strip / from prefix to normalize: "/" means "all objects" and "/foo"
	// means the same as "foo"
	prefix = strings.TrimPrefix(prefix, "/")

	pc := c.requestClient(r)
	if delimiter != "" && delimiter != "/" {
		return nil, invalidDelimiterError(r)
	}

	bucket, err := c.driver.bucket(pc, r, bucketName)
	if err != nil {
		return nil, err
	}
	bucketCaps, err := c.driver.bucketCapabilities(pc, r, bucket)
	if err != nil {
		return nil, err
	}

	result := s2.ListObjectsResult{
		Contents:       []*s2.Contents{},
		CommonPrefixes: []*s2.CommonPrefixes{},
	}

	if !bucketCaps.readable {
		// serve empty results if we can't read the bucket; this helps with s3
		// conformance
		return &result, nil
	}

	recursive := delimiter == ""
	var pattern string
	if recursive {
		pattern = fmt.Sprintf("%s**", glob.QuoteMeta(prefix))
	} else {
		pattern = fmt.Sprintf("%s*", glob.QuoteMeta(prefix))
	}

	err = pc.GlobFile(bucket.Commit, pattern, func(fileInfo *pfsClient.FileInfo) error {
		if fileInfo.FileType == pfsClient.FileType_DIR {
			if fileInfo.File.Path == "/" {
				// skip the root directory
				return nil
			}
			if recursive {
				// skip directories if recursing
				return nil
			}
		} else if fileInfo.FileType != pfsClient.FileType_FILE {
			// skip anything that isn't a file or dir
			return nil
		}

		fileInfo.File.Path = fileInfo.File.Path[1:] // strip leading slash

		if !strings.HasPrefix(fileInfo.File.Path, prefix) {
			return nil
		}
		if fileInfo.File.Path <= marker {
			return nil
		}

		if len(result.Contents)+len(result.CommonPrefixes) >= maxKeys {
			if maxKeys > 0 {
				result.IsTruncated = true
			}
			return errutil.ErrBreak
		}
		if fileInfo.FileType == pfsClient.FileType_FILE {
			c, err := c.newContents(fileInfo)
			if err != nil {
				return err
			}

			result.Contents = append(result.Contents, &c)
		} else {
			result.CommonPrefixes = append(result.CommonPrefixes, &s2.CommonPrefixes{
				Prefix: fileInfo.File.Path,
				Owner:  defaultUser,
			})
		}

		return nil
	})

	return &result, err
}

func (c *controller) CreateBucket(r *http.Request, bucketName string) error {
	c.logger.Debugf("CreateBucket: %+v", bucketName)

	if !c.driver.canModifyBuckets() {
		return s2.NotImplementedError(r)
	}

	pc := c.requestClient(r)

	bucket, err := c.driver.bucket(pc, r, bucketName)
	if err != nil {
		return err
	}

	_, err = pc.PfsAPIClient.CreateRepo(pc.Ctx(), &pfs.CreateRepoRequest{Repo: bucket.Commit.Branch.Repo})
	if err != nil {
		if errutil.IsAlreadyExistError(err) {
			// Bucket already exists - this is not an error so long as the
			// branch being created is new. Verify if that is the case now,
			// since PFS' `CreateBranch` won't error out.
			_, err := pc.PfsAPIClient.InspectBranch(pc.Ctx(), &pfs.InspectBranchRequest{Branch: bucket.Commit.Branch})
			if err != nil {
				if !pfsServer.IsBranchNotFoundErr(err) {
					return s2.InternalError(r, grpcutil.ScrubGRPC(err))
				}
			} else {
				return s2.BucketAlreadyOwnedByYouError(r)
			}
		} else if ancestry.IsInvalidNameError(err) {
			return s2.InvalidBucketNameError(r)
		} else {
			return s2.InternalError(r, grpcutil.ScrubGRPC(err))
		}
	}

	_, err = pc.PfsAPIClient.CreateBranch(pc.Ctx(), &pfs.CreateBranchRequest{Branch: bucket.Commit.Branch})
	if err != nil {
		if ancestry.IsInvalidNameError(err) {
			return s2.InvalidBucketNameError(r)
		}
		return s2.InternalError(r, grpcutil.ScrubGRPC(err))
	}

	return nil
}

func (c *controller) DeleteBucket(r *http.Request, bucketName string) error {
	c.logger.Debugf("DeleteBucket: %+v", bucketName)

	if !c.driver.canModifyBuckets() {
		return s2.NotImplementedError(r)
	}

	pc := c.requestClient(r)

	bucket, err := c.driver.bucket(pc, r, bucketName)
	if err != nil {
		return err
	}

	// `DeleteBranch` does not return an error if a non-existing branch is
	// deleting. So first, we verify that the branch exists so we can
	// otherwise return a 404.
	branchInfo, err := pc.PfsAPIClient.InspectBranch(pc.Ctx(), &pfs.InspectBranchRequest{Branch: bucket.Commit.Branch})
	if err != nil {
		return maybeNotFoundError(r, grpcutil.ScrubGRPC(err))
	}

	hasFiles := false
	err = pc.WalkFile(branchInfo.Head, "", func(fileInfo *pfsClient.FileInfo) error {
		if fileInfo.FileType == pfsClient.FileType_FILE {
			hasFiles = true
			return errutil.ErrBreak
		}
		return nil
	})
	if err != nil {
		return s2.InternalError(r, err)
	}

	if hasFiles {
		return s2.BucketNotEmptyError(r)
	}

	_, err = pc.PfsAPIClient.DeleteBranch(pc.Ctx(), &pfs.DeleteBranchRequest{Branch: bucket.Commit.Branch})
	if err != nil {
		return s2.InternalError(r, grpcutil.ScrubGRPC(err))
	}

	repoInfo, err := pc.PfsAPIClient.InspectRepo(pc.Ctx(), &pfs.InspectRepoRequest{Repo: bucket.Commit.Branch.Repo})
	if err != nil {
		return s2.InternalError(r, grpcutil.ScrubGRPC(err))
	}

	// delete the repo if this was the last branch
	if len(repoInfo.Branches) == 0 {
		_, err = pc.PfsAPIClient.DeleteRepo(pc.Ctx(), &pfs.DeleteRepoRequest{Repo: bucket.Commit.Branch.Repo})
		if err != nil {
			return s2.InternalError(r, grpcutil.ScrubGRPC(err))
		}
	}

	return nil
}

func (c *controller) ListObjectVersions(r *http.Request, bucketName, prefix, keyMarker, versionIDMarker string, delimiter string, maxKeys int) (*s2.ListObjectVersionsResult, error) {
	// NOTE: because this endpoint isn't implemented, conformance tests will
	// fail on teardown. It's nevertheless unimplemented because it's too
	// expensive to pull off with PFS until this is implemented:
	// https://github.com/pachyderm/pachyderm/v2/issues/3896
	c.logger.Debugf("ListObjectVersions: bucketName=%+v, prefix=%+v, keyMarker=%+v, versionIDMarker=%+v, delimiter=%+v, maxKeys=%+v", bucketName, prefix, keyMarker, versionIDMarker, delimiter, maxKeys)
	return nil, s2.NotImplementedError(r)
}

func (c *controller) GetBucketVersioning(r *http.Request, bucketName string) (string, error) {
	c.logger.Debugf("GetBucketVersioning: %+v", bucketName)

	pc := c.requestClient(r)

	bucket, err := c.driver.bucket(pc, r, bucketName)
	if err != nil {
		return "", err
	}
	bucketCaps, err := c.driver.bucketCapabilities(pc, r, bucket)
	if err != nil {
		return "", err
	}

	if bucketCaps.historicVersions {
		return s2.VersioningEnabled, nil
	}
	return s2.VersioningDisabled, nil
}

func (c *controller) SetBucketVersioning(r *http.Request, bucketName, status string) error {
	c.logger.Debugf("SetBucketVersioning: bucketName=%+v, status=%+v", bucketName, status)

	pc := c.requestClient(r)

	bucket, err := c.driver.bucket(pc, r, bucketName)
	if err != nil {
		return err
	}
	bucketCaps, err := c.driver.bucketCapabilities(pc, r, bucket)
	if err != nil {
		return err
	}

	if bucketCaps.historicVersions {
		if status != s2.VersioningEnabled {
			return s2.NotImplementedError(r)
		}
	} else {
		if status != s2.VersioningDisabled {
			return s2.NotImplementedError(r)
		}
	}
	return nil
}
