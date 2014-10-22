package main

import (
	"io"
	"net/http"
	"net/url"
	"strings"
    "fmt"
    "github.com/coreos/go-etcd/etcd"
    "hash/adler32"
    "log"
    "os"
    "path"
    "strconv"
)

var modulos uint64;

func Route(w http.ResponseWriter, r *http.Request, etcdKey, prefix string) {
    log.Printf("Request to `Route`: %s.\n", r.URL.String())

    file := strings.TrimPrefix(r.URL.String(), prefix)
	log.Printf("file = %s.", file)

    if len(file) == len(r.URL.String()) {
		http.Error(w, "Incorrect prefix on request.", 500)
		log.Print("Incorrect prefix on request.")
		return
	}

	bucket := uint64(adler32.Checksum([]byte(file))) % modulos
	shard := fmt.Sprint(bucket, "-", os.Args[1])

    client := etcd.NewClient([]string{"http://172.17.42.1:4001"})
    _master, err := client.Get(path.Join(etcdKey, shard), false, false)
    if err != nil { log.Fatal(err) }
	master := _master.Node.Value

	httpClient := &http.Client{}
	r.RequestURI = ""
	r.URL, err = url.Parse("http://" + path.Join(master, r.URL.String()))
	if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }
	log.Print("Proxying to: " + r.URL.String())
	resp, err := httpClient.Do(r)
	if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }
	io.Copy(w, resp.Body)
}

func Multicast(w http.ResponseWriter, r *http.Request, etcdKey string) {
    log.Printf("Request to `Multicast`: %s.\n", r.URL.String())

	baseUrl := r.URL.String()
    client := etcd.NewClient([]string{"http://172.17.42.1:4001"})
    _endpoints, err := client.Get(etcdKey, false, true)
    if err != nil { log.Fatal(err) }
	endpoints := _endpoints.Node.Nodes
	log.Print(endpoints)

	for _, node := range endpoints {
		log.Print("Multicasting to: ", node, node.Value)
		httpClient := &http.Client{}
		r.RequestURI = ""
		r.URL, err = url.Parse("http://" + path.Join(node.Value, baseUrl))
		if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }
		log.Print("Proxying to: " + r.URL.String())
		resp, err := httpClient.Do(r)
		if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }
		io.Copy(w, resp.Body)
	}
}


func RouterMux() *http.ServeMux {
    mux := http.NewServeMux()

	pfsHandler := func(w http.ResponseWriter, r *http.Request) {
		Route(w, r, "/pfs/master", "/pfs")
	}
	commitHandler := func(w http.ResponseWriter,r *http.Request) {
		Multicast(w, r, "/pfs/master")
    }

    mux.HandleFunc("/pfs/", pfsHandler)
	mux.HandleFunc("/commit", commitHandler)
	mux.HandleFunc("/ping", func (w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "pong\n") })

    return mux;
}

func main() {
    log.SetFlags(log.Lshortfile)

	var err error
	modulos, err = strconv.ParseUint(os.Args[1], 10, 32)

	if err != nil { log.Fatalf("Failed to parse %s as Uint.") }
	http.ListenAndServe(":80", RouterMux())
}
