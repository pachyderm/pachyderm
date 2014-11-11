package main

import (
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"hash/adler32"
	"io"
	"log"
	"net/http"
	"net/url"
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

	client := etcd.NewClient([]string{"http://172.17.42.1:4001", "http://10.1.42.1:4001"})
	_master, err := client.Get(path.Join(etcdKey, shard), false, false)
	if err != nil {
		log.Fatal(err)
	}
	master := _master.Node.Value

	httpClient := &http.Client{}
	r.RequestURI = ""
	r.URL, err = url.Parse("http://" + path.Join(master, r.URL.String()))
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

	baseUrl := r.URL.String()
	client := etcd.NewClient([]string{"http://172.17.42.1:4001"})
	_endpoints, err := client.Get(etcdKey, false, true)
	if err != nil {
		log.Fatal(err)
	}
	endpoints := _endpoints.Node.Nodes
	log.Print(endpoints)

	for _, node := range endpoints {
		log.Print("Multicasting to: ", node, node.Value)
		httpClient := &http.Client{}
		r.RequestURI = ""
		r.URL, err = url.Parse("http://" + path.Join(node.Value, baseUrl))
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
}

func RouterMux() *http.ServeMux {
	mux := http.NewServeMux()

	pfsHandler := func(w http.ResponseWriter, r *http.Request) {
		Route(w, r, "/pfs/master")
	}
	commitHandler := func(w http.ResponseWriter, r *http.Request) {
		Multicast(w, r, "/pfs/master")
	}

	mux.HandleFunc("/pfs/", pfsHandler)
	mux.HandleFunc("/commit", commitHandler)
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "pong\n") })
	mux.HandleFunc("/index.html", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hi, you'd probably be happier curling at http://146.148.77.106/")
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
