package drive

import (
	"fmt"
	"io"
	"path"
	"sync"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/pfsutil"
	"go.pedge.io/pb/go/google/protobuf"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type driver struct {
	blockAddress    string
	blockClient     pfs.BlockAPIClient
	blockClientOnce sync.Once
	started         diffMap
	finished        diffMap
	leaves          diffMap // commits with no children
	lock            sync.RWMutex
}

func newDriver(blockAddress string) (Driver, error) {
	return &driver{
		blockAddress,
		nil,
		sync.Once{},
		make(diffMap),
		make(diffMap),
		make(diffMap),
		sync.RWMutex{},
	}, nil
}

func (d *driver) getBlockClient() (pfs.BlockAPIClient, error) {
	if d.blockClient == nil {
		var onceErr error
		d.blockClientOnce.Do(func() {
			clientConn, err := grpc.Dial(d.blockAddress, grpc.WithInsecure())
			if err != nil {
				onceErr = err
			}
			d.blockClient = pfs.NewBlockAPIClient(clientConn)
		})
		if onceErr != nil {
			return nil, onceErr
		}
	}
	return d.blockClient, nil
}

func (d *driver) CreateRepo(repo *pfs.Repo, created *google_protobuf.Timestamp, shards map[uint64]bool) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	if _, ok := d.finished[repo.Name]; ok {
		return fmt.Errorf("repo %s exists", repo.Name)
	}
	d.createRepoDiffMaps(repo)

	blockClient, err := d.getBlockClient()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	var loopErr error
	for shard := range shards {
		wg.Add(1)
		diffInfo := &pfs.DiffInfo{
			Diff: &pfs.Diff{
				Commit: &pfs.Commit{Repo: repo},
				Shard:  shard,
			},
			Finished: created,
		}
		if err := d.finished.insert(diffInfo); err != nil {
			return err
		}
		go func() {
			defer wg.Done()
			if _, err := blockClient.CreateDiff(context.Background(), diffInfo); err != nil && loopErr == nil {
				loopErr = err
			}
		}()
	}
	wg.Wait()
	return loopErr
}

func (d *driver) InspectRepo(repo *pfs.Repo, shards map[uint64]bool) (*pfs.RepoInfo, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.inspectRepo(repo, shards)
}

func (d *driver) ListRepo(shards map[uint64]bool) ([]*pfs.RepoInfo, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	var wg sync.WaitGroup
	var loopErr error
	var result []*pfs.RepoInfo
	var lock sync.Mutex
	for repoName := range d.finished {
		wg.Add(1)
		repoName := repoName
		go func() {
			defer wg.Done()
			repoInfo, err := d.inspectRepo(&pfs.Repo{Name: repoName}, shards)
			if err != nil && loopErr == nil {
				loopErr = err
			}
			lock.Lock()
			defer lock.Unlock()
			result = append(result, repoInfo)
		}()
	}
	wg.Wait()
	if loopErr != nil {
		return nil, loopErr
	}
	return result, nil
}

func (d *driver) DeleteRepo(repo *pfs.Repo, shards map[uint64]bool) error {
	var diffInfos []*pfs.DiffInfo
	d.lock.Lock()
	for shard := range shards {
		for _, diffInfo := range d.started[repo.Name][shard] {
			diffInfos = append(diffInfos, diffInfo)
		}
	}
	delete(d.started, repo.Name)
	delete(d.finished, repo.Name)
	delete(d.leaves, repo.Name)
	d.lock.Unlock()
	blockClient, err := d.getBlockClient()
	if err != nil {
		return err
	}
	var loopErr error
	var wg sync.WaitGroup
	for _, diffInfo := range diffInfos {
		diffInfo := diffInfo
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := blockClient.DeleteDiff(
				context.Background(),
				&pfs.DeleteDiffRequest{Diff: diffInfo.Diff},
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
		diffInfo := &pfs.DiffInfo{
			Diff: &pfs.Diff{
				Commit: commit,
				Shard:  shard,
			},
			Started:      started,
			ParentCommit: parent,
			Appends:      make(map[string]*pfs.Append),
		}
		if err := d.started.insert(diffInfo); err != nil {
			return err
		}
		if err := d.insertLeaf(diffInfo); err != nil {
			return err
		}
	}
	return nil
}

func (d *driver) FinishCommit(commit *pfs.Commit, finished *google_protobuf.Timestamp, shards map[uint64]bool) error {
	// closure so we can defer Unlock
	var diffInfos []*pfs.DiffInfo
	if err := func() error {
		d.lock.Lock()
		defer d.lock.Unlock()
		for shard := range shards {
			diffInfo := d.started.pop(&pfs.Diff{
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
	blockClient, err := d.getBlockClient()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	var loopErr error
	for _, diffInfo := range diffInfos {
		diffInfo := diffInfo
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := blockClient.CreateDiff(context.Background(), diffInfo); err != nil && loopErr == nil {
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
			_, ok := d.finished[repo.Name]
			if !ok {
				return nil, fmt.Errorf("repo %s not found", repo.Name)
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

func (d *driver) PutFile(file *pfs.File, shard uint64, offset int64, reader io.Reader) (retErr error) {
	d.lock.RLock()
	diffInfo, ok := d.started.get(&pfs.Diff{
		Commit: file.Commit,
		Shard:  shard,
	})
	d.lock.RUnlock()
	if !ok {
		return fmt.Errorf("commit %s/%s not found", file.Commit.Repo.Name, file.Commit.Id)
	}
	blockClient, err := d.getBlockClient()
	if err != nil {
		return err
	}
	blockRefs, err := pfsutil.PutBlock(blockClient, reader)
	if err != nil {
		return err
	}
	d.lock.Lock()
	defer d.lock.Unlock()
	diffInfo, ok = d.started.get(&pfs.Diff{
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
		_append = &pfs.Append{}
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

func (d *driver) GetFile(file *pfs.File, filterShard *pfs.Shard, offset int64, size int64, shard uint64) (io.ReadCloser, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	fileInfo, blockRefs, err := d.inspectFile(file, filterShard, shard)
	if err != nil {
		return nil, err
	}
	if fileInfo.FileType == pfs.FileType_FILE_TYPE_DIR {
		return nil, fmt.Errorf("file %s/%s/%s is directory", file.Commit.Repo.Name, file.Commit.Id, file.Path)
	}
	blockClient, err := d.getBlockClient()
	if err != nil {
		return nil, err
	}
	return newFileReader(blockClient, blockRefs, offset, size), nil
}

func (d *driver) InspectFile(file *pfs.File, filterShard *pfs.Shard, shard uint64) (*pfs.FileInfo, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	fileInfo, _, err := d.inspectFile(file, filterShard, shard)
	return fileInfo, err
}

func (d *driver) ListFile(file *pfs.File, filterShard *pfs.Shard, shard uint64) ([]*pfs.FileInfo, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	fileInfo, _, err := d.inspectFile(file, filterShard, shard)
	if err != nil {
		return nil, err
	}
	if fileInfo.FileType == pfs.FileType_FILE_TYPE_REGULAR {
		return []*pfs.FileInfo{fileInfo}, nil
	}
	var result []*pfs.FileInfo
	for _, child := range fileInfo.Children {
		fileInfo, _, err := d.inspectFile(child, filterShard, shard)
		if err != nil && err != pfs.ErrFileNotFound {
			return nil, err
		}
		if err == pfs.ErrFileNotFound {
			// how can a listed child return not found?
			// regular files without any blocks in this shard count as not found
			continue
		}
		result = append(result, fileInfo)
	}
	return result, nil
}

func (d *driver) DeleteFile(file *pfs.File, shard uint64) error {
	return nil
}

func (d *driver) AddShard(shard uint64) error {
	blockClient, err := d.getBlockClient()
	if err != nil {
		return err
	}
	listDiffClient, err := blockClient.ListDiff(context.Background(), &pfs.ListDiffRequest{Shard: shard})
	if err != nil {
		return err
	}
	leaves := make(diffMap)
	for {
		diffInfo, err := listDiffClient.Recv()
		if err != nil && err != io.EOF {
			return err
		}
		if err == io.EOF {
			break
		}
		if err := func() error {
			d.lock.Lock()
			defer d.lock.Lock()
			if _, ok := d.finished[diffInfo.Diff.Commit.Repo.Name]; !ok {
				d.createRepoDiffMaps(diffInfo.Diff.Commit.Repo)
			}
			if err := d.finished.insert(diffInfo); err != nil {
				return err
			}
			if diffInfo.ParentCommit != nil {
				parentDiff := &pfs.Diff{
					Commit: diffInfo.ParentCommit,
					Shard:  shard,
				}
				// try to pop the parent diff, if it we find one to pop then we
				// already saw the parent we're good, otherwise we record the
				// fact that this diff is not a leaf with a dummy diffInfo
				if parentDiffInfo := leaves.pop(parentDiff); parentDiffInfo == nil {
					if err := leaves.insert(&pfs.DiffInfo{Diff: parentDiff}); err != nil {
						return err
					}
				}
			}
			// check if there's a dummy diff info for this diff, if so then we
			// already saw a parent, so this isn't a leaf, otherwise it's a leaf
			// so we insert it
			if dummyDiffInfo := leaves.pop(diffInfo.Diff); dummyDiffInfo == nil {
				if err := leaves.insert(diffInfo); err != nil {
					return err
				}
			}
			return nil
		}(); err != nil {
			return err
		}
	}
	for repo, shardMap := range leaves {
		for _, commitMap := range shardMap {
			for commit, diffInfo := range commitMap {
				if diffInfo.Appends == nil {
					return fmt.Errorf("diffInfos reference a parent that doesn't exist %s/%s", repo, commit)
				}
				if err := d.leaves.insert(diffInfo); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (d *driver) DeleteShard(shard uint64) error {
	d.lock.Lock()
	defer d.lock.Lock()
	for _, shardMap := range d.finished {
		delete(shardMap, shard)
	}
	for _, shardMap := range d.started {
		delete(shardMap, shard)
	}
	return nil
}

func (d *driver) inspectRepo(repo *pfs.Repo, shards map[uint64]bool) (*pfs.RepoInfo, error) {
	result := &pfs.RepoInfo{
		Repo: repo,
	}
	_, ok := d.finished[repo.Name]
	if !ok {
		return nil, fmt.Errorf("repo %s not found", repo.Name)
	}
	for shard := range shards {
		diffInfos, ok := d.finished[repo.Name][shard]
		if !ok {
			return nil, fmt.Errorf("repo %s not found", repo.Name)
		}
		for _, diffInfo := range diffInfos {
			diffInfo := diffInfo
			if diffInfo.Diff.Commit.Id == "" {
				result.Created = diffInfo.Finished
			}
			result.SizeBytes += diffInfo.SizeBytes
		}
	}
	return result, nil
}

func (d *driver) getDiffInfo(diff *pfs.Diff) (_ *pfs.DiffInfo, read bool, ok bool) {
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
		if diffInfo, ok := d.finished.get(&pfs.Diff{
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
		if diffInfo, ok := d.started.get(&pfs.Diff{
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

func filterBlockRefs(filterShard *pfs.Shard, blockRefs []*pfs.BlockRef) []*pfs.BlockRef {
	var result []*pfs.BlockRef
	for _, blockRef := range blockRefs {
		if pfs.BlockInShard(filterShard, blockRef.Block) {
			result = append(result, blockRef)
		}
	}
	return result
}

func (d *driver) inspectFile(file *pfs.File, filterShard *pfs.Shard, shard uint64) (*pfs.FileInfo, []*pfs.BlockRef, error) {
	fileInfo := &pfs.FileInfo{File: file}
	var blockRefs []*pfs.BlockRef
	children := make(map[string]bool)
	commit := file.Commit
	for commit != nil {
		diffInfo, _, ok := d.getDiffInfo(&pfs.Diff{
			Commit: commit,
			Shard:  shard,
		})
		if !ok {
			return nil, nil, fmt.Errorf("diff %s/%s not found", commit.Repo.Name, commit.Id)
		}
		if _append, ok := diffInfo.Appends[path.Clean(file.Path)]; ok {
			if len(_append.BlockRefs) > 0 {
				if fileInfo.FileType == pfs.FileType_FILE_TYPE_DIR {
					return nil, nil,
						fmt.Errorf("mixed dir and regular file %s/%s/%s, (this is likely a bug)", file.Commit.Repo.Name, file.Commit.Id, file.Path)
				}
				if fileInfo.FileType == pfs.FileType_FILE_TYPE_NONE {
					// the first time we find out it's a regular file we check
					// the file shard, dirs get returned regardless of sharding,
					// since they might have children from any shard
					if !pfs.FileInShard(filterShard, file) {
						return nil, nil, pfs.ErrFileNotFound
					}
				}
				fileInfo.FileType = pfs.FileType_FILE_TYPE_REGULAR
				filtered := filterBlockRefs(filterShard, _append.BlockRefs)
				blockRefs = append(filtered, blockRefs...)
				for _, blockRef := range filtered {
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
	return fileInfo, blockRefs, nil
}

// lastRef assumes the diffInfo file exists in finished
func (d *driver) lastRef(file *pfs.File, shard uint64) *pfs.Commit {
	commit := file.Commit
	for commit != nil {
		diffInfo, _ := d.finished.get(&pfs.Diff{
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

func addDirs(diffInfo *pfs.DiffInfo, child *pfs.File) {
	childPath := child.Path
	dirPath := path.Dir(childPath)
	for {
		_append, ok := diffInfo.Appends[dirPath]
		if !ok {
			_append = &pfs.Append{}
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
	blockClient pfs.BlockAPIClient
	blockRefs   []*pfs.BlockRef
	index       int
	reader      io.Reader
	offset      int64
	size        int64
	ctx         context.Context
	cancel      context.CancelFunc
}

func newFileReader(blockClient pfs.BlockAPIClient, blockRefs []*pfs.BlockRef, offset int64, size int64) *fileReader {
	return &fileReader{
		blockClient: blockClient,
		blockRefs:   blockRefs,
		offset:      offset,
		size:        size,
	}
}

func (r *fileReader) Read(data []byte) (int, error) {
	if r.reader == nil {
		if r.index == len(r.blockRefs) {
			return 0, io.EOF
		}
		blockRef := r.blockRefs[r.index]
		for r.offset != 0 && r.offset > int64(pfs.ByteRangeSize(blockRef.Range)) {
			r.index++
			r.offset -= int64(pfs.ByteRangeSize(blockRef.Range))
		}
		var err error
		r.reader, err = pfsutil.GetBlock(r.blockClient,
			r.blockRefs[r.index].Block.Hash, uint64(r.offset), uint64(r.size))
		if err != nil {
			return 0, err
		}
		r.offset = 0
		r.index++
	}
	size, err := r.reader.Read(data)
	if err != nil && err != io.EOF {
		return size, err
	}
	if err == io.EOF {
		r.reader = nil
	}
	r.size -= int64(size)
	if r.size == 0 {
		return size, io.EOF
	}
	return size, nil
}

func (r *fileReader) Close() error {
	return nil
}

type diffMap map[string]map[uint64]map[string]*pfs.DiffInfo

func (d diffMap) get(diff *pfs.Diff) (_ *pfs.DiffInfo, ok bool) {
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

func (d diffMap) insert(diffInfo *pfs.DiffInfo) error {
	diff := diffInfo.Diff
	shardMap, ok := d[diff.Commit.Repo.Name]
	if !ok {
		return fmt.Errorf("repo %s not found", diff.Commit.Repo.Name)
	}
	commitMap, ok := shardMap[diff.Shard]
	if !ok {
		commitMap = make(map[string]*pfs.DiffInfo)
		shardMap[diff.Shard] = commitMap
	}
	if _, ok = commitMap[diff.Commit.Id]; ok {
		return fmt.Errorf("commit %s/%s already exists", diff.Commit.Repo.Name, diff.Commit.Id)
	}
	commitMap[diff.Commit.Id] = diffInfo
	return nil
}

func (d diffMap) pop(diff *pfs.Diff) *pfs.DiffInfo {
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

func (d *driver) insertLeaf(leaf *pfs.DiffInfo) error {
	if err := d.leaves.insert(leaf); err != nil {
		return err
	}
	if leaf.ParentCommit != nil {
		parentDiff := &pfs.Diff{
			Commit: leaf.ParentCommit,
			Shard:  leaf.Diff.Shard,
		}
		if _, ok := d.leaves.get(parentDiff); ok {
			d.leaves.pop(parentDiff)
		}
	}
	return nil
}

func (d *driver) createRepoDiffMaps(repo *pfs.Repo) {
	d.finished[repo.Name] = make(map[uint64]map[string]*pfs.DiffInfo)
	d.started[repo.Name] = make(map[uint64]map[string]*pfs.DiffInfo)
	d.leaves[repo.Name] = make(map[uint64]map[string]*pfs.DiffInfo)
}
