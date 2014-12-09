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
	"strings"
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
	log.Print("Proxying to: " + r.URL.String())
	resp, err := httpClient.Do(r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Print(err)
		return
	}
	defer resp.Body.Close()
	if _, err := io.Copy(w, resp.Body); err != nil {
		http.Error(w, err.Error(), 500)
		log.Print(err)
		return
	}
}

func Multicast(w http.ResponseWriter, r *http.Request, etcdKey string) {
	log.Printf("Request to `Multicast`: %s.\n", r.URL.String())
	_endpoints, err := etcache.Get(etcdKey, false, true)
	if err != nil {
		log.Fatal(err)
	}
	endpoints := _endpoints.Node.Nodes
	log.Print(endpoints)

	for i, node := range endpoints {
		log.Print("Multicasting to: ", node, node.Value)
		httpClient := &http.Client{}
		// `Do` will complain if r.RequestURI is set so we unset it
		r.RequestURI = ""
		r.URL.Scheme = "http"
		r.URL.Host = node.Value
		log.Print("Proxying to: " + r.URL.String())
		resp, err := httpClient.Do(r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			http.Error(w, fmt.Sprintf("Failed request (%s) to %s.", resp.Status, r.URL.String()), resp.StatusCode)
			log.Printf("Failed request (%s) to %s.\n", resp.Status, r.URL.String())
			return
		}
		if i == 0 {
			if _, err := io.Copy(w, resp.Body); err != nil {
				http.Error(w, err.Error(), 500)
				log.Print(err)
				return
			}
		}
	}
}

func RouterMux() *http.ServeMux {
	mux := http.NewServeMux()

	pfsHandler := func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "*") {
			Multicast(w, r, "/pfs/master")
		} else {
			Route(w, r, "/pfs/master")
		}
	}
	commitHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			values := r.URL.Query()
			values.Add("commit", uuid.New())
			r.URL.RawQuery = values.Encode()
		}
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
