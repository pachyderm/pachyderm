package router

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/pachyderm/pachyderm/src/common"
	"github.com/pachyderm/pachyderm/src/etcache"
	"github.com/pachyderm/pachyderm/src/route"
	"github.com/satori/go.uuid"
)

type Router struct {
	modulos uint64
	cache   etcache.Cache
}

func NewRouter(modulos uint64, cache etcache.Cache) *Router {
	return &Router{
		modulos,
		cache,
	}
}

func RouterFromArgs() (*Router, error) {
	modulos, err := strconv.ParseUint(os.Args[1], 10, 32)

	if err != nil {
		return nil, err
	}
	return NewRouter(modulos, etcache.NewCache()), nil
}

func (ro *Router) RouterMux() *http.ServeMux {
	mux := http.NewServeMux()

	fileHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "*") {
			route.MulticastHttp(ro.cache, w, r, "/pfs/master", route.ReturnAll)
		} else {
			route.RouteHttp(ro.cache, w, r, "/pfs/master", ro.modulos)
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
		route.MulticastHttp(ro.cache, w, r, "/pfs/master", route.ReturnOne)
	}
	branchHandler := func(w http.ResponseWriter, r *http.Request) {
		route.MulticastHttp(ro.cache, w, r, "/pfs/master", route.ReturnOne)
	}
	pipelineHandler := func(w http.ResponseWriter, r *http.Request) {
		route.MulticastHttp(ro.cache, w, r, "/pfs/master", route.ReturnOne)
	}

	mux.HandleFunc("/file/", fileHandler)
	mux.HandleFunc("/commit", commitHandler)
	mux.HandleFunc("/branch", branchHandler)
	mux.HandleFunc("/pipeline/", pipelineHandler)
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "pong\n") })
	mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintf(w, "%s\n", common.VersionString()) })
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to pfs!\n")
	})

	return mux
}

func (r *Router) RunServer() error {
	return http.ListenAndServe(":80", r.RouterMux())
}
