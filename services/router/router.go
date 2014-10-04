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

func callonleaves(node *etcd.Node, cb func (leaf *etcd.Node)) {
    if !node.Dir {
        cb(node)
    } else {
        for _, n := range node.Nodes {
            callonleaves(n, cb)
        }
    }
}

func Handle(w http.ResponseWriter, r *http.Request) {
    fmt.Printf("Routing request to: %s.\n", r.URL.String())
	hash := adler32.Checksum([]byte(r.URL.String()))

    client := etcd.NewClient([]string{"http://172.17.42.1:4001"})
    endpoints, err := client.Get(os.Args[1], false, true)
    if err != nil { log.Fatal(err) }

    callonleaves(endpoints.Node,
			func (n *etcd.Node) {
				key := strings.Split(n.Key, "/")
				bucket, err := strconv.ParseUint(key[len(key) - 2], 10, 32)
				if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }
				modulos, err := strconv.ParseUint(key[len(key) - 1], 10, 32)
				if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }
				if  hash % uint32(modulos) == uint32(bucket) {
					fmt.Print(n.Value + "/" + os.Args[2] + r.URL.String())
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

    mux.HandleFunc("/", Handle)

    return mux;
}

func main() {
    log.SetFlags(log.Lshortfile)
    etcd_path := os.Args[1]
    endpoint_prefix := os.Args[2]
	log.Printf("Routing from etcd path %s with prefix %s.\n", etcd_path,
			endpoint_prefix)

	http.ListenAndServe(":80", RouterMux())
}
