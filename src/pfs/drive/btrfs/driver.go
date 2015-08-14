// +build linux

/*

directory structure

  .
  |-- repositoryName
	  |-- scratch
		  |-- shardNum // the read-only read created on InitRepository, this is where to start branching
      |-- commitID
	      |-- shardNum // this is where subvolumes are

*/

package btrfs

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pkg/executil"
	"github.com/peter-edge/go-google-protobuf"
	"github.com/satori/go.uuid"
)

const (
	metadataDir = ".pfs"
)

type driver struct {
	rootDir string
}

func newDriver(rootDir string) *driver {
	return &driver{rootDir}
}

func (d *driver) InitRepository(repository *pfs.Repository, replica bool, shards map[int]bool) error {
	if err := execSubvolumeCreate(d.repositoryPath(repository)); err != nil && !execSubvolumeExists(d.repositoryPath(repository)) {
		return err
	}
	// All we do for a replica is initialize the repository subvolume, we don't
	// create an initial commit because that initial commit will be created as
	// part of replication.
	if replica {
		return nil
	}
	initialCommit := &pfs.Commit{
		Repository: repository,
		Id:         drive.InitialCommitID,
	}
	if _, err := d.Branch(nil, initialCommit, shards); err != nil {
		return err
	}
	if err := d.Commit(initialCommit, shards); err != nil {
		return err
	}
	return nil
}

func (d *driver) GetFile(path *pfs.Path, shard int) (drive.ReaderAtCloser, error) {
	filePath, err := d.filePath(path, shard)
	if err != nil {
		return nil, err
	}
	return os.Open(filePath)
}

func (d *driver) GetFileInfo(path *pfs.Path, shard int) (_ *pfs.FileInfo, ok bool, _ error) {
	filePath, err := d.stat(path, shard)
	if err != nil && os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return filePath, true, nil
}

func (d *driver) MakeDirectory(path *pfs.Path, shards map[int]bool) error {
	// TODO(pedge): if PutFile fails here or on another shard, the directories
	// will still exist and be returned from ListFiles, we want to do this
	// iteratively and with rollback
	for shard := range shards {
		if err := d.checkWrite(path.Commit, shard); err != nil {
			return err
		}
		filePath, err := d.filePath(path, shard)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filePath, 0700); err != nil {
			return err
		}
	}
	return nil
}

func (d *driver) PutFile(path *pfs.Path, shard int, offset int64, reader io.Reader) error {
	if err := d.checkWrite(path.Commit, shard); err != nil {
		return err
	}
	filePath, err := d.filePath(path, shard)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.Seek(offset, 0); err != nil { // 0 means relative to start
		return err
	}
	_, err = bufio.NewReader(reader).WriteTo(file)
	return err
}

func (d *driver) ListFiles(path *pfs.Path, shard int) (_ []*pfs.FileInfo, retErr error) {
	filePath, err := d.filePath(path, shard)
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
	for names, err := dir.Readdirnames(100); err != io.EOF; names, err = dir.Readdirnames(100) {
		if err != nil {
			return nil, err
		}
		for _, name := range names {
			if inMetadataDir(name) {
				continue
			}
			fileInfo, err := d.stat(
				&pfs.Path{
					Commit: path.Commit,
					Path:   filepath.Join(path.Path, name),
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

func (d *driver) stat(path *pfs.Path, shard int) (*pfs.FileInfo, error) {
	filePath, err := d.filePath(path, shard)
	if err != nil {
		return nil, err
	}
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	fileType := pfs.FileType_FILE_TYPE_OTHER
	if stat.Mode().IsRegular() {
		fileType = pfs.FileType_FILE_TYPE_REGULAR
	}
	if stat.Mode().IsDir() {
		fileType = pfs.FileType_FILE_TYPE_DIR
	}
	return &pfs.FileInfo{
		Path:      path,
		FileType:  fileType,
		SizeBytes: uint64(stat.Size()),
		Perm:      uint32(stat.Mode() & os.ModePerm),
		LastModified: &google_protobuf.Timestamp{
			Seconds: stat.ModTime().UnixNano() / int64(time.Second),
			Nanos:   int32(stat.ModTime().UnixNano() % int64(time.Second)),
		},
	}, nil
}

func (d *driver) Branch(commit *pfs.Commit, newCommit *pfs.Commit, shards map[int]bool) (*pfs.Commit, error) {
	if commit == nil && newCommit == nil {
		return nil, fmt.Errorf("pachyderm: must specify either commit or newCommit")
	}
	if newCommit == nil {
		newCommit = &pfs.Commit{
			Repository: commit.Repository,
			Id:         newCommitID(),
		}
	}
	if err := execSubvolumeCreate(d.commitPathNoShard(newCommit)); err != nil && !execSubvolumeExists(d.commitPathNoShard(newCommit)) {
		return nil, err
	}
	for shard := range shards {
		newCommitPath := d.writeCommitPath(newCommit, shard)
		if commit != nil {
			if err := d.checkReadOnly(commit, shard); err != nil {
				return nil, err
			}
			commitPath, err := d.commitPath(commit, shard)
			if err != nil {
				return nil, err
			}
			if err := execSubvolumeSnapshot(commitPath, newCommitPath, false); err != nil {
				return nil, err
			}
			filePath, err := d.filePath(&pfs.Path{Commit: newCommit, Path: filepath.Join(metadataDir, "parent")}, shard)
			if err != nil {
				return nil, err
			}
			if err := ioutil.WriteFile(filePath, []byte(commit.Id), 0600); err != nil {
				return nil, err
			}
		} else {
			if err := execSubvolumeCreate(newCommitPath); err != nil {
				return nil, err
			}
			filePath, err := d.filePath(&pfs.Path{Commit: newCommit, Path: metadataDir}, shard)
			if err != nil {
				return nil, err
			}
			if err := os.Mkdir(filePath, 0700); err != nil {
				return nil, err
			}
		}
	}
	return newCommit, nil
}

func (d *driver) Commit(commit *pfs.Commit, shards map[int]bool) error {
	for shard := range shards {
		if err := execSubvolumeSnapshot(d.writeCommitPath(commit, shard), d.readCommitPath(commit, shard), true); err != nil {
			return err
		}
	}
	return nil
}

func (d *driver) PullDiff(commit *pfs.Commit, shard int, diff io.Writer) error {
	parent, err := d.getParent(commit, shard)
	if err != nil {
		return err
	}
	if parent == nil {
		return execSend(d.readCommitPath(commit, shard), "", diff)
	}
	return execSend(d.readCommitPath(commit, shard), d.readCommitPath(parent, shard), diff)
}

func (d *driver) PushDiff(repository *pfs.Repository, diff io.Reader) error {
	return execRecv(d.repositoryPath(repository), diff)
}

func (d *driver) GetCommitInfo(commit *pfs.Commit, shard int) (_ *pfs.CommitInfo, ok bool, _ error) {
	_, readErr := os.Stat(d.readCommitPath(commit, shard))
	_, writeErr := os.Stat(d.writeCommitPath(commit, shard))
	if readErr != nil && os.IsNotExist(readErr) && writeErr != nil && os.IsNotExist(writeErr) {
		return nil, false, nil
	}
	parent, err := d.getParent(commit, shard)
	if err != nil {
		return nil, false, err
	}
	readOnly, err := d.getReadOnly(commit, shard)
	if err != nil {
		return nil, false, err
	}
	commitType := pfs.CommitType_COMMIT_TYPE_WRITE
	if readOnly {
		commitType = pfs.CommitType_COMMIT_TYPE_READ
	}
	return &pfs.CommitInfo{
		Commit:       commit,
		CommitType:   commitType,
		ParentCommit: parent,
	}, true, nil
}

func (d *driver) ListCommits(repository *pfs.Repository, shard int) (_ []*pfs.CommitInfo, retErr error) {

	dir, err := os.Open(d.repositoryPath(repository))
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := dir.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	var commitInfos []*pfs.CommitInfo
	// TODO(pedge): constant
	for names, err := dir.Readdirnames(100); err != io.EOF; names, err = dir.Readdirnames(100) {
		if err != nil {
			return nil, err
		}
		for _, name := range names {
			commitInfo, ok, err := d.GetCommitInfo(
				&pfs.Commit{
					Repository: repository,
					Id:         path.Base(name),
				},
				shard,
			)
			if !ok {
				// This is a really weird error to get since we got this commit
				// name by listing commits. This is probably indicative of a
				// race condition.
				return nil, fmt.Errorf("Commit not found.")
			}
			if err != nil {
				return nil, err
			}
			commitInfos = append(commitInfos, commitInfo)
		}
	}
	return commitInfos, nil
}

func (d *driver) getParent(commit *pfs.Commit, shard int) (*pfs.Commit, error) {
	if commit.Id == drive.InitialCommitID {
		return nil, nil
	}
	filePath, err := d.filePath(&pfs.Path{Commit: commit, Path: filepath.Join(metadataDir, "parent")}, shard)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return &pfs.Commit{
		Repository: commit.Repository,
		Id:         string(data),
	}, nil
}

func (d *driver) checkReadOnly(commit *pfs.Commit, shard int) error {
	ok, err := d.getReadOnly(commit, shard)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("%+v is not a read only commit", commit)
	}
	return nil
}

func (d *driver) checkWrite(commit *pfs.Commit, shard int) error {
	ok, err := d.getReadOnly(commit, shard)
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf("%+v is not a write commit", commit)
	}
	return nil
}

func (d *driver) getReadOnly(commit *pfs.Commit, shard int) (bool, error) {
	if execSubvolumeExists(d.readCommitPath(commit, shard)) {
		return true, nil
	} else if execSubvolumeExists(d.writeCommitPath(commit, shard)) {
		return false, nil
	} else {
		return false, fmt.Errorf("pachyderm: commit %s doesn't exist", commit.Id)
	}
}

func (d *driver) repositoryPath(repository *pfs.Repository) string {
	return filepath.Join(d.rootDir, repository.Name)
}

func (d *driver) commitPathNoShard(commit *pfs.Commit) string {
	return filepath.Join(d.repositoryPath(commit.Repository), commit.Id)
}

func (d *driver) readCommitPath(commit *pfs.Commit, shard int) string {
	return filepath.Join(d.commitPathNoShard(commit), fmt.Sprint(shard))
}

func (d *driver) writeCommitPath(commit *pfs.Commit, shard int) string {
	return d.readCommitPath(commit, shard) + ".write"
}

func (d *driver) commitPath(commit *pfs.Commit, shard int) (string, error) {
	readOnly, err := d.getReadOnly(commit, shard)
	if err != nil {
		return "", err
	}
	if readOnly {
		return d.readCommitPath(commit, shard), nil
	}
	return d.writeCommitPath(commit, shard), nil
}

func (d *driver) filePath(path *pfs.Path, shard int) (string, error) {
	commitPath, err := d.commitPath(path.Commit, shard)
	if err != nil {
		return "", nil
	}
	return filepath.Join(commitPath, path.Path), nil
}

func newCommitID() string {
	return strings.Replace(uuid.NewV4().String(), "-", "", -1)
}

func inMetadataDir(name string) bool {
	parts := strings.Split(name, "/")
	return (len(parts) > 0 && parts[0] == metadataDir)
}

func execSubvolumeCreate(path string) error {
	return executil.Run("btrfs", "subvolume", "create", path)
}

func execSubvolumeDelete(path string) error {
	return executil.Run("btrfs", "subvolume", "delete", path)
}

func execSubvolumeExists(path string) bool {
	if err := executil.Run("btrfs", "subvolume", "show", path); err != nil {
		return false
	}
	return true
}

func execSubvolumeSnapshot(src string, dest string, readOnly bool) error {
	if readOnly {
		return executil.Run("btrfs", "subvolume", "snapshot", "-r", src, dest)
	}
	return executil.Run("btrfs", "subvolume", "snapshot", src, dest)
}

func execSend(path string, parent string, diff io.Writer) error {
	if parent == "" {
		return executil.RunStdout(diff, "btrfs", "send", path)
	}
	return executil.RunStdout(diff, "btrfs", "send", "-p", parent, path)
}

func execRecv(path string, diff io.Reader) error {
	return executil.RunStdin(diff, "btrfs", "receive", path)
}
