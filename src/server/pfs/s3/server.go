package main

import (
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/pachyderm/pachyderm/src/client"
)

func Serve(pc *client.APIClient, port uint16) {
	fmt.Println("Serve")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		repo, file := parts[1], path.Join(parts[2:]...)
		fmt.Printf("repo: %s, file: %s\n", repo, file)
		if err := pc.GetFile(repo, "master", file, 0, 0, w); err != nil {
			http.Error(w, fmt.Sprintf("%v", err), 500)
		}
		return
		// fmt.Printf("%+v\n", r)
		// http.Error(w, "not implemented", http.StatusNotImplemented)
	})
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
