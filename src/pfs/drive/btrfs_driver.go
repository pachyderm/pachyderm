package drive

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
)

const (
	// TODO(pedge): abstract for all drivers
	initialCommitName = "scratch"
)

type btrfsDriver struct {
	rootDir string
}

func newBtrfsDriver(rootDir string) *btrfsDriver {
	return &btrfsDriver{rootDir}
}

func (b *btrfsDriver) Init() error {
	return nil
}

func (b *btrfsDriver) DriverType() pfs.DriverType {
	return pfs.DriverType_DRIVER_TYPE_BTRFS
}

func (b *btrfsDriver) InitRepository(repository *pfs.Repository, shard int) error {
	commitPath := b.commitPath(
		&pfs.Commit{
			Repository: repository,
			Id:         initialCommitName,
		},
		shard,
	)
	if err := os.MkdirAll(filepath.Dir(commitPath), 0700); err != nil {
		return err
	}
	return subvolumeCreate(commitPath)
}

func (b *btrfsDriver) GetFile(path *pfs.Path, shard int) (io.ReadCloser, error) {
	return os.Open(b.filePath(path, shard))
}

func (b *btrfsDriver) MakeDirectory(path *pfs.Path, shard int) error {
	// TODO(pedge): if PutFile fails here or on another shard, the directories
	// will still exist and be returned from ListFiles, we want to do this
	// iteratively and with rollback
	return os.MkdirAll(b.filePath(path, shard), 0700)
}

func (b *btrfsDriver) PutFile(path *pfs.Path, shard int, reader io.Reader) error {
	file, err := os.Create(b.filePath(path, shard))
	if err != nil {
		return err
	}
	_, err = bufio.NewReader(reader).WriteTo(file)
	return err
}

func (b *btrfsDriver) ListFiles(path *pfs.Path, shard int) ([]*pfs.Path, error) {
	return nil, nil
}

func (b *btrfsDriver) GetParent(commit *pfs.Commit) (*pfs.Commit, error) {
	return nil, nil
}

func (b *btrfsDriver) Branch(commit *pfs.Commit) (*pfs.Commit, error) {
	return nil, nil
}

func (b *btrfsDriver) Commit(commit *pfs.Commit) error {
	return nil
}

func (b *btrfsDriver) PullDiff(commit *pfs.Commit, shard int) (io.Reader, error) {
	return nil, nil
}

func (b *btrfsDriver) PushDiff(commit *pfs.Commit, shard int, reader io.Reader) error {
	return nil
}

func (b *btrfsDriver) GetCommitInfo(commit *pfs.Commit) (*pfs.CommitInfo, error) {
	return nil, nil
}

func (b *btrfsDriver) commitPath(commit *pfs.Commit, shard int) string {
	return filepath.Join(b.rootDir, commit.Repository.Name, commit.Id, fmt.Sprintf("%d", shard))
}

func (b *btrfsDriver) filePath(path *pfs.Path, shard int) string {
	return filepath.Join(b.rootDir, path.Commit.Repository.Name, path.Commit.Id, fmt.Sprintf("%d", shard), path.Path)
}

func snapshotPropertyGet(path string) (io.Reader, error) {
	return runStdout("btrfs", "property", "get", "-t", "s", path)
}

func subvolumeCreate(path string) error {
	return run("btrfs", "subvolume", "create", path)
}

func subvolumeSnapshot(path string) error {
	return run("btrfs", "subvolume", "snapshot", path)
}

func subvolumeSnapshotReadonly(path string) error {
	return run("btrfs", "subvolume", "snapshot", "-r", path)
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
