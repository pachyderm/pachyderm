/*

directory structure

  .
  |-- repositoryName
	  |-- scratch
		  |-- shardNum // the read-only read created on InitRepository, this is where to start branching
      |-- commitID
	      |-- shardNum // this is where subvolumes are

*/
package drive

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/src/common"
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pkg/btrfs"
	"github.com/peter-edge/go-google-protobuf"
)

type btrfsDriver struct {
	rootDir  string
	btrfsAPI btrfs.API
}

func newBtrfsDriver(rootDir string, btrfsAPI btrfs.API) *btrfsDriver {
	return &btrfsDriver{rootDir, btrfsAPI}
}

func (d *btrfsDriver) Init() error {
	return nil
}

func (d *btrfsDriver) InitRepository(repository *pfs.Repository, shards map[int]bool) error {
	if err := os.MkdirAll(d.repositoryPath(repository), 0700); err != nil {
		return err
	}
	initialCommit := &pfs.Commit{
		Repository: repository,
		Id:         InitialCommitID,
	}
	if err := os.MkdirAll(d.commitPathNoShard(initialCommit), 0700); err != nil {
		return err
	}
	for shard := range shards {
		if err := d.btrfsAPI.SubvolumeCreate(d.commitPath(initialCommit, shard)); err != nil {
			return err
		}
		if err := os.Mkdir(d.filePath(&pfs.Path{Commit: initialCommit, Path: ".pfs"}, shard), 0700); err != nil {
			return err
		}
		if err := d.btrfsAPI.PropertySetReadonly(d.commitPath(initialCommit, shard), true); err != nil {
			return err
		}
	}
	return nil
}

func (d *btrfsDriver) GetFile(path *pfs.Path, shard int) (io.ReadCloser, error) {
	return os.Open(d.filePath(path, shard))
}

func (d *btrfsDriver) GetFileInfo(path *pfs.Path, shard int) (*pfs.FileInfo, error) {
	return d.stat(path, shard)
}

func (d *btrfsDriver) MakeDirectory(path *pfs.Path, shards map[int]bool) error {
	// TODO(pedge): if PutFile fails here or on another shard, the directories
	// will still exist and be returned from ListFiles, we want to do this
	// iteratively and with rollback
	// TODO(pedge): check that commit exists and is a write commit
	for shard := range shards {
		if err := os.MkdirAll(d.filePath(path, shard), 0700); err != nil {
			return err
		}
	}
	return nil
}

func (d *btrfsDriver) PutFile(path *pfs.Path, shard int, reader io.Reader) error {
	file, err := os.Create(d.filePath(path, shard))
	if err != nil {
		return err
	}
	_, err = bufio.NewReader(reader).WriteTo(file)
	return err
}

func (d *btrfsDriver) ListFiles(path *pfs.Path, shard int) (retValue []*pfs.FileInfo, retErr error) {
	filePath := d.filePath(path, shard)
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

func (d *btrfsDriver) stat(path *pfs.Path, shard int) (*pfs.FileInfo, error) {
	stat, err := os.Stat(d.filePath(path, shard))
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

func (d *btrfsDriver) Branch(commit *pfs.Commit, newCommit *pfs.Commit, shards map[int]bool) (*pfs.Commit, error) {
	if newCommit == nil {
		newCommit = &pfs.Commit{
			Repository: commit.Repository,
			Id:         newCommitID(),
		}
	}
	if err := os.MkdirAll(d.commitPathNoShard(newCommit), 0700); err != nil {
		return nil, err
	}
	for shard := range shards {
		if err := d.checkReadOnly(commit, shard); err != nil {
			return nil, err
		}
		commitPath := d.commitPath(commit, shard)
		newCommitPath := d.commitPath(newCommit, shard)
		if err := d.btrfsAPI.SubvolumeSnapshot(commitPath, newCommitPath, false); err != nil {
			return nil, err
		}
		if err := ioutil.WriteFile(d.filePath(&pfs.Path{Commit: newCommit, Path: ".pfs/parent"}, shard), []byte(commit.Id), 0600); err != nil {
			return nil, err
		}
	}
	return newCommit, nil
}

func (d *btrfsDriver) Commit(commit *pfs.Commit, shards map[int]bool) error {
	for shard := range shards {
		if err := d.checkWrite(commit, shard); err != nil {
			return err
		}
		if err := d.btrfsAPI.PropertySetReadonly(d.commitPath(commit, shard), true); err != nil {
			return err
		}
	}
	return nil
}

func (d *btrfsDriver) PullDiff(commit *pfs.Commit, shard int) (io.Reader, error) {
	return nil, nil
}

func (d *btrfsDriver) PushDiff(commit *pfs.Commit, shard int, reader io.Reader) error {
	return nil
}

func (d *btrfsDriver) GetCommitInfo(commit *pfs.Commit, shard int) (*pfs.CommitInfo, error) {
	parent, err := d.getParent(commit, shard)
	if err != nil {
		return nil, err
	}
	readOnly, err := d.btrfsAPI.PropertyGetReadonly(d.commitPath(commit, shard))
	if err != nil {
		return nil, err
	}
	commitType := pfs.CommitType_COMMIT_TYPE_WRITE
	if readOnly {
		commitType = pfs.CommitType_COMMIT_TYPE_READ
	}
	return &pfs.CommitInfo{
		Commit:       commit,
		CommitType:   commitType,
		ParentCommit: parent,
	}, nil
}

func (d *btrfsDriver) getParent(commit *pfs.Commit, shard int) (*pfs.Commit, error) {
	if commit.Id == InitialCommitID {
		return nil, nil
	}
	data, err := ioutil.ReadFile(d.filePath(&pfs.Path{Commit: commit, Path: ".pfs/parent"}, shard))
	if err != nil {
		return nil, err
	}
	return &pfs.Commit{
		Repository: commit.Repository,
		Id:         string(data),
	}, nil
}

func (d *btrfsDriver) checkReadOnly(commit *pfs.Commit, shard int) error {
	ok, err := d.btrfsAPI.PropertyGetReadonly(d.commitPath(commit, shard))
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("%+v is not a read only commit", commit)
	}
	return nil
}

func (d *btrfsDriver) checkWrite(commit *pfs.Commit, shard int) error {
	ok, err := d.btrfsAPI.PropertyGetReadonly(d.commitPath(commit, shard))
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf("%+v is not a write commit", commit)
	}
	return nil
}

func (d *btrfsDriver) repositoryPath(repository *pfs.Repository) string {
	return filepath.Join(d.rootDir, repository.Name)
}

func (d *btrfsDriver) commitPathNoShard(commit *pfs.Commit) string {
	return filepath.Join(d.repositoryPath(commit.Repository), commit.Id)
}

func (d *btrfsDriver) commitPath(commit *pfs.Commit, shard int) string {
	return filepath.Join(d.commitPathNoShard(commit), fmt.Sprintf("%d", shard))
}

func (d *btrfsDriver) filePath(path *pfs.Path, shard int) string {
	return filepath.Join(d.commitPath(path.Commit, shard), path.Path)
}

func newCommitID() string {
	return strings.Replace(common.NewUUID(), "-", "", -1)
}
