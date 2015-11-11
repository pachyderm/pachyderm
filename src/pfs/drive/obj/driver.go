package obj

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"io"
	"path"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pkg/obj"
)

type driver struct {
	objClient      obj.Client
	cacheDir       string
	namespace      string
	commits        map[pfs.Repo]map[pfs.Commit]map[uint64]*drive.Commit
	startedCommits map[pfs.Repo]map[pfs.Commit]map[uint64]*drive.Commit
	lock           sync.RWMutex
}

func newDriver(objClient obj.Client, cacheDir string, namespace string) (drive.Driver, error) {
	return &driver{
		objClient,
		cacheDir,
		namespace,
		make(map[pfs.Repo]map[pfs.Commit]map[uint64]*drive.Commit),
		make(map[pfs.Repo]map[pfs.Commit]map[uint64]*drive.Commit),
		sync.RWMutex{},
	}, nil
}

func (d *driver) CreateRepo(repo *pfs.Repo) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.commits[*repo] = make(map[pfs.Commit]map[uint64]*drive.Commit)
	d.startedCommits[*repo] = make(map[pfs.Commit]map[uint64]*drive.Commit)
	return nil
}

func (d *driver) InspectRepo(repo *pfs.Repo, shard uint64) (*pfs.RepoInfo, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	if _, ok := d.commits[*repo]; !ok {
		return nil, fmt.Errorf("repo %s not found", repo.Name)
	}
	return &pfs.RepoInfo{
		Repo: repo,
	}, nil
}

func (d *driver) ListRepo(shard uint64) ([]*pfs.RepoInfo, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	var result []*pfs.RepoInfo
	for repo := range d.commits {
		repoInfo, err := d.InspectRepo(&repo, shard)
		if err != nil {
			return nil, err
		}
		result = append(result, repoInfo)
	}
	return result, nil
}

func (d *driver) DeleteRepo(repo *pfs.Repo, shards map[uint64]bool) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	delete(d.commits, *repo)
	delete(d.startedCommits, *repo)
	d.objClient.Delete(d.repoPath(repo))
	return nil
}

func (d *driver) StartCommit(parent *pfs.Commit, commit *pfs.Commit, shards map[uint64]bool) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	if err := d.repoExists(commit.Repo); err != nil {
		return err
	}
	d.startedCommits[*commit.Repo][*commit] = make(map[uint64]*drive.Commit)
	for shard := range shards {
		d.startedCommits[*commit.Repo][*commit][shard] = &drive.Commit{
			Parent: parent,
		}
	}
	return nil
}

func (d *driver) FinishCommit(commit *pfs.Commit, shards map[uint64]bool) error {
	// closure so we can defer Unlock
	if err := func() error {
		d.lock.Lock()
		defer d.lock.Unlock()
		if err := d.repoExists(commit.Repo); err != nil {
			return err
		}
		d.commits[*commit.Repo][*commit] = d.startedCommits[*commit.Repo][*commit]
		delete(d.startedCommits[*commit.Repo], *commit)
	}(); err != nil {
		return err
	}
	var wg sync.WaitGroup
	var loopErr error
	for shard := range shards {
		shard := shard
		wg.Add(1)
		go func() {
			defer wg.Done()
			commit := d.commits[*commit.Repo][*commit][shard]
			data, err := proto.Marshal(commit)
			if err != nil && loopErr == nil {
				loopErr = err
				return
			}
			sha := sha512.Sum512(data)
			hash := base64.URLEncoding.EncodeToString(sha[:])
			writer, err := d.objClient.Writer(hash)
			if err != nil && loopErr == nil {
				loopErr = err
				return
			}
			if _, err := writer.Write(data); err != nil && loopErr == nil {
				loopErr = err
				return
			}
		}()
	}
	wg.Wait()
	return loopErr
}

func (d *driver) InspectCommit(commit *pfs.Commit, shards map[uint64]bool) (*pfs.CommitInfo, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	result := &pfs.CommitInfo{
		Commit: commit,
	}
	var commits map[uint64]*drive.Commit
	if err := d.commitExists(commit); err != nil {
		if err := d.startedCommitExists(commit); err != nil {
			return nil, err
		}
		result.CommitType = pfs.CommitType_COMMIT_TYPE_WRITE
		commits = d.startedCommits[*commit.Repo][*commit]
	} else {
		result.CommitType = pfs.CommitType_COMMIT_TYPE_READ
		commits = d.commits[*commit.Repo][*commit]
	}
	for _, commit := range shards {
		result.Parent = commit.Parent
	}
	return result, nil
}

func (d *driver) ListCommit(repo *pfs.Repo, from *pfs.Commit, shards map[uint64]bool) ([]*pfs.CommitInfo, error) {
	var commits map[pfs.Commit]bool
	func() {
		d.lock.RLock()
		defer d.lock.RUnlock()
		for commit := range d.commits[*repo] {
			commits[*commit] = true
		}
		for commit := range d.startCommits[*repo] {
			commits[*commit] = true
		}
	}()
	var result []*pfs.CommitInfo
	for commit := range commits {
		commitInfo, err := inspectCommit(commit, shards)
		if err != nil {
			return nil, err
		}
		result = append(result, commitInfo)
	}
	return result, nil
}

func (d *driver) DeleteCommit(commit *pfs.Commit, shards map[uint64]bool) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	if err := d.commitExists(commit); err != nil {
		if err := d.startedCommitExists(commit); err != nil {
			return err
		}
		delete(d.startCommits, *commit)
	} else {
		delete(d.commits, *commit)
	}
}

func (d *driver) PutBlock(file *pfs.File, block *pfs.Block, shard uint64, reader io.Reader) error {
	return nil
}

func (d *driver) GetBlock(block *pfs.Block, shard uint64) (drive.ReaderAtCloser, error) {
	return nil, nil
}

func (d *driver) InspectBlock(block *pfs.Block, shard uint64) (*pfs.BlockInfo, error) {
	return nil, nil
}

func (d *driver) ListBlock(shard uint64) ([]*pfs.BlockInfo, error) {
	return nil, nil
}

func (d *driver) PutFile(file *pfs.File, shard uint64, offset int64, reader io.Reader) error {
	return nil
}

func (d *driver) MakeDirectory(file *pfs.File, shards map[uint64]bool) error {
	return nil
}

func (d *driver) GetFile(file *pfs.File, shard uint64) (drive.ReaderAtCloser, error) {
	return nil, nil
}

func (d *driver) InspectFile(file *pfs.File, shard uint64) (*pfs.FileInfo, error) {
	return nil, nil
}

func (d *driver) ListFile(file *pfs.File, shard uint64) ([]*pfs.FileInfo, error) {
	return nil, nil
}

func (d *driver) ListChange(file *pfs.File, from *pfs.Commit, shard uint64) ([]*pfs.Change, error) {
	return nil, nil
}

func (d *driver) DeleteFile(file *pfs.File, shard uint64) error {
	return nil
}

func (d *driver) PullDiff(commit *pfs.Commit, shard uint64, diff io.Writer) error {
	return nil
}

func (d *driver) PushDiff(commit *pfs.Commit, shard uint64, diff io.Reader) error {
	return nil
}

func (d *driver) repoPath(repo *pfs.Repo) string {
	return path.Join(d.namespace, repo.Name)
}

func (d *driver) commitPath(commit *pfs.Commit) string {
	return path.Join(d.repoPath(commit.Repo), commit.Id)
}

func (d *driver) shardPath(commit *pfs.Commit, shard uint64) string {
	return path.Join(d.commitPath(commit), fmt.Sprint(shard))
}

func (d *driver) repoExists(repo *pfs.Repo) error {
	if _, ok := d.commits[*repo]; !ok {
		return fmt.Errorf("repo %s not found", repo.Name)
	}
	return nil
}

func (d *driver) commitExists(commit *pfs.Commit) error {
	if err := repoExists(*commit.Repo); err != nil {
		return err
	}
	if _, ok := d.commits[*commit.Repo][*commit]; !ok {
		return fmt.Errorf("commit %s/%s not found", commit.Repo.Name, commit.Id)
	}
}

func (d *driver) startedCommitExists(commit *pfs.Commit) error {
	if err := repoExists(*commit.Repo); err != nil {
		return err
	}
	if _, ok := d.startCommits[*commit.Repo][*commit]; !ok {
		return fmt.Errorf("commit %s/%s not found", commit.Repo.Name, commit.Id)
	}
}
