package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/src/client"
)

// const shutdownTimeoutSecs = 10
const locationResponse = `<?xml version="1.0" encoding="UTF-8"?>
<LocationConstraint xmlns="http://s3.amazonaws.com/doc/2006-03-01/">PACHYDERM</LocationConstraint>`

// type S3Server struct {
// 	server *http.Server
// }

// func NewS3Server() *S3Server {
// 	return *S3Server{
// 		server: *http.Server{
// 			Addr: ,
// 			Handler: 
// 		}
// 	}
// }

// func (s *S3Server) Listen() error {
// 	if err := s.server.ListenAndServe(); err == http.ErrServerClosed {
// 		return nil
// 	} else {
// 		return err
// 	}
// }

// func (s *S3Server) Close() error {
// 	ctx := context.WithTimeout(context.Background(), shutdownTimeoutSecs * time.Second)
// 	return s.server.Shutdown(ctx)
// }

type handler struct {
	pc *client.APIClient
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		h.serveRoot(w, r, repo)
	} else {
		h.serveFile(w, r, repo, file)
	}
}

func (h handler) serveRoot(w http.ResponseWriter, r *http.Request, repo string) {
	if _, ok := r.Form["location"]; ok {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(locationResponse))
	} else {
		w.Write([]byte("OK"))
	}
}

func (h handler) serveFile(w http.ResponseWriter, r *http.Request, repo, name string) {
	fileInfo, err := h.pc.InspectFile(repo, "master", name)
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

	reader, err := h.pc.GetFileReadSeeker(repo, "master", name)
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
