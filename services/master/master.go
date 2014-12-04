package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/pachyderm-io/pfs/lib/btrfs"
)

var repo string

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

func cat(w http.ResponseWriter, name string) {
	f, err := btrfs.Open(name)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Print(err)
	}
	defer f.Close()

	if _, err := io.Copy(w, f); err != nil {
		http.Error(w, err.Error(), 500)
		log.Print(err)
	}
}

// PfsHandler is the core route for modifying the contents of the fileystem.
// Changes are not replicated until a call to CommitHandler.
func PfsHandler(w http.ResponseWriter, r *http.Request) {
	url := strings.Split(r.URL.Path, "/")
	// commitFile is used for read methods (GET)
	commitFile := path.Join(append([]string{path.Join("repo", commitParam(r))}, url[2:]...)...)
	// branchFile is used for write methods (POST, PUT, DELETE)
	branchFile := path.Join(append([]string{path.Join("repo", branchParam(r))}, url[2:]...)...)

	if r.Method == "GET" {
		if strings.Contains(commitFile, "*") {
			if !strings.HasSuffix(commitFile, "*") {
				http.Error(w, "Illegal path containing internal `*`. `*` is currently only allowed as the last character of a path.", 400)
			} else {
				dir := path.Dir(commitFile)
				files, err := btrfs.ReadDir(dir)
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}
				for _, fi := range files {
					if fi.IsDir() {
						continue
					} else {
						cat(w, path.Join(dir, fi.Name()))
					}
				}
			}
		} else {
			cat(w, commitFile)
		}
	} else if r.Method == "POST" {
		btrfs.MkdirAll(path.Dir(branchFile))
		size, err := btrfs.CreateFromReader(branchFile, r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}
		fmt.Fprintf(w, "Created %s, size: %d.\n", branchFile, size)
	} else if r.Method == "PUT" {
		btrfs.MkdirAll(path.Dir(branchFile))
		size, err := btrfs.WriteFile(branchFile, r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}
		fmt.Fprintf(w, "Created %s, size: %d.\n", branchFile, size)
	} else if r.Method == "DELETE" {
		if err := btrfs.Remove(branchFile); err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}
		fmt.Fprintf(w, "Deleted %s.\n", branchFile)
	}
}

// CommitHandler creates a snapshot of outstanding changes.
func CommitHandler(w http.ResponseWriter, r *http.Request) {
	var commit string
	var err error
	if commit, err = btrfs.Commit("repo", commitParam(r), branchParam(r)); err != nil {
		http.Error(w, err.Error(), 500)
		log.Print(err)
		return
	}

	fmt.Fprintf(w, "Create commit: %s.\n", commit)
}

// BranchHandler creates a new branch from commit.
func BranchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid method.", 405)
		log.Print("Invalid method %s.", r.Method)
		return
	}

	if err := btrfs.Branch("repo", commitParam(r), branchParam(r)); err != nil {
		http.Error(w, err.Error(), 500)
		log.Print(err)
		return
	}
	fmt.Fprintf(w, "Created branch. (%s) -> %s.\n", commitParam(r), branchParam(r))
}

// MasterMux creates a multiplexer for a Master writing to the passed in FS.
func MasterMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/commit", CommitHandler)
	mux.HandleFunc("/pfs/", PfsHandler)
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "pong\n") })
	mux.HandleFunc("/branch", BranchHandler)

	return mux
}

// RunServer runs a master server listening on port 80
func RunServer() {
	http.ListenAndServe(":80", MasterMux())
}

func main() {
	log.SetFlags(log.Lshortfile)
	repo = "master-" + os.Args[1] + btrfs.RandSeq(4)
	if err := btrfs.Ensure(repo); err != nil {
		log.Fatal(err)
	}
	log.Print("Listening on port 80...")
	RunServer()
}
