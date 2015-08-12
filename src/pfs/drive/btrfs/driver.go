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

/*
#include <stdlib.h>
#include <dirent.h>
#include <btrfs/ioctl.h>
*/
import "C"

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/peter-edge/go-google-protobuf"
	"github.com/satori/go.uuid"
)

const (
	metadataDir = ".pfs"
)

var (
	callStrings = map[uintptr]string{
		C.BTRFS_IOC_SUBVOL_CREATE:  "btrfs subvolume create",
		C.BTRFS_IOC_SNAP_CREATE_V2: "btrfs subvolume snapshot",
	}
)

type driver struct {
	rootDir string
}

func newDriver(rootDir string) *driver {
	return &driver{rootDir}
}

func (d *driver) InitRepository(repository *pfs.Repository, shards map[int]bool) error {
	if err := os.MkdirAll(d.repositoryPath(repository), 0700); err != nil {
		return err
	}
	initialCommit := &pfs.Commit{
		Repository: repository,
		Id:         drive.InitialCommitID,
	}
	if err := os.MkdirAll(d.commitPathNoShard(initialCommit), 0700); err != nil {
		return err
	}
	for shard := range shards {
		if err := ffiSubvolumeCreate(d.commitPath(initialCommit, shard)); err != nil {
			return err
		}
		if err := os.Mkdir(d.filePath(&pfs.Path{Commit: initialCommit, Path: metadataDir}, shard), 0700); err != nil {
			return err
		}
		if err := d.setReadOnly(initialCommit, shard, true); err != nil {
			return err
		}
	}
	return nil
}

func (d *driver) GetFile(path *pfs.Path, shard int) (drive.ReaderAtCloser, error) {
	return os.Open(d.filePath(path, shard))
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
	// TODO(pedge): check that commit exists and is a write commit
	for shard := range shards {
		if err := d.checkWrite(path.Commit, shard); err != nil {
			return err
		}
		if err := os.MkdirAll(d.filePath(path, shard), 0700); err != nil {
			return err
		}
	}
	return nil
}

func (d *driver) PutFile(path *pfs.Path, shard int, offset int64, reader io.Reader) error {
	if err := d.checkWrite(path.Commit, shard); err != nil {
		return err
	}
	file, err := os.OpenFile(d.filePath(path, shard), os.O_CREATE|os.O_WRONLY, 0666)
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

func (d *driver) Branch(commit *pfs.Commit, newCommit *pfs.Commit, shards map[int]bool) (*pfs.Commit, error) {
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
		if err := ffiSubvolumeSnapshot(commitPath, newCommitPath); err != nil {
			return nil, err
		}
		if err := ioutil.WriteFile(d.filePath(&pfs.Path{Commit: newCommit, Path: filepath.Join(metadataDir, "parent")}, shard), []byte(commit.Id), 0600); err != nil {
			return nil, err
		}
		if err := d.setReadOnly(newCommit, shard, false); err != nil {
			return nil, err
		}
	}
	return newCommit, nil
}

func (d *driver) Commit(commit *pfs.Commit, shards map[int]bool) error {
	for shard := range shards {
		if err := d.checkWrite(commit, shard); err != nil {
			return err
		}
		if err := d.setReadOnly(commit, shard, true); err != nil {
			return err
		}
	}
	return nil
}

func (d *driver) PullDiff(commit *pfs.Commit, shard int) (io.Reader, error) {
	return nil, nil
}

func (d *driver) PushDiff(commit *pfs.Commit, shard int, reader io.Reader) error {
	return nil
}

func (d *driver) GetCommitInfo(commit *pfs.Commit, shard int) (_ *pfs.CommitInfo, ok bool, _ error) {
	_, err := os.Stat(d.commitPath(commit, shard))
	if err != nil && os.IsNotExist(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
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
	data, err := ioutil.ReadFile(d.filePath(&pfs.Path{Commit: commit, Path: filepath.Join(metadataDir, "parent")}, shard))
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
	data, err := ioutil.ReadFile(d.filePath(&pfs.Path{Commit: commit, Path: filepath.Join(metadataDir, "readonly")}, shard))
	if err != nil {
		return false, err
	}
	switch string(data) {
	case "0":
		return false, nil
	case "1":
		return true, nil
	default:
		return false, fmt.Errorf("unknown data for readonly metadata file: %s", string(data))
	}
}

func (d *driver) setReadOnly(commit *pfs.Commit, shard int, readOnly bool) error {
	data := []byte("0")
	if readOnly {
		data = []byte("1")
	}
	return ioutil.WriteFile(d.filePath(&pfs.Path{Commit: commit, Path: filepath.Join(metadataDir, "readonly")}, shard), data, 0400)
}

func (d *driver) repositoryPath(repository *pfs.Repository) string {
	return filepath.Join(d.rootDir, repository.Name)
}

func (d *driver) commitPathNoShard(commit *pfs.Commit) string {
	return filepath.Join(d.repositoryPath(commit.Repository), commit.Id)
}

func (d *driver) commitPath(commit *pfs.Commit, shard int) string {
	return filepath.Join(d.commitPathNoShard(commit), fmt.Sprintf("%d", shard))
}

func (d *driver) filePath(path *pfs.Path, shard int) string {
	return filepath.Join(d.commitPath(path.Commit, shard), path.Path)
}

func newCommitID() string {
	return strings.Replace(uuid.NewV4().String(), "-", "", -1)
}

func inMetadataDir(name string) bool {
	parts := strings.Split(name, "/")
	return (len(parts) > 0 && parts[0] == metadataDir)
}

func ffiSubvolumeCreate(path string) error {
	var args C.struct_btrfs_ioctl_vol_args
	for i, c := range []byte(filepath.Base(path)) {
		args.name[i] = C.char(c)
	}
	return ffiIoctl(filepath.Dir(path), C.BTRFS_IOC_SUBVOL_CREATE, uintptr(unsafe.Pointer(&args)))
}

func ffiSubvolumeSnapshot(src string, dest string) error {
	srcDir, err := ffiOpenDir(src)
	if err != nil {
		return err
	}
	defer ffiCloseDir(srcDir)
	var args C.struct_btrfs_ioctl_vol_args_v2
	args.fd = C.__s64(ffiGetDirFd(srcDir))
	for i, c := range []byte(filepath.Base(dest)) {
		args.name[i] = C.char(c)
	}
	return ffiIoctl(filepath.Dir(dest), C.BTRFS_IOC_SNAP_CREATE_V2, uintptr(unsafe.Pointer(&args)))
}

func ffiIoctl(path string, call uintptr, args uintptr) error {
	dir, err := ffiOpenDir(path)
	if err != nil {
		return err
	}
	defer ffiCloseDir(dir)
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, ffiGetDirFd(dir), call, args)
	if errno != 0 {
		return fmt.Errorf("%s failed for %s: %v", callStrings[call], path, errno.Error())
	}
	return nil
}

func ffiOpenDir(path string) (*C.DIR, error) {
	Cpath := C.CString(path)
	defer C.free(unsafe.Pointer(Cpath))
	dir := C.opendir(Cpath)
	if dir == nil {
		return nil, fmt.Errorf("cannot open dir: %s", path)
	}
	return dir, nil
}

func ffiCloseDir(dir *C.DIR) {
	if dir != nil {
		C.closedir(dir)
	}
}

func ffiGetDirFd(dir *C.DIR) uintptr {
	return uintptr(C.dirfd(dir))
}
