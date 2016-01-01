package obj

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"path"
	"sort"
	"strings"
	"sync"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pfs/pfsutil"
	"golang.org/x/net/context"
)

var (
	blockSize = 128 * 1024 * 1024 // 128 Megabytes
)

type driver struct {
	driveClient drive.APIClient
	started     diffMap
	finished    diffMap
	// branches indexes commits with no children
	branches diffMap
	lock     sync.RWMutex
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
	d.branches[repo.Name] = make(map[uint64]map[string]*drive.DiffInfo)
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
			LastRefs:     make(map[string]*pfs.Commit),
		}
		if err := d.started.insert(diffInfo); err != nil {
			return err
		}
		if err := d.branches.insert(diffInfo); err != nil {
			return err
		}
		if parent != nil {
			d.branches.pop(&drive.Diff{
				Commit: parent,
				Shard:  shard,
			})
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
			if _, err := d.driveClient.CreateDiff(
				context.Background(),
				&drive.CreateDiffRequest{
					Diff:         diffInfo.Diff,
					ParentCommit: diffInfo.ParentCommit,
					Appends:      diffInfo.Appends,
					LastRefs:     diffInfo.LastRefs,
					NewFiles:     diffInfo.NewFiles,
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

func (d *driver) ListCommit(repos []*pfs.Repo, fromCommit []*pfs.Commit, shards map[uint64]bool) ([]*pfs.CommitInfo, error) {
	breakCommitIds := make(map[string]bool)
	for _, commit := range fromCommit {
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
			for commitID := range d.branches[repo.Name][shard] {
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
	var blockRefs []*drive.BlockRef
	scanner := bufio.NewScanner(reader)
	scanner.Split(blockSplitFunc)
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
				return
			}
			blockRef.Block = block
		}()
	}
	wg.Wait()
	if err := scanner.Err(); err != nil {
		return err
	}
	if loopErr != nil {
		return loopErr
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
	blockRefsMsg, ok := diffInfo.Appends[file.Path]
	if !ok {
		blockRefsMsg = &drive.BlockRefs{}
		diffInfo.Appends[file.Path] = blockRefsMsg
	}
	blockRefsMsg.BlockRef = append(blockRefsMsg.BlockRef, blockRefs...)
	d.addFileIndexes(diffInfo, file, shard)
	return nil
}

func (d *driver) MakeDirectory(file *pfs.File, shards map[uint64]bool) error {
	return nil
}

func (d *driver) GetFile(file *pfs.File, shard uint64) (drive.ReaderAtCloser, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	blockRefs, isDir, err := d.fileBlockRefsOrDir(file, shard)
	if err != nil {
		return nil, err
	}
	if isDir {
		return nil, fmt.Errorf("file %s/%s/%s is directory", file.Commit.Repo.Name, file.Commit.Id, file.Path)
	}
	return newFileReader(d.driveClient, blockRefs), nil
}

func (d *driver) InspectFile(file *pfs.File, shard uint64) (*pfs.FileInfo, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.inspectFile(file, shard)
}

func (d *driver) ListFile(file *pfs.File, shard uint64) ([]*pfs.FileInfo, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	fileInfos := make(map[string]*pfs.FileInfo)
	commit := file.Commit
	for commit != nil {
		diffInfo, read, ok := d.getDiffInfo(&drive.Diff{
			Commit: commit,
			Shard:  shard,
		})
		if !ok {
			return nil, fmt.Errorf("diff %s/%s not found", commit.Repo.Name, commit.Id)
		}
		commit = diffInfo.ParentCommit
		if !read {
			// don't list files for read commits
			continue
		}
		for _, path := range diffInfo.NewFiles[sort.SearchStrings(diffInfo.NewFiles, file.Path):] {
			if path = relevantPath(file, path); path != "" {
				if _, ok := fileInfos[path]; ok {
					continue
				}
				fileInfo, err := d.inspectFile(&pfs.File{
					Commit: diffInfo.Diff.Commit,
					Path:   path,
				}, shard)
				if err != nil {
					return nil, err
				}
				fileInfos[path] = fileInfo
			}
		}
	}
	var result []*pfs.FileInfo
	for _, fileInfo := range fileInfos {
		result = append(result, fileInfo)
	}
	return result, nil
}

func relevantPath(file *pfs.File, foundPath string) string {
	if foundPath == file.Path {
		return foundPath
	}
	if strings.HasPrefix(foundPath, file.Path+"/") {
		return path.Join(
			file.Path,
			strings.Split(
				strings.TrimPrefix(foundPath, file.Path+"/"),
				"/",
			)[0],
		)
	}
	return ""
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
	commitInfo := pfs.Reduce(commitInfos)
	if len(commitInfo) < 1 {
		return nil, fmt.Errorf("commit %s/%s not found", commit.Repo.Name, commit.Id)
	}
	if len(commitInfo) > 1 {
		return nil, fmt.Errorf("multiple commitInfos, (this is likely a bug)")
	}
	return commitInfo[0], nil
}

func (d *driver) fileBlockRefsOrDir(file *pfs.File, shard uint64) (_ []*drive.BlockRef, isDir bool, _ error) {
	var result []*drive.BlockRef
	commit := file.Commit
	writeable := false
	for commit != nil {
		diffInfo, read, ok := d.getDiffInfo(&drive.Diff{
			Commit: commit,
			Shard:  shard,
		})
		log.Printf("fileBlockRefsOrDir\nfile: %+v\ndiffInfo: %+v\n", file, diffInfo)
		if !ok {
			return nil, false, fmt.Errorf("diff %s/%s not found", commit.Repo.Name, commit.Id)
		}
		if !read {
			writeable = true
		}
		iNewFile := sort.SearchStrings(diffInfo.NewFiles, file.Path)
		if blockRefsMsg, ok := diffInfo.Appends[file.Path]; ok {
			result = append(blockRefsMsg.BlockRef, result...)
		} else if len(diffInfo.NewFiles) > iNewFile && strings.HasPrefix(diffInfo.NewFiles[iNewFile], file.Path+"/") {
			return nil, true, nil
		}
		if lastRef, ok := diffInfo.LastRefs[file.Path]; ok {
			log.Printf("Using lastRef.")
			commit = lastRef
			continue
		}
		commit = diffInfo.ParentCommit
	}
	if result == nil && !writeable {
		return nil, false, fmt.Errorf("file %s/%s/%s not found", file.Commit.Repo.Name, file.Commit.Id, file.Path)
	}
	log.Printf("result: %+v", result)
	return result, false, nil
}

func (d *driver) inspectFile(file *pfs.File, shard uint64) (*pfs.FileInfo, error) {
	result := &pfs.FileInfo{File: file}
	blockRefs, isDir, err := d.fileBlockRefsOrDir(file, shard)
	if err != nil {
		return nil, err
	}
	if isDir {
		result.FileType = pfs.FileType_FILE_TYPE_DIR
	}
	result.FileType = pfs.FileType_FILE_TYPE_REGULAR
	for _, blockRef := range blockRefs {
		result.SizeBytes += (blockRef.Range.Upper - blockRef.Range.Lower)
	}
	return result, nil
}

// addFileIndexes fills in some in memory indexes we use
func (d *driver) addFileIndexes(diffInfo *drive.DiffInfo, file *pfs.File, shard uint64) {
	commit := diffInfo.ParentCommit
	for commit != nil {
		ancestorDiffInfo, _ := d.finished.get(&drive.Diff{
			Commit: commit,
			Shard:  shard,
		})
		if _, ok := ancestorDiffInfo.Appends[file.Path]; ok {
			diffInfo.LastRefs[file.Path] = commit
			return
		}
		commit = ancestorDiffInfo.ParentCommit
	}
	i := sort.SearchStrings(diffInfo.NewFiles, file.Path)
	diffInfo.NewFiles = append(diffInfo.NewFiles, "")
	copy(diffInfo.NewFiles[i+1:], diffInfo.NewFiles[i:])
	diffInfo.NewFiles[i] = file.Path
}

func blockSplitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if len(data) < blockSize && !atEOF {
		return 0, nil, nil
	}
	if len(data) == 0 && atEOF {
		return 0, nil, nil
	}
	if len(data) < blockSize && atEOF {
		return len(data), data, nil
	}
	for i := len(data) - 1; i >= 0; i-- {
		if data[i] == '\n' {
			return i + 1, data[:i+1], nil
		}
	}
	return 0, nil, fmt.Errorf("line too long")
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
