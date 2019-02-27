package s3

import (
	"fmt"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"
)

// this is a var instead of a const so that we can make a pointer to it
var defaultMaxKeys int = 1000

// the raw XML returned for a request to get the location of a bucket
const locationSource = `
<?xml version="1.0" encoding="UTF-8"?>
<LocationConstraint xmlns="http://s3.amazonaws.com/doc/2006-03-01/">PACHYDERM</LocationConstraint>
`

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

const globSpecialCharacters = "*?[\\"

type bucketHandler struct {
	pc *client.APIClient
}

func newBucketHandler(pc *client.APIClient) bucketHandler {
	return bucketHandler{pc: pc}
}

// rootDirs determines the root directories (common prefixes in s3 parlance)
// of a bucket by finding branches that have a given prefix
func (h bucketHandler) rootDirs(repo string, prefix string) ([]string, error) {
	dirs := []string{}
	branchInfos, err := h.pc.ListBranch(repo)
	if err != nil {
		return nil, err
	}

	// `branchInfos` is not sorted by default, but we need to sort in order to
	// match the s3 spec
	sort.Slice(branchInfos, func(i, j int) bool {
		return branchInfos[i].Branch.Name < branchInfos[j].Branch.Name
	})

	for _, branchInfo := range branchInfos {
		if !strings.HasPrefix(branchInfo.Branch.Name, prefix) {
			continue
		}
		if branchInfo.Head == nil {
			continue
		}
		dirs = append(dirs, branchInfo.Branch.Name)
	}

	return dirs, nil
}

func (h bucketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repo := vars["repo"]

	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		h.get(w, r, repo)
	} else if r.Method == http.MethodPut {
		h.put(w, r, repo)
	} else if r.Method == http.MethodDelete {
		h.delete(w, r, repo)
	} else {
		// method filtering on the mux router should prevent this
		panic("unreachable")
	}
}

func (h bucketHandler) get(w http.ResponseWriter, r *http.Request, repo string) {
	_, err := h.pc.InspectRepo(repo)
	if err != nil {
		writeMaybeNotFound(w, r, err)
		return
	}

	if err := r.ParseForm(); err != nil {
		writeBadRequest(w, err)
		return
	}

	if _, ok := r.Form["location"]; ok {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(locationSource))
		return
	}

	marker := r.FormValue("marker")
	maxKeys, err := intFormValue(r, "max-keys", 1, defaultMaxKeys, &defaultMaxKeys)
	if err != nil {
		writeBadRequest(w, err)
		return
	}

	recursive := false
	delimiter := r.FormValue("delimiter")
	if delimiter == "" {
		recursive = true
	} else if delimiter != "/" {
		writeBadRequest(w, fmt.Errorf("invalid delimiter '%s'; only '/' is allowed", delimiter))
		return
	}

	prefix := r.FormValue("prefix")
	prefixParts := strings.SplitN(prefix, "/", 2)

	result := &ListBucketResult{
		Name:        repo,
		Prefix:      prefix,
		Marker:      marker,
		MaxKeys:     maxKeys,
		IsTruncated: false,
	}

	if len(prefixParts) == 1 {
		branchPrefix := prefixParts[0]

		if recursive {
			h.listRecursive(w, r, result, branchPrefix)
		} else {
			h.list(w, r, result, branchPrefix)
		}
	} else {
		branch := prefixParts[0]
		filePrefix := prefixParts[1]
		isEmpty := false

		// ensure the branch exists and has a head
		branchInfo, err := h.pc.InspectBranch(repo, branch)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				isEmpty = true
			} else {
				writeServerError(w, err)
				return
			}
		} else if branchInfo.Head == nil {
			isEmpty = true
		}

		if isEmpty {
			// render an empty list if this branch doesn't exist or doesn't have a
			// head
			writeXML(w, http.StatusOK, &result)
		} else if recursive {
			h.listBranchRecursive(w, r, result, branch, filePrefix)
		} else {
			h.listBranch(w, r, result, branch, filePrefix)
		}
	}
}

func (h bucketHandler) listRecursive(w http.ResponseWriter, r *http.Request, result *ListBucketResult, branchPrefix string) {
	branches, err := h.rootDirs(result.Name, branchPrefix)
	if err != nil {
		writeServerError(w, err)
		return
	}

	for _, branch := range branches {
		result.CommonPrefixes = append(result.CommonPrefixes, newCommonPrefixes(branch))

		if result.isFull() {
			result.IsTruncated = true
			break
		}

		err = h.pc.Walk(result.Name, branch, "", func(fileInfo *pfs.FileInfo) error {
			fileInfo = updateFileInfo(branch, result.Marker, fileInfo)
			if fileInfo == nil {
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
			writeServerError(w, err)
			return
		}
		if result.IsTruncated {
			break
		}
	}

	writeXML(w, http.StatusOK, result)
}

func (h bucketHandler) list(w http.ResponseWriter, r *http.Request, result *ListBucketResult, branchPrefix string) {
	dirs, err := h.rootDirs(result.Name, branchPrefix)
	if err != nil {
		writeServerError(w, err)
		return
	}

	if len(dirs) > result.MaxKeys {
		dirs = dirs[:result.MaxKeys]
		result.IsTruncated = true
	}

	for _, dir := range dirs {
		result.CommonPrefixes = append(result.CommonPrefixes, newCommonPrefixes(dir))
	}

	writeXML(w, http.StatusOK, result)
}

func (h bucketHandler) listBranchRecursive(w http.ResponseWriter, r *http.Request, result *ListBucketResult, branch string, filePrefix string) {
	err := h.pc.Walk(result.Name, branch, filepath.Dir(filePrefix), func(fileInfo *pfs.FileInfo) error {
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
		writeServerError(w, err)
		return
	}

	writeXML(w, http.StatusOK, result)
}

func (h bucketHandler) listBranch(w http.ResponseWriter, r *http.Request, result *ListBucketResult, branch string, filePrefix string) {
	// ensure that we can globify the prefix string
	if !isGlobless(filePrefix) {
		writeBadRequest(w, fmt.Errorf("file prefix (everything after the first `/` in the prefix) cannot contain glob special characters"))
		return
	}

	fileInfos, err := h.pc.GlobFile(result.Name, branch, fmt.Sprintf("%s*", filePrefix))
	if err != nil {
		writeServerError(w, err)
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
				writeServerError(w, err)
				return
			}
			result.Contents = append(result.Contents, contents)
		} else {
			result.CommonPrefixes = append(result.CommonPrefixes, newCommonPrefixes(fileInfo.File.Path))
		}
	}

	writeXML(w, http.StatusOK, result)
}

func (h bucketHandler) put(w http.ResponseWriter, r *http.Request, repo string) {
	err := h.pc.CreateRepo(repo)
	if err != nil {
		if strings.Contains(err.Error(), "as it already exists") {
			writeBadRequest(w, err)
		} else {
			writeServerError(w, err)
		}

		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h bucketHandler) delete(w http.ResponseWriter, r *http.Request, repo string) {
	err := h.pc.DeleteRepo(repo, false)

	if err != nil {
		writeMaybeNotFound(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// updateFileInfo takes in a `FileInfo`, and updates it to be used in s3
// object listings:
// 1) if nil is returned, the `FileInfo` should not be included in the list
// 2) the path is updated to be prefixed by the branch name
func updateFileInfo(branch, marker string, fileInfo *pfs.FileInfo) *pfs.FileInfo {
	if fileInfo.FileType != pfs.FileType_FILE && fileInfo.FileType != pfs.FileType_DIR {
		// skip anything that isn't a file or dir
		return nil
	}
	if fileInfo.FileType == pfs.FileType_DIR && fileInfo.File.Path == "/" {
		// skip the root directory
		return nil
	}

	// update the path to match s3
	fileInfo.File.Path = fmt.Sprintf("%s%s", branch, fileInfo.File.Path)

	if fileInfo.File.Path <= marker {
		// skip file paths below the marker
		return nil
	}

	return fileInfo
}

// isGlobless checks whether any a string contains any characters used in
// glob file path matching
func isGlobless(s string) bool {
	return !strings.ContainsAny(s, globSpecialCharacters)
}
