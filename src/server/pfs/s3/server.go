package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/src/client"
)

const LOCATION_RESPONSE = `<?xml version="1.0" encoding="UTF-8"?>
<LocationConstraint xmlns="http://s3.amazonaws.com/doc/2006-03-01/">PACHYDERM</LocationConstraint>`

func serveRoot(pc *client.APIClient, w http.ResponseWriter, r *http.Request, repo string) {
	if _, ok := r.Form["location"]; ok {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(LOCATION_RESPONSE))
	} else {
		w.Write([]byte("OK"))
	}
}

func serveFile(pc *client.APIClient, w http.ResponseWriter, r *http.Request, repo, name string) {
	fileInfo, err := pc.InspectFile(repo, "master", name)
	if err != nil {
		code := 500

		if strings.Contains(err.Error(), "not found in repo") {
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

	reader, err := pc.GetFileReadSeeker(repo, "master", name)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), 500)
		return
	}

	http.ServeContent(w, r, "", timestamp, reader)
}

// Serve runs an HTTP server with an S3-like API for PFS
func Serve(pc *client.APIClient, port uint16) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.SplitN(r.URL.Path, "/", 3)
		if err := r.ParseForm(); err != nil {
			http.Error(w, fmt.Sprintf("%v", err), 400)
			return
		}

		if len(parts) < 3 {
			http.Error(w, "Invalid path", 404)
			return
		}

		repo, file := parts[1], parts[2]

		if file == "" {
			serveRoot(pc, w, r, repo)
		} else {
			serveFile(pc, w, r, repo, file)
		}
	})

	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
