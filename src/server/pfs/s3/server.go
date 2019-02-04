package s3

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/src/client"
)

const locationResponse = `<?xml version="1.0" encoding="UTF-8"?>
<LocationConstraint xmlns="http://s3.amazonaws.com/doc/2006-03-01/">PACHYDERM</LocationConstraint>`

type handler struct {
	pc *client.APIClient
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	parts := strings.SplitN(r.URL.Path, "/", 4)

	if len(parts) == 2 && parts[1] == "_ping" {
		// matches `/_ping`
		w.Write([]byte("OK"))
	} else if len(parts) == 3 && parts[2] == "" {
		// matches `/repo/`
		h.serveRoot(w, r, parts[1])
	} else if len(parts) == 4 {
		// matches /repo/branch/path/to/file.txt
		repo := parts[1]
		branch := parts[2]
		file := parts[3]
		h.serveFile(w, r, repo, branch, file)
	} else {
		http.Error(w, "not found", 404)
	}
}

func (h handler) serveRoot(w http.ResponseWriter, r *http.Request, repo string) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("%v", err), 400)
		return
	}

	if _, err := h.pc.InspectRepo(repo); err != nil {
		code := 500
		if strings.Contains(err.Error(), "not found") {
			code = 404
		}
		http.Error(w, fmt.Sprintf("%v", err), code)
		return
	}
	
	if _, ok := r.Form["location"]; ok {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(locationResponse))
	} else {
		w.Write([]byte("OK"))
	}
}

func (h handler) serveFile(w http.ResponseWriter, r *http.Request, repo, branch, file string) {
	fileInfo, err := h.pc.InspectFile(repo, branch, file)
	if err != nil {
		code := 500
		if strings.Contains(err.Error(), "not found") {
			// captures both missing branches and missing files
			code = 404
		}
		http.Error(w, fmt.Sprintf("%v", err), code)
		return
	}

	timestamp, err := types.TimestampFromProto(fileInfo.Committed)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), 500)
		return
	}

	reader, err := h.pc.GetFileReadSeeker(repo, branch, file)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), 500)
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
