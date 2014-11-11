package main

import (
	"fmt"
	"github.com/pachyderm-io/pfs/lib/btrfs"
	"log"
	"net/http"
	"os"
	"strings"
)

//TODO these functions can be merge right?

// RecvHandler takes the output of btrfs send and applies it to the local
// filesystem.
func RecvHandler(w http.ResponseWriter, r *http.Request, fs *btrfs.FS) {
	log.Print("Recv.")
	err := fs.Recv(".", r.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Print(err)
		return
	}
}

// DelCommitHandler deletes a commit.
func DelCommitHandler(w http.ResponseWriter, r *http.Request, fs *btrfs.FS) {
	url := strings.Split(r.URL.Path, "/")
	log.Print("Del.")
	err := fs.SubvolumeDelete(url[2])
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Print(err)
		return
	}
}

// SlaveMux creates a multiplexer for a Slave writing to the passed in FS.
func SlaveMux(fs *btrfs.FS) *http.ServeMux {
	mux := http.NewServeMux()

	recvHandler := func(w http.ResponseWriter, r *http.Request) {
		RecvHandler(w, r, fs)
	}

	delCommitHandler := func(w http.ResponseWriter, r *http.Request) {
		DelCommitHandler(w, r, fs)
	}

	mux.HandleFunc("/recv", recvHandler)
	mux.HandleFunc("/del", delCommitHandler)
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "pong\n") })

	return mux
}

// RunServer runs a replica server listening on port 80
func RunServer(fs *btrfs.FS) {
	http.ListenAndServe(":80", SlaveMux(fs))
}

func main() {
	log.SetFlags(log.Lshortfile)
	fs := btrfs.NewFS("replica-" + os.Args[1])
	fs.EnsureNamespace()
	log.Print("Listening on port 80...")
	RunServer(fs)
}
