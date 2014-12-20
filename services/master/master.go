package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"code.google.com/p/go-uuid/uuid"
	"github.com/pachyderm-io/pfs/lib/btrfs"
)

var dataRepo string
var compRepo string

var jobDir string = "jobs"

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

// FileHandler is the core route for modifying the contents of the fileystem.
// Changes are not replicated until a call to CommitHandler.
func FileHandler(w http.ResponseWriter, r *http.Request) {
	url := strings.Split(r.URL.Path, "/")
	// commitFile is used for read methods (GET)
	commitFile := path.Join(append([]string{path.Join(dataRepo, commitParam(r))}, url[2:]...)...)
	// branchFile is used for write methods (POST, PUT, DELETE)
	branchFile := path.Join(append([]string{path.Join(dataRepo, branchParam(r))}, url[2:]...)...)

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
	if r.Method == "GET" {
		commits, err := btrfs.ReadDir(dataRepo)
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}

		for _, ci := range commits {
			if uuid.Parse(ci.Name()) != nil {
				fmt.Fprintf(w, "%s    %s\n", ci.Name(), ci.ModTime().Format("2006-01-02T15:04:05.999999-07:00"))
			}
		}
	} else if r.Method == "POST" {
		err := btrfs.Commit(dataRepo, commitParam(r), branchParam(r))
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}

		fmt.Fprint(w, commitParam)
	} else {
		http.Error(w, "Unsupported method.", http.StatusMethodNotAllowed)
		log.Printf("Unsupported method %s in request to %s.", r.Method, r.URL.String())
		return
	}
}

// BranchHandler creates a new branch from commit.
func BranchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		branches, err := btrfs.ReadDir(dataRepo)
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}

		for _, bi := range branches {
			if uuid.Parse(bi.Name()) == nil {
				fmt.Fprintf(w, "%s    %s\n", bi.Name(), bi.ModTime().Format("2006-01-02T15:04:05.999999-07:00"))
			}
		}
	} else if r.Method == "POST" {
		if err := btrfs.Branch(dataRepo, commitParam(r), branchParam(r)); err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}
		fmt.Fprintf(w, "Created branch. (%s) -> %s.\n", commitParam(r), branchParam(r))
	} else {
		http.Error(w, "Invalid method.", 405)
		log.Print("Invalid method %s.", r.Method)
		return
	}
}

func JobHandler(w http.ResponseWriter, r *http.Request) {
	url := strings.Split(r.URL.Path, "/")
	log.Print(url)
	if r.Method == "GET" {
	} else if r.Method == "POST" {
		r.URL.Path = path.Join("/file", jobDir, url[2])
		FileHandler(w, r)
	}
}

// MasterMux creates a multiplexer for a Master writing to the passed in FS.
func MasterMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/branch", BranchHandler)
	mux.HandleFunc("/commit", CommitHandler)
	mux.HandleFunc("/file/", FileHandler)
	mux.HandleFunc("/job/", JobHandler)
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "pong\n") })

	return mux
}

// RunServer runs a master server listening on port 80
func RunServer() {
	http.ListenAndServe(":80", MasterMux())
}

func main() {
	log.SetFlags(log.Lshortfile)
	dataRepo = "data-" + os.Args[1] + btrfs.RandSeq(4)
	compRepo = "comp-" + os.Args[1] + btrfs.RandSeq(4)
	if err := btrfs.Ensure(dataRepo); err != nil {
		log.Fatal(err)
	}
	if err := btrfs.Ensure(compRepo); err != nil {
		log.Fatal(err)
	}
	log.Print("Listening on port 80...")
	RunServer()
}
