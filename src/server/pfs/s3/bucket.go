package s3

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/glob"
	"github.com/pachyderm/pachyderm/src/client"
	pfsClient "github.com/pachyderm/pachyderm/src/client/pfs"
	pfsServer "github.com/pachyderm/pachyderm/src/server/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"
	"github.com/pachyderm/s3server"
	"github.com/sirupsen/logrus"
)

func newContents(fileInfo *pfsClient.FileInfo) (s3server.Contents, error) {
	t, err := types.TimestampFromProto(fileInfo.Committed)
	if err != nil {
		return s3server.Contents{}, err
	}

	return s3server.Contents{
		Key:          fileInfo.File.Path,
		LastModified: t,
		ETag:         fmt.Sprintf("%x", fileInfo.Hash),
		Size:         fileInfo.SizeBytes,
		StorageClass: storageClass,
		Owner:        defaultUser,
	}, nil
}

func newCommonPrefixes(dir string) s3server.CommonPrefixes {
	return s3server.CommonPrefixes{
		Prefix: fmt.Sprintf("%s/", dir),
		Owner:  defaultUser,
	}
}

type bucketController struct {
	pc     *client.APIClient
	logger *logrus.Entry
}

func (c bucketController) GetLocation(r *http.Request, bucket string, result *s3server.LocationConstraint) *s3server.Error {
	repo, branch, s3Err := bucketArgs(r, bucket)
	if s3Err != nil {
		return s3Err
	}

	_, err := c.pc.InspectBranch(repo, branch)
	if err != nil {
		return maybeNotFoundError(r, err)
	}

	result.Location = "PACHYDERM"
	return nil
}

func (c bucketController) List(r *http.Request, bucket string, result *s3server.ListBucketResult) *s3server.Error {
	repo, branch, s3Err := bucketArgs(r, bucket)
	if s3Err != nil {
		return s3Err
	}

	if result.Delimiter != "" && result.Delimiter != "/" {
		return invalidDelimiterError(r)
	}

	// ensure the branch exists and has a head
	branchInfo, err := c.pc.InspectBranch(repo, branch)
	if err != nil {
		return maybeNotFoundError(r, err)
	}
	if branchInfo.Head == nil {
		// if there's no head commit, just print an empty list of files
		return nil
	}

	recursive := result.Delimiter == ""
	var pattern string
	if recursive {
		pattern = fmt.Sprintf("%s**", glob.QuoteMeta(result.Prefix))
	} else {
		pattern = fmt.Sprintf("%s*", glob.QuoteMeta(result.Prefix))
	}

	if err = c.pc.GlobFileF(repo, branch, pattern, func(fileInfo *pfsClient.FileInfo) error {
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

		if !strings.HasPrefix(fileInfo.File.Path, result.Prefix) {
			return nil
		}
		if fileInfo.File.Path <= result.Marker {
			return nil
		}

		if result.IsFull() {
			if result.MaxKeys > 0 {
				result.IsTruncated = true
			}
			return errutil.ErrBreak
		}
		if fileInfo.FileType == pfsClient.FileType_FILE {
			contents, err := newContents(fileInfo)
			if err != nil {
				return err
			}

			result.Contents = append(result.Contents, contents)
		} else {
			result.CommonPrefixes = append(result.CommonPrefixes, newCommonPrefixes(fileInfo.File.Path))
		}

		return nil
	}); err != nil {
		return s3server.InternalError(r, err)
	}

	if result.IsTruncated {
		if len(result.Contents) > 0 && len(result.CommonPrefixes) == 0 {
			result.NextMarker = result.Contents[len(result.Contents)-1].Key
		} else if len(result.Contents) == 0 && len(result.CommonPrefixes) > 0 {
			result.NextMarker = result.CommonPrefixes[len(result.CommonPrefixes)-1].Prefix
		} else if len(result.Contents) > 0 && len(result.CommonPrefixes) > 0 {
			lastContents := result.Contents[len(result.Contents)-1].Key
			lastCommonPrefixes := result.CommonPrefixes[len(result.CommonPrefixes)-1].Prefix

			if lastContents > lastCommonPrefixes {
				result.NextMarker = lastContents
			} else {
				result.NextMarker = lastCommonPrefixes
			}
		}
	}

	return nil
}

func (c bucketController) Create(r *http.Request, bucket string) *s3server.Error {
	repo, branch, s3Err := bucketArgs(r, bucket)
	if s3Err != nil {
		return s3Err
	}

	err := c.pc.CreateRepo(repo)
	if err != nil {
		if errutil.IsAlreadyExistError(err) {
			// Bucket already exists - this is not an error so long as the
			// branch being created is new. Verify if that is the case now,
			// since PFS' `CreateBranch` won't error out.
			_, err := c.pc.InspectBranch(repo, branch)
			if err != nil {
				if !pfsServer.IsBranchNotFoundErr(err) {
					return s3server.InternalError(r, err)
				}
			} else {
				return s3server.BucketAlreadyOwnedByYouError(r)
			}
		} else {
			return s3server.InternalError(r, err)
		}
	}

	err = c.pc.CreateBranch(repo, branch, "", nil)
	if err != nil {
		return s3server.InternalError(r, err)
	}

	return nil
}

func (c bucketController) Delete(r *http.Request, bucket string) *s3server.Error {
	repo, branch, s3Err := bucketArgs(r, bucket)
	if s3Err != nil {
		return s3Err
	}

	// `DeleteBranch` does not return an error if a non-existing branch is
	// deleting. So first, we verify that the branch exists so we can
	// otherwise return a 404.
	branchInfo, err := c.pc.InspectBranch(repo, branch)
	if err != nil {
		return maybeNotFoundError(r, err)
	}

	if branchInfo.Head != nil {
		hasFiles := false
		err = c.pc.Walk(branchInfo.Branch.Repo.Name, branchInfo.Head.ID, "", func(fileInfo *pfsClient.FileInfo) error {
			if fileInfo.FileType == pfsClient.FileType_FILE {
				hasFiles = true
				return errutil.ErrBreak
			}
			return nil
		})
		if err != nil {
			return s3server.InternalError(r, err)
		}

		if hasFiles {
			return s3server.BucketNotEmptyError(r)
		}
	}

	err = c.pc.DeleteBranch(repo, branch, false)
	if err != nil {
		return s3server.InternalError(r, err)
	}

	repoInfo, err := c.pc.InspectRepo(repo)
	if err != nil {
		return s3server.InternalError(r, err)
	}

	// delete the repo if this was the last branch
	if len(repoInfo.Branches) == 0 {
		err = c.pc.DeleteRepo(repo, false)
		if err != nil {
			return s3server.InternalError(r, err)
		}
	}

	return nil
}
