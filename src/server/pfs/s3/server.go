package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/src/client"
)

// Serve runs an HTTP server with an S3-like API for PFS
func Serve(pc *client.APIClient, port uint16) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.SplitN(r.URL.Path, "/", 3)

		if len(parts) < 3 {
			http.Error(w, "Invalid path", 404)
			return
		}

		repo, file := parts[1], parts[2]

		fileInfo, err := pc.InspectFile(repo, "master", file)
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

		reader, err := pc.GetFileReadSeeker(repo, "master", file)
		if err != nil {
			http.Error(w, fmt.Sprintf("%v", err), 500)
			return
		}

		http.ServeContent(w, r, "", timestamp, reader)
	})

	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
