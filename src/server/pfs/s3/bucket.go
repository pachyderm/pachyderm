package s3

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

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
	repo, branch := bucketArgs(r)
	branchInfo, err := h.pc.InspectBranch(repo, branch)
	if err != nil {
		newNotFoundError(r, err).write(w)
		return
	}
	if branchInfo.Head == nil {
		newNoSuchBucketError(r).write(w)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(locationSource))
}

func (h bucketHandler) get(w http.ResponseWriter, r *http.Request) {
	repo, branch := bucketArgs(r)

	// ensure the branch exists and has a head
	branchInfo, err := h.pc.InspectBranch(repo, branch)
	if err != nil {
		newNotFoundError(r, err).write(w)
		return
	}
	if branchInfo.Head == nil {
		newNoSuchBucketError(r).write(w)
		return
	}

	result := &ListBucketResult{
		Name:        repo,
		Prefix:      r.FormValue("prefix"),
		Marker:      r.FormValue("marker"),
		MaxKeys:     intFormValue(r, "max-keys", 1, defaultMaxKeys, defaultMaxKeys),
		IsTruncated: false,
	}

	delimiter := r.FormValue("delimiter")
	if delimiter != "" && delimiter != "/" {
		newInvalidDelimiterError(r).write(w)
		return
	}

	if delimiter == "" {
		h.listRecursive(w, r, result, branch)
	} else {
		h.list(w, r, result, branch)
	}
}

func (h bucketHandler) listRecursive(w http.ResponseWriter, r *http.Request, result *ListBucketResult, branch string) {
	err := h.pc.Walk(result.Name, branch, filepath.Dir(result.Prefix), func(fileInfo *pfs.FileInfo) error {
		if !shouldShowFileInfo(branch, result.Marker, fileInfo) {
			return nil
		}
		fileInfo.File.Path = fileInfo.File.Path[1:] // strip leading slash
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
		newInternalError(r, err).write(w)
		return
	}

	writeXML(w, http.StatusOK, result)
}

func (h bucketHandler) list(w http.ResponseWriter, r *http.Request, result *ListBucketResult, branch string) {
	pattern := fmt.Sprintf("%s*", glob.QuoteMeta(result.Prefix))
	fileInfos, err := h.pc.GlobFile(result.Name, branch, pattern)
	if err != nil {
		newInternalError(r, err).write(w)
		return
	}

	for _, fileInfo := range fileInfos {
		if !shouldShowFileInfo(branch, result.Marker, fileInfo) {
			continue
		}
		fileInfo.File.Path = fileInfo.File.Path[1:] // strip leading slash
		if result.isFull() {
			result.IsTruncated = true
			break
		}
		if fileInfo.FileType == pfs.FileType_FILE {
			contents, err := newContents(fileInfo)
			if err != nil {
				newInternalError(r, err).write(w)
				return
			}
			result.Contents = append(result.Contents, contents)
		} else {
			result.CommonPrefixes = append(result.CommonPrefixes, newCommonPrefixes(fileInfo.File.Path))
		}
	}

	writeXML(w, http.StatusOK, result)
}

// TODO: handling of branches on put/del bucket
func (h bucketHandler) put(w http.ResponseWriter, r *http.Request) {
	repo, _ := bucketArgs(r)

	err := h.pc.CreateRepo(repo)
	if err != nil {
		if strings.Contains(err.Error(), "as it already exists") {
			newBucketAlreadyExistsError(r).write(w)
		} else {
			newInternalError(r, err).write(w)
		}

		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h bucketHandler) del(w http.ResponseWriter, r *http.Request) {
	repo, _ := bucketArgs(r)

	err := h.pc.DeleteRepo(repo, false)

	if err != nil {
		newNotFoundError(r, err).write(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func shouldShowFileInfo(branch, marker string, fileInfo *pfs.FileInfo) bool {
	if fileInfo.FileType != pfs.FileType_FILE && fileInfo.FileType != pfs.FileType_DIR {
		// skip anything that isn't a file or dir
		return false
	}
	if fileInfo.FileType == pfs.FileType_DIR && fileInfo.File.Path == "/" {
		// skip the root directory
		return false
	}
	if fileInfo.File.Path <= marker {
		// skip file paths below the marker
		return false
	}

	return true
}
