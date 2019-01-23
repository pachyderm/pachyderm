package main

import (
	"fmt"
	"net/http"

	"github.com/pachyderm/pachyderm/src/client"
)

func Serve(pc *client.APIClient, port uint16) {
	fmt.Println("Serve")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%+v\n", r)
		http.Error(w, "not implemented", http.StatusNotImplemented)
	})
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
