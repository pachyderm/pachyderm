package main

import (
	"fmt"
	"github.com/pachyderm-io/pfs/lib/btrfs"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

//TODO commits should be content based

func commitParam(r *http.Request) string {
	if c := r.URL.Query().Get("commit"); c != "" {
		return c
	}
	return "master"
}

func branchParam(r *http.Request) string {
	if c := r.URL.Query().Get("branch"); c != "" {
		return c
	}
	return "master"
}

// PfsHandler is the core route for modifying the contents of the fileystem.
// Changes are not replicated until a call to CommitHandler.
func PfsHandler(w http.ResponseWriter, r *http.Request, fs *btrfs.FS) {
	url := strings.Split(r.URL.Path, "/")
	commitPath := path.Join("repo", commitParam(r))
	file := path.Join(append([]string{commitPath}, url[2:]...)...)

	if r.Method == "GET" {
		if f, err := fs.Open(file); err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		} else {
			io.Copy(w, f)
		}
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
	var commit string
	var err error
	if commit, err = fs.Commit("repo", branchParam(r)); err != nil {
		http.Error(w, err.Error(), 500)
		log.Print(err)
		return
	}

	fmt.Fprintf(w, "Create commit: %s.\n", commit)
}

// BranchHandler creates a new branch from commit.
func BranchHandler(w http.ResponseWriter, r *http.Request, fs *btrfs.FS) {
	if r.Method != "POST" {
		http.Error(w, "Invalid method.", 405)
		log.Print("Invalid method %s.", r.Method)
		return
	}

	if err := fs.Branch("repo", commitParam(r), branchParam(r)); err != nil {
		http.Error(w, err.Error(), 500)
		log.Print(err)
		return
	}
	fmt.Fprintf(w, "Created branch. (%s) -> %s.\n", commitParam(r), branchParam(r))
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

	branchHandler := func(w http.ResponseWriter, r *http.Request) {
		BranchHandler(w, r, fs)
	}

	mux.HandleFunc("/commit", commitHandler)
	mux.HandleFunc("/pfs/", pfsHandler)
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "pong\n") })
	mux.HandleFunc("/branch", branchHandler)

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
