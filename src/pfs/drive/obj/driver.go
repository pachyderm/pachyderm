package obj

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"sync"

	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pfs/pfsutil"
)

var (
	blockSize = 128 * 1024 * 1024 // 128 Megabytes
)

type diffMap map[string]map[uint64]map[string]*drive.DiffInfo

func (d diffMap) get(diff *drive.Diff) (_ *drive.DiffInfo, ok bool) {
	shardMap, ok := d[diff.Commit.Repo.Name]
	if !ok {
		return nil, false
	}
	commitMap, ok := shardMap[diff.Shard]
	if !ok {
		return nil, false
	}
	diffInfo, ok := commitMap[diff.Commit.Id]
	return diffInfo, ok
}

func (d diffMap) insert(diffInfo *drive.DiffInfo) error {
	diff := diffInfo.Diff
	shardMap, ok := d[diff.Commit.Repo.Name]
	if !ok {
		return fmt.Errorf("repo %s not found")
	}
	commitMap, ok := shardMap[diff.Shard]
	if !ok {
		commitMap = make(map[string]*drive.DiffInfo)
		shardMap[diff.Shard] = commitMap
	}
	if _, ok = commitMap[diff.Commit.Id]; ok {
		return fmt.Errorf("commit %s/%d/%s already exists")
	}
	commitMap[diff.Commit.Id] = diffInfo
	return nil
}

func (d diffMap) pop(diff *drive.Diff) *drive.DiffInfo {
	shardMap, ok := d[diff.Commit.Repo.Name]
	if !ok {
		return nil
	}
	commitMap, ok := shardMap[diff.Shard]
	if !ok {
		return nil
	}
	diffInfo := commitMap[diff.Commit.Id]
	delete(commitMap, diff.Commit.Id)
	return diffInfo
}

type driver struct {
	driveClient drive.APIClient
	started     diffMap
	finished    diffMap
	lock        sync.RWMutex
}

func newDriver(driveClient drive.APIClient) (drive.Driver, error) {
	return &driver{
		driveClient,
		make(diffMap),
		make(diffMap),
		sync.RWMutex{},
	}, nil
}

func (d *driver) CreateRepo(repo *pfs.Repo) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	if _, ok := d.finished[repo.Name]; ok {
		return fmt.Errorf("repo %s exists", repo.Name)
	}
	d.finished[repo.Name] = make(map[uint64]map[string]*drive.DiffInfo)
	d.started[repo.Name] = make(map[uint64]map[string]*drive.DiffInfo)
	return nil
}

func (d *driver) InspectRepo(repo *pfs.Repo, shard uint64) (*pfs.RepoInfo, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.inspectRepo(repo, shard)
}

func (d *driver) ListRepo(shard uint64) ([]*pfs.RepoInfo, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	var result []*pfs.RepoInfo
	for repoName := range d.finished {
		repoInfo, err := d.inspectRepo(&pfs.Repo{Name: repoName}, shard)
		if err != nil {
			return nil, err
		}
		result = append(result, repoInfo)
	}
	return result, nil
}

func (d *driver) DeleteRepo(repo *pfs.Repo, shards map[uint64]bool) error {
	var diffInfos []*drive.DiffInfo
	d.lock.Lock()
	for shard := range shards {
		for _, diffInfo := range d.started[repo.Name][shard] {
			diffInfos = append(diffInfos, diffInfo)
		}
	}
	delete(d.started, repo.Name)
	delete(d.finished, repo.Name)
	d.lock.Unlock()
	var loopErr error
	var wg sync.WaitGroup
	for _, diffInfo := range diffInfos {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := d.driveClient.DeleteDiff(
				context.Background(),
				&drive.DeleteDiffRequest{Diff: diffInfo.Diff},
			); err != nil && loopErr == nil {
				loopErr = err
			}
		}()
	}
	wg.Wait()
	return loopErr
}

func (d *driver) StartCommit(parent *pfs.Commit, commit *pfs.Commit, shards map[uint64]bool) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	for shard := range shards {
		diffInfo := &drive.DiffInfo{
			Diff: &drive.Diff{
				Commit: commit,
				Shard:  shard,
			},
			ParentCommit: parent,
			Appends:      make(map[string]*drive.BlockRefs),
			LastRef:      make(map[string]*drive.Diff),
		}
		if err := d.started.insert(diffInfo); err != nil {
			return err
		}
	}
	return nil
}

func (d *driver) FinishCommit(commit *pfs.Commit, shards map[uint64]bool) error {
	// closure so we can defer Unlock
	var diffInfos []*drive.DiffInfo
	if err := func() error {
		d.lock.Lock()
		defer d.lock.Unlock()
		for shard := range shards {
			diffInfo := d.started.pop(&drive.Diff{
				Commit: commit,
				Shard:  shard,
			})
			if diffInfo == nil {
				return fmt.Errorf("commit %s/%s not found", commit.Repo.Name, commit.Id)
			}
			diffInfos = append(diffInfos)
			if err := d.finished.insert(diffInfo); err != nil {
				return err
			}
		}
		return nil
	}(); err != nil {
		return err
	}
	var wg sync.WaitGroup
	var loopErr error
	for _, diffInfo := range diffInfos {
		diffInfo := diffInfo
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := d.driveClient.CreateDiff(
				context.Background(),
				&drive.CreateDiffRequest{
					Diff:         diffInfo.Diff,
					ParentCommit: diffInfo.ParentCommit,
					Appends:      diffInfo.Appends,
					LastRef:      diffInfo.LastRef,
				}); err != nil && loopErr == nil {
				loopErr = err
			}
		}()
	}
	wg.Wait()
	return loopErr
}

func (d *driver) InspectCommit(commit *pfs.Commit, shards map[uint64]bool) (*pfs.CommitInfo, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.inspectCommit(commit, shards)
}

func (d *driver) ListCommit(repo *pfs.Repo, from *pfs.Commit, shards map[uint64]bool) ([]*pfs.CommitInfo, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	var result []*pfs.CommitInfo
	for shard := range shards {
		if _, err := d.inspectRepo(repo, shard); err != nil {
			return nil, err
		}
		var commitIds []string
		for commitId := range d.finished[repo.Name][shard] {
			commitIds = append(commitIds, commitId)
		}
		for commitId := range d.started[repo.Name][shard] {
			commitIds = append(commitIds, commitId)
		}
		for _, commitId := range commitIds {
			commitInfo, err := d.inspectCommit(&pfs.Commit{
				Repo: repo,
				Id:   commitId,
			}, shards)
			if err != nil {
				return nil, err
			}
			result = append(result, commitInfo)
		}
	}
	return result, nil
}

func (d *driver) DeleteCommit(commit *pfs.Commit, shards map[uint64]bool) error {
	return nil
}

func (d *driver) PutBlock(file *pfs.File, block *drive.Block, shard uint64, reader io.Reader) error {
	return nil
}

func (d *driver) GetBlock(block *drive.Block, shard uint64) (drive.ReaderAtCloser, error) {
	return nil, nil
}

func (d *driver) InspectBlock(block *drive.Block, shard uint64) (*drive.BlockInfo, error) {
	return nil, nil
}

func (d *driver) ListBlock(shard uint64) ([]*drive.BlockInfo, error) {
	return nil, nil
}

func (d *driver) PutFile(file *pfs.File, shard uint64, offset int64, reader io.Reader) (retErr error) {
	d.lock.RLock()
	diffInfo, ok := d.started.get(&drive.Diff{
		Commit: file.Commit,
		Shard:  shard,
	})
	d.lock.RUnlock()
	if !ok {
		return fmt.Errorf("commit %s/%s not found", file.Commit.Repo.Name, file.Commit.Id)
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
	var loopErr error
	for scanner.Scan() {
		data := scanner.Bytes()
		blockRef := &drive.BlockRef{
			Range: &drive.ByteRange{
				Lower: 0,
				Upper: uint64(len(data)),
			},
		}
		blockRefs = append(blockRefs, blockRef)
		wg.Add(1)
		go func() {
			defer wg.Done()
			block, err := pfsutil.PutBlock(d.driveClient, bytes.NewReader(data))
			if err != nil && loopErr == nil {
				loopErr = err
			}
			blockRef.Block = block
		}()
	}
	wg.Wait()
	if err := scanner.Err(); err != nil {
		return err
	}
	d.lock.Lock()
	defer d.lock.Unlock()
	diffInfo, ok = d.started.get(&drive.Diff{
		Commit: file.Commit,
		Shard:  shard,
	})
	if !ok {
		return fmt.Errorf("commit %s/%s not found", file.Commit.Repo.Name, file.Commit.Id)
	}
	diffInfo.Appends[file.Path] = &drive.BlockRefs{
		BlockRef: append(diffInfo.Appends[file.Path].BlockRef, blockRefs...),
	}
	return nil
}

func (d *driver) MakeDirectory(file *pfs.File, shards map[uint64]bool) error {
	return nil
}

func (d *driver) GetFile(file *pfs.File, shard uint64) (drive.ReaderAtCloser, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	var blockRefs []*drive.BlockRef
	commit := file.Commit
	for commit != nil {
		diffInfo, ok := d.finished.get(&drive.Diff{
			Commit: commit,
			Shard:  shard,
		})
		if !ok {
			return nil, fmt.Errorf("file %s/%s/%s not found", file.Commit.Repo.Name, file.Commit.Id, file.Path)
		}
		blockRefs = append(blockRefs, diffInfo.Appends[file.Path].BlockRef...)
		commit = diffInfo.ParentCommit
	}
	return &fileReader{
		driveClient: d.driveClient,
		blockRefs:   blockRefs,
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

func (d *driver) inspectRepo(repo *pfs.Repo, shard uint64) (*pfs.RepoInfo, error) {
	_, ok := d.finished[repo.Name]
	if !ok {
		return nil, fmt.Errorf("repo %s not found", repo.Name)
	}
	return &pfs.RepoInfo{
		Repo: repo,
		// Created: TODO,
		// SizeBytes: TODO,
	}, nil
}

func (d *driver) inspectCommit(commit *pfs.Commit, shards map[uint64]bool) (*pfs.CommitInfo, error) {
	result := &pfs.CommitInfo{
		Commit: commit,
	}
	for shard := range shards {
		if diffInfo, ok := d.finished.get(&drive.Diff{
			Commit: commit,
			Shard:  shard,
		}); ok {
			result.CommitType = pfs.CommitType_COMMIT_TYPE_READ
			result.ParentCommit = diffInfo.ParentCommit
			break
		}
		if diffInfo, ok := d.started.get(&drive.Diff{
			Commit: commit,
			Shard:  shard,
		}); ok {
			result.CommitType = pfs.CommitType_COMMIT_TYPE_WRITE
			result.ParentCommit = diffInfo.ParentCommit
			break
		}
		return nil, fmt.Errorf("commit %s/%s not found", commit.Repo.Name, commit.Id)
	}
	return result, nil
}

type fileReader struct {
	driveClient drive.APIClient
	blockRefs   []*drive.BlockRef
	index       int
	reader      io.Reader
}

func (r *fileReader) Read(data []byte) (int, error) {
	if r.reader == nil {
		if r.index == len(r.blockRefs) {
			return 0, io.EOF
		}
		var err error
		r.reader, err = pfsutil.GetBlock(r.driveClient, r.blockRefs[r.index].Block.Hash)
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
	return nil
}
