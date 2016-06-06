package drive

import (
	"fmt"
	"io"
	"path"
	"regexp"
	"sync"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/dag"
	"github.com/pachyderm/pachyderm/src/server/pkg/metrics"
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
	// used for signaling the completion (i.e. finishing) of a commit
	commitConds map[string]*sync.Cond
}

func newDriver(blockAddress string) (Driver, error) {
	return &driver{
		blockAddress:    blockAddress,
		blockClient:     nil,
		blockClientOnce: sync.Once{},
		diffs:           make(diffMap),
		dags:            make(map[string]*dag.DAG),
		branches:        make(map[string]map[string]string),
		lock:            sync.RWMutex{},
		commitConds:     make(map[string]*sync.Cond),
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

func validateRepoName(name string) error {
	match, _ := regexp.MatchString("^[a-zA-Z0-9_]+$", name)

	if !match {
		return fmt.Errorf("Repo name (%v) invalid. Only alphanumeric and underscore characters allowed.", name)
	}

	return nil
}

func (d *driver) CreateRepo(repo *pfs.Repo, created *google_protobuf.Timestamp,
	provenance []*pfs.Repo, shards map[uint64]bool) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	if _, ok := d.diffs[repo.Name]; ok {
		return fmt.Errorf("repo %s exists", repo.Name)
	}
	if err := validateRepoName(repo.Name); err != nil {
		return err
	}
	for _, provRepo := range provenance {
		if _, err := d.inspectRepo(provRepo, shards); err != nil {
			return nil
		}
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
			Diff:     client.NewDiff(repo.Name, "", shard),
			Finished: created,
		}
		for _, provRepo := range provenance {
			diffInfo.Provenance = append(diffInfo.Provenance, client.NewCommit(provRepo.Name, ""))
		}
		if err := d.diffs.insert(diffInfo); err != nil {
			return err
		}
		go func() {
			defer wg.Done()
			if _, err := blockClient.CreateDiff(context.Background(), diffInfo); err != nil {
				select {
				case errCh <- err:
				default:
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

func (d *driver) ListRepo(provenance []*pfs.Repo, shards map[uint64]bool) ([]*pfs.RepoInfo, error) {
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
				default:
				}
			}
			provSet := repoSet(repoInfo.Provenance)
			for _, repo := range provenance {
				if !provSet[repo.Name] {
					// this repo doesn't match the provenance we want, ignore it
					return
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
	}
	return result, nil
}

func (d *driver) DeleteRepo(repo *pfs.Repo, shards map[uint64]bool) error {
	// Make sure that this repo is not the provenance of any other repo
	repoInfos, err := d.ListRepo([]*pfs.Repo{repo}, shards)
	if err != nil {
		return err
	}

	var diffInfos []*pfs.DiffInfo
	err = func() error {
		d.lock.Lock()
		defer d.lock.Unlock()
		if _, ok := d.diffs[repo.Name]; !ok {
			return pfsserver.NewErrRepoNotFound(repo.Name)
		}
		if len(repoInfos) > 0 {
			var repoNames []string
			for _, repoInfo := range repoInfos {
				repoNames = append(repoNames, repoInfo.Repo.Name)
			}
			return fmt.Errorf("cannot delete repo %v; it's the provenance of the following repos: %v", repo.Name, repoNames)
		}

		for shard := range shards {
			for _, diffInfo := range d.diffs[repo.Name][shard] {
				diffInfos = append(diffInfos, diffInfo)
			}
		}
		delete(d.diffs, repo.Name)
		return nil
	}()
	if err != nil {
		return err
	}

	blockClient, err := d.getBlockClient()
	if err != nil {
		return err
	}
	errCh := make(chan error, 1)
	var wg sync.WaitGroup
	for _, diffInfo := range diffInfos {
		diffInfo := diffInfo
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := blockClient.DeleteDiff(
				context.Background(),
				&pfs.DeleteDiffRequest{Diff: diffInfo.Diff},
			); err != nil {
				select {
				case errCh <- err:
				default:
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

func (d *driver) StartCommit(repo *pfs.Repo, commitID string, parentID string, branch string,
	started *google_protobuf.Timestamp, provenance []*pfs.Commit, shards map[uint64]bool) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	// make sure that the parent commit exists
	if parentID != "" {
		_, err := d.inspectCommit(client.NewCommit(repo.Name, parentID), shards)
		if err != nil {
			return err
		}
	}

	for shard := range shards {
		if len(provenance) != 0 {
			diffInfo, ok := d.diffs.get(client.NewDiff(repo.Name, "", shard))
			if !ok {
				return pfsserver.NewErrRepoNotFound(repo.Name)
			}
			provRepos := repoSetFromCommits(diffInfo.Provenance)
			for _, provCommit := range provenance {
				if !provRepos[provCommit.Repo.Name] {
					return fmt.Errorf("cannot use %s/%s as provenance, %s is not provenance of %s",
						provCommit.Repo.Name, provCommit.ID, provCommit.Repo.Name, repo.Name)
				}
			}
		}
		diffInfo := &pfs.DiffInfo{
			Diff:       client.NewDiff(repo.Name, commitID, shard),
			Started:    started,
			Appends:    make(map[string]*pfs.Append),
			Branch:     branch,
			Provenance: provenance,
		}
		if branch != "" {
			parentCommit, err := d.branchParent(client.NewCommit(repo.Name, commitID), branch)
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
			diffInfo.ParentCommit = client.NewCommit(repo.Name, parentID)
		}
		if err := d.insertDiffInfo(diffInfo); err != nil {
			return err
		}
	}
	d.commitConds[commitID] = sync.NewCond(&d.lock)
	return nil
}

// FinishCommit blocks until its parent has been finished/cancelled
func (d *driver) FinishCommit(commit *pfs.Commit, finished *google_protobuf.Timestamp, cancel bool, shards map[uint64]bool) error {
	canonicalCommit, err := d.canonicalCommit(commit)
	if err != nil {
		return err
	}
	// closure so we can defer Unlock
	var diffInfos []*pfs.DiffInfo
	if err := func() error {
		d.lock.Lock()
		defer d.lock.Unlock()
		for shard := range shards {
			diffInfo, ok := d.diffs.get(client.NewDiff(canonicalCommit.Repo.Name, canonicalCommit.ID, shard))
			if !ok {
				return pfsserver.NewErrCommitNotFound(canonicalCommit.Repo.Name, canonicalCommit.ID)
			}
			if diffInfo.ParentCommit != nil {
				parentDiffInfo, ok := d.diffs.get(client.NewDiff(canonicalCommit.Repo.Name, diffInfo.ParentCommit.ID, shard))
				if !ok {
					return pfsserver.NewErrParentCommitNotFound(canonicalCommit.Repo.Name, diffInfo.ParentCommit.ID)
				}
				// Wait for parent to finish
				for parentDiffInfo.Finished == nil {
					cond, ok := d.commitConds[diffInfo.ParentCommit.ID]
					if !ok {
						return fmt.Errorf("parent commit %s/%s was not finished but a corresponding conditional variable could not be found; this is likely a bug", canonicalCommit.Repo.Name, diffInfo.ParentCommit.ID)
					}
					cond.Wait()
				}

				diffInfo.Cancelled = parentDiffInfo.Cancelled
			}
			diffInfo.Finished = finished
			for _, _append := range diffInfo.Appends {
				coalesceHandles(_append)
			}
			diffInfo.Cancelled = diffInfo.Cancelled || cancel
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
	errCh := make(chan error, 1)
	for _, diffInfo := range diffInfos {
		diffInfo := diffInfo
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := blockClient.CreateDiff(context.Background(), diffInfo); err != nil {
				select {
				case errCh <- err:
				default:
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

	d.lock.Lock()
	defer d.lock.Unlock()
	cond, ok := d.commitConds[canonicalCommit.ID]
	if !ok {
		return fmt.Errorf("could not found a conditional variable to signal commit completion; this is likely a bug")
	}
	cond.Broadcast()
	delete(d.commitConds, canonicalCommit.ID)

	return nil
}

func (d *driver) InspectCommit(commit *pfs.Commit, shards map[uint64]bool) (*pfs.CommitInfo, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.inspectCommit(commit, shards)
}

func (d *driver) ListCommit(repos []*pfs.Repo, commitType pfs.CommitType, fromCommit []*pfs.Commit,
	provenance []*pfs.Commit, all bool, shards map[uint64]bool) ([]*pfs.CommitInfo, error) {
	repoSet := repoSet(repos)
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
			return nil, pfsserver.NewErrRepoNotFound(repo.Name)
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
				commit = commitInfo.ParentCommit
				if commitInfo.Cancelled && !all {
					continue
				}
				if !MatchProvenance(provenance, commitInfo.Provenance) {
					continue
				}
				if commitType != pfs.CommitType_COMMIT_TYPE_NONE &&
					commitType != commitInfo.CommitType {
					continue
				}
				result = append(result, commitInfo)
			}
		}
	}
	return result, nil
}

func MatchProvenance(want []*pfs.Commit, have []*pfs.Commit) bool {
	repoToCommit := make(map[string]*pfs.Commit)
	for _, haveCommit := range have {
		repoToCommit[haveCommit.Repo.Name] = haveCommit
	}
	for _, wantCommit := range want {
		haveCommit, ok := repoToCommit[wantCommit.Repo.Name]
		if !ok || wantCommit.ID != haveCommit.ID {
			return false
		}
	}
	return true
}

func (d *driver) ListBranch(repo *pfs.Repo, shards map[uint64]bool) ([]*pfs.CommitInfo, error) {
	var result []*pfs.CommitInfo

	_, ok := d.branches[repo.Name]
	if !ok {
		return nil, pfsserver.NewErrRepoNotFound(repo.Name)
	}

	for commitID := range d.branches[repo.Name] {
		commitInfo, err := d.inspectCommit(client.NewCommit(repo.Name, commitID), shards)
		if err != nil {
			return nil, err
		}
		result = append(result, commitInfo)
	}
	return result, nil
}

func (d *driver) DeleteCommit(commit *pfs.Commit, shards map[uint64]bool) error {
	return fmt.Errorf("DeleteCommit is not implemented")
}

func (d *driver) PutFile(file *pfs.File, handle string,
	delimiter pfs.Delimiter, shard uint64, reader io.Reader) (retErr error) {
	blockClient, err := d.getBlockClient()
	if err != nil {
		return err
	}
	_client := client.APIClient{BlockAPIClient: blockClient}
	blockRefs, err := _client.PutBlock(delimiter, reader)
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

	fileType, err := d.getFileType(file, shard)
	if err != nil {
		return err
	}

	if fileType == pfs.FileType_FILE_TYPE_DIR {
		return fmt.Errorf("%s is a directory", file.Path)
	}

	canonicalCommit, err := d.canonicalCommit(file.Commit)
	if err != nil {
		return err
	}
	diffInfo, ok := d.diffs.get(client.NewDiff(canonicalCommit.Repo.Name, canonicalCommit.ID, shard))
	if !ok {
		// This is a weird case since the commit existed above, it means someone
		// deleted the commit while the above code was running
		return pfsserver.NewErrCommitNotFound(canonicalCommit.Repo.Name, canonicalCommit.ID)
	}
	if diffInfo.Finished != nil {
		return fmt.Errorf("commit %s/%s has already been finished", canonicalCommit.Repo.Name, canonicalCommit.ID)
	}
	d.addDirs(diffInfo, file, shard)
	_append, ok := diffInfo.Appends[path.Clean(file.Path)]
	if !ok {
		_append = &pfs.Append{
			Handles:  make(map[string]*pfs.BlockRefs),
			FileType: pfs.FileType_FILE_TYPE_REGULAR,
		}
	}
	if diffInfo.ParentCommit != nil {
		_append.LastRef = d.lastRef(
			client.NewFile(diffInfo.ParentCommit.Repo.Name, diffInfo.ParentCommit.ID, file.Path),
			shard,
		)
	}
	diffInfo.Appends[path.Clean(file.Path)] = _append
	if handle == "" {
		_append.BlockRefs = append(_append.BlockRefs, blockRefs.BlockRef...)
	} else {
		handleBlockRefs, ok := _append.Handles[handle]
		if !ok {
			handleBlockRefs = &pfs.BlockRefs{}
			_append.Handles[handle] = handleBlockRefs
		}
		handleBlockRefs.BlockRef = append(handleBlockRefs.BlockRef, blockRefs.BlockRef...)
	}
	for _, blockRef := range blockRefs.BlockRef {
		diffInfo.SizeBytes += blockRef.Range.Upper - blockRef.Range.Lower
	}
	return nil
}

func (d *driver) MakeDirectory(file *pfs.File, shard uint64) (retErr error) {
	defer func() {
		if retErr == nil {
			metrics.AddFiles(1)
		}
	}()
	d.lock.Lock()
	defer d.lock.Unlock()

	fileType, err := d.getFileType(file, shard)
	if err != nil {
		return err
	}

	if fileType == pfs.FileType_FILE_TYPE_REGULAR {
		return fmt.Errorf("%s already exists and is a file", file.Path)
	} else if fileType == pfs.FileType_FILE_TYPE_DIR {
		return nil
	}

	canonicalCommit, err := d.canonicalCommit(file.Commit)
	if err != nil {
		return err
	}
	diffInfo, ok := d.diffs.get(client.NewDiff(canonicalCommit.Repo.Name, canonicalCommit.ID, shard))
	if !ok {
		return pfsserver.NewErrCommitNotFound(canonicalCommit.Repo.Name, canonicalCommit.ID)
	}
	if diffInfo.Finished != nil {
		return fmt.Errorf("commit %s/%s has already been finished", canonicalCommit.Repo.Name, canonicalCommit.ID)
	}
	d.addDirs(diffInfo, file, shard)
	_append, ok := diffInfo.Appends[path.Clean(file.Path)]
	if !ok {
		_append = &pfs.Append{FileType: pfs.FileType_FILE_TYPE_DIR}
	}
	if diffInfo.ParentCommit != nil {
		_append.LastRef = d.lastRef(
			client.NewFile(
				diffInfo.ParentCommit.Repo.Name,
				diffInfo.ParentCommit.ID,
				file.Path,
			),
			shard,
		)
	}
	diffInfo.Appends[path.Clean(file.Path)] = _append
	// The fact that this is a directory is signified by setting Children
	// to non-nil
	_append.Children = make(map[string]bool)
	return nil
}

func (d *driver) GetFile(file *pfs.File, filterShard *pfs.Shard, offset int64, size int64, from *pfs.Commit, shard uint64, unsafe bool) (io.ReadCloser, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	fileInfo, blockRefs, err := d.inspectFile(file, filterShard, shard, from, false, unsafe)
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

func (d *driver) InspectFile(file *pfs.File, filterShard *pfs.Shard, from *pfs.Commit, shard uint64, unsafe bool) (*pfs.FileInfo, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	fileInfo, _, err := d.inspectFile(file, filterShard, shard, from, false, unsafe)
	return fileInfo, err
}

func (d *driver) ListFile(file *pfs.File, filterShard *pfs.Shard, from *pfs.Commit, shard uint64, recurse bool, unsafe bool) ([]*pfs.FileInfo, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	fileInfo, _, err := d.inspectFile(file, filterShard, shard, from, false, unsafe)
	if err != nil {
		return nil, err
	}
	if fileInfo.FileType == pfs.FileType_FILE_TYPE_REGULAR {
		return []*pfs.FileInfo{fileInfo}, nil
	}
	var result []*pfs.FileInfo
	for _, child := range fileInfo.Children {
		fileInfo, _, err := d.inspectFile(child, filterShard, shard, from, recurse, unsafe)
		_, ok := err.(*pfsserver.ErrFileNotFound)
		if err != nil && !ok {
			return nil, err
		}
		if ok {
			// how can a listed child return not found?
			// regular files without any blocks in this shard count as not found
			continue
		}
		result = append(result, fileInfo)
	}
	return result, nil
}

func (d *driver) DeleteFile(file *pfs.File, shard uint64) error {
	d.lock.RLock()
	// We don't want to be able to delete files that are only added in the current
	// commit, which is why we set unsafe to false.
	fileInfo, _, err := d.inspectFile(file, nil, shard, nil, false, false)
	if err != nil {
		d.lock.RUnlock()
		return err
	}
	d.lock.RUnlock()

	if fileInfo.FileType == pfs.FileType_FILE_TYPE_DIR {
		fileInfos, err := d.ListFile(file, nil, nil, shard, false, false)
		if err != nil {
			return err
		}

		for _, info := range fileInfos {
			// We are deleting the file from the current commit, not whatever
			// commit they were last modified in
			info.File.Commit = file.Commit
			if err := d.DeleteFile(info.File, shard); err != nil {
				return err
			}
		}
	}

	return d.deleteFile(file, shard)
}

func (d *driver) deleteFile(file *pfs.File, shard uint64) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	canonicalCommit, err := d.canonicalCommit(file.Commit)
	if err != nil {
		return err
	}
	diffInfo, ok := d.diffs.get(client.NewDiff(canonicalCommit.Repo.Name, canonicalCommit.ID, shard))
	if !ok {
		// This is a weird case since the commit existed above, it means someone
		// deleted the commit while the above code was running
		return pfsserver.NewErrCommitNotFound(canonicalCommit.Repo.Name, canonicalCommit.ID)
	}
	if diffInfo.Finished != nil {
		return fmt.Errorf("commit %s/%s has already been finished", canonicalCommit.Repo.Name, canonicalCommit.ID)
	}

	cleanPath := path.Clean(file.Path)
	if _, ok := diffInfo.Appends[cleanPath]; !ok {
		diffInfo.Appends[cleanPath] = &pfs.Append{Handles: make(map[string]*pfs.BlockRefs)}
	}
	diffInfo.Appends[cleanPath].Delete = true
	d.deleteFromDir(diffInfo, file, shard)

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
		if diffInfo.Diff == nil || diffInfo.Diff.Commit == nil || diffInfo.Diff.Commit.Repo == nil {
			return fmt.Errorf("broken diff info: %v; this is likely a bug", diffInfo)
		}
		repoName := diffInfo.Diff.Commit.Repo.Name
		if _, ok := diffInfos[repoName]; !ok {
			diffInfos[repoName] = make(map[uint64]map[string]*pfs.DiffInfo)
			dags[repoName] = dag.NewDAG(nil)
		}
		updateDAG(diffInfo, dags[repoName])
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
			d.createRepoState(client.NewRepo(repoName))
			if diffInfo, ok := diffInfos.get(client.NewDiff(repoName, commitID, shard)); ok {
				if err := d.insertDiffInfo(diffInfo); err != nil {
					return err
				}
				if diffInfo.Finished == nil {
					return fmt.Errorf("diff %s/%s/%d is not finished; this is likely a bug", repoName, commitID, shard)
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

func (d *driver) Dump() {
	d.lock.RLock()
	defer d.lock.RUnlock()
	fmt.Printf("%p.Dump()\n", d)
	for repoName, dag := range d.dags {
		fmt.Printf("%s:\n", repoName)
		for _, commitID := range dag.Sorted() {
			fmt.Printf("\t%s: ", commitID)
			for shard, commitToDiffInfo := range d.diffs[repoName] {
				if _, ok := commitToDiffInfo[commitID]; ok {
					fmt.Printf("%d, ", shard)
				}
			}
			fmt.Printf("\n")
		}
	}
}

// inspectRepo assumes that the lock is being held
func (d *driver) inspectRepo(repo *pfs.Repo, shards map[uint64]bool) (*pfs.RepoInfo, error) {
	result := &pfs.RepoInfo{
		Repo: repo,
	}
	shardToDiffInfo, ok := d.diffs[repo.Name]
	if !ok {
		return nil, pfsserver.NewErrRepoNotFound(repo.Name)
	}
	for shard := range shards {
		diffInfos, ok := shardToDiffInfo[shard]
		if !ok {
			continue
		}
		for _, diffInfo := range diffInfos {
			diffInfo := diffInfo
			if diffInfo.Diff.Commit.ID == "" && result.Created == nil {
				result.Created = diffInfo.Finished
			}
			result.SizeBytes += diffInfo.SizeBytes
		}
	}
	provenance, err := d.fullRepoProvenance(repo, shards)
	if err != nil {
		return nil, err
	}
	result.Provenance = provenance
	return result, nil
}

// fullCommitProvenance recursively computes the provenance of the commit,
// starting with the immediate provenance that was declared when the commit was
// created, then our immediate provenance's immediate provenance etc.
func (d *driver) fullCommitProvenance(commit *pfs.Commit, repoSet map[string]bool,
	shards map[uint64]bool) ([]*pfs.Commit, error) {
	shardToDiffInfo, ok := d.diffs[commit.Repo.Name]
	if !ok {
		return nil, pfsserver.NewErrRepoNotFound(commit.Repo.Name)
	}
	var result []*pfs.Commit
	for shard := range shards {
		diffInfos := shardToDiffInfo[shard]
		if !ok {
			return nil, fmt.Errorf("missing shard %d (this is likely a bug)")
		}
		diffInfo, ok := diffInfos[commit.ID]
		if !ok {
			return nil, fmt.Errorf("missing \"\" diff (this is likely a bug)")
		}
		for _, provCommit := range diffInfo.Provenance {
			if !repoSet[provCommit.Repo.Name] {
				repoSet[provCommit.Repo.Name] = true
				result = append(result, provCommit)
				provCommits, err := d.fullCommitProvenance(provCommit, repoSet, shards)
				if err != nil {
					return nil, err
				}
				result = append(result, provCommits...)
			}
		}
		break // we only need to consider 1 shard
	}
	return result, nil
}

// fullRepoProvenance recursively computes the provenance of a repo
func (d *driver) fullRepoProvenance(repo *pfs.Repo, shards map[uint64]bool) ([]*pfs.Repo, error) {
	provCommits, err := d.fullCommitProvenance(client.NewCommit(repo.Name, ""), make(map[string]bool), shards)
	if err != nil {
		return nil, err
	}
	var result []*pfs.Repo
	for _, provCommit := range provCommits {
		result = append(result, provCommit.Repo)
	}
	return result, nil
}

func (d *driver) commitProvenance(commit *pfs.Commit, shards map[uint64]bool) ([]*pfs.Commit, error) {
	return d.fullCommitProvenance(commit, make(map[string]bool), shards)
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
		diff := client.NewDiff(canonicalCommit.Repo.Name, canonicalCommit.ID, shard)
		if diffInfo, ok = d.diffs.get(diff); !ok {
			return nil, pfsserver.NewErrCommitNotFound(canonicalCommit.Repo.Name, canonicalCommit.ID)
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
		commitInfo.Cancelled = diffInfo.Cancelled
		commitInfos = append(commitInfos, commitInfo)
	}
	commitInfo := pfsserver.ReduceCommitInfos(commitInfos)
	if len(commitInfo) < 1 {
		// we should have caught this above
		return nil, pfsserver.NewErrCommitNotFound(canonicalCommit.Repo.Name, canonicalCommit.ID)
	}
	if len(commitInfo) > 1 {
		return nil, fmt.Errorf("multiple commitInfos, (this is likely a bug)")
	}
	result := commitInfo[0]
	provenance, err := d.commitProvenance(canonicalCommit, shards)
	if err != nil {
		return nil, err
	}
	result.Provenance = provenance
	return commitInfo[0], nil
}

func filterBlockRefs(filterShard *pfs.Shard, blockRefs []*pfs.BlockRef) []*pfs.BlockRef {
	var result []*pfs.BlockRef
	for _, blockRef := range blockRefs {
		if pfsserver.BlockInShard(filterShard, blockRef.Block) {
			result = append(result, blockRef)
		}
	}
	return result
}

func (d *driver) getFileType(file *pfs.File, shard uint64) (pfs.FileType, error) {
	commit, err := d.canonicalCommit(file.Commit)
	if err != nil {
		return pfs.FileType_FILE_TYPE_NONE, err
	}
	for commit != nil {
		diffInfo, ok := d.diffs.get(client.NewDiff(commit.Repo.Name, commit.ID, shard))
		if !ok {
			return pfs.FileType_FILE_TYPE_NONE, pfsserver.NewErrCommitNotFound(commit.Repo.Name, commit.ID)
		}
		if _append, ok := diffInfo.Appends[path.Clean(file.Path)]; ok {
			if _append.Delete {
				break
			} else if len(_append.BlockRefs) > 0 || len(_append.Handles) > 0 {
				return pfs.FileType_FILE_TYPE_REGULAR, nil
			} else {
				return pfs.FileType_FILE_TYPE_DIR, nil
			}
		}
		commit = diffInfo.ParentCommit
	}
	return pfs.FileType_FILE_TYPE_NONE, nil
}

// If recurse is set to true, and if the file being inspected is a directory,
// its children will have the correct sizes.  If recurse is false and the file
// is a directory, its children will have size of 0.
// If unsafe is set to true, you can inspect files in an open commit
func (d *driver) inspectFile(file *pfs.File, filterShard *pfs.Shard, shard uint64, from *pfs.Commit, recurse bool, unsafe bool) (*pfs.FileInfo, []*pfs.BlockRef, error) {
	fileInfo := &pfs.FileInfo{File: file}
	var blockRefs []*pfs.BlockRef
	children := make(map[string]bool)
	deletedChildren := make(map[string]bool)
	commit, err := d.canonicalCommit(file.Commit)
	if err != nil {
		return nil, nil, err
	}
	for commit != nil && (from == nil || commit.ID != from.ID) {
		diffInfo, ok := d.diffs.get(client.NewDiff(commit.Repo.Name, commit.ID, shard))
		if !ok {
			return nil, nil, pfsserver.NewErrCommitNotFound(commit.Repo.Name, commit.ID)
		}
		if !unsafe && diffInfo.Finished == nil {
			commit = diffInfo.ParentCommit
			continue
		}
		if _append, ok := diffInfo.Appends[path.Clean(file.Path)]; ok {
			if _append.FileType == pfs.FileType_FILE_TYPE_NONE && !_append.Delete {
				return nil, nil, fmt.Errorf("the append for %s has file type NONE, this is likely a bug", path.Clean(file.Path))
			}
			if len(_append.BlockRefs) > 0 || len(_append.Handles) > 0 {
				if fileInfo.FileType == pfs.FileType_FILE_TYPE_DIR {
					return nil, nil,
						fmt.Errorf("mixed dir and regular file %s/%s/%s, (this is likely a bug)", file.Commit.Repo.Name, file.Commit.ID, file.Path)
				}
				if fileInfo.FileType == pfs.FileType_FILE_TYPE_NONE {
					// the first time we find out it's a regular file we check
					// the file shard, dirs get returned regardless of sharding,
					// since they might have children from any shard
					if !pfsserver.FileInShard(filterShard, file) {
						return nil, nil, pfsserver.NewErrFileNotFound(file.Path, file.Commit.Repo.Name, file.Commit.ID)
					}
				}
				fileInfo.FileType = pfs.FileType_FILE_TYPE_REGULAR
				filtered := filterBlockRefs(filterShard, _append.BlockRefs)
				for _, handleBlockRefs := range _append.Handles {
					filtered = append(filtered, filterBlockRefs(filterShard, handleBlockRefs.BlockRef)...)
				}
				blockRefs = append(filtered, blockRefs...)
				for _, blockRef := range filtered {
					fileInfo.SizeBytes += (blockRef.Range.Upper - blockRef.Range.Lower)
				}
			} else if _append.Children != nil {
				// With non-nil Children, this Append is for a directory, even if
				// Children is empty.  This is because we sometimes
				// have an Append with an empty children just to signify that
				// this is a directory.
				if fileInfo.FileType == pfs.FileType_FILE_TYPE_REGULAR {
					return nil, nil,
						fmt.Errorf("mixed dir and regular file %s/%s/%s, (this is likely a bug)", file.Commit.Repo.Name, file.Commit.ID, file.Path)
				}
				fileInfo.FileType = pfs.FileType_FILE_TYPE_DIR
				for child, add := range _append.Children {
					if !add {
						deletedChildren[child] = true
						continue
					}

					if !children[child] && !deletedChildren[child] {
						childFile := client.NewFile(commit.Repo.Name, commit.ID, child)
						if pfsserver.FileInShard(filterShard, childFile) {
							fileInfo.Children = append(
								fileInfo.Children,
								client.NewFile(commit.Repo.Name, commit.ID, child),
							)
							if recurse {
								childFileInfo, _, err := d.inspectFile(&pfs.File{
									Commit: file.Commit,
									Path:   child,
								}, filterShard, shard, from, recurse, unsafe)
								if err != nil {
									return nil, nil, err
								}
								fileInfo.SizeBytes += childFileInfo.SizeBytes
							}
						}
					}
					children[child] = true
				}
			}
			// If Delete is true, then everything before this commit is irrelevant
			if _append.Delete {
				break
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
		return nil, nil, pfsserver.NewErrFileNotFound(file.Path, file.Commit.Repo.Name, file.Commit.ID)
	}
	return fileInfo, blockRefs, nil
}

// lastRef assumes the diffInfo file exists in finished
func (d *driver) lastRef(file *pfs.File, shard uint64) *pfs.Commit {
	commit := file.Commit
	for commit != nil {
		diffInfo, _ := d.diffs.get(client.NewDiff(commit.Repo.Name, commit.ID, shard))
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
		return nil, pfsserver.NewErrRepoNotFound(commit.Repo.Name)
	}
	if commitID, ok := d.branches[commit.Repo.Name][commit.ID]; ok {
		return client.NewCommit(commit.Repo.Name, commitID), nil
	}
	return commit, nil
}

// branchParent finds the parent that should be used for a new commit being started on a branch
func (d *driver) branchParent(commit *pfs.Commit, branch string) (*pfs.Commit, error) {
	// canonicalCommit is the head of branch
	canonicalCommit, err := d.canonicalCommit(client.NewCommit(commit.Repo.Name, branch))
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
		updateDAG(diffInfo, d.dags[commit.Repo.Name])
	}
	return nil
}

func updateDAG(diffInfo *pfs.DiffInfo, dag *dag.DAG) {
	if diffInfo.ParentCommit != nil {
		dag.NewNode(diffInfo.Diff.Commit.ID, []string{diffInfo.ParentCommit.ID})
	} else {
		dag.NewNode(diffInfo.Diff.Commit.ID, nil)
	}
}

func (d *driver) addDirs(diffInfo *pfs.DiffInfo, child *pfs.File, shard uint64) {
	childPath := child.Path
	dirPath := path.Dir(childPath)
	for {
		_append, ok := diffInfo.Appends[dirPath]
		if !ok {
			_append = &pfs.Append{FileType: pfs.FileType_FILE_TYPE_DIR}
			diffInfo.Appends[dirPath] = _append
		}
		if _append.Children == nil {
			_append.Children = make(map[string]bool)
		}
		_append.Children[childPath] = true
		if diffInfo.ParentCommit != nil {
			_append.LastRef = d.lastRef(
				client.NewFile(diffInfo.ParentCommit.Repo.Name, diffInfo.ParentCommit.ID, dirPath),
				shard,
			)
		}
		if dirPath == "." {
			break
		}
		childPath = dirPath
		dirPath = path.Dir(childPath)
	}
}

func (d *driver) deleteFromDir(diffInfo *pfs.DiffInfo, child *pfs.File, shard uint64) {
	childPath := child.Path
	dirPath := path.Dir(childPath)

	_append, ok := diffInfo.Appends[dirPath]
	if !ok {
		_append = &pfs.Append{FileType: pfs.FileType_FILE_TYPE_DIR}
		diffInfo.Appends[dirPath] = _append
	}
	if _append.Children == nil {
		_append.Children = make(map[string]bool)
	}
	// Basically, we only set the entry to false if it's not been
	// set to true.  If it's been set to true, that means that there
	// is a PutFile operation in this commit for this very same file,
	// so we don't want to remove the file from the directory.
	if !_append.Children[childPath] {
		_append.Children[childPath] = false
		if diffInfo.ParentCommit != nil {
			_append.LastRef = d.lastRef(
				client.NewFile(diffInfo.ParentCommit.Repo.Name, diffInfo.ParentCommit.ID, dirPath),
				shard,
			)
		}
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

func (r *fileReader) blockRef() *pfs.BlockRef {
	return r.blockRefs[r.index]
}

func (r *fileReader) Read(data []byte) (int, error) {
	if r.reader == nil {
		// skip blocks as long as our offset is past the end of the current block
		for r.offset != 0 && r.index < len(r.blockRefs) && r.offset >= int64(pfsserver.ByteRangeSize(r.blockRef().Range)) {
			r.offset -= int64(pfsserver.ByteRangeSize(r.blockRef().Range))
			r.index++
		}
		if r.index == len(r.blockRefs) {
			return 0, io.EOF
		}
		var err error
		client := client.APIClient{BlockAPIClient: r.blockClient}
		r.reader, err = client.GetBlock(r.blockRef().Block.Hash, uint64(r.offset), uint64(r.size))
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
	if r.size < 0 {
		return 0, fmt.Errorf("read more than we need; this is likely a bug")
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
		return pfsserver.NewErrRepoNotFound(diff.Commit.Repo.Name)
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

func coalesceHandles(_append *pfs.Append) {
	for _, blockRefs := range _append.Handles {
		_append.BlockRefs = append(_append.BlockRefs, blockRefs.BlockRef...)
	}
	_append.Handles = nil
}

func repoSet(repos []*pfs.Repo) map[string]bool {
	result := make(map[string]bool)
	for _, repo := range repos {
		result[repo.Name] = true
	}
	return result
}

func repoSetFromCommits(commits []*pfs.Commit) map[string]bool {
	result := make(map[string]bool)
	for _, commit := range commits {
		result[commit.Repo.Name] = true
	}
	return result
}
