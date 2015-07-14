package storage

import (
	"errors"
	"fmt"
	"io"
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
)

const (
	pipelineDir = "pipeline"
)

var (
	ErrInvalidObject = errors.New("pfs: invalid object")
	ErrIsDirectory   = errors.New("pfs: is directory")
	ErrOverallocated = errors.New("pfs: overallocated")
	ErrNoShards      = errors.New("pfs: no shards")
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
	roles              []string
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
		nil,
	}
}

func (s *shard) FileGet(name string, commit string) (File, error) {
	return s.fileGet(path.Join(s.dataRepo, commit), name)
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
		if err == ErrIsDirectory {
			continue
		}
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

func (s *shard) CommitList() ([]Commit, error) {
	var result []Commit
	if err := btrfs.Commits(s.dataRepo, "", btrfs.Desc, func(name string) error {
		isCommit, err := btrfs.IsCommit(path.Join(s.dataRepo, name))
		if err != nil {
			return err
		}
		if isCommit {
			fi, err := btrfs.Stat(path.Join(s.dataRepo, name))
			if err != nil {
				return err
			}
			result = append(result, Commit{name, fi.ModTime()})
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *shard) CommitCreate(name string, branch string) (Commit, error) {
	if err := btrfs.Commit(s.dataRepo, name, branch); err != nil {
		return Commit{}, err
	}
	// We lock the guard so that we can remove the oldRunner from the map
	// and add the newRunner in.
	s.guard.Lock()
	oldRunner, ok := s.runners[branch]
	newRunner := pipeline.NewRunner("pipeline", s.dataRepo, s.pipelinePrefix, name, branch, s.shardStr, s.cache)
	s.runners[branch] = newRunner
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
	go s.syncToPeers()
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

func (s *shard) BranchList() ([]Branch, error) {
	var result []Branch
	if err := btrfs.Commits(s.dataRepo, "", btrfs.Desc, func(name string) error {
		isCommit, err := btrfs.IsCommit(path.Join(s.dataRepo, name))
		if err != nil {
			return err
		}
		if !isCommit {
			fi, err := btrfs.Stat(path.Join(s.dataRepo, name))
			if err != nil {
				return err
			}
			result = append(result, Branch{name, fi.ModTime()})
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *shard) BranchCreate(name string, commit string) (Branch, error) {
	if err := btrfs.Branch(s.dataRepo, commit, name); err != nil {
		return Branch{}, nil
	}
	return s.BranchGet(name)
}

func (s *shard) PipelineCreate(name string, content io.Reader, branch string) error {
	pipelinePath := path.Join(s.dataRepo, branch, "pipeline", name)
	btrfs.MkdirAll(path.Dir(pipelinePath))
	_, err := btrfs.CreateFromReader(pipelinePath, content)
	return err
}

func (s *shard) PipelineWait(name string, commit string) error {
	return pipeline.WaitPipeline(s.pipelinePrefix, name, commit)
}

func (s *shard) PipelineFileGet(pipelineName string, fileName string, commit string) (File, error) {
	return s.fileGet(path.Join(s.pipelinePrefix, pipelineName, commit), fileName)
}

func (s *shard) PipelineFileGetAll(pipelineName string, fileName string, commit string, shard string) ([]File, error) {
	matches, err := btrfs.Glob(path.Join(s.pipelinePrefix, pipelineName, commit, fileName))
	if err != nil {
		return nil, err
	}
	var result []File
	for _, match := range matches {
		prefix := path.Join("/", s.pipelinePrefix, pipelineName, commit)
		if !strings.HasSuffix(prefix, "/") {
			prefix = prefix + "/"
		}
		name := strings.TrimPrefix(match, prefix)
		if shard != "" {
			ok, err := route.Match(name, shard)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
		}
		file, err := s.PipelineFileGet(pipelineName, name, commit)
		if err == ErrIsDirectory {
			continue
		}
		if err != nil {
			return nil, err
		}
		result = append(result, file)
	}
	return result, nil
}

func (s *shard) From() (string, error) {
	commits, err := s.CommitList()
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

func (s *shard) EnsureRepos() error {
	if err := btrfs.Ensure(s.dataRepo); err != nil {
		return err
	}
	if err := btrfs.Ensure(s.compRepo); err != nil {
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
		defer client.Close()
		// First we attempt to become the master for this shard
		if !amMaster {
			// We're not master, so we attempt to claim it, this will error if
			// another shard is already master
			backfillingKey := "[backfilling]" + s.url
			if _, err := client.Create(masterKey, backfillingKey, 5*60); err == nil {
				// no error means we succesfully claimed master
				_ = s.syncFromPeers()
				// Attempt to finalize ourselves as master
				_, _ = client.CompareAndSwap(masterKey, s.url, 60, backfillingKey, 0)
				if err == nil {
					// no error means that we succusfully announced ourselves as master
					// Sync the new data we pulled to peers
					go s.syncToPeers()
					//Record that we're master
					amMaster = true
				}
			} else {
				log.Print(err)
			}
		} else {
			// We're already master, renew our lease
			_, err := client.CompareAndSwap(masterKey, s.url, 60, s.url, 0)
			if err != nil { // error means we failed to reclaim master
				log.Print(err)
				amMaster = false
			}
		}

		// We didn't claim master, so we add ourselves as replica instead.
		if replicaKey == "" {
			if resp, err := client.CreateInOrder(replicaDir, s.url, 60); err == nil {
				replicaKey = resp.Node.Key
				// Get ourselves up to date
				go s.syncFromPeers()
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

func (s *shard) FindRole() {
	client := etcd.NewClient([]string{"http://172.17.42.1:4001", "http://10.1.42.1:4001"})
	defer client.Close()
	role := ""
	for {
		// renew our role if we have one
		if role != "" {
			_, _ = client.CompareAndSwap(role, s.url, 60, s.url, 0)
			continue
		}
		// figure out if we should take on a new role
		role, err := s.freeRole()
		if err != nil {
			<-time.After(time.Second * 30)
			role = ""
			continue
		}
		_, err = client.Create(role, s.url, 60)
		if err != nil {
			role = ""
			continue
		}
	}
}

func (s *shard) masters() ([]string, error) {
	result := make([]string, s.modulos)
	client := etcd.NewClient([]string{"http://172.17.42.1:4001", "http://10.1.42.1:4001"})
	defer client.Close()

	response, err := client.Get("/pachyderm.io/pfs", false, true)
	if err == nil {
		for _, node := range response.Node.Nodes {
			// node.Key looks like " /pachyderm.io/pfs/0-5"
			//                      0 1            2   3
			key := strings.Split(node.Key, "/")
			shard, _, err := route.ParseShard(key[3])
			if err != nil {
				return nil, err
			}
			result[shard] = node.Value
		}

	}
	return result, nil
}

func counts(masters []string) map[string]int {
	result := make(map[string]int)
	for _, host := range masters {
		if host != "" {
			result[host] = result[host] + 1
		}
	}
	return result
}

// bestRole returns the best role for us to fill in the cluster right now. If
// no shards are available it returns ErrNoShards.
func (s *shard) freeRole() (string, error) {
	masters, err := s.masters()
	if err != nil {
		return "", err
	}
	// First we check if there's an empty shard
	for i, master := range masters {
		if master == "" {
			return fmt.Sprintf("/pachyderm.io/pfs/%d-%d", i, int(s.modulos)), nil
		}
	}
	return "", ErrNoShards
}

func (s *shard) syncFromPeers() error {
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

func (s *shard) syncToPeers() error {
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

func (s *shard) peers() ([]string, error) {
	var peers []string
	client := etcd.NewClient([]string{"http://172.17.42.1:4001", "http://10.1.42.1:4001"})
	defer client.Close()
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

func (s *shard) fileGet(dir string, name string) (File, error) {
	path := path.Join(dir, name)
	info, err := btrfs.Stat(path)
	if err != nil {
		return File{}, err
	}
	if info.IsDir() {
		return File{}, ErrIsDirectory
	}
	file, err := btrfs.Open(path)
	if err != nil {
		return File{}, err
	}
	return File{
		name,
		info.ModTime(),
		file,
	}, nil
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
