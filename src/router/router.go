package router

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/src/log"
	"github.com/pachyderm/pachyderm/src/route"
	"github.com/pachyderm/pachyderm/src/shard"
	"github.com/pachyderm/pachyderm/src/traffic"
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

func testHandler(w http.ResponseWriter, r *http.Request) {
	pipeline := `
image ubuntu

input data

run mkdir -p /out/counts
run cat /in/data/* | tr -cs "A-Za-z'" "\n" | sort | uniq -c | sort -n -r | while read count; do echo ${count% *} >/out/counts/${count#* }; done
shuffle counts
run find /out/counts | while read count; do cat $count | awk '{ sum+=$1} END {print sum}' >/tmp/count; mv /tmp/count $count; done
`
	// used to prevent collisions
	rand := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	url := "http://localhost"
	var _w traffic.Workload
	// Run the workload
	workload := _w.Generate(rand, 20).Interface().(traffic.Workload)
	shard.RunWorkload(url, workload, nil)
	// Make sure we see the changes we should
	facts := workload.Facts()
	shard.RunWorkload(url, facts, nil)
	log.Print("Workload done.")
	// Install the pipeline
	log.Print("Installing pipeline:")
	res, err := http.Post(url+"/pipeline/wc", "application/text", strings.NewReader(pipeline))
	log.Print("Done.")
	if err != nil {
		log.Print(err)
	}
	res.Body.Close()
	// Make a commit
	log.Print("Committing.")
	res, err = http.Post(url+"/commit?commit=commit1", "", nil)
	log.Print("Done.")
	if err != nil {
		log.Print(err)
	}
	res.Body.Close()
	// TODO(jd) make this check for correctness, not just that the request
	// completes. It's a bit hard because the input is random. Probably the
	// right idea is to modify the traffic package so that it keeps track of
	// this.
	res, err = http.Get(url + "/pipeline/wc/file/counts/*?commit=commit1")
	if err != nil {
		log.Print(err)
	}
	if res.StatusCode != 200 {
		log.Print("Bad status code.")
	}
	res.Body.Close()
	fmt.Fprint(w, "Tests Complete.")
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
	mux.HandleFunc("/test", testHandler)
	mux.HandleFunc("/log", logHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Welcome to pfs!\n")
	})

	return mux
}

func (r *Router) RunServer() error {
	return http.ListenAndServe(":80", r.RouterMux())
}
