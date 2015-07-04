package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-etcd/etcd"
	"github.com/pachyderm/pachyderm/src/btrfs"
	"github.com/pachyderm/pachyderm/src/etcache"
	"github.com/pachyderm/pachyderm/src/log"
	"github.com/pachyderm/pachyderm/src/pipeline"
	"github.com/pachyderm/pachyderm/src/route"
	"github.com/satori/go.uuid"
)

const (
	pipelineDir = "pipeline"
)

var (
	ErrInvalidObject = errors.New("pfs: invalid object")
)

type shard struct {
	url                string
	dataRepo, compRepo string
	pipelinePrefix     string
	shard, modulos     uint64
	shardStr           string
	runners            map[string]*pipeline.Runner
	guard              *sync.Mutex
	cache              etcache.Cache
}

func newShard(
	url string,
	dataRepo string,
	compRepo string,
	pipelinePrefix string,
	shardNum uint64,
	modulos uint64,
	cache etcache.Cache,
) *shard {
	return &shard{
		url,
		dataRepo,
		compRepo,
		pipelinePrefix,
		shardNum,
		modulos,
		fmt.Sprint(shardNum, "-", modulos),
		make(map[string]*pipeline.Runner),
		&sync.Mutex{},
		cache,
	}
}

func (s *shard) EnsureRepos() error {
	if err := btrfs.Ensure(s.dataRepo); err != nil {
		return err
	}
	if err := btrfs.Ensure(s.compRepo); err != nil {
		return err
	}
	return nil
}

func (s *shard) SyncFromPeers() error {
	peers, err := s.peers()
	if err != nil {
		return err
	}

	err = syncFrom(s.dataRepo, peers)
	if err != nil {
		return err
	}

	return nil
}

func (s *shard) SyncToPeers() error {
	peers, err := s.peers()
	if err != nil {
		return err
	}

	err = syncTo(s.dataRepo, peers)
	if err != nil {
		return err
	}

	return nil
}

// FillRole attempts to find a role in the cluster. Once on is found it
// prepares the local storage for the role and announces the shard to the rest
// of the cluster. This function will loop until `cancel` is closed.
func (s *shard) FillRole(cancel chan bool) error {
	shard := fmt.Sprintf("%d-%d", s.shard, s.modulos)
	masterKey := path.Join("/pfs/master", shard)
	replicaDir := path.Join("/pfs/replica", shard)

	amMaster := false //true if we're master
	replicaKey := ""
	for {
		client := etcd.NewClient([]string{"http://172.17.42.1:4001", "http://10.1.42.1:4001"})
		// First we attempt to become the master for this shard
		if !amMaster {
			// We're not master, so we attempt to claim it, this will error if
			// another shard is already master
			backfillingKey := "[backfilling]" + s.url
			if _, err := client.Create(masterKey, backfillingKey, 5*60); err == nil {
				// no error means we succesfully claimed master
				_ = s.SyncFromPeers()
				// Attempt to finalize ourselves as master
				_, _ = client.CompareAndSwap(masterKey, s.url, 60, backfillingKey, 0)
				if err == nil {
					// no error means that we succusfully announced ourselves as master
					// Sync the new data we pulled to peers
					go s.SyncToPeers()
					//Record that we're master
					amMaster = true
				}
			}
		} else {
			// We're already master, renew our lease
			_, err := client.CompareAndSwap(masterKey, s.url, 60, s.url, 0)
			if err != nil { // error means we failed to reclaim master
				amMaster = false
			}
		}

		// We didn't claim master, so we add ourselves as replica instead.
		if replicaKey == "" {
			if resp, err := client.CreateInOrder(replicaDir, s.url, 60); err == nil {
				replicaKey = resp.Node.Key
				// Get ourselves up to date
				go s.SyncFromPeers()
			}
		} else {
			_, err := client.CompareAndSwap(replicaKey, s.url, 60, s.url, 0)
			if err != nil {
				replicaKey = ""
			}
		}

		select {
		case <-time.After(time.Second * 45):
			continue
		case <-cancel:
			break
		}
	}
}

// FileHandler is the core route for modifying the contents of the fileystem.
func (s *shard) fileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" || r.Method == "DELETE" || r.Method == "PUT" {
		genericFileHandler(path.Join(s.dataRepo, branchParam(r)), w, r)
	} else if r.Method == "GET" {
		genericFileHandler(path.Join(s.dataRepo, commitParam(r)), w, r)
	} else {
		http.Error(w, "Invalid method.", 405)
	}
}

// CommitHandler creates a snapshot of outstanding changes.
func (s *shard) commitHandler(w http.ResponseWriter, r *http.Request) {
	url := strings.Split(r.URL.Path, "/")
	// url looks like [, commit, <commit>, file, <file>]
	if len(url) > 3 && url[3] == "file" {
		genericFileHandler(path.Join(s.dataRepo, url[2]), w, r)
		return
	}
	if r.Method == "GET" {
		encoder := json.NewEncoder(w)
		btrfs.Commits(s.dataRepo, "", btrfs.Desc, func(name string) error {
			isReadOnly, err := btrfs.IsCommit(path.Join(s.dataRepo, name))
			if err != nil {
				return err
			}
			if isReadOnly {
				fi, err := btrfs.Stat(path.Join(s.dataRepo, name))
				if err != nil {
					return err
				}
				err = encoder.Encode(commitMsg{Name: fi.Name(), TStamp: fi.ModTime().Format("2006-01-02T15:04:05.999999-07:00")})
				if err != nil {
					return err
				}
			}
			return nil
		})
	} else if r.Method == "POST" && r.ContentLength == 0 {
		// Create a commit from local data
		var commit string
		if commit = r.URL.Query().Get("commit"); commit == "" {
			commit = uuid.NewV4().String()
		}
		err := btrfs.Commit(s.dataRepo, commit, branchParam(r))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// We lock the guard so that we can remove the oldRunner from the map
		// and add the newRunner in.
		s.guard.Lock()
		oldRunner, ok := s.runners[branchParam(r)]
		newRunner := pipeline.NewRunner("pipeline", s.dataRepo, s.pipelinePrefix, commit, branchParam(r), s.shardStr, s.cache)
		s.runners[branchParam(r)] = newRunner
		s.guard.Unlock()
		go func() {
			// cancel oldRunner if it exists
			if ok {
				err := oldRunner.Cancel()
				if err != nil {
					log.Print(err)
				}
			}
			err := newRunner.Run()
			if err != nil {
				log.Print(err)
			}
		}()
		go s.SyncToPeers()
		fmt.Fprintf(w, "%s\n", commit)
	} else if r.Method == "POST" {
		// Commit being pushed via a diff
		replica := btrfs.NewLocalReplica(s.dataRepo)
		if err := replica.Push(r.Body); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "Unsupported method.", http.StatusMethodNotAllowed)
		log.Printf("Unsupported method %s in request to %s.", r.Method, r.URL.String())
		return
	}
}

// BranchHandler creates a new branch from commit.
func (s *shard) branchHandler(w http.ResponseWriter, r *http.Request) {
	url := strings.Split(r.URL.Path, "/")
	// url looks like [, commit, <commit>, file, <file>]
	if len(url) > 3 && url[3] == "file" {
		genericFileHandler(path.Join(s.dataRepo, url[2]), w, r)
		return
	}
	if r.Method == "GET" {
		encoder := json.NewEncoder(w)
		btrfs.Commits(s.dataRepo, "", btrfs.Desc, func(name string) error {
			isReadOnly, err := btrfs.IsCommit(path.Join(s.dataRepo, name))
			if err != nil {
				return err
			}
			if !isReadOnly {
				fi, err := btrfs.Stat(path.Join(s.dataRepo, name))
				if err != nil {
					return err
				}
				err = encoder.Encode(branchMsg{Name: fi.Name(), TStamp: fi.ModTime().Format("2006-01-02T15:04:05.999999-07:00")})
				if err != nil {
					return err
				}
			}
			return nil
		})
	} else if r.Method == "POST" {
		if err := btrfs.Branch(s.dataRepo, commitParam(r), branchParam(r)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Created branch. (%s) -> %s.\n", commitParam(r), branchParam(r))
	} else {
		http.Error(w, "Invalid method.", 405)
		log.Printf("Invalid method %s.", r.Method)
		return
	}
}

func (s *shard) pipelineHandler(w http.ResponseWriter, r *http.Request) {
	url := strings.Split(r.URL.Path, "/")
	if r.Method == "GET" && len(url) > 3 && url[3] == "file" {
		// First wait for the commit to show up
		err := pipeline.WaitPipeline(s.pipelinePrefix, url[2], commitParam(r))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// url looks like [, pipeline, <pipeline>, file, <file>]
		genericFileHandler(path.Join(s.pipelinePrefix, url[2], commitParam(r)), w, r)
		return
	} else if r.Method == "POST" {
		r.URL.Path = path.Join("/file", pipelineDir, url[2])
		genericFileHandler(path.Join(s.dataRepo, branchParam(r)), w, r)
	} else {
		http.Error(w, "Invalid method.", 405)
		log.Printf("Invalid method %s.", r.Method)
		return
	}
}

func (s *shard) pullHandler(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	mpw := multipart.NewWriter(w)
	defer mpw.Close()
	cb := newMultipartPusher(mpw)
	w.Header().Add("Boundary", mpw.Boundary())
	localReplica := btrfs.NewLocalReplica(s.dataRepo)
	if err := localReplica.Pull(from, cb); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *shard) peers() ([]string, error) {
	var peers []string
	client := etcd.NewClient([]string{"http://172.17.42.1:4001", "http://10.1.42.1:4001"})
	resp, err := client.Get(fmt.Sprintf("/pfs/replica/%d-%d", s.shard, s.modulos), false, true)
	if err != nil {
		return peers, err
	}
	for _, node := range resp.Node.Nodes {
		if node.Value != s.url {
			peers = append(peers, node.Value)
		}
	}
	return peers, err
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
		files, err := btrfs.Glob(file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		switch len(files) {
		case 0:
			http.Error(w, "404 page not found", 404)
			return
		case 1:
			http.ServeFile(w, r, btrfs.FilePath(files[0]))
		default:
			writer := multipart.NewWriter(w)
			defer writer.Close()
			w.Header().Add("Boundary", writer.Boundary())
			for _, file := range files {
				info, err := btrfs.Stat(file)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				if info.IsDir() {
					// We don't do anything with directories.
					continue
				}
				name := strings.TrimPrefix(file, "/"+fs+"/")
				if shardParam(r) != "" {
					// We have a shard param, check if the file matches the shard.
					match, err := route.Match(name, shardParam(r))
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					if !match {
						continue
					}
				}
				fWriter, err := writer.CreateFormFile(name, name)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				err = rawCat(fWriter, file)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}
	} else if r.Method == "POST" {
		btrfs.MkdirAll(path.Dir(file))
		size, err := btrfs.CreateFromReader(file, r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Created %s, size: %d.\n", path.Join(url[fileStart:]...), size)
	} else if r.Method == "PUT" {
		btrfs.MkdirAll(path.Dir(file))
		size, err := btrfs.CopyFile(file, r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Created %s, size: %d.\n", path.Join(url[fileStart:]...), size)
	} else if r.Method == "DELETE" {
		if err := btrfs.Remove(file); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Deleted %s.\n", file)
	}
}

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

func shardParam(r *http.Request) string {
	return r.URL.Query().Get("shard")
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

func rawCat(w io.Writer, name string) error {
	f, err := btrfs.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(w, f); err != nil {
		return err
	}
	return nil
}

// syncTo syncs the contents in p to all of the shards in urls
// returns the first error if there are multiple
func syncTo(dataRepo string, urls []string) error {
	var errs []error
	var lock sync.Mutex
	addErr := func(err error) {
		lock.Lock()
		errs = append(errs, err)
		lock.Unlock()
	}
	lr := btrfs.NewLocalReplica(dataRepo)
	var wg sync.WaitGroup
	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			sr := newShardReplica(url)
			from, err := sr.From()
			if err != nil {
				addErr(err)
			}
			err = lr.Pull(from, sr)
			if err != nil {
				addErr(err)
			}
		}(url)
	}
	wg.Wait()
	if len(errs) != 0 {
		return errs[0]
	}
	return nil
}

// syncFrom syncs from the most up to date replica in urls
// returns the first error if ALL urls error.
func syncFrom(dataRepo string, urls []string) error {
	if len(urls) == 0 {
		return nil
	}
	errCount := 0
	for _, url := range urls {
		if err := syncFromUrl(dataRepo, url); err != nil {
			errCount++
		}
	}
	if errCount == len(urls) {
		return fmt.Errorf("all urls %v had errors", urls)
	}
	return nil
}

func syncFromUrl(dataRepo string, url string) error {
	sr := newShardReplica(url)
	lr := btrfs.NewLocalReplica(dataRepo)
	from, err := lr.From()
	if err != nil {
		return err
	}
	return sr.Pull(from, lr)
}

func (s *shard) FileGet(name string, commit string) (File, error) {
	info, err := btrfs.Stat(path.Join(s.dataRepo, commit, name))
	if err != nil {
		return File{}, err
	}
	file, err := btrfs.Open(path.Join(s.dataRepo, commit, name))
	if err != nil {
		return File{}, err
	}
	return File{name, info.ModTime(), file}, nil
}
func (s *shard) FileGetAll(name string, commit string) ([]File, error) {
	matches, err := btrfs.Glob(path.Join(s.dataRepo, commit, name))
	if err != nil {
		return nil, err
	}
	var result []File
	for _, match := range matches {
		name := strings.TrimPrefix(match, path.Join(s.dataRepo, commit))
		file, err := s.FileGet(name, commit)
		if err != nil {
			return nil, err
		}
		result = append(result, file)
	}
	return result, nil
}
func (s *shard) FileCreate(name string, content io.Reader, branch string) error {
	filePath := path.Join(s.dataRepo, branch, name)
	btrfs.MkdirAll(path.Dir(filePath))
	_, err := btrfs.CreateFromReader(filePath, content)
	return err
}

func (s *shard) CommitGet(name string) (Commit, error) {
	commitPath := path.Join(s.dataRepo, name)
	isCommit, err := btrfs.IsCommit(commitPath)
	if err != nil {
		return Commit{}, err
	}
	if !isCommit {
		return Commit{}, ErrInvalidObject
	}
	info, err := btrfs.Stat(commitPath)
	if err != nil {
		return Commit{}, err
	}
	return Commit{name, info.ModTime()}, nil
}
func (s *shard) CommitGetAll(name string) ([]Commit, error) {
	matches, err := btrfs.Glob(path.Join(s.dataRepo, name))
	if err != nil {
		return nil, err
	}
	var result []Commit
	for _, match := range matches {
		name := strings.TrimPrefix(match, s.dataRepo)
		commit, err := s.CommitGet(name)
		if err == ErrInvalidObject {
			continue
		}
		if err != nil {
			return nil, err
		}
		result = append(result, commit)
	}
	return result, nil
}
func (s *shard) CommitCreate(name string, branch string) (Commit, error) {
	if err := btrfs.Commit(s.dataRepo, name, branch); err != nil {
		return Commit{}, err
	}
	return s.CommitGet(name)
}

func (s *shard) BranchGet(name string) (Branch, error) {
	branchPath := path.Join(s.dataRepo, name)
	isCommit, err := btrfs.IsCommit(branchPath)
	if err != nil {
		return Branch{}, err
	}
	if isCommit {
		return Branch{}, ErrInvalidObject
	}
	info, err := btrfs.Stat(branchPath)
	if err != nil {
		return Branch{}, err
	}
	return Branch{name, info.ModTime()}, nil
}
func (s *shard) BranchGetAll(name string) ([]Branch, error) {
	matches, err := btrfs.Glob(path.Join(s.dataRepo, name))
	if err != nil {
		return nil, err
	}
	var result []Branch
	for _, match := range matches {
		name := strings.TrimPrefix(match, s.dataRepo)
		branch, err := s.BranchGet(name)
		if err == ErrInvalidObject {
			continue
		}
		if err != nil {
			return nil, err
		}
		result = append(result, branch)
	}
	return result, nil
}
func (s *shard) BranchCreate(name string, commit string) (Branch, error) {
	if err := btrfs.Branch(s.dataRepo, commit, name); err != nil {
		return Branch{}, nil
	}
	return s.BranchGet(name)
}

func (s *shard) From() (string, error) {
	commits, err := s.CommitGetAll("")
	if err != nil {
		return "", err
	}
	if len(commits) == 0 {
		return "", nil
	}
	return commits[0].Name, nil
}
func (s *shard) Push(diff io.Reader) error {
	return btrfs.NewLocalReplica(s.dataRepo).Push(diff)
}
func (s *shard) Pull(from string, p btrfs.Pusher) error {
	return btrfs.NewLocalReplica(s.dataRepo).Pull(from, p)
}
