package router

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/pachyderm/pachyderm/src/route"
	"github.com/satori/go.uuid"
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
			route.MulticastHttp(w, r, "/pfs/master", route.ReturnAll)
		} else {
			route.RouteHttp(w, r, "/pfs/master", ro.modulos)
		}
	}
	commitHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			values := r.URL.Query()
			if values.Get("commit") == "" {
				values.Add("commit", uuid.NewV4().String())
				r.URL.RawQuery = values.Encode()
			}
		}
		route.MulticastHttp(w, r, "/pfs/master", route.ReturnOne)
	}
	branchHandler := func(w http.ResponseWriter, r *http.Request) {
		route.MulticastHttp(w, r, "/pfs/master", route.ReturnOne)
	}
	jobHandler := func(w http.ResponseWriter, r *http.Request) {
		route.MulticastHttp(w, r, "/pfs/master", route.ReturnOne)
	}
	pipelineHandler := func(w http.ResponseWriter, r *http.Request) {
		route.MulticastHttp(w, r, "/pfs/master", route.ReturnOne)
	}
	logHandler := func(w http.ResponseWriter, r *http.Request) {
		route.MulticastHttp(w, r, "/pfs/master", route.ReturnAll)
	}

	mux.HandleFunc("/file/", fileHandler)
	mux.HandleFunc("/commit", commitHandler)
	mux.HandleFunc("/branch", branchHandler)
	mux.HandleFunc("/job/", jobHandler)
	mux.HandleFunc("/pipeline/", pipelineHandler)
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "pong\n") })
	mux.HandleFunc("/log", logHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to pfs!\n")
	})

	return mux
}

func (r *Router) RunServer() error {
	return http.ListenAndServe(":80", r.RouterMux())
}
