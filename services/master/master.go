package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"code.google.com/p/go-uuid/uuid"
	"github.com/pachyderm-io/pfs/lib/btrfs"
	"github.com/pachyderm-io/pfs/lib/mapreduce"
)

var dataRepo, compRepo string
var shard, modulos uint64

func parseArgs() {
	// os.Args[1] looks like 2-16
	dataRepo = "data-" + os.Args[1] + btrfs.RandSeq(4)
	compRepo = "comp-" + os.Args[1] + btrfs.RandSeq(4)
	s_m := strings.Split(os.Args[1], "-")
	var err error
	shard, err = strconv.ParseUint(s_m[0], 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	modulos, err = strconv.ParseUint(s_m[1], 10, 64)
	if err != nil {
		log.Fatal(err)
	}
}

var jobDir string = "job"

func commitParam(r *http.Request) string {
	if p := r.URL.Query().Get("commit"); p != "" {
		return p
	}
	return "master"
}

func branchParam(r *http.Request) string {
	if p := r.URL.Query().Get("branch"); p != "" {
		return p
	}
	return "master"
}

func materializeParam(r *http.Request) string {
	if _, ok := r.URL.Query()["run"]; ok {
		return "true"
	}
	return "false"
}

func indexOf(haystack []string, needle string) int {
	for i, s := range haystack {
		if s == needle {
			return i
		}
	}
	return -1
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
func genericFileHandler(fs string, w http.ResponseWriter, r *http.Request) {
	url := strings.Split(r.URL.Path, "/")
	// url looks like: /foo/bar/.../file/<file>
	fileStart := indexOf(url, "file") + 1
	// file is the path in the filesystem we're getting
	file := path.Join(append([]string{fs}, url[fileStart:]...)...)

	if r.Method == "GET" {
		if strings.Contains(file, "*") {
			if !strings.HasSuffix(file, "*") {
				http.Error(w, "Illegal path containing internal `*`. `*` is currently only allowed as the last character of a path.", 400)
			} else {
				dir := path.Dir(file)
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
			cat(w, file)
		}
	} else if r.Method == "POST" {
		btrfs.MkdirAll(path.Dir(file))
		size, err := btrfs.CreateFromReader(file, r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}
		fmt.Fprintf(w, "Created %s, size: %d.\n", file, size)
	} else if r.Method == "PUT" {
		btrfs.MkdirAll(path.Dir(file))
		size, err := btrfs.WriteFile(file, r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}
		fmt.Fprintf(w, "Created %s, size: %d.\n", file, size)
	} else if r.Method == "DELETE" {
		if err := btrfs.Remove(file); err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}
		fmt.Fprintf(w, "Deleted %s.\n", file)
	}
}

// FileHandler is the core route for modifying the contents of the fileystem.
// Changes are not replicated until a call to CommitHandler.
func FileHandler(w http.ResponseWriter, r *http.Request) {
	genericFileHandler(path.Join(dataRepo, "master"), w, r)
}

// CommitHandler creates a snapshot of outstanding changes.
func CommitHandler(w http.ResponseWriter, r *http.Request) {
	url := strings.Split(r.URL.Path, "/")
	// url looks like [, commit, <commit>, file, <file>]
	if len(url) > 3 && url[3] == "file" {
		genericFileHandler(path.Join(dataRepo, url[2]), w, r)
		return
	}
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

		if materializeParam(r) == "true" {
			go func() {
				err := mapreduce.Materialize(dataRepo, branchParam(r), commitParam(r), compRepo, jobDir, shard, modulos)
				if err != nil {
					log.Print(err)
				}
			}()
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
	url := strings.Split(r.URL.Path, "/")
	// url looks like [, commit, <commit>, file, <file>]
	if len(url) > 3 && url[3] == "file" {
		genericFileHandler(path.Join(dataRepo, url[2]), w, r)
		return
	}
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
	// url looks like [, job, <job>, file, <file>]
	if len(url) > 3 && url[3] == "file" {
		mapreduce.WaitJob(dataRepo, commitParam(r), url[2])
		genericFileHandler(path.Join(compRepo, "master", url[2]), w, r)
		return
	}
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
	parseArgs()
	if err := btrfs.Ensure(dataRepo); err != nil {
		log.Fatal(err)
	}
	if err := btrfs.Ensure(compRepo); err != nil {
		log.Fatal(err)
	}
	log.Print("Listening on port 80...")
	log.Printf("dataRepo: %s, compRepo: %s.", dataRepo, compRepo)
	RunServer()
}
