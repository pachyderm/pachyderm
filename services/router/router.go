package main

import (
	"io"
	"net/http"
	"strings"
    "fmt"
    "github.com/coreos/go-etcd/etcd"
    "hash/adler32"
    "log"
    "os"
    "strconv"
)

var modulos uint64;

func callonleaves(node *etcd.Node, cb func (leaf *etcd.Node)) {
    if !node.Dir {
        cb(node)
    } else {
        for _, n := range node.Nodes {
            callonleaves(n, cb)
        }
    }
}

func Route(w http.ResponseWriter, r *http.Request, prefix string) {
    log.Printf("Request to `Route`: %s.\n", r.URL.String())

    file := strings.TrimPrefix(r.URL.String(), prefix)
	log.Printf("file = %s.", file)

    if len(file) == len(r.URL.String()) {
		http.Error(w, "Incorrect prefix on request.", 500)
		log.Print("Incorrect prefix on request.")
		return
	}

	bucket := uint64(adler32.Checksum([]byte(file))) % modulos

    client := etcd.NewClient([]string{"http://172.17.42.1:4001"})
    endpoints, err := client.Get("/pfs/master", false, true)
    if err != nil { log.Fatal(err) }

    callonleaves(endpoints.Node,
			func (n *etcd.Node) {
                // key := ["pfs", "master", "n-m"]
				key := strings.Split(n.Key, "/")
                log.Print("Full Key: ", key)
				shard := key[2]
				log.Print("Shard: ", shard)
				if  fmt.Sprintf("%d-%d", int(bucket), int(modulos)) == shard {
					fmt.Print("Posting to: http://", n.Value + "/" + os.Args[2] + r.URL.String())
					resp, err := http.Post("http://" + n.Value + "/" + os.Args[2] + r.URL.String(),
						"text/plain", r.Body)
					if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }
					io.Copy(w, resp.Body)
					return
				}
            })
}

func RouterMux() *http.ServeMux {
    mux := http.NewServeMux()

	pfsHandler := func(w http.ResponseWriter, r *http.Request) {
		Route(w, r, "/pfs")
	}

    mux.HandleFunc("/pfs", pfsHandler)
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
