package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"code.google.com/p/go-uuid/uuid"
	"github.com/pachyderm-io/pfs/lib/route"
)

var modulos uint64

func RouterMux() *http.ServeMux {
	mux := http.NewServeMux()

	fileHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "*") {
			route.MulticastHttp(w, r, "/pfs/master")
		} else {
			route.RouteHttp(w, r, "/pfs/master", modulos)
		}
	}
	commitHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			values := r.URL.Query()
			values.Add("commit", uuid.New())
			r.URL.RawQuery = values.Encode()
		}
		route.MulticastHttp(w, r, "/pfs/master")
	}
	branchHandler := func(w http.ResponseWriter, r *http.Request) {
		route.MulticastHttp(w, r, "/pfs/master")
	}
	jobHandler := func(w http.ResponseWriter, r *http.Request) {
		route.MulticastHttp(w, r, "/pfs/master")
	}
	materializeHandler := func(w http.ResponseWriter, r *http.Request) {
		route.MulticastHttp(w, r, "/pfs/master")
	}

	mux.HandleFunc("/file/", fileHandler)
	mux.HandleFunc("/commit", commitHandler)
	mux.HandleFunc("/branch", branchHandler)
	mux.HandleFunc("/job/", jobHandler)
	mux.HandleFunc("/materialize", materializeHandler)
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "pong\n") })
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to pfs!\n")
	})

	return mux
}

func main() {
	log.SetFlags(log.Lshortfile)
	log.Print("Starting up...")

	var err error
	modulos, err = strconv.ParseUint(os.Args[1], 10, 32)

	if err != nil {
		log.Fatalf("Failed to parse %s as Uint.")
	}
	log.Fatal(http.ListenAndServe(":80", RouterMux()))
}
