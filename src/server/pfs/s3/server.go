package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/src/client"
)

const locationResponse = `<?xml version="1.0" encoding="UTF-8"?>
<LocationConstraint xmlns="http://s3.amazonaws.com/doc/2006-03-01/">PACHYDERM</LocationConstraint>`

var (
	pathMatcher = regexp.MustCompile("^/(([A-Za-z0-9_-]+)\\.)?([A-Za-z0-9_-]+)/(.*)$")
)

type handler struct {
	pc *client.APIClient
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	match := pathMatcher.FindStringSubmatch(r.URL.Path)
	if len(match) == 0 {
		http.Error(w, "Invalid path", 404)
		return
	}

	branch := match[2]
	if branch == "" {
		branch = "master"
	}
	repo := match[3]
	file := match[4]

	_, err := h.pc.InspectBranch(repo, branch)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), 500)
		return
	}

	if file == "" {
		h.serveRoot(w, r, repo)
	} else {
		h.serveFile(w, r, repo, branch, file)
	}
}

func (h handler) serveRoot(w http.ResponseWriter, r *http.Request, repo string) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("%v", err), 400)
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

	reader, err := h.pc.GetFileReadSeeker(repo, branch, file)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), 500)
		return
	}

	http.ServeContent(w, r, "", timestamp, reader)
}

// Server runs an HTTP server with an S3-like API for PFS
func Server(pc *client.APIClient, port uint16) *http.Server {
	return &http.Server {
		Addr: fmt.Sprintf(":%d", port),
		Handler: handler{ pc: pc },
	}
}
