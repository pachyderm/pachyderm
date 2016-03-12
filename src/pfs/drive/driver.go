package drive

import (
	"fmt"
	"io"
	"path"
	"sync"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/pfsutil"
	"github.com/pachyderm/pachyderm/src/pkg/dag"
	"github.com/pachyderm/pachyderm/src/pkg/metrics"
	"go.pedge.io/pb/go/google/protobuf"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type driver struct {
	blockAddress    string
	blockClient     pfs.BlockAPIClient
	blockClientOnce sync.Once
	diffs           diffMap
	dags            map[string]*dag.DAG
	branches        map[string]map[string]string
	lock            sync.RWMutex
}

func newDriver(blockAddress string) (Driver, error) {
	return &driver{
		blockAddress,
		nil,
		sync.Once{},
		make(diffMap),
		make(map[string]*dag.DAG),
		make(map[string]map[string]string),
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
	if _, ok := d.diffs[repo.Name]; ok {
		return fmt.Errorf("repo %s exists", repo.Name)
	}
	d.createRepoState(repo)

	blockClient, err := d.getBlockClient()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	errCh := make(chan error, 1)
	for shard := range shards {
		wg.Add(1)
		diffInfo := &pfs.DiffInfo{
			Diff:     pfsutil.NewDiff(repo.Name, "", shard),
			Finished: created,
		}
		if err := d.diffs.insert(diffInfo); err != nil {
			return err
		}
		go func() {
			defer wg.Done()
			if _, err := blockClient.CreateDiff(context.Background(), diffInfo); err != nil {
				select {
				case errCh <- err:
					// error reported
				default:
					// not the first error
				}
			}
		}()
	}
	wg.Wait()
	select {
	case err := <-errCh:
		return err
	default:
	}
	return nil
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
	errCh := make(chan error, 1)
	var result []*pfs.RepoInfo
	var lock sync.Mutex
	for repoName := range d.diffs {
		wg.Add(1)
		repoName := repoName
		go func() {
			defer wg.Done()
			repoInfo, err := d.inspectRepo(&pfs.Repo{Name: repoName}, shards)
			if err != nil {
				select {
				case errCh <- err:
					// error reported
				default:
					// not the first error
				}
			}
			lock.Lock()
			defer lock.Unlock()
			result = append(result, repoInfo)
		}()
	}
	wg.Wait()
	select {
	case err := <-errCh:
		return nil, err
	default:
		// no error
	}
	return result, nil
}

func (d *driver) DeleteRepo(repo *pfs.Repo, shards map[uint64]bool) error {
	var diffInfos []*pfs.DiffInfo
	d.lock.Lock()
	for shard := range shards {
		for _, diffInfo := range d.diffs[repo.Name][shard] {
			diffInfos = append(diffInfos, diffInfo)
		}
	}
	delete(d.diffs, repo.Name)
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

func (d *driver) StartCommit(repo *pfs.Repo, commitID string, parentID string, branch string,
	started *google_protobuf.Timestamp, shards map[uint64]bool) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	for shard := range shards {
		diffInfo := &pfs.DiffInfo{
			Diff:    pfsutil.NewDiff(repo.Name, commitID, shard),
			Started: started,
			Appends: make(map[string]*pfs.Append),
			Branch:  branch,
		}
		if branch != "" {
			parentCommit, err := d.branchParent(pfsutil.NewCommit(repo.Name, commitID), branch)
			if err != nil {
				return err
			}
			if parentCommit != nil && parentID != "" {
				return fmt.Errorf("branch %s already exists as %s, can't create with %s as parent",
					branch, parentCommit.ID, parentID)
			}
			diffInfo.ParentCommit = parentCommit
		}
		if diffInfo.ParentCommit == nil && parentID != "" {
			diffInfo.ParentCommit = pfsutil.NewCommit(repo.Name, parentID)
		}
		if err := d.insertDiffInfo(diffInfo); err != nil {
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
		canonicalCommit, err := d.canonicalCommit(commit)
		if err != nil {
			return err
		}
		for shard := range shards {
			diffInfo, ok := d.diffs.get(pfsutil.NewDiff(canonicalCommit.Repo.Name, canonicalCommit.ID, shard))
			if !ok {
				return fmt.Errorf("commit %s/%s not found", canonicalCommit.Repo.Name, canonicalCommit.ID)
			}
			diffInfo.Finished = finished
			diffInfos = append(diffInfos, diffInfo)
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
	breakCommitIDs := make(map[string]bool)
	for _, commit := range fromCommit {
		if !repoSet[commit.Repo.Name] {
			return nil, fmt.Errorf("Commit %s/%s is from a repo that isn't being listed.", commit.Repo.Name, commit.ID)
		}
		breakCommitIDs[commit.ID] = true
	}
	d.lock.RLock()
	defer d.lock.RUnlock()
	var result []*pfs.CommitInfo
	for _, repo := range repos {
		_, ok := d.diffs[repo.Name]
		if !ok {
			return nil, fmt.Errorf("repo %s not found", repo.Name)
		}
		for _, commitID := range d.dags[repo.Name].Leaves() {
			commit := &pfs.Commit{
				Repo: repo,
				ID:   commitID,
			}
			for commit != nil && !breakCommitIDs[commit.ID] {
				// we add this commit to breakCommitIDs so we won't see it twice
				breakCommitIDs[commit.ID] = true
				commitInfo, err := d.inspectCommit(commit, shards)
				if err != nil {
					return nil, err
				}
				result = append(result, commitInfo)
				commit = commitInfo.ParentCommit
			}
		}
	}
	return result, nil
}

func (d *driver) ListBranch(repo *pfs.Repo, shards map[uint64]bool) ([]*pfs.CommitInfo, error) {
	var result []*pfs.CommitInfo
	for commitID := range d.branches[repo.Name] {
		commitInfo, err := d.inspectCommit(pfsutil.NewCommit(repo.Name, commitID), shards)
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

func (d *driver) PutFile(file *pfs.File, shard uint64, offset int64, reader io.Reader) (retErr error) {
	blockClient, err := d.getBlockClient()
	if err != nil {
		return err
	}
	blockRefs, err := pfsutil.PutBlock(blockClient, reader)
	if err != nil {
		return err
	}
	defer func() {
		if retErr == nil {
			metrics.AddFiles(1)
			for _, blockRef := range blockRefs.BlockRef {
				metrics.AddBytes(int64(blockRef.Range.Upper - blockRef.Range.Lower))
			}
		}
	}()
	d.lock.Lock()
	defer d.lock.Unlock()
	canonicalCommit, err := d.canonicalCommit(file.Commit)
	if err != nil {
		return err
	}
	diffInfo, ok := d.diffs.get(pfsutil.NewDiff(canonicalCommit.Repo.Name, canonicalCommit.ID, shard))
	if !ok {
		// This is a weird case since the commit existed above, it means someone
		// deleted the commit while the above code was running
		return fmt.Errorf("commit %s/%s not found", canonicalCommit.Repo.Name, canonicalCommit.ID)
	}
	if diffInfo.Finished != nil {
		return fmt.Errorf("commit %s/%s has already been finished", canonicalCommit.Repo.Name, canonicalCommit.ID)
	}
	addDirs(diffInfo, file)
	_append, ok := diffInfo.Appends[path.Clean(file.Path)]
	if !ok {
		_append = &pfs.Append{}
		if diffInfo.ParentCommit != nil {
			_append.LastRef = d.lastRef(
				pfsutil.NewFile(
					diffInfo.ParentCommit.Repo.Name,
					diffInfo.ParentCommit.ID,
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

func (d *driver) GetFile(file *pfs.File, filterShard *pfs.Shard, offset int64, size int64, from *pfs.Commit, shard uint64) (io.ReadCloser, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	fileInfo, blockRefs, err := d.inspectFile(file, filterShard, shard, from)
	if err != nil {
		return nil, err
	}
	if fileInfo.FileType == pfs.FileType_FILE_TYPE_DIR {
		return nil, fmt.Errorf("file %s/%s/%s is directory", file.Commit.Repo.Name, file.Commit.ID, file.Path)
	}
	blockClient, err := d.getBlockClient()
	if err != nil {
		return nil, err
	}
	return newFileReader(blockClient, blockRefs, offset, size), nil
}

func (d *driver) InspectFile(file *pfs.File, filterShard *pfs.Shard, from *pfs.Commit, shard uint64) (*pfs.FileInfo, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	fileInfo, _, err := d.inspectFile(file, filterShard, shard, from)
	return fileInfo, err
}

func (d *driver) ListFile(file *pfs.File, filterShard *pfs.Shard, from *pfs.Commit, shard uint64) ([]*pfs.FileInfo, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	fileInfo, _, err := d.inspectFile(file, filterShard, shard, from)
	if err != nil {
		return nil, err
	}
	if fileInfo.FileType == pfs.FileType_FILE_TYPE_REGULAR {
		return []*pfs.FileInfo{fileInfo}, nil
	}
	var result []*pfs.FileInfo
	for _, child := range fileInfo.Children {
		fileInfo, _, err := d.inspectFile(child, filterShard, shard, from)
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
	diffInfos := make(diffMap)
	dags := make(map[string]*dag.DAG)
	for {
		diffInfo, err := listDiffClient.Recv()
		if err != nil && err != io.EOF {
			return err
		}
		if err == io.EOF {
			break
		}
		if _, ok := dags[diffInfo.Diff.Commit.Repo.Name]; !ok {
			dags[diffInfo.Diff.Commit.Repo.Name] = dag.NewDAG(nil)
		}
		dags[diffInfo.Diff.Commit.Repo.Name].NewNode(diffInfo.Diff.Commit.ID, []string{diffInfo.ParentCommit.ID})
		if err := diffInfos.insert(diffInfo); err != nil {
			return err
		}
	}
	for repoName, dag := range dags {
		if ghosts := dag.Ghosts(); len(ghosts) != 0 {
			return fmt.Errorf("error adding shard %d, repo %s has ghost commits: %+v", shard, repoName, ghosts)
		}
	}
	d.lock.Lock()
	defer d.lock.Unlock()
	for repoName, dag := range dags {
		for _, commitID := range dag.Sorted() {
			if _, ok := d.diffs[repoName]; !ok {
				d.createRepoState(pfsutil.NewRepo(repoName))
			}
			if diffInfo, ok := diffInfos.get(pfsutil.NewDiff(repoName, commitID, shard)); ok {
				if err := d.insertDiffInfo(diffInfo); err != nil {
					return err
				}
			} else {
				return fmt.Errorf("diff %s/%s/%d not found; this is likely a bug", repoName, commitID, shard)
			}
		}
	}
	return nil
}

func (d *driver) DeleteShard(shard uint64) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	for _, shardMap := range d.diffs {
		delete(shardMap, shard)
	}
	return nil
}

func (d *driver) inspectRepo(repo *pfs.Repo, shards map[uint64]bool) (*pfs.RepoInfo, error) {
	result := &pfs.RepoInfo{
		Repo: repo,
	}
	_, ok := d.diffs[repo.Name]
	if !ok {
		return nil, fmt.Errorf("repo %s not found", repo.Name)
	}
	for shard := range shards {
		diffInfos, ok := d.diffs[repo.Name][shard]
		if !ok {
			return nil, fmt.Errorf("repo %s not found", repo.Name)
		}
		for _, diffInfo := range diffInfos {
			diffInfo := diffInfo
			if diffInfo.Diff.Commit.ID == "" {
				result.Created = diffInfo.Finished
			}
			result.SizeBytes += diffInfo.SizeBytes
		}
	}
	return result, nil
}

func (d *driver) inspectCommit(commit *pfs.Commit, shards map[uint64]bool) (*pfs.CommitInfo, error) {
	var commitInfos []*pfs.CommitInfo
	canonicalCommit, err := d.canonicalCommit(commit)
	if err != nil {
		return nil, err
	}
	for shard := range shards {
		var diffInfo *pfs.DiffInfo
		var ok bool
		commitInfo := &pfs.CommitInfo{Commit: canonicalCommit}
		diff := pfsutil.NewDiff(canonicalCommit.Repo.Name, canonicalCommit.ID, shard)
		if diffInfo, ok = d.diffs.get(diff); !ok {
			return nil, fmt.Errorf("commit %s/%s not found", canonicalCommit.Repo.Name, canonicalCommit.ID)
		}
		if diffInfo.Finished == nil {
			commitInfo.CommitType = pfs.CommitType_COMMIT_TYPE_WRITE
		} else {
			commitInfo.CommitType = pfs.CommitType_COMMIT_TYPE_READ
		}
		commitInfo.Branch = diffInfo.Branch
		commitInfo.ParentCommit = diffInfo.ParentCommit
		commitInfo.Started = diffInfo.Started
		commitInfo.Finished = diffInfo.Finished
		commitInfo.SizeBytes = diffInfo.SizeBytes
		commitInfos = append(commitInfos, commitInfo)
	}
	commitInfo := pfs.ReduceCommitInfos(commitInfos)
	if len(commitInfo) < 1 {
		// we should have caught this above
		return nil, fmt.Errorf("commit %s/%s not found", canonicalCommit.Repo.Name, canonicalCommit.ID)
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

func (d *driver) inspectFile(file *pfs.File, filterShard *pfs.Shard, shard uint64, from *pfs.Commit) (*pfs.FileInfo, []*pfs.BlockRef, error) {
	fileInfo := &pfs.FileInfo{File: file}
	var blockRefs []*pfs.BlockRef
	children := make(map[string]bool)
	commit, err := d.canonicalCommit(file.Commit)
	if err != nil {
		return nil, nil, err
	}
	for commit != nil && (from == nil || commit.ID != from.ID) {
		diffInfo, ok := d.diffs.get(pfsutil.NewDiff(commit.Repo.Name, commit.ID, shard))
		if !ok {
			return nil, nil, fmt.Errorf("diff %s/%s not found", commit.Repo.Name, commit.ID)
		}
		if diffInfo.Finished == nil {
			commit = diffInfo.ParentCommit
			continue
		}
		if _append, ok := diffInfo.Appends[path.Clean(file.Path)]; ok {
			if len(_append.BlockRefs) > 0 {
				if fileInfo.FileType == pfs.FileType_FILE_TYPE_DIR {
					return nil, nil,
						fmt.Errorf("mixed dir and regular file %s/%s/%s, (this is likely a bug)", file.Commit.Repo.Name, file.Commit.ID, file.Path)
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
						fmt.Errorf("mixed dir and regular file %s/%s/%s, (this is likely a bug)", file.Commit.Repo.Name, file.Commit.ID, file.Path)
				}
				fileInfo.FileType = pfs.FileType_FILE_TYPE_DIR
				for child := range _append.Children {
					if !children[child] {
						fileInfo.Children = append(
							fileInfo.Children,
							pfsutil.NewFile(commit.Repo.Name, commit.ID, child),
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
		diffInfo, _ := d.diffs.get(pfsutil.NewDiff(commit.Repo.Name, commit.ID, shard))
		if _, ok := diffInfo.Appends[path.Clean(file.Path)]; ok {
			return commit
		}
		commit = diffInfo.ParentCommit
	}
	return nil
}

func (d *driver) createRepoState(repo *pfs.Repo) {
	if _, ok := d.diffs[repo.Name]; ok {
		return // this function is idempotent
	}
	d.diffs[repo.Name] = make(map[uint64]map[string]*pfs.DiffInfo)
	d.dags[repo.Name] = dag.NewDAG(nil)
	d.branches[repo.Name] = make(map[string]string)
}

// canonicalCommit finds the canonical way of referring to a commit
func (d *driver) canonicalCommit(commit *pfs.Commit) (*pfs.Commit, error) {
	if _, ok := d.branches[commit.Repo.Name]; !ok {
		return nil, fmt.Errorf("repo %s not found", commit.Repo.Name)
	}
	if commitID, ok := d.branches[commit.Repo.Name][commit.ID]; ok {
		return pfsutil.NewCommit(commit.Repo.Name, commitID), nil
	}
	return commit, nil
}

// branchParent finds the parent that should be used for a new commit being started on a branch
func (d *driver) branchParent(commit *pfs.Commit, branch string) (*pfs.Commit, error) {
	// canonicalCommit is the head of branch
	canonicalCommit, err := d.canonicalCommit(pfsutil.NewCommit(commit.Repo.Name, branch))
	if err != nil {
		return nil, err
	}
	if canonicalCommit.ID == branch {
		// first commit on this branch, return nil
		return nil, nil
	}
	if canonicalCommit.ID == commit.ID {
		// this commit is the head of branch
		// that's because this isn't the first shard of this commit we've seen
		for _, commitToDiffInfo := range d.diffs[commit.Repo.Name] {
			if diffInfo, ok := commitToDiffInfo[commit.ID]; ok {
				return diffInfo.ParentCommit, nil
			}
		}
		// reaching this code means that canonicalCommit resolved the branch to
		// a commit we've never seen (on any shard) which indicates a bug
		// elsewhere
		return nil, fmt.Errorf("unreachable")
	}
	return canonicalCommit, nil
}

func (d *driver) insertDiffInfo(diffInfo *pfs.DiffInfo) error {
	commit := diffInfo.Diff.Commit
	updateIndexes := true
	for _, commitToDiffInfo := range d.diffs[commit.Repo.Name] {
		if _, ok := commitToDiffInfo[diffInfo.Diff.Commit.ID]; ok {
			// we've already seen this diff, no need to update indexes
			updateIndexes = false
		}
	}
	if err := d.diffs.insert(diffInfo); err != nil {
		return err
	}
	if updateIndexes {
		if diffInfo.Branch != "" {
			if _, ok := d.diffs[commit.Repo.Name][diffInfo.Diff.Shard][diffInfo.Branch]; ok {
				return fmt.Errorf("branch %s conflicts with commit of the same name", diffInfo.Branch)
			}
			d.branches[commit.Repo.Name][diffInfo.Branch] = commit.ID
		}
		if diffInfo.ParentCommit != nil {
			d.dags[commit.Repo.Name].NewNode(commit.ID, []string{diffInfo.ParentCommit.ID})
		} else {
			d.dags[commit.Repo.Name].NewNode(commit.ID, nil)
		}
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
	diffInfo, ok := commitMap[diff.Commit.ID]
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
	if _, ok = commitMap[diff.Commit.ID]; ok {
		return fmt.Errorf("commit %s/%s already exists", diff.Commit.Repo.Name, diff.Commit.ID)
	}
	commitMap[diff.Commit.ID] = diffInfo
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
	diffInfo := commitMap[diff.Commit.ID]
	delete(commitMap, diff.Commit.ID)
	return diffInfo
}
