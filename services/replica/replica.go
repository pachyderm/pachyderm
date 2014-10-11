package main

import (
	"log"
	"net/http"
	"os"
    "pfs/lib/btrfs"
    "strings"
)

//TODO these functions can be merge right?

// http://host/recvbase

func RecvBaseHandler(w http.ResponseWriter, r *http.Request, fs *btrfs.FS) {
	log.Print("RecvBase.")
    err := fs.Recv(".", r.Body)
    if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }
}

// http://host/recv

func RecvHandler(w http.ResponseWriter, r *http.Request, fs *btrfs.FS) {
	log.Print("Recv.")
	err := fs.Recv(".", r.Body)
    if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }
}

// http://host/del

func DelCommitHandler(w http.ResponseWriter, r *http.Request, fs *btrfs.FS) {
    url := strings.Split(r.URL.String(), "/")
    log.Print("Del.")
    err := fs.SubvolumeDelete(url[2])
    if err != nil { http.Error(w, err.Error(), 500); log.Print(err); return }
}

func SlaveMux(fs *btrfs.FS) *http.ServeMux {
    mux := http.NewServeMux()

	// http://host/recvbase/fs
	recvBaseHandler := func (w http.ResponseWriter, r *http.Request) {
		RecvBaseHandler(w, r, fs)
	}

	// http://host/recv/fs

	recvHandler := func (w http.ResponseWriter, r *http.Request) {
        RecvHandler(w, r, fs)
    }

    delCommitHandler := func (w http.ResponseWriter, r *http.Request) {
        DelCommitHandler(w, r, fs)
    }

	mux.HandleFunc("/recvbase", recvBaseHandler)
	mux.HandleFunc("/recv", recvHandler)
    mux.HandleFunc("/del", delCommitHandler)

    return mux;
}

func RunServer(fs *btrfs.FS) {
    http.ListenAndServe(":80", SlaveMux(fs))
}

func main() {
    log.SetFlags(log.Lshortfile)
	fs := btrfs.ExistingFS("pfs", "replica-" + os.Args[1])
    fs.EnsureNamespace()
    log.Print("Listening on port 80...")
    RunServer(fs)
}
