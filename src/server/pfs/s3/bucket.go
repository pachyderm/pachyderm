package s3

import (
	"fmt"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/gorilla/mux"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"
)

const defaultMaxKeys = 1000

const locationSource = `<?xml version="1.0" encoding="UTF-8"?>
<LocationConstraint xmlns="http://s3.amazonaws.com/doc/2006-03-01/">PACHYDERM</LocationConstraint>`

const listObjectsSource = `<?xml version="1.0" encoding="UTF-8"?>
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

type bucketHandler struct {
	pc           *client.APIClient
	listTemplate *template.Template
}

// TODO: support xml escaping
func newBucketHandler(pc *client.APIClient) bucketHandler {
	funcMap := template.FuncMap{
		"formatTime": formatTime,
	}

	listTemplate := template.Must(template.New("list-objects").
		Funcs(funcMap).
		Parse(listObjectsSource))

	return bucketHandler{
		pc:           pc,
		listTemplate: listTemplate,
	}
}

func (h bucketHandler) renderList(w http.ResponseWriter, bucket, prefix, marker string, maxKeys int, isTruncated bool, files []*pfs.FileInfo, dirs []string) {
	args := map[string]interface{}{
		"bucket":      bucket,
		"prefix":      prefix,
		"marker":      marker,
		"maxKeys":     maxKeys,
		"isTruncated": isTruncated,
		"files":       files,
		"dirs":        dirs,
	}
	if err := h.listTemplate.Execute(w, args); err != nil {
		writeServerError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
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
		h.getRepo(w, r, repo)
	} else if r.Method == http.MethodPut {
		h.putRepo(w, r, repo)
	} else if r.Method == http.MethodDelete {
		h.deleteRepo(w, r, repo)
	} else {
		// method filtering on the mux router should prevent this
		panic("unreachable")
	}
}

func (h bucketHandler) getRepo(w http.ResponseWriter, r *http.Request, repo string) {
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
		w.Write([]byte(locationSource))
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

	prefix := r.FormValue("prefix")
	prefixParts := strings.SplitN(prefix, "/", 2)
	delimiter := r.FormValue("delimiter")
	if delimiter == "" {
		// a delimiter was not specified; recurse into subdirectories
		if len(prefixParts) == 1 {
			h.listRepoRecursive(w, r, repo, prefixParts[0], prefix, marker, maxKeys)
		} else {
			h.listFilesRecursive(w, r, repo, prefixParts[0], prefixParts[1], prefix, marker, maxKeys)
		}
	} else if delimiter == "/" {
		// a delimiter was specified; do not recurse into subdirectories
		if len(prefixParts) == 1 {
			h.listRepo(w, r, repo, prefixParts[0], prefix, marker, maxKeys)
		} else {
			h.listFiles(w, r, repo, prefixParts[0], prefixParts[1], prefix, marker, maxKeys)
		}
	} else {
		writeBadRequest(w, fmt.Errorf("invalid delimiter '%s'; only '/' is allowed", delimiter))
	}
}

func (h bucketHandler) listRepo(w http.ResponseWriter, r *http.Request, repo, branchPrefix, completePrefix, marker string, maxKeys int) {
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

func (h bucketHandler) listFiles(w http.ResponseWriter, r *http.Request, repo, branch, filePrefix, completePrefix, marker string, maxKeys int) {
	var files []*pfs.FileInfo
	var dirs []string

	// ensure the branch exists and has a head
	branchInfo, err := h.pc.InspectBranch(repo, branch)
	if err != nil {
		writeMaybeNotFound(w, r, err)
		return
	}
	if branchInfo.Head == nil {
		http.NotFound(w, r)
		return
	}

	isTruncated := false
	fileInfos, err := h.pc.GlobFile(repo, branch, fmt.Sprintf("%s*", filePrefix)) //TODO: escape filePrefix
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

func (h bucketHandler) listRepoRecursive(w http.ResponseWriter, r *http.Request, repo, branchPrefix, completePrefix, marker string, maxKeys int) {
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
			writeMaybeNotFound(w, r, err)
			return
		}
		if isTruncated {
			break
		}
	}
	
	h.renderList(w, repo, completePrefix, marker, maxKeys, isTruncated, files, dirs)
}

func (h bucketHandler) listFilesRecursive(w http.ResponseWriter, r *http.Request, repo string, branch string, filePrefix string, completePrefix string, marker string, maxKeys int) {
	var files []*pfs.FileInfo
	var dirs []string
	isTruncated := false

	// ensure the branch exists and has a head
	branchInfo, err := h.pc.InspectBranch(repo, branch)
	if err != nil {
		writeMaybeNotFound(w, r, err)
		return
	}
	if branchInfo.Head == nil {
		http.NotFound(w, r)
		return
	}

	// get what directory we should be recursing into
	dir := filepath.Dir(filePrefix)
	if dir == "." {
		dir = ""
	}

	err = h.pc.Walk(repo, branch, dir, func(fileInfo *pfs.FileInfo) error {
		fileInfo = updateFileInfo(branch, marker, fileInfo)
		if fileInfo == nil {
			return nil
		}
		if !strings.HasPrefix(fileInfo.File.Path, filePrefix) {
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
		writeMaybeNotFound(w, r, err)
		return
	}
	
	h.renderList(w, repo, completePrefix, marker, maxKeys, isTruncated, files, dirs)
}

func (h bucketHandler) putRepo(w http.ResponseWriter, r *http.Request, repo string) {
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

func (h bucketHandler) deleteRepo(w http.ResponseWriter, r *http.Request, repo string) {
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
