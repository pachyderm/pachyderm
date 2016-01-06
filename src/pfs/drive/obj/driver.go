package obj

import (
	"fmt"
	"io"
	"log"
	"path"
	"sync"

	"go.pedge.io/google-protobuf"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pfs/pfsutil"
	"golang.org/x/net/context"
)

type driver struct {
	driveClient drive.APIClient
	started     diffMap
	finished    diffMap
	leaves      diffMap // commits with no children
	lock        sync.RWMutex
}

func newDriver(driveClient drive.APIClient) (drive.Driver, error) {
	return &driver{
		driveClient,
		make(diffMap),
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
	d.leaves[repo.Name] = make(map[uint64]map[string]*drive.DiffInfo)
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
		diffInfo := diffInfo
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

func (d *driver) StartCommit(parent *pfs.Commit, commit *pfs.Commit, started *google_protobuf.Timestamp, shards map[uint64]bool) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	for shard := range shards {
		diffInfo := &drive.DiffInfo{
			Diff: &drive.Diff{
				Commit: commit,
				Shard:  shard,
			},
			Started:      started,
			ParentCommit: parent,
			Appends:      make(map[string]*drive.Append),
		}
		if err := d.started.insert(diffInfo); err != nil {
			return err
		}
		if err := d.leaves.insert(diffInfo); err != nil {
			return err
		}
		if parent != nil {
			d.leaves.pop(&drive.Diff{
				Commit: parent,
				Shard:  shard,
			})
		}
	}
	return nil
}

func (d *driver) FinishCommit(commit *pfs.Commit, finished *google_protobuf.Timestamp, shards map[uint64]bool) error {
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
			diffInfo.Finished = finished
			diffInfos = append(diffInfos, diffInfo)
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
			if _, err := d.driveClient.CreateDiff(context.Background(), diffInfo); err != nil && loopErr == nil {
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

func (d *driver) ListCommit(repos []*pfs.Repo, fromCommit []*pfs.Commit, shards map[uint64]bool) ([]*pfs.CommitInfo, error) {
	repoSet := make(map[string]bool)
	for _, repo := range repos {
		repoSet[repo.Name] = true
	}
	breakCommitIds := make(map[string]bool)
	for _, commit := range fromCommit {
		if !repoSet[commit.Repo.Name] {
			return nil, fmt.Errorf("Commit %s/%s is from a repo that isn't being listed.", commit.Repo.Name, commit.Id)
		}
		breakCommitIds[commit.Id] = true
	}
	d.lock.RLock()
	defer d.lock.RUnlock()
	var result []*pfs.CommitInfo
	for _, repo := range repos {
		for shard := range shards {
			if _, err := d.inspectRepo(repo, shard); err != nil {
				return nil, err
			}
			for commitID := range d.leaves[repo.Name][shard] {
				commit := &pfs.Commit{
					Repo: repo,
					Id:   commitID,
				}
				for commit != nil && !breakCommitIds[commit.Id] {
					// we add this commit to breakCommitIds so we won't see it twice
					breakCommitIds[commit.Id] = true
					commitInfo, err := d.inspectCommit(commit, shards)
					if err != nil {
						return nil, err
					}
					result = append(result, commitInfo)
					commit = commitInfo.ParentCommit
				}
			}
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
	blockRefs, err := pfsutil.PutBlock(d.driveClient, reader)
	if err != nil {
		return err
	}
	d.lock.Lock()
	defer d.lock.Unlock()
	diffInfo, ok = d.started.get(&drive.Diff{
		Commit: file.Commit,
		Shard:  shard,
	})
	if !ok {
		// This is a weird case since the commit existed above, it means someone
		// deleted the commit while the above code was running
		return fmt.Errorf("commit %s/%s not found", file.Commit.Repo.Name, file.Commit.Id)
	}
	addDirs(diffInfo, file)
	_append, ok := diffInfo.Appends[path.Clean(file.Path)]
	if !ok {
		_append = &drive.Append{}
		if diffInfo.ParentCommit != nil {
			_append.LastRef = d.lastRef(
				pfsutil.NewFile(
					diffInfo.ParentCommit.Repo.Name,
					diffInfo.ParentCommit.Id,
					file.Path,
				),
				shard,
			)
		}
		diffInfo.Appends[path.Clean(file.Path)] = _append
	}
	_append.BlockRefs = append(_append.BlockRefs, blockRefs.BlockRef...)
	for _, blockRef := range blockRefs.BlockRef {
		diffInfo.SizeBytes += blockRef.Range.Upper - blockRef.Range.Lower
	}
	return nil
}

func (d *driver) MakeDirectory(file *pfs.File, shards map[uint64]bool) error {
	return nil
}

func (d *driver) GetFile(file *pfs.File, shard uint64) (drive.ReaderAtCloser, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	fileInfo, blockRefs, err := d.inspectFile(file, shard)
	if err != nil {
		return nil, err
	}
	if fileInfo.FileType == pfs.FileType_FILE_TYPE_DIR {
		return nil, fmt.Errorf("file %s/%s/%s is directory", file.Commit.Repo.Name, file.Commit.Id, file.Path)
	}
	return newFileReader(d.driveClient, blockRefs), nil
}

func (d *driver) InspectFile(file *pfs.File, shard uint64) (*pfs.FileInfo, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	fileInfo, _, err := d.inspectFile(file, shard)
	return fileInfo, err
}

func (d *driver) ListFile(file *pfs.File, shard uint64) ([]*pfs.FileInfo, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	fileInfo, _, err := d.inspectFile(file, shard)
	if err != nil {
		return nil, err
	}
	if fileInfo.FileType == pfs.FileType_FILE_TYPE_REGULAR {
		return []*pfs.FileInfo{fileInfo}, nil
	}
	var result []*pfs.FileInfo
	for _, child := range fileInfo.Children {
		fileInfo, _, err := d.inspectFile(child, shard)
		if err != nil {
			return nil, err
		}
		result = append(result, fileInfo)
	}
	return result, nil
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

func (d *driver) getDiffInfo(diff *drive.Diff) (_ *drive.DiffInfo, read bool, ok bool) {
	if diffInfo, ok := d.finished.get(diff); ok {
		return diffInfo, true, true
	}
	if diffInfo, ok := d.started.get(diff); ok {
		return diffInfo, false, true
	}
	return nil, false, false
}

func (d *driver) inspectCommit(commit *pfs.Commit, shards map[uint64]bool) (*pfs.CommitInfo, error) {
	var commitInfos []*pfs.CommitInfo
	for shard := range shards {
		if diffInfo, ok := d.finished.get(&drive.Diff{
			Commit: commit,
			Shard:  shard,
		}); ok {
			commitInfos = append(commitInfos,
				&pfs.CommitInfo{
					Commit:       commit,
					CommitType:   pfs.CommitType_COMMIT_TYPE_READ,
					ParentCommit: diffInfo.ParentCommit,
					Started:      diffInfo.Started,
					Finished:     diffInfo.Finished,
					SizeBytes:    diffInfo.SizeBytes,
				})
		}
		if diffInfo, ok := d.started.get(&drive.Diff{
			Commit: commit,
			Shard:  shard,
		}); ok {
			commitInfos = append(commitInfos,
				&pfs.CommitInfo{
					Commit:       commit,
					CommitType:   pfs.CommitType_COMMIT_TYPE_WRITE,
					ParentCommit: diffInfo.ParentCommit,
				})
		}
	}
	commitInfo := pfs.ReduceCommitInfos(commitInfos)
	if len(commitInfo) < 1 {
		return nil, fmt.Errorf("commit %s/%s not found", commit.Repo.Name, commit.Id)
	}
	if len(commitInfo) > 1 {
		return nil, fmt.Errorf("multiple commitInfos, (this is likely a bug)")
	}
	return commitInfo[0], nil
}

func (d *driver) inspectFile(file *pfs.File, shard uint64) (*pfs.FileInfo, []*drive.BlockRef, error) {
	fileInfo := &pfs.FileInfo{File: file}
	var blockRefs []*drive.BlockRef
	children := make(map[string]bool)
	commit := file.Commit
	for commit != nil {
		diffInfo, _, ok := d.getDiffInfo(&drive.Diff{
			Commit: commit,
			Shard:  shard,
		})
		if !ok {
			return nil, nil, fmt.Errorf("diff %s/%s not found", commit.Repo.Name, commit.Id)
		}
		log.Printf("Appends: %+v", diffInfo.Appends)
		if _append, ok := diffInfo.Appends[path.Clean(file.Path)]; ok {
			log.Printf("using: %+v", _append)
			if len(_append.BlockRefs) > 0 {
				if fileInfo.FileType == pfs.FileType_FILE_TYPE_DIR {
					return nil, nil,
						fmt.Errorf("mixed dir and regular file %s/%s/%s, (this is likely a bug)", file.Commit.Repo.Name, file.Commit.Id, file.Path)
				}
				fileInfo.FileType = pfs.FileType_FILE_TYPE_REGULAR
				blockRefs = append(_append.BlockRefs, blockRefs...)
				for _, blockRef := range _append.BlockRefs {
					fileInfo.SizeBytes += (blockRef.Range.Upper - blockRef.Range.Lower)
				}
			} else if len(_append.Children) > 0 {
				if fileInfo.FileType == pfs.FileType_FILE_TYPE_REGULAR {
					return nil, nil,
						fmt.Errorf("mixed dir and regular file %s/%s/%s, (this is likely a bug)", file.Commit.Repo.Name, file.Commit.Id, file.Path)
				}
				fileInfo.FileType = pfs.FileType_FILE_TYPE_DIR
				for child := range _append.Children {
					if !children[child] {
						fileInfo.Children = append(
							fileInfo.Children,
							pfsutil.NewFile(commit.Repo.Name, commit.Id, child),
						)
					}
					children[child] = true
				}
			}
			if fileInfo.CommitModified == nil {
				fileInfo.CommitModified = commit
				fileInfo.Modified = diffInfo.Finished
			}
			commit = _append.LastRef
			continue
		}
		commit = diffInfo.ParentCommit
	}
	if fileInfo.FileType == pfs.FileType_FILE_TYPE_NONE {
		return nil, nil, pfs.ErrFileNotFound
	}
	log.Printf("returning: %+v", fileInfo)
	return fileInfo, blockRefs, nil
}

// lastRef assumes the diffInfo file exists in finished
func (d *driver) lastRef(file *pfs.File, shard uint64) *pfs.Commit {
	commit := file.Commit
	for commit != nil {
		diffInfo, _ := d.finished.get(&drive.Diff{
			Commit: commit,
			Shard:  shard,
		})
		if _, ok := diffInfo.Appends[path.Clean(file.Path)]; ok {
			return commit
		}
		commit = diffInfo.ParentCommit
	}
	return nil
}

func addDirs(diffInfo *drive.DiffInfo, child *pfs.File) {
	childPath := child.Path
	dirPath := path.Dir(childPath)
	for {
		_append, ok := diffInfo.Appends[dirPath]
		if !ok {
			_append = &drive.Append{}
			diffInfo.Appends[dirPath] = _append
		}
		if _append.Children == nil {
			_append.Children = make(map[string]bool)
		}
		_append.Children[childPath] = true
		if dirPath == "." {
			break
		}
		childPath = dirPath
		dirPath = path.Dir(childPath)
	}
}

type fileReader struct {
	driveClient drive.APIClient
	blockRefs   []*drive.BlockRef
	index       int
	reader      io.Reader
}

func newFileReader(driveClient drive.APIClient, blockRefs []*drive.BlockRef) *fileReader {
	return &fileReader{
		driveClient: driveClient,
		blockRefs:   blockRefs,
	}
}

func (r *fileReader) Read(data []byte) (int, error) {
	if r.reader == nil {
		if r.index == len(r.blockRefs) {
			return 0, io.EOF
		}
		var err error
		r.reader, err = pfsutil.GetBlock(r.driveClient, r.blockRefs[r.index].Block.Hash, 0)
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

func (r *fileReader) ReadAt(p []byte, off int64) (int, error) {
	var read int
	for _, blockRef := range r.blockRefs {
		blockSize := int64(blockRef.Range.Upper - blockRef.Range.Lower)
		if off >= blockSize {
			off -= blockSize
			continue
		}
		reader, err := pfsutil.GetBlock(r.driveClient, blockRef.Block.Hash, uint64(off))
		off = 0
		if err != nil {
			return 0, err
		}
		for read < len(p) {
			n, err := reader.Read(p[read:])
			read += n
			if err == io.EOF {
				break
			}
			if err != nil {
				return read, err
			}
		}
		if read == len(p) {
			break
		}
	}
	if read <= len(p) {
		return read, io.EOF
	}
	return read, nil
}

func (r *fileReader) Close() error {
	return nil
}

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
		return fmt.Errorf("repo %s not found", diff.Commit.Repo.Name)
	}
	commitMap, ok := shardMap[diff.Shard]
	if !ok {
		commitMap = make(map[string]*drive.DiffInfo)
		shardMap[diff.Shard] = commitMap
	}
	if _, ok = commitMap[diff.Commit.Id]; ok {
		return fmt.Errorf("commit %s/%s already exists", diff.Commit.Repo.Name, diff.Commit.Id)
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
