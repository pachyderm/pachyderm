package s3

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gogo/protobuf/types"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"
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
	    <CommonPrefixes>
	    	{{ range .dirs }}
	    		<Prefix>{{ .File.Path }}</Prefix>
	    	{{ end }}
	    </CommonPrefixes>
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

func (h handler) ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
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

// TODO: handle no head commits in a branch
func (h handler) getRepo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repo := vars["repo"]

	if err := r.ParseForm(); err != nil {
		writeBadRequest(w, err)
		return
	}

	repoInfo, err := h.pc.InspectRepo(repo)
	if err != nil {
		writeMaybeNotFound(w, r, err)
		return
	}

	if _, ok := r.Form["location"]; ok {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(locationResponse))
		return
	}

	delimiter := r.FormValue("delimiter")
	if delimiter != "" && delimiter != "/" {
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
			writeBadRequest(w, fmt.Errorf("max-keys value '%d' is too small", maxKeys))
			return
		}
		if maxKeys > defaultMaxKeys {
			writeBadRequest(w, fmt.Errorf("max-keys value '%d' is too large; it can only go up to %d", maxKeys, defaultMaxKeys))
			return
		}
	}

	fullPrefix := r.FormValue("prefix")
	fullPrefixParts := strings.SplitN(fullPrefix, "/", 2)
	var branches []string
	var pathPrefix string
	if len(fullPrefixParts) >= 1 && fullPrefixParts[0] != "" {
		branches = []string{fullPrefixParts[0]}
	}
	if len(fullPrefixParts) >= 2 {
		pathPrefix = fullPrefixParts[1]
	}
	if branches == nil {
		for _, branchInfo := range repoInfo.Branches {
			branches = append(branches, branchInfo.Name)
		}
	}
	sort.Strings(branches)

	var files []*pfs.FileInfo
	var dirs []*pfs.FileInfo
	isTruncated := false

	for _, branch := range branches {
		err = h.pc.Walk(repo, branch, pathPrefix, func(fileInfo *pfs.FileInfo) error {
			if fileInfo.FileType == pfs.FileType_FILE {
				// update the path to match s3
				fileInfo.File.Path = fmt.Sprintf("%s/%s", branch, fileInfo.File.Path)

				if fileInfo.File.Path <= marker {
					return nil
				}

				if len(files) == maxKeys {
					isTruncated = true
					return errutil.ErrBreak
				}

				files = append(files, fileInfo)
			} else if fileInfo.FileType == pfs.FileType_DIR {
				// skip the root directory
				if fileInfo.File.Path == "" {
					return nil
				}

				// update the path to match s3
				fileInfo.File.Path = fmt.Sprintf("%s/%s/", branch, fileInfo.File.Path)

				if fileInfo.File.Path <= marker {
					return nil
				}

				dirs = append(dirs, fileInfo)
			}

			return nil
		})

		if err != nil {
			writeMaybeNotFound(w, r, err)
			return
		}
	}

	args := map[string]interface{}{
		"bucket":      repo,
		"prefix":      fullPrefix,
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

func (h handler) putRepo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repo := vars["repo"]

	err := h.pc.CreateRepo(repo)

	if err != nil {
		writeServerError(w, err)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func (h handler) deleteRepo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repo := vars["repo"]

	err := h.pc.DeleteRepo(repo, false)

	if err != nil {
		writeMaybeNotFound(w, r, err)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h handler) getObject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repo := vars["repo"]
	branch := vars["branch"]
	file := vars["file"]

	fileInfo, err := h.pc.InspectFile(repo, branch, file)
	if err != nil {
		if strings.Contains(err.Error(), "has no head") {
			// occurs if the branch exists, but there's no head commit
			http.NotFound(w, r)
		} else {
			writeMaybeNotFound(w, r, err)
		}

		return
	}

	timestamp, err := types.TimestampFromProto(fileInfo.Committed)
	if err != nil {
		writeServerError(w, err)
		return
	}

	reader, err := h.pc.GetFileReadSeeker(repo, branch, file)
	if err != nil {
		writeServerError(w, err)
		return
	}

	http.ServeContent(w, r, "", timestamp, reader)
}

func (h handler) putObject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repo := vars["repo"]
	branch := vars["branch"]
	file := vars["file"]

	expectedHash := r.Header.Get("Content-MD5")

	if expectedHash != "" {
		expectedHashBytes, err := base64.StdEncoding.DecodeString(expectedHash)

		if err != nil {
			writeBadRequest(w, fmt.Errorf("could not decode `Content-MD5`, as it is not base64-encoded"))
			return
		}

		h.putObjectVerifying(w, r, repo, branch, file, expectedHashBytes)
		return
	}

	h.putObjectUnverified(w, r, repo, branch, file)
}

func (h handler) putObjectVerifying(w http.ResponseWriter, r *http.Request, repo, branch, file string, expectedHash []byte) {
	hasher := md5.New()
	reader := io.TeeReader(r.Body, hasher)

	_, err := h.pc.PutFileOverwrite(repo, branch, file, reader, 0)
	if err != nil {
		// the error may be because the repo or branch does not exist -
		// double-check that by inspecting the branch, so we can serve a 404
		// instead
		_, inspectError := h.pc.InspectBranch(repo, branch)
		writeMaybeNotFound(w, r, inspectError)
		return
	}

	actualHash := hasher.Sum(nil)

	if !bytes.Equal(expectedHash, actualHash) {
		err = fmt.Errorf("content checksums differ; expected=%x, actual=%x", expectedHash, actualHash)
		writeBadRequest(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h handler) putObjectUnverified(w http.ResponseWriter, r *http.Request, repo, branch, file string) {
	_, err := h.pc.PutFileOverwrite(repo, branch, file, r.Body, 0)
	if err != nil {
		// the error may be because the repo or branch does not exist -
		// double-check that by inspecting the branch, so we can serve a 404
		// instead
		_, inspectError := h.pc.InspectBranch(repo, branch)
		writeMaybeNotFound(w, r, inspectError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h handler) removeObject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repo := vars["repo"]
	branch := vars["branch"]
	file := vars["file"]

	if err := h.pc.DeleteFile(repo, branch, file); err != nil {
		writeMaybeNotFound(w, r, err)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
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
	router.HandleFunc(`/{repo:[a-z0-9][a-z0-9\.\-]{1,61}[a-z0-9]}/`, handler.getRepo).Methods("GET", "HEAD")
	router.HandleFunc(`/{repo:[a-z0-9][a-z0-9\.\-]{1,61}[a-z0-9]}/`, handler.putRepo).Methods("PUT")
	router.HandleFunc(`/{repo:[a-z0-9][a-z0-9\.\-]{1,61}[a-z0-9]}/`, handler.deleteRepo).Methods("DELETE")
	router.HandleFunc(`/{repo:[a-z0-9][a-z0-9\.\-]{1,61}[a-z0-9]}/{branch}/{file:.+}`, handler.getObject).Methods("GET", "HEAD")
	router.HandleFunc(`/{repo:[a-z0-9][a-z0-9\.\-]{1,61}[a-z0-9]}/{branch}/{file:.+}`, handler.putObject).Methods("PUT")
	router.HandleFunc(`/{repo:[a-z0-9][a-z0-9\.\-]{1,61}[a-z0-9]}/{branch}/{file:.+}`, handler.removeObject).Methods("DELETE")
	router.HandleFunc(`/_ping`, handler.ping).Methods("GET", "HEAD")

	// Note: error log is not customized on this `http.Server`, which means
	// it'll default to using the stdlib logger and produce log messages that
	// don't look like the ones produced elsewhere, and aren't configured
	// properly. In testing, this didn't seem to be a big deal because it's
	// rather hard to trigger it anyways, but if we find a reliable way to
	// create error logs, it might be worthwhile to fix.
	return &http.Server{
		Addr: fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Debugf("s3 gateway request: %s %s", r.Method, r.RequestURI)
			router.ServeHTTP(w, r)
		}),
	}
}
