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

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pkg/executil"
	"github.com/satori/go.uuid"
	"go.pedge.io/proto/time"
	"go.pedge.io/protolog"
)

const (
	metadataDir = ".pfs"
)

type driver struct {
	rootDir   string
	namespace string
}

func newDriver(rootDir string, namespace string) (*driver, error) {
	if namespace != "" {
		if err := os.MkdirAll(filepath.Join(rootDir, namespace), 0700); err != nil {
			return nil, err
		}
	}
	return &driver{rootDir, namespace}, nil
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
	repositories, err := ioutil.ReadDir(d.basePath())
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

func (d *driver) StartCommit(parent *pfs.Commit, commit *pfs.Commit, shards map[uint64]bool) (*pfs.Commit, error) {
	if parent == nil && commit == nil {
		return nil, fmt.Errorf("pachyderm: must specify either parent or commit")
	}
	if commit == nil {
		commit = &pfs.Commit{
			Repo: parent.Repo,
			Id:   newCommitID(),
		}
	}
	if err := execSubvolumeCreate(d.commitPathNoShard(commit)); err != nil && !execSubvolumeExists(d.commitPathNoShard(commit)) {
		return nil, err
	}
	for shard := range shards {
		commitPath := d.writeCommitPath(commit, shard)
		if parent != nil {
			if err := d.checkReadOnly(parent, shard); err != nil {
				return nil, err
			}
			parentPath, err := d.commitPath(parent, shard)
			if err != nil {
				return nil, err
			}
			if err := execSubvolumeSnapshot(parentPath, commitPath, false); err != nil {
				return nil, err
			}
			filePath, err := d.filePath(&pfs.File{Commit: commit, Path: filepath.Join(metadataDir, "parent")}, shard)
			if err != nil {
				return nil, err
			}
			if err := ioutil.WriteFile(filePath, []byte(parent.Id), 0600); err != nil {
				return nil, err
			}
		} else {
			if err := execSubvolumeCreate(commitPath); err != nil {
				return nil, err
			}
			filePath, err := d.filePath(&pfs.File{Commit: commit, Path: metadataDir}, shard)
			if err != nil {
				return nil, err
			}
			if err := os.Mkdir(filePath, 0700); err != nil {
				return nil, err
			}
		}
	}
	return commit, nil
}

func (d *driver) FinishCommit(commit *pfs.Commit, shards map[uint64]bool) error {
	for shard := range shards {
		if err := execSubvolumeSnapshot(d.writeCommitPath(commit, shard), d.readCommitPath(commit, shard), true); err != nil {
			return err
		}
		if err := execSubvolumeDelete(d.writeCommitPath(commit, shard)); err != nil {
			return err
		}
	}
	return nil
}

func (d *driver) InspectCommit(commit *pfs.Commit, shard uint64) (*pfs.CommitInfo, error) {
	if !execSubvolumeExists(d.readCommitPath(commit, shard)) && !execSubvolumeExists(d.writeCommitPath(commit, shard)) {
		return nil, nil // returning nil means not found
	}
	parent, err := d.getParent(commit, shard)
	if err != nil {
		return nil, err
	}
	readOnly, err := d.getReadOnly(commit, shard)
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

func (d *driver) ListCommit(repo *pfs.Repo, from *pfs.Commit, shard uint64) ([]*pfs.CommitInfo, error) {
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
			shard,
		)
		if commitInfo == nil {
			// This is a really weird error to get since we got this commit
			// name by listing commits. This is probably indicative of a
			// race condition.
			return nil, fmt.Errorf("commit %s should exist", commitID)
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

func (d *driver) PutFile(file *pfs.File, shard uint64, offset int64, reader io.Reader) error {
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
	defer osFile.Close()
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
	for names, err := dir.Readdirnames(100); err != io.EOF; names, err = dir.Readdirnames(100) {
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
	parent, err := d.getParent(commit, shard)
	if err != nil {
		return err
	}
	if parent == nil {
		return execSend(d.readCommitPath(commit, shard), "", diff)
	}
	return execSend(d.readCommitPath(commit, shard), d.readCommitPath(parent, shard), diff)
}

func (d *driver) PushDiff(commit *pfs.Commit, diff io.Reader) error {
	if err := execSubvolumeCreate(d.commitPathNoShard(commit)); err != nil && !execSubvolumeExists(d.commitPathNoShard(commit)) {
		return err
	}
	return execRecv(d.commitPathNoShard(commit), diff)
}

func newCommitID() string {
	return strings.Replace(uuid.NewV4().String(), "-", "", -1)
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
		LastModified: prototime.TimeToTimestamp(
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
		return false, fmt.Errorf("pachyderm: commit %s doesn't exist", commit.Id)
	}
}

func (d *driver) basePath() string {
	return filepath.Join(d.rootDir, d.namespace)
}

func (d *driver) repoPath(repo *pfs.Repo) string {
	return filepath.Join(d.basePath(), repo.Name)
}

func (d *driver) commitPathNoShard(commit *pfs.Commit) string {
	return filepath.Join(d.repoPath(commit.Repo), commit.Id)
}

func (d *driver) readCommitPath(commit *pfs.Commit, shard uint64) string {
	return filepath.Join(d.commitPathNoShard(commit), fmt.Sprint(shard))
}

func (d *driver) writeCommitPath(commit *pfs.Commit, shard uint64) string {
	return d.readCommitPath(commit, shard) + ".write"
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

func inMetadataDir(name string) bool {
	parts := strings.Split(name, "/")
	return (len(parts) > 0 && parts[0] == metadataDir)
}

func execSubvolumeCreate(path string) (retErr error) {
	defer func() {
		protolog.Info(&SubvolumeCreate{path, errorToString(retErr)})
	}()
	return executil.Run("btrfs", "subvolume", "create", path)
}

func execSubvolumeDelete(path string) (retErr error) {
	defer func() {
		protolog.Info(&SubvolumeDelete{path, errorToString(retErr)})
	}()
	return executil.Run("btrfs", "subvolume", "delete", path)
}

func execSubvolumeExists(path string) (result bool) {
	defer func() {
		protolog.Info(&SubvolumeExists{path, result})
	}()
	if err := executil.Run("btrfs", "subvolume", "show", path); err != nil {
		return false
	}
	return true
}

func execSubvolumeSnapshot(src string, dest string, readOnly bool) (retErr error) {
	defer func() {
		protolog.Info(&SubvolumeSnapshot{src, dest, readOnly, errorToString(retErr)})
	}()
	if readOnly {
		return executil.Run("btrfs", "subvolume", "snapshot", "-r", src, dest)
	}
	return executil.Run("btrfs", "subvolume", "snapshot", src, dest)
}

func execTransID(path string) (result string, retErr error) {
	defer func() {
		protolog.Info(&TransID{path, result, errorToString(retErr)})
	}()
	//  "9223372036854775810" == 2 ** 63 we use a very big number there so that
	//  we get the transid of the from path. According to the internet this is
	//  the nicest way to get it from btrfs.
	var buffer bytes.Buffer
	if err := executil.RunStdout(&buffer, "btrfs", "subvolume", "find-new", path, "9223372036854775808"); err != nil {
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
		protolog.Info(&SubvolumeList{path, fromCommit, ascending, errorToString(retErr)})
	}()
	var sort string
	if ascending {
		sort = "+ogen"
	} else {
		sort = "-ogen"
	}

	if fromCommit == "" {
		return executil.RunStdout(out, "btrfs", "subvolume", "list", "-a", "--sort", sort, path)
	}
	transid, err := execTransID(fromCommit)
	if err != nil {
		return err
	}
	return executil.RunStdout(out, "btrfs", "subvolume", "list", "-aC", "+"+transid, "--sort", sort, path)
}

func execSubvolumeFindNew(commit string, fromCommit string, out io.Writer) (retErr error) {
	defer func() {
		protolog.Info(&SubvolumeFindNew{commit, fromCommit, errorToString(retErr)})
	}()
	if fromCommit == "" {
		return executil.RunStdout(out, "btrfs", "subvolume", "find-new", commit, "0")
	}
	transid, err := execTransID(fromCommit)
	if err != nil {
		return err
	}
	return executil.RunStdout(out, "btrfs", "subvolume", "find-new", commit, transid)
}

func execSend(path string, parent string, diff io.Writer) (retErr error) {
	defer func() {
		protolog.Info(&Send{path, parent, errorToString(retErr)})
	}()
	if parent == "" {
		return executil.RunStdout(diff, "btrfs", "send", path)
	}
	return executil.RunStdout(diff, "btrfs", "send", "-p", parent, path)
}

func execRecv(path string, diff io.Reader) (retErr error) {
	defer func() {
		protolog.Info(&Recv{path, errorToString(retErr)})
	}()
	return executil.RunStdin(diff, "btrfs", "receive", path)
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
	return commit
}

func (c *commitScanner) parseCommit() (string, bool) {
	// c.textScanner.Text() looks like:
	// ID 905 gen 865 top level 5 path <FS_TREE>/[namespace/]repo/commit
	// 0  1   2   3   4   5     6 7    8
	tokens := strings.Split(c.textScanner.Text(), " ")
	if len(tokens) != 9 {
		return "", false
	}
	if c.namespace == "" {
		if strings.HasPrefix(tokens[8], filepath.Join("<FS_TREE>", c.repo)) && len(strings.Split(tokens[8], "/")) == 3 {
			return strings.Split(tokens[8], "/")[2], true
		}
	} else {
		if strings.HasPrefix(tokens[8], filepath.Join("<FS_TREE>", c.namespace, c.repo)) && len(strings.Split(tokens[8], "/")) == 4 {
			return strings.Split(tokens[8], "/")[3], true
		}
	}
	return "", false
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
