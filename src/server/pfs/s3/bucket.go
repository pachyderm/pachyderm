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
		pc:                  pc,
		listTemplate: listTemplate,
	}
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
	repoInfo, err := h.pc.InspectRepo(repo)
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

	delimiter := r.FormValue("delimiter")
	if delimiter == "" {
		// Just return OK so the callee knows that the bucket exists. This is
		// to support `BucketExists`, which is a `HEAD` request on `/<repo>/`.
		// TODO: should this be implemented more thoroughly?
		w.WriteHeader(http.StatusOK)
		return
	}
	if delimiter != "/" {
		writeBadRequest(w, fmt.Errorf("invalid delimiter '%s'; only '/' is allowed", delimiter))
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

	if len(prefixParts) == 1 {
		branchPattern := fmt.Sprintf("%s*", prefixParts[0])
		h.getBranches(w, r, repoInfo, branchPattern, prefix, marker, maxKeys)
	} else {
		branch := prefixParts[0]
		filePattern := fmt.Sprintf("%s*", prefixParts[1])
		h.getFiles(w, r, repo, branch, filePattern, prefix, marker, maxKeys)
	}
}

func (h bucketHandler) getBranches(w http.ResponseWriter, r *http.Request, repoInfo *pfs.RepoInfo, pattern string, prefix string, marker string, maxKeys int) {
	var dirs []string
	isTruncated := false

	// While `repoInfo` has the list of branch names, it doesn't include head
	// commit info. We need this to remove branches without a head commit.
	branchInfos, err := h.pc.ListBranch(repoInfo.Repo.Name)
	if err != nil {
		writeServerError(w, err)
		return
	}

	// `branchInfos` is not sorted by default, but we need to sort in order to
	// match the s3 spec
	sort.Slice(branchInfos, func(i, j int) bool {
		return branchInfos[i].Branch.Name < branchInfos[j].Branch.Name
	})

	for _, branchInfo := range branchInfos {
		match, err := filepath.Match(pattern, branchInfo.Branch.Name)
		if err != nil {
			writeBadRequest(w, fmt.Errorf("invalid prefix '%s' (compiled to pattern '%s'): %v", prefix, pattern, err))
			return
		}
		if !match {
			continue
		}

		if branchInfo.Head == nil {
			continue
		}
		if len(dirs) == maxKeys {
			isTruncated = true
			break
		}

		dirs = append(dirs, branchInfo.Branch.Name)
	}

	args := map[string]interface{}{
		"bucket":      repoInfo.Repo.Name,
		"prefix":      prefix,
		"marker":      marker,
		"maxKeys":     maxKeys,
		"isTruncated": isTruncated,
		"files":       []*pfs.FileInfo{},
		"dirs":        dirs,
	}
	if err := h.listTemplate.Execute(w, args); err != nil {
		writeServerError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
}

func (h bucketHandler) getFiles(w http.ResponseWriter, r *http.Request, repo string, branch string, pattern string, prefix string, marker string, maxKeys int) {
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
	fileInfos, err := h.pc.GlobFile(repo, branch, pattern)
	if err != nil {
		writeServerError(w, err)
		return
	}

	for _, fileInfo := range fileInfos {
		isFile := fileInfo.FileType == pfs.FileType_FILE
		isDir := fileInfo.FileType == pfs.FileType_DIR
		if !isFile && !isDir {
			continue
		}
		if len(files)+len(dirs) == maxKeys {
			isTruncated = true
			break
		}
		if isDir && fileInfo.File.Path == "" {
			// skip the root directory
			continue
		}

		// update the path to match s3
		fileInfo.File.Path = fmt.Sprintf("%s%s", branch, fileInfo.File.Path)
		if fileInfo.File.Path <= marker {
			continue
		}

		if isFile {
			files = append(files, fileInfo)
		} else {
			dirs = append(dirs, fileInfo.File.Path)
		}
	}

	args := map[string]interface{}{
		"bucket":      repo,
		"prefix":      prefix,
		"marker":      marker,
		"maxKeys":     maxKeys,
		"isTruncated": isTruncated,
		"files":       files,
		"dirs":        dirs,
	}
	if err = h.listTemplate.Execute(w, args); err != nil {
		writeServerError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
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
