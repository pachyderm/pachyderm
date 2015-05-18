package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"

	"code.google.com/p/go-uuid/uuid"
	"github.com/pachyderm/pfs/lib/btrfs"
	"github.com/pachyderm/pfs/lib/mapreduce"
)

var (
	dataRepo, compRepo string
	logFile            string
	shard, modulos     uint64
)

func parseArgs() {
	// os.Args[1] looks like 2-16
	dataRepo = "data-" + os.Args[1]
	compRepo = "comp-" + os.Args[1]
	logFile = "log-" + os.Args[1]
	m := strings.Split(os.Args[1], "-")
	var err error
	shard, err = strconv.ParseUint(m[0], 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	modulos, err = strconv.ParseUint(m[1], 10, 64)
	if err != nil {
		log.Fatal(err)
	}
}

var jobDir = "job"

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

func hasBranch(r *http.Request) bool {
	return (r.URL.Query().Get("branch") == "")
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

func cat(w io.Writer, name string) (n int64, err error) {
	f, err := btrfs.Open(name)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	return io.Copy(w, f)
}

func notAllowed(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Unsupported method.", http.StatusMethodNotAllowed)
	log.Printf("Unsupported method %s in request to %s.", r.Method, r.URL)
}

func internalError(err error, w http.ResponseWriter, r *http.Request) {
	pc, fn, line, _ := runtime.Caller(1)
	http.Error(w, err.Error(), http.StatusInternalServerError)
	log.Println("%s[%s:%d] %v", runtime.FuncForPC(pc).Name(), fn, line, err)
}

// genericFileHandler serves files from fs. It's used after branch and commit
// info have already been extracted and ignores those aspects of the URL.
func genericFileHandler(fs string, w http.ResponseWriter, r *http.Request) {
	url := strings.Split(r.URL.Path, "/")
	// url looks like: /foo/bar/.../file/<file>
	fileStart := indexOf(url, "file") + 1
	// file is the path in the filesystem we're getting
	file := path.Join(append([]string{fs}, url[fileStart:]...)...)
	switch r.Method {
	case "GET":
		if !strings.Contains(file, "*") {
			if _, err := cat(w, file); err != nil {
				internalError(err, w, r)
			}
			return
		}

		if !strings.HasSuffix(file, "*") {
			http.Error(w, "Illegal path containing internal `*`. `*` is currently only allowed as the last character of a path.", http.StatusBadRequest)
			return
		}

		dir := path.Dir(file)
		files, err := btrfs.ReadDir(dir)
		if err != nil {
			internalError(err, w, r)
			return
		}

		for _, fi := range files {
			if fi.IsDir() {
				continue
			}

			name := path.Join(dir, fi.Name())
			if _, err := cat(w, name); err != nil {
				internalError(err, w, r)
				return
			}
		}

	case "POST":
		btrfs.MkdirAll(path.Dir(file))
		size, err := btrfs.CreateFromReader(file, r.Body)
		if err != nil {
			internalError(err, w, r)
			return
		}
		fmt.Fprintf(w, "Created %s, size: %d.\n", file, size)
	case "PUT":
		btrfs.MkdirAll(path.Dir(file))
		size, err := btrfs.WriteFile(file, r.Body)
		if err != nil {
			internalError(err, w, r)
			return
		}
		fmt.Fprintf(w, "Created %s, size: %d.\n", file, size)
	case "DELETE":
		if err := btrfs.Remove(file); err != nil {
			internalError(err, w, r)
			return
		}
		fmt.Fprintf(w, "Deleted %s.\n", file)
	default:
		notAllowed(w, r)
		return
	}
}

// fileHandler is the core route for modifying the contents of the filesystem.
func fileHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST", "DELETE", "PUT":
		genericFileHandler(path.Join(dataRepo, branchParam(r)), w, r)
	case "GET":
		genericFileHandler(path.Join(dataRepo, commitParam(r)), w, r)
	default:
		notAllowed(w, r)
		return
	}
}

// commitHandler creates a snapshot of outstanding changes.
func commitHandler(w http.ResponseWriter, r *http.Request) {
	url := strings.Split(r.URL.Path, "/")
	// url looks like [, commit, <commit>, file, <file>]
	if len(url) > 3 && url[3] == "file" {
		genericFileHandler(path.Join(dataRepo, url[2]), w, r)
		return
	}

	switch r.Method {
	case "GET":
		commits, err := btrfs.ReadDir(dataRepo)
		if err != nil {
			internalError(err, w, r)
			return
		}

		for _, ci := range commits {
			if uuid.Parse(ci.Name()) != nil {
				fmt.Fprintf(w, "%s    %s\n", ci.Name(), ci.ModTime().Format("2006-01-02T15:04:05.999999-07:00"))
			}
		}
	case "POST":
		var commit string
		if commit = r.URL.Query().Get("commit"); commit == "" {
			commit = uuid.New()
		}
		err := btrfs.Commit(dataRepo, commit, branchParam(r))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Print(err)
			return
		}

		if materializeParam(r) == "true" {
			go func() {
				err := mapreduce.Materialize(dataRepo, branchParam(r), commit, compRepo, jobDir, shard, modulos)
				if err != nil {
					log.Print(err)
				}
			}()
		}

		fmt.Fprint(w, commit)
	default:
		notAllowed(w, r)
		return
	}
}

// branchHandler creates a new branch from commit.
func branchHandler(w http.ResponseWriter, r *http.Request) {
	url := strings.Split(r.URL.Path, "/")
	// url looks like [, commit, <commit>, file, <file>]
	if len(url) > 3 && url[3] == "file" {
		genericFileHandler(path.Join(dataRepo, url[2]), w, r)
		return
	}
	switch r.Method {
	case "GET":
		branches, err := btrfs.ReadDir(dataRepo)
		if err != nil {
			internalError(err, w, r)
			return
		}

		for _, bi := range branches {
			if uuid.Parse(bi.Name()) == nil {
				fmt.Fprintf(w, "%s    %s\n", bi.Name(), bi.ModTime().Format("2006-01-02T15:04:05.999999-07:00"))
			}
		}
	case "POST":
		if err := btrfs.Branch(dataRepo, commitParam(r), branchParam(r)); err != nil {
			internalError(err, w, r)
			return
		}
		fmt.Fprintf(w, "Created branch. (%s) -> %s.\n", commitParam(r), branchParam(r))
	default:
		notAllowed(w, r)
		return
	}
}

func jobHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("URL in job handler:\n", r.URL)
	url := strings.Split(r.URL.Path, "/")
	// url looks like [, job, <job>, file, <file>]
	switch r.Method {
	case "GET":
		if len(url) < 3 || url[3] != "file" {
			http.Error(w, "GET url must contain 'file'", http.StatusBadRequest)
			return
		}

		if hasBranch(r) {
			if err := mapreduce.WaitJob(compRepo, branchParam(r), commitParam(r), url[2]); err != nil {
				internalError(err, w, r)
				return
			}
			genericFileHandler(path.Join(compRepo, branchParam(r), url[2]), w, r)
		} else {
			genericFileHandler(path.Join(compRepo, commitParam(r), url[2]), w, r)
		}
		return
	case "POST":
		r.URL.Path = path.Join("/file", jobDir, url[2])
		log.Print("URL with reset path:\n", r.URL)
		genericFileHandler(path.Join(dataRepo, branchParam(r)), w, r)
	default:
		notAllowed(w, r)
		return
	}
}

// masterMux creates a multiplexer for a Master writing to the passed in FS.
func masterMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/branch", branchHandler)
	mux.HandleFunc("/commit", commitHandler)
	mux.HandleFunc("/file/", fileHandler)
	mux.HandleFunc("/job/", jobHandler)
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, "pong") })

	return mux
}

// RunServer runs a master server listening on port 80
func runServer() {
	http.ListenAndServe(":80", masterMux())
}

func main() {
	log.SetFlags(log.Lshortfile)
	parseArgs()
	if err := os.MkdirAll("/var/lib/pfs/log", 0777); err != nil {
		log.Fatal(err)
	}
	logF, err := os.Create(path.Join("/var/lib/pfs/log", logFile))
	if err != nil {
		log.Fatal(err)
	}
	defer logF.Close()
	log.SetOutput(logF)
	if err := btrfs.Ensure(dataRepo); err != nil {
		log.Fatal(err)
	}
	if err := btrfs.Ensure(compRepo); err != nil {
		log.Fatal(err)
	}
	log.Print("Listening on port 80...")
	log.Printf("dataRepo: %s, compRepo: %s.", dataRepo, compRepo)
	runServer()
}
