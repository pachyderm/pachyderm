/*

directory structure

  .
  |-- repoName
	  |-- scratch
		  |-- shardNum // the read-only read created on InitRepo, this is where to start branching
      |-- commitID
	      |-- shardNum // this is where subvolumes are

*/

package btrfs

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"go.pachyderm.com/pachyderm/src/pfs"
	"go.pachyderm.com/pachyderm/src/pfs/drive"
	"go.pedge.io/pkg/exec"
	"go.pedge.io/proto/time"
	"go.pedge.io/protolog"
)

const (
	metadataDir  = ".pfs"
	blockDir     = "block"
	repoDir      = "repo"
	writeSuffix  = ".write"
	infoSuffix   = ".info"
	readDirBatch = 100
)

type driver struct {
	rootDir   string
	namespace string
}

func newDriver(rootDir string, namespace string) (*driver, error) {
	driver := &driver{
		rootDir,
		namespace,
	}
	if err := os.MkdirAll(driver.blockDir(), 0700); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(driver.repoDir(), 0700); err != nil {
		_ = os.Remove(driver.blockDir())
		return nil, err
	}
	return driver, nil
}

func (d *driver) CreateRepo(repo *pfs.Repo) error {
	if err := execSubvolumeCreate(d.repoPath(repo)); err != nil && !execSubvolumeExists(d.repoPath(repo)) {
		return err
	}
	return nil
}

func (d *driver) InspectRepo(repo *pfs.Repo, shard uint64) (*pfs.RepoInfo, error) {
	stat, err := os.Stat(d.repoPath(repo))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("repo %s not found", d.repoPath(repo))
		}
		return nil, err
	}
	return &pfs.RepoInfo{
			Repo: repo,
			Created: prototime.TimeToTimestamp(
				stat.ModTime(),
			),
		},
		nil

}

func (d *driver) ListRepo(shard uint64) ([]*pfs.RepoInfo, error) {
	repositories, err := ioutil.ReadDir(d.repoDir())
	if err != nil {
		return nil, err
	}
	var result []*pfs.RepoInfo
	for _, repo := range repositories {
		repoInfo, err := d.InspectRepo(&pfs.Repo{repo.Name()}, shard)
		if err != nil {
			return nil, err
		}
		if repoInfo == nil {
			return nil, fmt.Errorf("repo %s should exist", repo.Name())
		}
		result = append(result, repoInfo)
	}
	return result, nil
}

func (d *driver) DeleteRepo(repo *pfs.Repo, shard map[uint64]bool) error {
	return fmt.Errorf("not implemented")
}

func (d *driver) StartCommit(parent *pfs.Commit, commit *pfs.Commit, shards map[uint64]bool) error {
	if commit == nil {
		return fmt.Errorf("pachyderm: nil commit")
	}
	if err := execSubvolumeCreate(d.commitPathNoShard(commit)); err != nil && !execSubvolumeExists(d.commitPathNoShard(commit)) {
		return err
	}
	for shard := range shards {
		commitPath := d.writeCommitPath(commit, shard)
		if parent != nil {
			if err := d.checkReadOnly(parent, shard); err != nil {
				return err
			}
			parentPath, err := d.commitPath(parent, shard)
			if err != nil {
				return err
			}
			if err := execSubvolumeSnapshot(parentPath, commitPath, false); err != nil {
				return err
			}
			filePath, err := d.filePath(&pfs.File{Commit: commit, Path: filepath.Join(metadataDir, "parent")}, shard)
			if err != nil {
				return err
			}
			if err := ioutil.WriteFile(filePath, []byte(parent.Id), 0600); err != nil {
				return err
			}
		} else {
			if err := execSubvolumeCreate(commitPath); err != nil {
				return err
			}
			filePath, err := d.filePath(&pfs.File{Commit: commit, Path: metadataDir}, shard)
			if err != nil {
				return err
			}
			if err := os.Mkdir(filePath, 0700); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *driver) FinishCommit(commit *pfs.Commit, shards map[uint64]bool) error {
	for shard := range shards {
		if err := execSubvolumeSnapshot(d.writeCommitPath(commit, shard), d.readCommitPath(commit, shard), true); err != nil {
			return err
		}
	}
	// TODO we don't delete the writeCommit here because we want to use it to
	// figure out when the commit started. However if we recorded the ModTime
	// on that directory then we could go back to deleting it. This isn't
	// hugely important, due to COW the write snapshot doesn't take up much
	// space.
	return nil
}

func (d *driver) InspectCommit(commit *pfs.Commit, shards map[uint64]bool) (*pfs.CommitInfo, error) {
	var wg sync.WaitGroup
	var lock sync.Mutex
	result := &pfs.CommitInfo{Commit: commit, CommitType: pfs.CommitType_COMMIT_TYPE_WRITE}
	var loopErr error
	notFound := false
	for shard := range shards {
		shard := shard
		wg.Add(1)
		go func() {
			defer wg.Done()
			if !execSubvolumeExists(d.readCommitPath(commit, shard)) && !execSubvolumeExists(d.writeCommitPath(commit, shard)) {
				notFound = true
				return
			}
			parent, err := d.getParent(commit, shard)
			if err != nil {
				if loopErr == nil {
					loopErr = err
				}
				return
			}
			readOnly, err := d.getReadOnly(commit, shard)
			if err != nil {
				if loopErr == nil {
					loopErr = err
				}
				return
			}
			commitType := pfs.CommitType_COMMIT_TYPE_WRITE
			if readOnly {
				commitType = pfs.CommitType_COMMIT_TYPE_READ
			}
			writeStat, err := os.Stat(d.writeCommitPath(commit, shard))
			if err != nil {
				if loopErr == nil {
					loopErr = err
				}
				return
			}
			startTime := writeStat.ModTime()
			var finishTime *time.Time
			if commitType == pfs.CommitType_COMMIT_TYPE_READ {
				readStat, err := os.Stat(d.readCommitPath(commit, shard))
				if err != nil && loopErr == nil {
					loopErr = err
					return
				}
				_finishTime := readStat.ModTime()
				finishTime = &_finishTime
			}
			changes, err := d.ListChange(&pfs.File{Commit: commit}, parent, shard)
			if err != nil {
				if loopErr == nil {
					loopErr = err
				}
				return
			}
			var commitBytes uint64
			for _, change := range changes {
				commitBytes += change.SizeBytes
			}
			commitPath, err := d.commitPath(commit, shard)
			if err != nil {
				if loopErr == nil {
					loopErr = err
				}
				return
			}
			totalBytes, err := d.recursiveSize(commitPath)
			if err != nil {
				if loopErr == nil {
					loopErr = err
				}
				return
			}
			lock.Lock()
			defer lock.Unlock()
			if result.CommitType != pfs.CommitType_COMMIT_TYPE_READ {
				result.CommitType = commitType
			}
			result.ParentCommit = parent
			if result.Started == nil || startTime.Before(prototime.TimestampToTime(result.Started)) {
				result.Started = prototime.TimeToTimestamp(startTime)
			}
			if finishTime != nil && (result.Finished == nil || finishTime.After(prototime.TimestampToTime(result.Finished))) {
				result.Finished = prototime.TimeToTimestamp(*finishTime)
			}
			result.CommitBytes += commitBytes
			result.TotalBytes += totalBytes
		}()
	}
	wg.Wait()
	if notFound {
		return nil, nil
	}
	if loopErr != nil {
		return nil, loopErr
	}
	return result, nil
}

func (d *driver) ListCommit(repo *pfs.Repo, from *pfs.Commit, shards map[uint64]bool) ([]*pfs.CommitInfo, error) {
	var commitInfos []*pfs.CommitInfo
	//TODO this buffer might get too big
	var buffer bytes.Buffer
	var fromCommit string
	if from != nil {
		fromCommit = from.Id
	}
	if err := execSubvolumeList(d.repoPath(repo), fromCommit, false, &buffer); err != nil {
		return nil, err
	}
	commitScanner := newCommitScanner(&buffer, d.namespace, repo.Name)
	for commitScanner.Scan() {
		commitID := commitScanner.Commit()
		commitInfo, err := d.InspectCommit(
			&pfs.Commit{
				Repo: repo,
				Id:   commitID,
			},
			shards,
		)
		if commitInfo == nil {
			// It's possible for us to not find a commit if it's in the middle
			// of being committed, we ignore partial commits.
			continue
		}
		if err != nil {
			return nil, err
		}
		commitInfos = append(commitInfos, commitInfo)
	}
	return commitInfos, nil
}

func (d *driver) DeleteCommit(commit *pfs.Commit, shard map[uint64]bool) error {
	return fmt.Errorf("not implemented")
}

func (d *driver) PutBlock(file *pfs.File, block *pfs.Block, shard uint64, reader io.Reader) (retErr error) {
	if err := d.checkWrite(file.Commit, shard); err != nil {
		return err
	}
	filePath, err := d.filePath(file, shard)
	if err != nil {
		return err
	}
	fileFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer func() {
		if err := fileFile.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	encodedBlock, err := proto.Marshal(block)
	if err != nil {
		return err
	}
	if _, err := fileFile.Write(encodedBlock); err != nil {
		return err
	}
	if err := os.MkdirAll(d.blockShardDir(shard), 0700); err != nil {
		return err
	}
	_, err = os.Stat(d.blockPath(block, shard))
	if err == nil {
		// No error means the block already exists
		return nil
	}
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	blockFile, err := os.OpenFile(d.blockPath(block, shard), os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer func() {
		if err := blockFile.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	sizeBytes, err := bufio.NewReader(reader).WriteTo(blockFile)
	if err != nil {
		return err
	}
	blockInfo := pfs.BlockInfo{
		Block:     block,
		SizeBytes: uint64(sizeBytes),
	}
	encodedBlockInfo, err := proto.Marshal(&blockInfo)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(d.blockInfoPath(block, shard), encodedBlockInfo, 0666); err != nil {
		return err
	}
	return nil
}

func (d *driver) GetBlock(block *pfs.Block, shard uint64) (drive.ReaderAtCloser, error) {
	return os.Open(d.blockPath(block, shard))
}

func (d *driver) InspectBlock(block *pfs.Block, shard uint64) (*pfs.BlockInfo, error) {
	encodedBlockInfo, err := ioutil.ReadFile(d.blockInfoPath(block, shard))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var result pfs.BlockInfo
	if err := proto.Unmarshal(encodedBlockInfo, &result); err != nil {
		return nil, err
	}
	stat, err := os.Stat(d.blockPath(block, shard))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	result.Created = prototime.TimeToTimestamp(stat.ModTime())
	return &result, nil
}

func (d *driver) ListBlock(shard uint64) (_ []*pfs.BlockInfo, retErr error) {
	var result []*pfs.BlockInfo
	dir, err := os.Open(d.blockShardDir(shard))
	if os.IsNotExist(err) {
		return result, nil
	}
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := dir.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	var names []string
	for names, err = dir.Readdirnames(readDirBatch); err == nil; names, err = dir.Readdirnames(readDirBatch) {
		for _, name := range names {
			if strings.HasSuffix(name, infoSuffix) {
				continue
			}
			blockInfo, err := d.InspectBlock(&pfs.Block{Hash: filepath.Base(name)}, shard)
			if err != nil {
				return nil, err
			}
			result = append(result, blockInfo)
		}
	}
	if err != io.EOF {
		return nil, err
	}
	return result, nil
}

func (d *driver) PutFile(file *pfs.File, shard uint64, offset int64, reader io.Reader) (retErr error) {
	if err := d.checkWrite(file.Commit, shard); err != nil {
		return err
	}
	filePath, err := d.filePath(file, shard)
	if err != nil {
		return err
	}
	osFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer func() {
		if err := osFile.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	if _, err := osFile.Seek(offset, 0); err != nil { // 0 means relative to start
		return err
	}
	_, err = bufio.NewReader(reader).WriteTo(osFile)
	return err
}

func (d *driver) MakeDirectory(file *pfs.File, shards map[uint64]bool) error {
	// TODO(pedge): if PutFile fails here or on another shard, the directories
	// will still exist and be returned from ListFiles, we want to do this
	// iteratively and with rollback
	for shard := range shards {
		if err := d.checkWrite(file.Commit, shard); err != nil {
			return err
		}
		filePath, err := d.filePath(file, shard)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filePath, 0700); err != nil {
			return err
		}
	}
	return nil
}

func (d *driver) GetFile(file *pfs.File, shard uint64) (drive.ReaderAtCloser, error) {
	filePath, err := d.filePath(file, shard)
	if err != nil {
		return nil, err
	}
	return os.Open(filePath)
}

func (d *driver) InspectFile(file *pfs.File, shard uint64) (*pfs.FileInfo, error) {
	fileInfo, err := d.stat(file, shard)
	if err != nil && os.IsNotExist(err) {
		return nil, fmt.Errorf("file %s not found", file.Path)
	}
	if err != nil {
		return nil, err
	}
	return fileInfo, nil
}

func (d *driver) ListFile(file *pfs.File, shard uint64) (_ []*pfs.FileInfo, retErr error) {
	filePath, err := d.filePath(file, shard)
	if err != nil {
		return nil, err
	}
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", filePath)
	}
	dir, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := dir.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	var fileInfos []*pfs.FileInfo
	// TODO(pedge): constant
	for names, err := dir.Readdirnames(readDirBatch); err != io.EOF; names, err = dir.Readdirnames(readDirBatch) {
		if err != nil {
			return nil, err
		}
		for _, name := range names {
			if inMetadataDir(name) {
				continue
			}
			fileInfo, err := d.stat(
				&pfs.File{
					Commit: file.Commit,
					Path:   path.Join(file.Path, name),
				},
				shard,
			)
			if err != nil {
				return nil, err
			}
			fileInfos = append(fileInfos, fileInfo)
		}
	}
	return fileInfos, nil
}

func (d *driver) ListChange(file *pfs.File, from *pfs.Commit, shard uint64) ([]*pfs.Change, error) {
	//TODO this buffer might get too big
	var buffer bytes.Buffer
	commitPath, err := d.commitPath(file.Commit, shard)
	if err != nil {
		return nil, err
	}
	var fromCommitPath string
	if from != nil {
		fromCommitPath, err = d.commitPath(from, shard)
		if err != nil {
			return nil, err
		}
	}
	if err := execSubvolumeFindNew(commitPath, fromCommitPath, &buffer); err != nil {
		return nil, err
	}
	var changes []*pfs.Change
	changeScanner := newChangeScanner(&buffer, file.Commit)
	for changeScanner.Scan() {
		change := changeScanner.Change()
		changes = append(changes, change)
	}
	return changes, nil
}

func (d *driver) DeleteFile(file *pfs.File, shard uint64) error {
	if err := d.checkWrite(file.Commit, shard); err != nil {
		return err
	}
	filePath, err := d.filePath(file, shard)
	if err != nil {
		return err
	}
	return os.Remove(filePath)
}

func (d *driver) PullDiff(commit *pfs.Commit, shard uint64, diff io.Writer) error {
	protolog.Printf("PullDiff: %s", d.readCommitPath(commit, shard))
	parent, err := d.getParent(commit, shard)
	if err != nil {
		return err
	}
	if parent == nil {
		return execSend(d.readCommitPath(commit, shard), "", diff)
	}
	return execSend(d.readCommitPath(commit, shard), d.readCommitPath(parent, shard), diff)
}

func (d *driver) PushDiff(commit *pfs.Commit, shard uint64, diff io.Reader) error {
	protolog.Printf("PushDiff: %s", d.readCommitPath(commit, shard))
	if err := execSubvolumeCreate(d.commitPathNoShard(commit)); err != nil && !execSubvolumeExists(d.commitPathNoShard(commit)) {
		return err
	}
	if err := execRecv(d.commitPathNoShard(commit), diff); err != nil {
		return err
	}
	if !execSubvolumeExists(d.readCommitPath(commit, shard)) {
		return fmt.Errorf("PushDiff failed %s", d.readCommitPath(commit, shard))
	}
	return nil
}

func (d *driver) stat(file *pfs.File, shard uint64) (*pfs.FileInfo, error) {
	filePath, err := d.filePath(file, shard)
	if err != nil {
		return nil, err
	}
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	var fileType pfs.FileType
	if stat.Mode().IsDir() {
		fileType = pfs.FileType_FILE_TYPE_DIR
	} else {
		fileType = pfs.FileType_FILE_TYPE_REGULAR
	}
	return &pfs.FileInfo{
		File:      file,
		FileType:  fileType,
		SizeBytes: uint64(stat.Size()),
		Perm:      uint32(stat.Mode() & os.ModePerm),
		Modified: prototime.TimeToTimestamp(
			stat.ModTime(),
		),
	}, nil
}

func (d *driver) getParent(commit *pfs.Commit, shard uint64) (*pfs.Commit, error) {
	filePath, err := d.filePath(&pfs.File{Commit: commit, Path: filepath.Join(metadataDir, "parent")}, shard)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadFile(filePath)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &pfs.Commit{
		Repo: commit.Repo,
		Id:   string(data),
	}, nil
}

func (d *driver) checkReadOnly(commit *pfs.Commit, shard uint64) error {
	ok, err := d.getReadOnly(commit, shard)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("%+v is not a read only commit", commit)
	}
	return nil
}

func (d *driver) checkWrite(commit *pfs.Commit, shard uint64) error {
	ok, err := d.getReadOnly(commit, shard)
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf("%+v is not a write commit", commit)
	}
	return nil
}

func (d *driver) getReadOnly(commit *pfs.Commit, shard uint64) (bool, error) {
	if execSubvolumeExists(d.readCommitPath(commit, shard)) {
		return true, nil
	} else if execSubvolumeExists(d.writeCommitPath(commit, shard)) {
		return false, nil
	} else {
		return false, fmt.Errorf("Commit not found: %s", d.readCommitPath(commit, shard))
	}
}

func (d *driver) basePath() string {
	return filepath.Join(d.rootDir, d.namespace)
}

func (d *driver) blockDir() string {
	return filepath.Join(d.basePath(), blockDir)
}

func (d *driver) blockShardDir(shard uint64) string {
	return filepath.Join(d.blockDir(), fmt.Sprint(shard))
}

func (d *driver) blockPath(block *pfs.Block, shard uint64) string {
	return filepath.Join(d.blockShardDir(shard), block.Hash)
}

func (d *driver) blockInfoPath(block *pfs.Block, shard uint64) string {
	return filepath.Join(d.basePath(), blockDir, fmt.Sprint(shard), block.Hash+infoSuffix)
}

func (d *driver) repoDir() string {
	return filepath.Join(d.basePath(), repoDir)
}

func (d *driver) repoPath(repo *pfs.Repo) string {
	return filepath.Join(d.repoDir(), repo.Name)
}

func (d *driver) commitPathNoShard(commit *pfs.Commit) string {
	return filepath.Join(d.repoPath(commit.Repo), commit.Id)
}

func (d *driver) readCommitPath(commit *pfs.Commit, shard uint64) string {
	return filepath.Join(d.commitPathNoShard(commit), fmt.Sprint(shard))
}

func (d *driver) writeCommitPath(commit *pfs.Commit, shard uint64) string {
	return d.readCommitPath(commit, shard) + writeSuffix
}

func (d *driver) commitPath(commit *pfs.Commit, shard uint64) (string, error) {
	readOnly, err := d.getReadOnly(commit, shard)
	if err != nil {
		return "", err
	}
	if readOnly {
		return d.readCommitPath(commit, shard), nil
	}
	return d.writeCommitPath(commit, shard), nil
}

func (d *driver) filePath(file *pfs.File, shard uint64) (string, error) {
	commitPath, err := d.commitPath(file.Commit, shard)
	if err != nil {
		return "", err
	}
	return filepath.Join(commitPath, file.Path), nil
}

func (d *driver) recursiveSize(root string) (uint64, error) {
	var result int64
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if inMetadataDir(path) {
			return filepath.SkipDir
		}
		if info.IsDir() {
			result += info.Size()
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return uint64(result), err
}

func inMetadataDir(name string) bool {
	parts := strings.Split(name, "/")
	return len(parts) > 0 && parts[0] == metadataDir
}

func execSubvolumeCreate(path string) (retErr error) {
	defer func() {
		protolog.Debug(&SubvolumeCreate{path, errorToString(retErr)})
	}()
	return pkgexec.Run("btrfs", "subvolume", "create", path)
}

func execSubvolumeDelete(path string) (retErr error) {
	defer func() {
		protolog.Debug(&SubvolumeDelete{path, errorToString(retErr)})
	}()
	return pkgexec.Run("btrfs", "subvolume", "delete", path)
}

func execSubvolumeExists(path string) (result bool) {
	defer func() {
		protolog.Debug(&SubvolumeExists{path, result})
	}()
	if err := pkgexec.Run("btrfs", "subvolume", "show", path); err != nil {
		return false
	}
	return true
}

func execSubvolumeSnapshot(src string, dest string, readOnly bool) (retErr error) {
	defer func() {
		protolog.Info(&SubvolumeSnapshot{src, dest, readOnly, errorToString(retErr)})
	}()
	if readOnly {
		return pkgexec.Run("btrfs", "subvolume", "snapshot", "-r", src, dest)
	}
	return pkgexec.Run("btrfs", "subvolume", "snapshot", src, dest)
}

func execTransID(path string) (result string, retErr error) {
	defer func() {
		protolog.Debug(&TransID{path, result, errorToString(retErr)})
	}()
	//  "9223372036854775810" == 2 ** 63 we use a very big number there so that
	//  we get the transid of the from path. According to the internet this is
	//  the nicest way to get it from btrfs.
	var buffer bytes.Buffer
	if err := pkgexec.RunStdout(&buffer, "btrfs", "subvolume", "find-new", path, "9223372036854775808"); err != nil {
		return "", err
	}
	scanner := bufio.NewScanner(&buffer)
	for scanner.Scan() {
		// scanner.Text() looks like this:
		// transid marker was 907
		// 0       1      2   3
		tokens := strings.Split(scanner.Text(), " ")
		if len(tokens) != 4 {
			return "", fmt.Errorf("pachyderm: failed to parse find-new output")
		}
		return tokens[3], nil
	}
	if scanner.Err() != nil {
		return "", scanner.Err()
	}
	return "", fmt.Errorf("pachyderm: empty output from find-new")
}

func execSubvolumeList(path string, fromCommit string, ascending bool, out io.Writer) (retErr error) {
	defer func() {
		protolog.Debug(&SubvolumeList{path, fromCommit, ascending, errorToString(retErr)})
	}()
	var sort string
	if ascending {
		sort = "+ogen"
	} else {
		sort = "-ogen"
	}

	if fromCommit == "" {
		return pkgexec.RunStdout(out, "btrfs", "subvolume", "list", "-a", "--sort", sort, path)
	}
	transid, err := execTransID(fromCommit)
	if err != nil {
		return err
	}
	return pkgexec.RunStdout(out, "btrfs", "subvolume", "list", "-aC", "+"+transid, "--sort", sort, path)
}

func execSubvolumeFindNew(commit string, fromCommit string, out io.Writer) (retErr error) {
	defer func() {
		protolog.Debug(&SubvolumeFindNew{commit, fromCommit, errorToString(retErr)})
	}()
	if fromCommit == "" {
		return pkgexec.RunStdout(out, "btrfs", "subvolume", "find-new", commit, "0")
	}
	transid, err := execTransID(fromCommit)
	if err != nil {
		return err
	}
	return pkgexec.RunStdout(out, "btrfs", "subvolume", "find-new", commit, transid)
}

func execSend(path string, parent string, diff io.Writer) (retErr error) {
	defer func() {
		protolog.Debug(&Send{path, parent, errorToString(retErr)})
	}()
	if parent == "" {
		return pkgexec.RunStdout(diff, "btrfs", "send", path)
	}
	return pkgexec.RunStdout(diff, "btrfs", "send", "-p", parent, path)
}

func execRecv(path string, diff io.Reader) (retErr error) {
	defer func() {
		protolog.Debug(&Recv{path, errorToString(retErr)})
	}()
	return pkgexec.RunStdin(diff, "btrfs", "receive", path)
}

type commitScanner struct {
	textScanner *bufio.Scanner
	namespace   string
	repo        string
}

func newCommitScanner(reader io.Reader, namespace string, repo string) *commitScanner {
	return &commitScanner{bufio.NewScanner(reader), namespace, repo}
}

func (c *commitScanner) Scan() bool {
	for {
		if !c.textScanner.Scan() {
			return false
		}
		if _, ok := c.parseCommit(); ok {
			return true
		}
	}
}

func (c *commitScanner) Err() error {
	return c.textScanner.Err()
}

func (c *commitScanner) Commit() string {
	commit, _ := c.parseCommit()
	protolog.Debug(&SubvolumeListLine{c.textScanner.Text()})
	return commit
}

func (c *commitScanner) parseCommit() (string, bool) {
	// c.textScanner.Text() looks like:
	// ID 905 gen 865 top level 5 path <FS_TREE>/[c.namespace/]repoDir/c.repo/commit
	// 0  1   2   3   4   5     6 7    8
	tokens := strings.Split(c.textScanner.Text(), " ")
	if len(tokens) != 9 {
		return "", false
	}
	var prefix string
	if c.namespace == "" {
		prefix = filepath.Join("<FS_TREE>", repoDir, c.repo) + "/"
	} else {
		prefix = filepath.Join("<FS_TREE>", c.namespace, repoDir, c.repo) + "/"
	}
	if !strings.HasPrefix(tokens[8], prefix) {
		return "", false
	}
	commit := strings.TrimPrefix(tokens[8], prefix)
	if len(strings.Split(commit, "/")) != 1 {
		// this happens when commit is `scratch/1.write` or something
		return "", false
	}
	return commit, true
}

type changeScanner struct {
	textScanner *bufio.Scanner
	commit      *pfs.Commit
}

func newChangeScanner(reader io.Reader, commit *pfs.Commit) *changeScanner {
	return &changeScanner{bufio.NewScanner(reader), commit}
}

func (c *changeScanner) Scan() bool {
	for {
		if !c.textScanner.Scan() {
			return false
		}
		if _, ok := c.parseChange(); ok {
			return true
		}
	}
}

func (c *changeScanner) Err() error {
	return c.textScanner.Err()
}

func (c *changeScanner) Change() *pfs.Change {
	change, _ := c.parseChange()
	protolog.Debug(&SubvolumeFindNewLine{c.textScanner.Text()})
	return change
}

func (c *changeScanner) parseChange() (*pfs.Change, bool) {
	// c.textScanner.Text() looks like:
	// inode 258 file offset 0 len 7 disk start 0 offset 0 gen 330 flags INLINE path/to/file
	// 0     1   2    3      4 5   6 7    8     9 10     1112  13  14    15     16
	tokens := strings.Split(c.textScanner.Text(), " ")
	if len(tokens) != 17 {
		return nil, false
	}
	if inMetadataDir(tokens[16]) {
		return nil, false
	}
	offset, err := strconv.ParseUint(tokens[4], 10, 64)
	if err != nil {
		return nil, false
	}
	length, err := strconv.ParseUint(tokens[6], 10, 64)
	if err != nil {
		return nil, false
	}
	return &pfs.Change{
		File: &pfs.File{
			Commit: c.commit,
			Path:   tokens[16],
		},
		OffsetBytes: offset,
		SizeBytes:   length,
	}, true
}

// TODO this code is duplicate elsewhere, we should put it somehwere.
func errorToString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
