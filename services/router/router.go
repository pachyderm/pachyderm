package main

import (
	"code.google.com/p/go-uuid/uuid"
	"fmt"
	"github.com/pachyderm-io/pfs/lib/etcache"
	"hash/adler32"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
)

var modulos uint64

func hashRequest(r *http.Request) uint64 {
	return uint64(adler32.Checksum([]byte(r.URL.Path))) % modulos
}

func Route(w http.ResponseWriter, r *http.Request, etcdKey string) {
	log.Printf("Request to `Route`: %s.\n", r.URL.Path)

	bucket := hashRequest(r)
	shard := fmt.Sprint(bucket, "-", os.Args[1])

	_master, err := etcache.Get(path.Join(etcdKey, shard), false, false)
	if err != nil {
		log.Fatal(err)
	}
	master := _master.Node.Value

	httpClient := &http.Client{}
	// `Do` will complain if r.RequestURI is set so we unset it
	r.RequestURI = ""
	r.URL.Scheme = "http"
	r.URL.Host = master
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Print(err)
		return
	}
	log.Print("Proxying to: " + r.URL.String())
	resp, err := httpClient.Do(r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Print(err)
		return
	}
	io.Copy(w, resp.Body)
}

func Multicast(w http.ResponseWriter, r *http.Request, etcdKey string) {
	log.Printf("Request to `Multicast`: %s.\n", r.URL.String())
	_endpoints, err := etcache.Get(etcdKey, false, true)
	if err != nil {
		log.Fatal(err)
	}
	endpoints := _endpoints.Node.Nodes
	log.Print(endpoints)

	for _, node := range endpoints {
		log.Print("Multicasting to: ", node, node.Value)
		httpClient := &http.Client{}
		// `Do` will complain if r.RequestURI is set so we unset it
		r.RequestURI = ""
		r.URL.Scheme = "http"
		r.URL.Host = node.Value
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}
		log.Print("Proxying to: " + r.URL.String())
		log.Print("Request:", r)
		resp, err := httpClient.Do(r)
		if err != nil {
			fmt.Fprint(w, r.URL)
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}
		io.Copy(w, resp.Body)
	}
}

func RouterMux() *http.ServeMux {
	mux := http.NewServeMux()

	pfsHandler := func(w http.ResponseWriter, r *http.Request) {
		Route(w, r, "/pfs/master")
	}
	commitHandler := func(w http.ResponseWriter, r *http.Request) {
		values := r.URL.Query()
		values.Add("commit", uuid.New())
		r.URL.RawQuery = values.Encode()
		Multicast(w, r, "/pfs/master")
	}
	branchHandler := func(w http.ResponseWriter, r *http.Request) {
		Multicast(w, r, "/pfs/master")
	}

	mux.HandleFunc("/pfs/", pfsHandler)
	mux.HandleFunc("/commit", commitHandler)
	mux.HandleFunc("/branch", branchHandler)
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "pong\n") })
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "PFS\n")
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
