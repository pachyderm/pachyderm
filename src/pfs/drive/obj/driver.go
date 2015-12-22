package obj

import (
	"bufio"
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

var (
	blockSize = 128 * 1024 * 1024 // 128 Megabytes
)

type commitMap map[pfs.Repo]map[pfs.Commit]map[uint64]*drive.DiffInfo

type driver struct {
	objClient      obj.Client
	cacheDir       string
	namespace      string
	commits        commitMap
	startedCommits commitMap
	lock           sync.RWMutex
}

func newDriver(objClient obj.Client, cacheDir string, namespace string) (drive.Driver, error) {
	return &driver{
		objClient,
		cacheDir,
		namespace,
		make(commitMap),
		make(commitMap),
		sync.RWMutex{},
	}, nil
}

func (d *driver) CreateRepo(repo *pfs.Repo) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	if repoToCommits, _ := d.commits.repo(repo); repoToCommits != nil {
		return fmt.Errorf("repo %s exists", repo.Name)
	}
	d.commits[*repo] = make(map[pfs.Commit]map[uint64]*drive.DiffInfo)
	d.startedCommits[*repo] = make(map[pfs.Commit]map[uint64]*drive.DiffInfo)
	return nil
}

func (d *driver) InspectRepo(repo *pfs.Repo, shard uint64) (*pfs.RepoInfo, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	if _, err := d.commits.repo(repo); err != nil {
		return nil, err
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
	if _, err := d.commits.commit(parent); err != nil {
		return err
	}
	repoToCommit, err := d.startedCommits.repo(commit.Repo)
	if err != nil {
		return err
	}
	repoToCommit[*commit] = make(map[uint64]*drive.DiffInfo)
	for shard := range shards {
		repoToCommit[*commit][shard] = &drive.DiffInfo{
			ParentCommit: parent,
		}
	}
	return nil
}

func (d *driver) FinishCommit(commit *pfs.Commit, shards map[uint64]bool) error {
	var shardToCommit map[uint64]*drive.DiffInfo
	// closure so we can defer Unlock
	if err := func() error {
		d.lock.Lock()
		defer d.lock.Unlock()
		var err error
		shardToCommit, err = d.startedCommits.commit(commit)
		if err != nil {
			return err
		}
		delete(d.startedCommits[*commit.Repo], *commit)
		return nil
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
			commit := shardToCommit[shard]
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
	d.lock.Lock()
	defer d.lock.Unlock()
	if loopErr != nil {
		d.startedCommits[*commit.Repo][*commit] = shardToCommit
		return loopErr
	}
	d.commits[*commit.Repo][*commit] = shardToCommit
	return nil
}

func (d *driver) InspectCommit(commit *pfs.Commit, shards map[uint64]bool) (*pfs.CommitInfo, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	result := &pfs.CommitInfo{
		Commit: commit,
	}
	var shardToDiffInfo map[uint64]*drive.DiffInfo
	var err error
	if shardToDiffInfo, err = d.commits.commit(commit); err != nil {
		if shardToDiffInfo, err = d.startedCommits.commit(commit); err != nil {
			return nil, err
		}
		result.CommitType = pfs.CommitType_COMMIT_TYPE_WRITE
	} else {
		result.CommitType = pfs.CommitType_COMMIT_TYPE_READ
	}
	for _, diffInfo := range shardToDiffInfo {
		result.ParentCommit = diffInfo.ParentCommit
		break
	}
	return result, nil
}

func (d *driver) ListCommit(repo *pfs.Repo, from *pfs.Commit, shards map[uint64]bool) ([]*pfs.CommitInfo, error) {
	var commits []*pfs.Commit
	func() {
		d.lock.RLock()
		defer d.lock.RUnlock()
		for commit := range d.commits[*repo] {
			commit := commit
			commits = append(commits, &commit)
		}
		for commit := range d.startedCommits[*repo] {
			commit := commit
			commits = append(commits, &commit)
		}
	}()
	var result []*pfs.CommitInfo
	for _, commit := range commits {
		commitInfo, err := d.InspectCommit(commit, shards)
		if err != nil {
			return nil, err
		}
		result = append(result, commitInfo)
	}
	return result, nil
}

func (d *driver) DeleteCommit(commit *pfs.Commit, shards map[uint64]bool) error {
	return nil
}

func (d *driver) PutBlock(file *pfs.File, block *drive.Block, shard uint64, reader io.Reader) error {
	return nil
}

func (d *driver) putBlock(block *drive.Block) (io.WriteCloser, error) {
	return d.objClient.Writer(d.blockPath(block))
}

func (d *driver) GetBlock(block *drive.Block, shard uint64) (drive.ReaderAtCloser, error) {
	return nil, nil
}

func (d *driver) getBlock(block *drive.Block) (io.ReadCloser, error) {
	return d.objClient.Reader(d.blockPath(block))
}

func (d *driver) InspectBlock(block *drive.Block, shard uint64) (*drive.BlockInfo, error) {
	return nil, nil
}

func (d *driver) ListBlock(shard uint64) ([]*drive.BlockInfo, error) {
	return nil, nil
}

func (d *driver) PutFile(file *pfs.File, shard uint64, offset int64, reader io.Reader) (retErr error) {
	d.lock.RLock()
	_, err := d.startedCommits.shard(file.Commit, shard)
	d.lock.RUnlock()
	if err != nil {
		return err
	}
	var blockRefs []*drive.BlockRef
	scanner := bufio.NewScanner(reader)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if len(data) < blockSize && !atEOF {
			return 0, nil, nil
		}
		if len(data) < blockSize && atEOF {
			if data[len(data)-1] != '\n' {
				return len(data), append(data, '\n'), nil
			}
			return len(data), data, nil
		}
		for i := len(data) - 1; i >= 0; i-- {
			if data[i] == '\n' {
				return i + 1, data[:i+1], nil
			}
		}
		return 0, nil, fmt.Errorf("line too long")
	})
	var wg sync.WaitGroup
	for scanner.Scan() {
		data := scanner.Bytes()
		sha := sha512.Sum512(data)
		block := &drive.Block{
			Hash: base64.URLEncoding.EncodeToString(sha[:]),
		}
		blockRefs = append(blockRefs, &drive.BlockRef{
			Block: block,
			Range: &drive.ByteRange{
				Lower: 0,
				Upper: uint64(len(data)),
			},
		})
		wg.Add(1)
		go func() {
			defer wg.Done()
			writer, err := d.putBlock(block)
			if err != nil && retErr == nil {
				retErr = err
			}
			defer func() {
				if err := writer.Close(); err != nil && retErr == nil {
					retErr = err
				}
			}()
			if _, err := writer.Write(data); err != nil && retErr == nil {
				retErr = err
			}
		}()
	}
	wg.Wait()
	if err := scanner.Err(); err != nil {
		return err
	}
	d.lock.Lock()
	defer d.lock.Unlock()
	changes, err := d.startedCommits.shard(file.Commit, shard)
	if err != nil {
		return err
	}
	existingBlockRefs, ok := changes.Appends[file.Path]
	if ok {
		existingBlockRefs.BlockRef = append(existingBlockRefs.BlockRef, blockRefs...)
	} else {
		changes.Appends[file.Path] = &drive.BlockRefs{
			BlockRef: blockRefs,
		}
	}
	return nil
}

func (d *driver) MakeDirectory(file *pfs.File, shards map[uint64]bool) error {
	return nil
}

func (d *driver) GetFile(file *pfs.File, shard uint64) (drive.ReaderAtCloser, error) {
	var blockRefs []*drive.BlockRef
	commit := file.Commit
	for commit != nil {
		if err := func() error {
			d.lock.RLock()
			defer d.lock.RUnlock()
			diffInfo, err := d.commits.shard(commit, shard)
			if err != nil {
				return err
			}
			blockRefs = append(blockRefs, diffInfo.Appends[file.Path].BlockRef...)
			commit = diffInfo.ParentCommit
			return nil
		}(); err != nil {
			return nil, err
		}
	}
	return &fileReader{
		driver:    d,
		blockRefs: blockRefs,
	}, nil
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
	return path.Join(d.namespace, "index", repo.Name)
}

func (d *driver) commitPath(commit *pfs.Commit) string {
	return path.Join(d.repoPath(commit.Repo), commit.Id)
}

func (d *driver) shardPath(commit *pfs.Commit, shard uint64) string {
	return path.Join(d.commitPath(commit), fmt.Sprint(shard))
}

func (d *driver) blockPath(block *drive.Block) string {
	return path.Join(d.namespace, "block", block.Hash)
}

func (m commitMap) repo(repo *pfs.Repo) (map[pfs.Commit]map[uint64]*drive.DiffInfo, error) {
	result, ok := m[*repo]
	if !ok {
		return nil, fmt.Errorf("repo %s not found", repo.Name)
	}
	return result, nil
}

func (m commitMap) commit(commit *pfs.Commit) (map[uint64]*drive.DiffInfo, error) {
	repoToCommit, err := m.repo(commit.Repo)
	if err != nil {
		return nil, err
	}
	result, ok := repoToCommit[*commit]
	if !ok {
		return nil, fmt.Errorf("commit %s/%s not found", commit.Repo.Name, commit.Id)
	}
	return result, nil
}

func (m commitMap) shard(commit *pfs.Commit, shard uint64) (*drive.DiffInfo, error) {
	shardToChanges, err := m.commit(commit)
	if err != nil {
		return nil, err
	}
	result, ok := shardToChanges[shard]
	if !ok {
		return nil, fmt.Errorf("shard %s/%s/%d not found", commit.Repo.Name, commit.Id, shard)
	}
	return result, nil
}

type fileReader struct {
	driver    *driver
	blockRefs []*drive.BlockRef
	index     int
	reader    io.ReadCloser
}

func (r *fileReader) Read(data []byte) (int, error) {
	if r.reader == nil {
		if r.index == len(r.blockRefs) {
			return 0, io.EOF
		}
		var err error
		r.reader, err = r.driver.getBlock(r.blockRefs[r.index].Block)
		if err != nil {
			return 0, err
		}
		r.index++
	}
	size, err := r.reader.Read(data)
	if err != nil && err != io.EOF {
		return size, err
	}
	if err == io.EOF {
		if err := r.reader.Close(); err != nil {
			return size, err
		}
		r.reader = nil
		recurseSize, err := r.Read(data[size:])
		if err != nil {
			return size + recurseSize, err
		}
		size += recurseSize
	}
	return size, nil
}

func (r *fileReader) ReadAt(p []byte, off int64) (n int, err error) {
	return 0, fmt.Errorf("fileReader.ReadAt: unimplemented")
}

func (r *fileReader) Close() error {
	if r.reader != nil {
		return r.reader.Close()
	}
	return nil
}
