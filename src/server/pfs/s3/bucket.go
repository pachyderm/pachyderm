package s3

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"strconv"

	"github.com/gobwas/glob"
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"
)

const defaultMaxKeys int = 1000

// the raw XML returned for a request to get the location of a bucket
const locationSource = `<?xml version="1.0" encoding="UTF-8"?>
<LocationConstraint xmlns="http://s3.amazonaws.com/doc/2006-03-01/">PACHYDERM</LocationConstraint>`

// ListBucketResult is an XML-encodable listing of files/objects in a
// repo/bucket
type ListBucketResult struct {
	Name           string           `xml:"Name"`
	Prefix         string           `xml:"Prefix"`
	Marker         string           `xml:"Marker"`
	MaxKeys        int              `xml:"MaxKeys"`
	IsTruncated    bool             `xml:"IsTruncated"`
	Contents       []Contents       `xml:"Contents"`
	CommonPrefixes []CommonPrefixes `xml:"CommonPrefixes"`
}

func (r *ListBucketResult) isFull() bool {
	return len(r.Contents)+len(r.CommonPrefixes) >= r.MaxKeys
}

// Contents is an individual file/object
type Contents struct {
	Key          string    `xml:"Key"`
	LastModified time.Time `xml:"LastModified"`
	ETag         string    `xml:"ETag"`
	Size         uint64    `xml:"Size"`
	StorageClass string    `xml:"StorageClass"`
	Owner        User      `xml:"Owner"`
}

func newContents(fileInfo *pfs.FileInfo) (Contents, error) {
	t, err := types.TimestampFromProto(fileInfo.Committed)
	if err != nil {
		return Contents{}, err
	}

	return Contents{
		Key:          fileInfo.File.Path,
		LastModified: t,
		ETag:         "",
		Size:         fileInfo.SizeBytes,
		StorageClass: storageClass,
		Owner:        defaultUser,
	}, nil
}

// CommonPrefixes is an individual PFS directory
type CommonPrefixes struct {
	Prefix string `xml:"Prefix"`
}

func newCommonPrefixes(dir string) CommonPrefixes {
	return CommonPrefixes{
		Prefix: fmt.Sprintf("%s/", dir),
	}
}

type bucketHandler struct {
	pc *client.APIClient
}

func newBucketHandler(pc *client.APIClient) bucketHandler {
	return bucketHandler{pc: pc}
}

func (h bucketHandler) location(w http.ResponseWriter, r *http.Request) {
	repo, branch := bucketArgs(w, r)
	
	_, err := h.pc.InspectBranch(repo, branch)
	if err != nil {
		notFoundError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(locationSource))
}

func (h bucketHandler) get(w http.ResponseWriter, r *http.Request) {
	repo, branch := bucketArgs(w, r)

	// ensure the branch exists and has a head
	branchInfo, err := h.pc.InspectBranch(repo, branch)
	if err != nil {
		notFoundError(w, r, err)
		return
	}

	maxKeys := defaultMaxKeys
	maxKeysStr := r.FormValue("max-keys")
	if maxKeysStr != "" {
		i, err := strconv.Atoi(maxKeysStr)
		if err != nil || i < 0 || i > defaultMaxKeys {
			invalidArgument(w, r)
			return
		}
		maxKeys = i	
	}

	result := &ListBucketResult{
		Name:        repo,
		Prefix:      r.FormValue("prefix"),
		Marker:      r.FormValue("marker"),
		MaxKeys:     maxKeys,
		IsTruncated: false,
	}

	delimiter := r.FormValue("delimiter")
	if delimiter != "" && delimiter != "/" {
		invalidDelimiterError(w, r)
		return
	}

	if branchInfo.Head == nil {
		// if there's no head commit, just print an empty list of files
		writeXML(w, http.StatusOK, result)
	} else if delimiter == "" {
		h.listRecursive(w, r, result, branch)
	} else {
		h.list(w, r, result, branch)
	}
}

func (h bucketHandler) listRecursive(w http.ResponseWriter, r *http.Request, result *ListBucketResult, branch string) {
	err := h.pc.Walk(result.Name, branch, filepath.Dir(result.Prefix), func(fileInfo *pfs.FileInfo) error {
		fileInfo = updateFileInfo(branch, result.Marker, fileInfo)
		if fileInfo == nil {
			return nil
		}
		if !strings.HasPrefix(fileInfo.File.Path, result.Prefix) {
			return nil
		}
		if result.isFull() {
			result.IsTruncated = true
			return errutil.ErrBreak
		}
		if fileInfo.FileType == pfs.FileType_FILE {
			contents, err := newContents(fileInfo)
			if err != nil {
				return err
			}
			result.Contents = append(result.Contents, contents)
		} else {
			result.CommonPrefixes = append(result.CommonPrefixes, newCommonPrefixes(fileInfo.File.Path))
		}
		return nil
	})

	if err != nil {
		internalError(w, r, err)
		return
	}

	writeXML(w, http.StatusOK, result)
}

func (h bucketHandler) list(w http.ResponseWriter, r *http.Request, result *ListBucketResult, branch string) {
	pattern := fmt.Sprintf("%s*", glob.QuoteMeta(result.Prefix))
	fileInfos, err := h.pc.GlobFile(result.Name, branch, pattern)
	if err != nil {
		internalError(w, r, err)
		return
	}

	for _, fileInfo := range fileInfos {
		fileInfo = updateFileInfo(branch, result.Marker, fileInfo)
		if fileInfo == nil {
			continue
		}
		if result.isFull() {
			result.IsTruncated = true
			break
		}
		if fileInfo.FileType == pfs.FileType_FILE {
			contents, err := newContents(fileInfo)
			if err != nil {
				internalError(w, r, err)
				return
			}
			result.Contents = append(result.Contents, contents)
		} else {
			result.CommonPrefixes = append(result.CommonPrefixes, newCommonPrefixes(fileInfo.File.Path))
		}
	}

	writeXML(w, http.StatusOK, result)
}

func (h bucketHandler) put(w http.ResponseWriter, r *http.Request) {
	repo, branch := bucketArgs(w, r)

	err := h.pc.CreateRepo(repo)
	if err != nil {
		if strings.Contains(err.Error(), "as it already exists") {
			// Bucket already exists - this is not an error so long as the
			// branch being created is new. Verify if that is the case now,
			// since PFS' `CreateBranch` won't error out.
			_, err := h.pc.InspectBranch(repo, branch)
			if err != nil {
				if !branchNotFoundMatcher.MatchString(err.Error()) {
					internalError(w, r, err)
					return
				}
			} else {
				bucketAlreadyExistsError(w, r)
				return
			}
		} else {
			internalError(w, r, err)
			return
		}
	}

	err = h.pc.CreateBranch(repo, branch, "", nil)
	if err != nil {
		internalError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h bucketHandler) del(w http.ResponseWriter, r *http.Request) {
	repo, branch := bucketArgs(w, r)

	// `DeleteBranch` does not return an error if a non-existing branch is
	// deleting. So first, we verify that the branch exists so we can
	// otherwise return a 404.
	_, err := h.pc.InspectBranch(repo, branch)
	if err != nil {
		notFoundError(w, r, err)
		return
	}

	err = h.pc.DeleteBranch(repo, branch, false)
	if err != nil {
		internalError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// updateFileInfo takes in a `FileInfo`, and updates it to be used in s3
// object listings:
// 1) if nil is returned, the `FileInfo` should not be included in the list
// 2) the path is updated to remove the leading slash
func updateFileInfo(branch, marker string, fileInfo *pfs.FileInfo) *pfs.FileInfo {
	if fileInfo.FileType == pfs.FileType_DIR {
		if fileInfo.File.Path == "/" {
			// skip the root directory
			return nil
		}
	} else if fileInfo.FileType == pfs.FileType_FILE {
		if strings.HasSuffix(fileInfo.File.Path, ".s3g.json") {
			// skip metadata files
			return nil
		}
	} else {
		// skip anything that isn't a file or dir
		return nil
	}
	fileInfo.File.Path = fileInfo.File.Path[1:] // strip leading slash
	if fileInfo.File.Path <= marker {
		// skip file paths below the marker
		return nil
	}

	return fileInfo
}
