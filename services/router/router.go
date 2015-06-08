package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"code.google.com/p/go-uuid/uuid"
	"github.com/pachyderm/pfs/lib/router"
)

type Router struct {
	modulos uint64
}

func NewRouter(modulos uint64) *Router {
	return &Router{
		modulos: modulos,
	}
}

func RouterFromArgs() (*Router, error) {
	modulos, err := strconv.ParseUint(os.Args[1], 10, 32)

	if err != nil {
		return nil, err
	}
	return NewRouter(modulos), nil
}

func (ro *Router) RouterMux() *http.ServeMux {
	mux := http.NewServeMux()

	fileHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "*") {
			router.MulticastHttp(w, r, "/pfs/master")
		} else {
			router.RouteHttp(w, r, "/pfs/master", ro.modulos)
		}
	}
	commitHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			values := r.URL.Query()
			values.Add("commit", uuid.New())
			r.URL.RawQuery = values.Encode()
		}
		router.MulticastHttp(w, r, "/pfs/master")
	}
	branchHandler := func(w http.ResponseWriter, r *http.Request) {
		router.MulticastHttp(w, r, "/pfs/master")
	}
	jobHandler := func(w http.ResponseWriter, r *http.Request) {
		router.MulticastHttp(w, r, "/pfs/master")
	}
	materializeHandler := func(w http.ResponseWriter, r *http.Request) {
		router.MulticastHttp(w, r, "/pfs/master")
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

func (r *Router) RunServer() {
	http.ListenAndServe(":80", r.RouterMux())
}

func main() {
	log.SetFlags(log.Lshortfile)
	log.Print("Starting up...")
	r, err := RouterFromArgs()
	if err != nil {
		log.Fatal(err)
	}
	r.RunServer()
}
