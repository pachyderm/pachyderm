package s3

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/src/client"
)

const locationResponse = `<?xml version="1.0" encoding="UTF-8"?>
<LocationConstraint xmlns="http://s3.amazonaws.com/doc/2006-03-01/">PACHYDERM</LocationConstraint>`

func writeOK(w http.ResponseWriter) {
	w.Write([]byte("OK"))
}

func writeBadRequest(w http.ResponseWriter, err error) {
	http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
}

func writeMaybeNotFound(w http.ResponseWriter, err error) {
	if strings.Contains(err.Error(), "not found") {
		// This error message matches what the mux router returns when it 404s
		http.Error(w, "404 page not found", http.StatusNotFound)
	} else {
		writeServerError(w, err)
	}
}

func writeServerError(w http.ResponseWriter, err error) {
	http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
}

type handler struct {
	pc *client.APIClient
}

func (h handler) ping(w http.ResponseWriter, r *http.Request) {
	writeOK(w)
}

func (h handler) repo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repo := vars["repo"]

	if err := r.ParseForm(); err != nil {
		writeBadRequest(w, err)
		return
	}

	if _, err := h.pc.InspectRepo(repo); err != nil {
		writeMaybeNotFound(w, err)
		return
	}

	if _, ok := r.Form["location"]; ok {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(locationResponse))
	} else {
		writeOK(w)
	}
}

func (h handler) getObject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repo := vars["repo"]
	branch := vars["branch"]
	file := vars["file"]

	fileInfo, err := h.pc.InspectFile(repo, branch, file)
	if err != nil {
		writeMaybeNotFound(w, err)
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
		writeMaybeNotFound(w, inspectError)
		return
	}

	actualHash := hasher.Sum(nil)

	if !bytes.Equal(expectedHash, actualHash) {
		err = fmt.Errorf("content checksums differ; expected=%x, actual=%x", expectedHash, actualHash)
		writeServerError(w, err)
		return
	}

	writeOK(w)
}

func (h handler) putObjectUnverified(w http.ResponseWriter, r *http.Request, repo, branch, file string) {
	_, err := h.pc.PutFileOverwrite(repo, branch, file, r.Body, 0)
	if err != nil {
		// the error may be because the repo or branch does not exist -
		// double-check that by inspecting the branch, so we can serve a 404
		// instead
		_, inspectError := h.pc.InspectBranch(repo, branch)
		writeMaybeNotFound(w, inspectError)
		return
	}

	writeOK(w)
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
func Server(pc *client.APIClient, port uint16) *http.Server {
	handler := handler{pc: pc}
	router := mux.NewRouter()
	router.HandleFunc("/_ping", handler.ping).Methods("GET")
	router.HandleFunc("/{repo}/", handler.repo).Methods("GET")
	router.HandleFunc("/{repo}/{branch}/{file:.+}", handler.getObject).Methods("GET")
	router.HandleFunc("/{repo}/{branch}/{file:.+}", handler.putObject).Methods("PUT")

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}
}
