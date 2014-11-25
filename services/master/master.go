package main

import (
	"fmt"
	"github.com/pachyderm-io/pfs/lib/btrfs"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

//TODO commits should be content based

func commitForRequest(r *http.Request) string {
	if c := r.URL.Query().Get("commit"); c != "" {
		return c
	}
	return "master"
}

// PfsHandler is the core route for modifying the contents of the fileystem.
// Changes are not replicated until a call to CommitHandler.
func PfsHandler(w http.ResponseWriter, r *http.Request, fs *btrfs.FS) {
	url := strings.Split(r.URL.Path, "/")
	commitPath := path.Join("repo", commitForRequest(r))
	file := path.Join(append([]string{commitPath}, url[2:]...)...)

	if r.Method == "GET" {
		http.StripPrefix("/pfs/", http.FileServer(http.Dir(commitPath))).ServeHTTP(w, r)
	} else if r.Method == "POST" {
		size, err := fs.CreateFromReader(file, r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}
		fmt.Fprintf(w, "Added %s, size: %d.\n", file, size)
	} else if r.Method == "PUT" {
		size, err := fs.WriteFile(file, r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}
		fmt.Fprintf(w, "Wrote %s, size: %d.\n", file, size)
	} else if r.Method == "DELETE" {
		if err := fs.Remove(file); err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}
		fmt.Fprintf(w, "Deleted %s.\n", file)
	}
}

// CommitHandler creates a snapshot of outstanding changes.
func CommitHandler(w http.ResponseWriter, r *http.Request, fs *btrfs.FS) {
	if commit, err := fs.Commit("repo", commitForRequest(r)); err != nil {
		http.Error(w, err.Error(), 500)
		log.Print(err)
		return
	} else {
		fmt.Fprintf(w, "Create commit: %s.\n", commit)
	}
}

// MasterMux creates a multiplexer for a Master writing to the passed in FS.
func MasterMux(fs *btrfs.FS) *http.ServeMux {
	mux := http.NewServeMux()

	commitHandler := func(w http.ResponseWriter, r *http.Request) {
		CommitHandler(w, r, fs)
	}

	pfsHandler := func(w http.ResponseWriter, r *http.Request) {
		PfsHandler(w, r, fs)
	}

	mux.HandleFunc("/commit", commitHandler)
	mux.HandleFunc("/pfs/", pfsHandler)
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "pong\n") })

	return mux
}

// RunServer runs a master server listening on port 80
func RunServer(fs *btrfs.FS) {
	http.ListenAndServe(":80", MasterMux(fs))
}

func main() {
	log.SetFlags(log.Lshortfile)
	fs := btrfs.NewFSWithRandSeq("master-" + os.Args[1])
	if err := fs.EnsureNamespace(); err != nil {
		log.Fatal(err)
	}
	if err := fs.Init("repo"); err != nil {
		log.Fatal(err)
	}
	log.Print("Listening on port 80...")
	RunServer(fs)
}
