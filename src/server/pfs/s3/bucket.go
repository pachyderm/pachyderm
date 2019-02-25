package s3

import (
	"fmt"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"
)

const defaultMaxKeys = 1000

const locationSource = `
<LocationConstraint xmlns="http://s3.amazonaws.com/doc/2006-03-01/">PACHYDERM</LocationConstraint>`

const listObjectsSource = `
<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
    <Name>{{ .bucket }}</Name>
    <Prefix>{{ .prefix }}</Prefix>
    <Marker>{{ .marker }}</Marker>
    <MaxKeys>{{ .maxKeys }}</MaxKeys>
    <IsTruncated>{{ .isTruncated }}</IsTruncated>
    {{ range .files }}
	    <Contents>
	        <Key>{{ .File.Path }}</Key>
	        <LastModified>{{ formatTime .Committed }}</LastModified>
	        <ETag></ETag>
	        <Size>{{ .SizeBytes }}</Size>
	        <StorageClass>STANDARD</StorageClass>
	        <Owner>
		    	<ID>000000000000000000000000000000</ID>
		    	<DisplayName>pachyderm</DisplayName>
	        </Owner>
	    </Contents>
    {{ end }}
    {{ if .dirs }}
    	{{ range .dirs }}
		    <CommonPrefixes>
		    	<Prefix>{{ . }}/</Prefix>
		    </CommonPrefixes>
	    {{ end }}
    {{ end }}
</ListBucketResult>`

const globSpecialCharacters = "*?[\\"

type bucketHandler struct {
	pc               *client.APIClient
	locationTemplate xmlTemplate
	listTemplate     xmlTemplate
}

func newBucketHandler(pc *client.APIClient) bucketHandler {
	return bucketHandler{
		pc:               pc,
		locationTemplate: newXmlTemplate(http.StatusOK, "location", locationSource),
		listTemplate:     newXmlTemplate(http.StatusOK, "list-objects", listObjectsSource),
	}
}

func (h bucketHandler) renderList(w http.ResponseWriter, bucket, prefix, marker string, maxKeys int, isTruncated bool, files []*pfs.FileInfo, dirs []string) {
	h.listTemplate.render(w, map[string]interface{}{
		"bucket":      bucket,
		"prefix":      prefix,
		"marker":      marker,
		"maxKeys":     maxKeys,
		"isTruncated": isTruncated,
		"files":       files,
		"dirs":        dirs,
	})
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
		h.locationTemplate.render(w, nil)
		return
	}

	marker := r.FormValue("marker")
	maxKeys := defaultMaxKeys
	maxKeysStr := r.FormValue("max-keys")
	if maxKeysStr != "" {
		maxKeys, err = strconv.Atoi(maxKeysStr)
		if err != nil {
			writeBadRequest(w, fmt.Errorf("invalid max-keys value '%s': %s", maxKeysStr, err))
			return
		}
		if maxKeys <= 0 {
			writeBadRequest(w, fmt.Errorf("max-keys value %d cannot be less than 1", maxKeys))
			return
		}
		if maxKeys > defaultMaxKeys {
			writeBadRequest(w, fmt.Errorf("max-keys value %d is too large; it can only go up to %d", maxKeys, defaultMaxKeys))
			return
		}
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

	if len(prefixParts) == 1 {
		branchPrefix := prefixParts[0]

		if recursive {
			h.listRecursive(w, r, repo, branchPrefix, prefix, marker, maxKeys)
		} else {
			h.list(w, r, repo, branchPrefix, prefix, marker, maxKeys)
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
			h.renderList(w, repo, prefix, marker, maxKeys, false, []*pfs.FileInfo{}, []string{})
		} else if recursive {
			h.listBranchRecursive(w, r, repo, branch, filePrefix, prefix, marker, maxKeys)
		} else {
			h.listBranch(w, r, repo, branch, filePrefix, prefix, marker, maxKeys)
		}
	}
}

func (h bucketHandler) listRecursive(w http.ResponseWriter, r *http.Request, repo, branchPrefix, completePrefix, marker string, maxKeys int) {
	var files []*pfs.FileInfo
	var dirs []string
	isTruncated := false
	branches, err := h.rootDirs(repo, branchPrefix)
	if err != nil {
		writeServerError(w, err)
		return
	}

	for _, branch := range branches {
		dirs = append(dirs, branch)
		if len(files)+len(dirs) == maxKeys {
			isTruncated = true
			break
		}

		err = h.pc.Walk(repo, branch, "", func(fileInfo *pfs.FileInfo) error {
			fileInfo = updateFileInfo(branch, marker, fileInfo)
			if fileInfo == nil {
				return nil
			}
			if len(files)+len(dirs) == maxKeys {
				isTruncated = true
				return errutil.ErrBreak
			}
			if fileInfo.FileType == pfs.FileType_FILE {
				files = append(files, fileInfo)
			} else {
				dirs = append(dirs, fileInfo.File.Path)
			}
			return nil
		})
		if err != nil {
			writeServerError(w, err)
			return
		}
		if isTruncated {
			break
		}
	}

	h.renderList(w, repo, completePrefix, marker, maxKeys, isTruncated, files, dirs)
}

func (h bucketHandler) list(w http.ResponseWriter, r *http.Request, repo, branchPrefix, completePrefix, marker string, maxKeys int) {
	var dirs []string
	isTruncated := false
	dirs, err := h.rootDirs(repo, branchPrefix)
	if err != nil {
		writeServerError(w, err)
		return
	}

	if len(dirs) > maxKeys {
		dirs = dirs[:maxKeys]
		isTruncated = true
	}

	h.renderList(w, repo, completePrefix, marker, maxKeys, isTruncated, []*pfs.FileInfo{}, dirs)
}

func (h bucketHandler) listBranchRecursive(w http.ResponseWriter, r *http.Request, repo string, branch string, filePrefix string, completePrefix string, marker string, maxKeys int) {
	var files []*pfs.FileInfo
	var dirs []string
	isTruncated := false

	// get what directory we should be recursing into
	dir := filepath.Dir(filePrefix)
	err := h.pc.Walk(repo, branch, dir, func(fileInfo *pfs.FileInfo) error {
		fileInfo = updateFileInfo(branch, marker, fileInfo)
		if fileInfo == nil {
			return nil
		}
		if !strings.HasPrefix(fileInfo.File.Path, completePrefix) {
			return nil
		}
		if len(files)+len(dirs) == maxKeys {
			isTruncated = true
			return errutil.ErrBreak
		}
		if fileInfo.FileType == pfs.FileType_FILE {
			files = append(files, fileInfo)
		} else {
			dirs = append(dirs, fileInfo.File.Path)
		}
		return nil
	})

	if err != nil {
		writeServerError(w, err)
		return
	}

	h.renderList(w, repo, completePrefix, marker, maxKeys, isTruncated, files, dirs)
}

func (h bucketHandler) listBranch(w http.ResponseWriter, r *http.Request, repo, branch, filePrefix, completePrefix, marker string, maxKeys int) {
	var files []*pfs.FileInfo
	var dirs []string
	isTruncated := false

	// ensure that we can globify the prefix string
	if !isGlobless(filePrefix) {
		writeBadRequest(w, fmt.Errorf("file prefix (everything after the first `/` in the prefix) cannot contain glob special characters"))
		return
	}

	fileInfos, err := h.pc.GlobFile(repo, branch, fmt.Sprintf("%s*", filePrefix))
	if err != nil {
		writeServerError(w, err)
		return
	}

	for _, fileInfo := range fileInfos {
		fileInfo = updateFileInfo(branch, marker, fileInfo)
		if fileInfo == nil {
			continue
		}
		if len(files)+len(dirs) == maxKeys {
			isTruncated = true
			break
		}
		if fileInfo.FileType == pfs.FileType_FILE {
			files = append(files, fileInfo)
		} else {
			dirs = append(dirs, fileInfo.File.Path)
		}
	}

	h.renderList(w, repo, completePrefix, marker, maxKeys, isTruncated, files, dirs)
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
