package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

	"code.google.com/p/go-uuid/uuid"
	"github.com/pachyderm/pfs/lib/btrfs"
	"github.com/pachyderm/pfs/lib/mapreduce"
	"github.com/pachyderm/pfs/lib/pipeline"
)

var jobDir string = "job"
var pipelineDir string = "pipeline"

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

func cat(w http.ResponseWriter, name string) {
	exists, err := btrfs.FileExists(name)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Print(err)
	}
	if !exists {
		http.Error(w, "404 page not found", 404)
		return
	}

	f, err := btrfs.Open(name)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Print(err)
		return
	}
	defer f.Close()

	if _, err := io.Copy(w, f); err != nil {
		http.Error(w, err.Error(), 500)
		log.Print(err)
		return
	}
}

type Shard struct {
	url                string
	dataRepo, compRepo string
	pipelinePrefix     string
	shard, modulos     uint64
	runners            map[string]*pipeline.Runner
	guard              sync.Mutex
}

func ShardFromArgs() (*Shard, error) {
	s_m := strings.Split(os.Args[1], "-")
	shard, err := strconv.ParseUint(s_m[0], 10, 64)
	if err != nil {
		return nil, err
	}
	modulos, err := strconv.ParseUint(s_m[1], 10, 64)
	if err != nil {
		return nil, err
	}
	return &Shard{
		url:      "http://" + os.Args[2],
		dataRepo: "data-" + os.Args[1],
		compRepo: "comp-" + os.Args[1],
		shard:    shard,
		modulos:  modulos,
		runners:  make(map[string]*pipeline.Runner),
	}, nil
}

func NewShard(dataRepo, compRepo, pipelinePrefix string, shard, modulos uint64) *Shard {
	return &Shard{
		dataRepo:       dataRepo,
		compRepo:       compRepo,
		pipelinePrefix: pipelinePrefix,
		shard:          shard,
		modulos:        modulos,
		runners:        make(map[string]*pipeline.Runner),
	}
}

func (s *Shard) EnsureRepos() error {
	if err := btrfs.Ensure(s.dataRepo); err != nil {
		return err
	}
	if err := btrfs.Ensure(s.compRepo); err != nil {
		return err
	}
	return nil
}

// genericFileHandler serves files from fs. It's used after branch and commit
// info have already been extracted and ignores those aspects of the URL.
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
		fmt.Fprintf(w, "Created %s, size: %d.\n", path.Join(url[fileStart:]...), size)
	} else if r.Method == "PUT" {
		btrfs.MkdirAll(path.Dir(file))
		size, err := btrfs.CopyFile(file, r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}
		fmt.Fprintf(w, "Created %s, size: %d.\n", path.Join(url[fileStart:]...), size)
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
func (s *Shard) FileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" || r.Method == "DELETE" || r.Method == "PUT" {
		genericFileHandler(path.Join(s.dataRepo, branchParam(r)), w, r)
	} else if r.Method == "GET" {
		genericFileHandler(path.Join(s.dataRepo, commitParam(r)), w, r)
	} else {
		http.Error(w, "Invalid method.", 405)
	}
}

// CommitHandler creates a snapshot of outstanding changes.
func (s *Shard) CommitHandler(w http.ResponseWriter, r *http.Request) {
	url := strings.Split(r.URL.Path, "/")
	// url looks like [, commit, <commit>, file, <file>]
	if len(url) > 3 && url[3] == "file" {
		genericFileHandler(path.Join(s.dataRepo, url[2]), w, r)
		return
	}
	if r.Method == "GET" {
		encoder := json.NewEncoder(w)
		btrfs.Commits(s.dataRepo, "", btrfs.Desc, func(c btrfs.CommitInfo) error {
			isReadOnly, err := btrfs.IsReadOnly(path.Join(s.dataRepo, c.Path))
			if err != nil {
				log.Print(err)
				return err
			}
			if isReadOnly {
				fi, err := btrfs.Stat(path.Join(s.dataRepo, c.Path))
				if err != nil {
					log.Print(err)
					return err
				}
				err = encoder.Encode(CommitMsg{Name: fi.Name(), TStamp: fi.ModTime().Format("2006-01-02T15:04:05.999999-07:00")})
				if err != nil {
					log.Print(err)
					return err
				}
			}
			return nil
		})
	} else if r.Method == "POST" && r.ContentLength == 0 {
		// Create a commit from local data
		var commit string
		if commit = r.URL.Query().Get("commit"); commit == "" {
			commit = uuid.New()
		}
		err := btrfs.Commit(s.dataRepo, commit, branchParam(r))
		if err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}

		// We lock the guard so that we can remove the oldRunner from the map
		// and add the newRunner in.
		s.guard.Lock()
		oldRunner, ok := s.runners[branchParam(r)]
		newRunner := pipeline.NewRunner("pipeline", s.dataRepo, s.pipelinePrefix, commit, branchParam(r))
		s.runners[branchParam(r)] = newRunner
		s.guard.Unlock()
		go func() {
			// cancel oldRunner if it exists
			if ok {
				err := oldRunner.Cancel()
				if err != nil {
					log.Print(err)
					return
				}
			}
			err := newRunner.Run()
			if err != nil {
				log.Print(err)
			}
		}()

		if materializeParam(r) == "true" {
			go func() {
				err := mapreduce.Materialize(s.dataRepo, branchParam(r), commit,
					s.compRepo, jobDir, s.shard, s.modulos)
				if err != nil {
					log.Print(err)
				}
			}()
		}
		// Sync changes to peers
		go s.SyncToPeers()
		fmt.Fprintf(w, "%s\n", commit)
	} else if r.Method == "POST" {
		// Commit being pushed via a diff
		replica := btrfs.NewLocalReplica(s.dataRepo)
		if err := replica.Push(r.Body); err != nil {
			http.Error(w, err.Error(), 500)
			log.Print(err)
			return
		}
	} else {
		http.Error(w, "Unsupported method.", http.StatusMethodNotAllowed)
		log.Printf("Unsupported method %s in request to %s.", r.Method, r.URL.String())
		return
	}
}

// BranchHandler creates a new branch from commit.
func (s *Shard) BranchHandler(w http.ResponseWriter, r *http.Request) {
	url := strings.Split(r.URL.Path, "/")
	// url looks like [, commit, <commit>, file, <file>]
	if len(url) > 3 && url[3] == "file" {
		genericFileHandler(path.Join(s.dataRepo, url[2]), w, r)
		return
	}
	if r.Method == "GET" {
		encoder := json.NewEncoder(w)
		btrfs.Commits(s.dataRepo, "", btrfs.Desc, func(c btrfs.CommitInfo) error {
			isReadOnly, err := btrfs.IsReadOnly(path.Join(s.dataRepo, c.Path))
			if err != nil {
				return err
			}
			if !isReadOnly {
				fi, err := btrfs.Stat(path.Join(s.dataRepo, c.Path))
				if err != nil {
					return err
				}
				err = encoder.Encode(BranchMsg{Name: fi.Name(), TStamp: fi.ModTime().Format("2006-01-02T15:04:05.999999-07:00")})
				if err != nil {
					log.Print(err)
					return err
				}
			}
			return nil
		})
	} else if r.Method == "POST" {
		if err := btrfs.Branch(s.dataRepo, commitParam(r), branchParam(r)); err != nil {
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

func (s *Shard) JobHandler(w http.ResponseWriter, r *http.Request) {
	url := strings.Split(r.URL.Path, "/")
	if r.Method == "GET" && len(url) > 3 && url[3] == "file" {
		// url looks like [, job, <job>, file, <file>]
		if hasBranch(r) {
			err := mapreduce.WaitJob(s.compRepo, branchParam(r), commitParam(r), url[2])
			if err != nil {
				http.Error(w, err.Error(), 500)
				log.Print(err)
				return
			}
			genericFileHandler(path.Join(s.compRepo, branchParam(r), url[2]), w, r)
		} else {
			genericFileHandler(path.Join(s.compRepo, commitParam(r), url[2]), w, r)
		}
		return
	} else if r.Method == "POST" {
		r.URL.Path = path.Join("/file", jobDir, url[2])
		log.Print("URL with reset path:\n", r.URL)
		genericFileHandler(path.Join(s.dataRepo, branchParam(r)), w, r)
	} else {
		http.Error(w, "Invalid method.", 405)
		log.Print("Invalid method %s.", r.Method)
		return
	}
}

func (s *Shard) PipelineHandler(w http.ResponseWriter, r *http.Request) {
	url := strings.Split(r.URL.Path, "/")
	if r.Method == "POST" {
		r.URL.Path = path.Join("/file", pipelineDir, url[2])
		genericFileHandler(path.Join(s.dataRepo, branchParam(r)), w, r)
	} else {
		http.Error(w, "Invalid method.", 405)
		log.Print("Invalid method %s.", r.Method)
		return
	}
}

func (s *Shard) PullHandler(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	mpw := multipart.NewWriter(w)
	defer mpw.Close()
	cb := NewMultiPartCommitBrancher(mpw)
	w.Header().Add("Boundary", mpw.Boundary())
	localReplica := btrfs.NewLocalReplica(s.dataRepo)
	err := localReplica.Pull(from, cb)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Print(err)
		return
	}
}

// ShardMux creates a multiplexer for a Shard writing to the passed in FS.
func (s *Shard) ShardMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/branch", s.BranchHandler)
	mux.HandleFunc("/commit", s.CommitHandler)
	mux.HandleFunc("/file/", s.FileHandler)
	mux.HandleFunc("/job/", s.JobHandler)
	mux.HandleFunc("/pipeline/", s.PipelineHandler)
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "pong\n") })
	mux.HandleFunc("/pull", s.PullHandler)

	return mux
}

// RunServer runs a shard server listening on port 80
func (s *Shard) RunServer() {
	http.ListenAndServe(":80", s.ShardMux())
}

func main() {
	log.SetFlags(log.Lshortfile)
	if err := os.MkdirAll("/var/lib/pfs/log", 0777); err != nil {
		log.Fatal(err)
	}
	logF, err := os.Create(path.Join("/var/lib/pfs/log", "log-"+os.Args[1]))
	if err != nil {
		log.Fatal(err)
	}
	defer logF.Close()
	log.SetOutput(logF)

	s, err := ShardFromArgs()
	if err != nil {
		log.Fatal(err)
	}
	if err := s.EnsureRepos(); err != nil {
		log.Fatal(err)
	}

	log.Print("Listening on port 80...")
	log.Printf("dataRepo: %s, compRepo: %s.", s.dataRepo, s.compRepo)
	cancel := make(chan struct{})
	defer close(cancel)
	go s.FillRole(cancel)
	s.RunServer()
}
