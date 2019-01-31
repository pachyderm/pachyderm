package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/src/client"
)

const locationResponse = `<?xml version="1.0" encoding="UTF-8"?>
<LocationConstraint xmlns="http://s3.amazonaws.com/doc/2006-03-01/">PACHYDERM</LocationConstraint>`

func writeOK(w http.ResponseWriter) {
	w.Write([]byte("OK"))
}

func writeMethodNotAllowed(w http.ResponseWriter) {
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func writeNotFound(w http.ResponseWriter) {
	http.Error(w, "not found", http.StatusNotFound)
}

func writeServerError(w http.ResponseWriter, err error) {
	http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
}

type handler struct {
	pc *client.APIClient
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	parts := strings.SplitN(r.URL.Path, "/", 4)

	if len(parts) == 2 && parts[1] == "_ping" {
		// matches `/_ping`
		if r.Method == http.MethodGet {
			writeOK(w)
		} else {
			writeMethodNotAllowed(w)
		}
	} else if len(parts) == 3 && parts[2] == "" {
		// matches `/repo/`
		if r.Method == http.MethodGet {
			h.serveRoot(w, r, parts[1])
		} else {
			writeMethodNotAllowed(w)
		}
	} else if len(parts) == 4 {
		// matches /repo/branch/path/to/file.txt
		repo := parts[1]
		branch := parts[2]
		file := parts[3]
		if r.Method == http.MethodGet {
			h.serveFile(w, r, repo, branch, file)
		} else {
			writeMethodNotAllowed(w)
		}
	} else {
		writeNotFound(w)
	}
}

func (h handler) serveRoot(w http.ResponseWriter, r *http.Request, repo string) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
		return
	}

	if _, err := h.pc.InspectRepo(repo); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeNotFound(w)
		} else {
			writeServerError(w, err)
		}
		return
	}
	
	if _, ok := r.Form["location"]; ok {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(locationResponse))
	} else {
		writeOK(w)
	}
}

func (h handler) serveFile(w http.ResponseWriter, r *http.Request, repo, branch, file string) {
	fileInfo, err := h.pc.InspectFile(repo, branch, file)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeNotFound(w)
		} else {
			writeServerError(w, err)
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

// Server runs an HTTP server with an S3-like API for PFS. This allows you to
// use s3 clients to acccess PFS contents.
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
	return &http.Server {
		Addr: fmt.Sprintf(":%d", port),
		Handler: handler{ pc: pc },
	}
}
