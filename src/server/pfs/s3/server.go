package s3

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/gogo/protobuf/types"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
)

const defaultMaxKeys = 1000

const locationResponse = `<?xml version="1.0" encoding="UTF-8"?>
<LocationConstraint xmlns="http://s3.amazonaws.com/doc/2006-03-01/">PACHYDERM</LocationConstraint>`

const listBucketsSource = `<?xml version="1.0" encoding="UTF-8"?>
<ListAllMyBucketsResult xmlns="http://s3.amazonaws.com/doc/2006-03-01">
    <Owner>
    	<ID>000000000000000000000000000000</ID>
    	<DisplayName>pachyderm</DisplayName>
    </Owner>
    <Buckets>
        {{ range . }}
            <Bucket>
                <Name>{{ .Repo.Name }}</Name>
                <CreationDate>{{ formatTime .Created }}</CreationDate>
            </Bucket>
        {{ end }}
    </Buckets>
</ListAllMyBucketsResult>`

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

func writeBadRequest(w http.ResponseWriter, err error) {
	http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
}

func writeMaybeNotFound(w http.ResponseWriter, r *http.Request, err error) {
	if strings.Contains(err.Error(), "not found") {
		http.NotFound(w, r)
	} else {
		writeServerError(w, err)
	}
}

func writeServerError(w http.ResponseWriter, err error) {
	http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
}

type handler struct {
	pc                  *client.APIClient
	listBucketsTemplate *template.Template
	listObjectsTemplate *template.Template
}

// TODO: support xml escaping
func newHandler(pc *client.APIClient) handler {
	funcMap := template.FuncMap{
		"formatTime": func(timestamp *types.Timestamp) string {
			return timestamp.String()
		},
	}

	listBucketsTemplate := template.Must(template.New("list-buckets").
		Funcs(funcMap).
		Parse(listBucketsSource))

	listObjectsTemplate := template.Must(template.New("list-objects").
		Funcs(funcMap).
		Parse(listObjectsSource))

	return handler{
		pc:                  pc,
		listBucketsTemplate: listBucketsTemplate,
		listObjectsTemplate: listObjectsTemplate,
	}
}

func (h handler) root(w http.ResponseWriter, r *http.Request) {
	buckets, err := h.pc.ListRepo()
	if err != nil {
		writeServerError(w, err)
		return
	}

	if err = h.listBucketsTemplate.Execute(w, buckets); err != nil {
		writeServerError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
}

func (h handler) repo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repo := vars["repo"]

	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		h.getRepo(w, r, repo)
	} else if r.Method == http.MethodPut {
		h.putRepo(w, r, repo)
	} else if r.Method == http.MethodDelete {
		h.deleteRepo(w, r, repo)
	} else {
		panic("unreachable")
	}
}

func (h handler) getRepo(w http.ResponseWriter, r *http.Request, repo string) {
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
		w.Write([]byte(locationResponse))
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

func (h handler) getBranches(w http.ResponseWriter, r *http.Request, repoInfo *pfs.RepoInfo, pattern string, prefix string, marker string, maxKeys int) {
	var dirs []string
	isTruncated := false

	// TODO: remove branches that don't have a head
	for _, branchInfo := range repoInfo.Branches {
		match, err := filepath.Match(pattern, branchInfo.Name)

		if err != nil {
			writeBadRequest(w, fmt.Errorf("invalid prefix '%s' (compiled to pattern '%s'): %v", prefix, pattern, err))
			return
		}

		if !match {
			continue
		}
		if len(dirs) == maxKeys {
			isTruncated = true
			break
		}

		dirs = append(dirs, branchInfo.Name)
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
	if err := h.listObjectsTemplate.Execute(w, args); err != nil {
		writeServerError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
}

func (h handler) getFiles(w http.ResponseWriter, r *http.Request, repo string, branch string, pattern string, prefix string, marker string, maxKeys int) {
	var files []*pfs.FileInfo
	var dirs []string
	isTruncated := false
	fileInfos, err := h.pc.GlobFile(repo, branch, pattern)

	// TODO: handle branches w/o heads
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
	if err = h.listObjectsTemplate.Execute(w, args); err != nil {
		writeServerError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
}

func (h handler) putRepo(w http.ResponseWriter, r *http.Request, repo string) {
	err := h.pc.CreateRepo(repo)

	if err != nil {
		writeServerError(w, err)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func (h handler) deleteRepo(w http.ResponseWriter, r *http.Request, repo string) {
	err := h.pc.DeleteRepo(repo, false)

	if err != nil {
		writeMaybeNotFound(w, r, err)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h handler) object(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repo := vars["repo"]
	branch := vars["branch"]
	file := vars["file"]

	branchInfo, err := h.pc.InspectBranch(repo, branch)
	if err != nil {
		writeMaybeNotFound(w, r, err)
		return
	}

	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		h.getObject(w, r, branchInfo, file)
	} else if r.Method == http.MethodPut {
		h.putObject(w, r, branchInfo, file)
	} else if r.Method == http.MethodDelete {
		h.deleteObject(w, r, branchInfo, file)
	} else {
		panic("unreachable")
	}
}

func (h handler) getObject(w http.ResponseWriter, r *http.Request, branchInfo *pfs.BranchInfo, file string) {
	if branchInfo.Head == nil {
		http.NotFound(w, r)
		return
	}

	fileInfo, err := h.pc.InspectFile(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, file)
	if err != nil {
		writeMaybeNotFound(w, r, err)
		return
	}

	timestamp, err := types.TimestampFromProto(fileInfo.Committed)
	if err != nil {
		writeServerError(w, err)
		return
	}

	reader, err := h.pc.GetFileReadSeeker(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, file)
	if err != nil {
		writeServerError(w, err)
		return
	}

	http.ServeContent(w, r, "", timestamp, reader)
}

func (h handler) putObject(w http.ResponseWriter, r *http.Request, branchInfo *pfs.BranchInfo, file string) {
	expectedHash := r.Header.Get("Content-MD5")

	if expectedHash != "" {
		expectedHashBytes, err := base64.StdEncoding.DecodeString(expectedHash)
		if err != nil {
			writeBadRequest(w, fmt.Errorf("could not decode `Content-MD5`, as it is not base64-encoded"))
			return
		}

		hasher := md5.New()
		reader := io.TeeReader(r.Body, hasher)

		_, err = h.pc.PutFileOverwrite(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, file, reader, 0)
		if err != nil {
			writeServerError(w, err)
			return
		}

		actualHash := hasher.Sum(nil)
		if !bytes.Equal(expectedHashBytes, actualHash) {
			err = fmt.Errorf("content checksums differ; expected=%x, actual=%x", expectedHash, actualHash)
			writeBadRequest(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	_, err := h.pc.PutFileOverwrite(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, file, r.Body, 0)
	if err != nil {
		writeServerError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h handler) deleteObject(w http.ResponseWriter, r *http.Request, branchInfo *pfs.BranchInfo, file string) {
	if branchInfo.Head == nil {
		http.NotFound(w, r)
		return
	}

	if err := h.pc.DeleteFile(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, file); err != nil {
		writeMaybeNotFound(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Server runs an HTTP server with an S3-like API for PFS. This allows you to
// use s3 clients to acccess PFS contents.
//
// This returns an `http.Server` instance. It is the responsibility of the
// caller to start the server. This also makes it possible for the caller to
// enable graceful shutdown if desired; see the `http` package for details.
//
// Bucket names correspond to repo names, and files are accessible via the s3
// key pattern "<branch>/<filepath>". For example, to get the file "a/b/c.txt"
// on the "foo" repo's "master" branch, you'd making an s3 get request with
// bucket = "foo", key = "master/a/b/c.txt".
//
// Note: in s3, bucket names are constrained by IETF RFC 1123, (and its
// predecessor RFC 952) but pachyderm's repo naming constraints are slightly
// more liberal. If the s3 client does any kind of bucket name validation
// (this includes minio), repos whose names do not comply with RFC 1123 will
// not be accessible.
//
// Note: In `s3cmd`, you must set the access key and secret key, even though
// this API will ignore them - otherwise, you'll get an opaque config error:
// https://github.com/s3tools/s3cmd/issues/845#issuecomment-464885959
func Server(pc *client.APIClient, port uint16) *http.Server {
	handler := newHandler(pc)

	// repo validation regex is the same as minio
	router := mux.NewRouter()
	router.HandleFunc(`/`, handler.root).Methods("GET", "HEAD")
	router.HandleFunc(`/{repo:[a-z0-9][a-z0-9\.\-]{1,61}[a-z0-9]}/`, handler.repo).Methods("GET", "HEAD", "PUT", "DELETE")
	router.HandleFunc(`/{repo:[a-z0-9][a-z0-9\.\-]{1,61}[a-z0-9]}/{branch}/{file:.+}`, handler.object).Methods("GET", "HEAD", "PUT", "DELETE")

	// Note: error log is not customized on this `http.Server`, which means
	// it'll default to using the stdlib logger and produce log messages that
	// don't look like the ones produced elsewhere, and aren't configured
	// properly. In testing, this didn't seem to be a big deal because it's
	// rather hard to trigger it anyways, but if we find a reliable way to
	// create error logs, it might be worthwhile to fix.
	return &http.Server{
		Addr: fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: reduce log level
			log.Infof("s3 gateway request: %s %s", r.Method, r.RequestURI)
			router.ServeHTTP(w, r)
		}),
	}
}
