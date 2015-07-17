/*

directory structure

  .
  |-- repositoryName
      |-- __root__
	      |-- shardNum // the writeable root created from btrfs snapshot create on InitRepository
	  |-- scratch
		  |-- shardNum // the read-only read created on InitRepository, this is where to start branching
      |-- commitID
	      |-- shardNum // this is where subvolumes are

*/
package btrfs

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
)

type driver struct {
	rootDir string
}

func newDriver(rootDir string) *driver {
	return &driver{rootDir}
}

func (d *driver) Init() error {
	return nil
}

func (d *driver) DriverType() pfs.DriverType {
	return pfs.DriverType_DRIVER_TYPE_BTRFS
}

func (d *driver) InitRepository(repository *pfs.Repository, shards map[int]bool) error {
	// syscall.Mkdir (which os.Mkdir directly calls) is atomic across processes, and since
	// we do not include shards as part of the repository path, this guaranteees
	// only one initialization will complete successfully
	if err := os.Mkdir(d.repositoryPath(repository), 0700); err != nil {
		return err
	}
	for shard := range shards {
		commitPath := d.commitPath(
			&pfs.Commit{
				Repository: repository,
				Id:         drive.SystemRootCommitID,
			},
			shard,
		)
		if err := os.MkdirAll(filepath.Dir(commitPath), 0700); err != nil {
			return err
		}
		if err := subvolumeCreate(commitPath); err != nil {
			return err
		}
	}
	return nil
}

func (d *driver) GetFile(path *pfs.Path, shard int) (io.ReadCloser, error) {
	return os.Open(d.filePath(path, shard))
}

func (d *driver) MakeDirectory(path *pfs.Path, shard int) error {
	// TODO(pedge): if PutFile fails here or on another shard, the directories
	// will still exist and be returned from ListFiles, we want to do this
	// iteratively and with rollback
	return os.MkdirAll(d.filePath(path, shard), 0700)
}

func (d *driver) PutFile(path *pfs.Path, shard int, reader io.Reader) error {
	file, err := os.Create(d.filePath(path, shard))
	if err != nil {
		return err
	}
	_, err = bufio.NewReader(reader).WriteTo(file)
	return err
}

func (d *driver) ListFiles(path *pfs.Path, shard int) ([]*pfs.Path, error) {
	return nil, nil
}

func (d *driver) GetParent(commit *pfs.Commit) (*pfs.Commit, error) {
	return nil, nil
}

func (d *driver) Branch(commit *pfs.Commit) (*pfs.Commit, error) {
	return nil, nil
}

func (d *driver) Commit(commit *pfs.Commit) error {
	return nil
}

func (d *driver) PullDiff(commit *pfs.Commit, shard int) (io.Reader, error) {
	return nil, nil
}

func (d *driver) PushDiff(commit *pfs.Commit, shard int, reader io.Reader) error {
	return nil
}

func (d *driver) GetCommitInfo(commit *pfs.Commit) (*pfs.CommitInfo, error) {
	return nil, nil
}

func (d *driver) repositoryPath(repository *pfs.Repository) string {
	return filepath.Join(d.rootDir, repository.Name)
}

func (d *driver) commitPath(commit *pfs.Commit, shard int) string {
	return filepath.Join(d.repositoryPath(commit.Repository), commit.Id, fmt.Sprintf("%d", shard))
}

func (d *driver) filePath(path *pfs.Path, shard int) string {
	return filepath.Join(d.commitPath(path.Commit, shard), path.Path)
}

func (d *driver) isReadOnly(commit *pfs.Commit, shard int) (bool, error) {
	reader, err := snapshotPropertyGet(d.commitPath(commit, shard))
	if err != nil {
		return false, err
	}
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "ro=true") {
			return true, nil
		}
	}
	return false, scanner.Err()
}

func snapshotPropertyGet(path string) (io.Reader, error) {
	return runStdout("btrfs", "property", "get", "-t", "s", path)
}

func subvolumeCreate(path string) error {
	return run("btrfs", "subvolume", "create", path)
}

func subvolumeSnapshot(src string, dest string) error {
	return run("btrfs", "subvolume", "snapshot", src, dest)
}

func subvolumeSnapshotReadonly(src string, dest string) error {
	return run("btrfs", "subvolume", "snapshot", "-r", src, dest)
}

func run(args ...string) error {
	return runWithOptions(runOptions{}, args...)
}

func runStdout(args ...string) (io.Reader, error) {
	stdout := bytes.NewBuffer(nil)
	err := runWithOptions(runOptions{stdout: stdout}, args...)
	return stdout, err
}

type runOptions struct {
	stdout io.Writer
	stderr io.Writer
}

func runWithOptions(runOptions runOptions, args ...string) error {
	if len(args) == 0 {
		return errors.New("runCmd called with no args")
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = runOptions.stdout
	cmd.Stderr = runOptions.stderr
	argsString := strings.Join(args, " ")
	log.Printf("shell: %s", argsString)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %s", argsString, err.Error())
	}
	return nil
}

func isFileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
