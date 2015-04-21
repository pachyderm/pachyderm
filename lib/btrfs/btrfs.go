package btrfs

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"code.google.com/p/go-uuid/uuid"
	"github.com/go-fsnotify/fsnotify"
	"github.com/pachyderm/pfs/lib/shell"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var once sync.Once

// volume is hardcoded because we can map any host directory in to this path
// using Docker's `volume`s.
var volume = "/var/lib/pfs/vol"

// Generates a random sequence of letters. Useful for making filesystems that won't interfere with each other.
// This should be factored out to another file.
func RandSeq(n int) string {
	once.Do(func() { rand.Seed(time.Now().UTC().UnixNano()) })
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func Sync() error {
	return shell.RunStderr(exec.Command("sync"))
}

func BasePath(name string) string {
	return path.Join(volume, name)
}

func FilePath(name string) string {
	return path.Join(volume, name)
}

func TrimFilePath(name string) string {
	return strings.TrimPrefix(name, volume)
}

func Create(name string) (*os.File, error) {
	return os.Create(FilePath(name))
}

func CreateAll(name string) (*os.File, error) {
	err := MkdirAll(path.Dir(name))
	if err != nil {
		return nil, err
	}
	return os.Create(FilePath(name))
}

func CreateFromReader(name string, r io.Reader) (int64, error) {
	f, err := Create(name)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return io.Copy(f, r)
}

func Open(name string) (*os.File, error) {
	return os.Open(FilePath(name))
}

func OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(FilePath(name), flag, perm)
}

func OpenFd(name string, mode int, perm uint32) (int, error) {
	return syscall.Open(FilePath(name), mode, perm)
}

func WriteFile(name string, r io.Reader) (int64, error) {
	f, err := Open(name)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return io.Copy(f, r)
}

func Remove(name string) error {
	return os.Remove(FilePath(name))
}

func RemoveAll(name string) error {
	return os.RemoveAll(FilePath(name))
}

func Rename(oldname, newname string) error {
	return os.Rename(FilePath(oldname), FilePath(newname))
}

func FileExists(name string) (bool, error) {
	_, err := os.Stat(FilePath(name))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func Mkdir(name string) error {
	return os.Mkdir(FilePath(name), 0777)
}

// TODO(rw,jd): check into atomicity/race conditions with multiple callers
func MkdirAll(name string) error {
	return os.MkdirAll(FilePath(name), 0777)
}

func Link(oldname, newname string) error {
	return os.Link(FilePath(oldname), FilePath(newname))
}

func Readlink(name string) (string, error) {
	p, err := os.Readlink(FilePath(name))
	if err != nil {
		return "", err
	}
	return TrimFilePath(p), nil
}

func Symlink(oldname, newname string) error {
	return os.Symlink(FilePath(oldname), FilePath(newname))
}

func ReadDir(name string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(FilePath(name))
}

var walkChunk int = 100

func LazyWalk(name string, f func(string) error) error {
	dir, err := os.Open(FilePath(name))
	if err != nil {
		return nil
	}
	defer dir.Close()
	var names []string
	for names, err = dir.Readdirnames(walkChunk); err == nil; names, err = dir.Readdirnames(walkChunk) {
		for _, fname := range names {
			err := f(fname)
			if err != nil {
				return err
			}
		}
	}
	if err != io.EOF {
		return err
	}
	return nil
}

func WaitForFile(name string) error {
	if err := MkdirAll(path.Dir(name)); err != nil {
		log.Print(err)
		return err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Print(err)
		return err
	}
	defer watcher.Close()

	if err := watcher.Add(FilePath(path.Dir(name))); err != nil {
		log.Print(err)
		return err
	}

	exists, err := FileExists(name)
	if err != nil {
		log.Print(err)
		return err
	}
	if exists {
		log.Print("File exists, returning.")
		return nil
	}

	log.Print("File doesn't exist. Waiting for it...")
	for {
		select {
		case event := <-watcher.Events:
			log.Print(event)
			if event.Op == fsnotify.Create && event.Name == FilePath(name) {
				return nil
			}
		case err := <-watcher.Errors:
			log.Print(err)
			return err
		}
	}
	return nil
}

func SubvolumeCreate(name string) error {
	return shell.RunStderr(exec.Command("btrfs", "subvolume", "create", FilePath(name)))
}

func SubvolumeDelete(name string) error {
	return shell.RunStderr(exec.Command("btrfs", "subvolume", "delete", FilePath(name)))
}

func Snapshot(volume string, dest string, readonly bool) error {
	if readonly {
		return shell.RunStderr(exec.Command("btrfs", "subvolume", "snapshot", "-r",
			FilePath(volume), FilePath(dest)))
	} else {
		return shell.RunStderr(exec.Command("btrfs", "subvolume", "snapshot",
			FilePath(volume), FilePath(dest)))
	}
}

func SetReadOnly(volume string) error {
	return shell.RunStderr(exec.Command("btrfs", "property", "set", FilePath(volume), "ro", "true"))
}

func UnsetReadOnly(volume string) error {
	return shell.RunStderr(exec.Command("btrfs", "property", "set", FilePath(volume), "ro", "false"))
}

func SendBase(to string, cont func(io.ReadCloser) error) error {
	c := exec.Command("btrfs", "send", FilePath(to))
	return shell.CallCont(c, cont)
}

func Send(from string, to string, cont func(io.ReadCloser) error) error {
	c := exec.Command("btrfs", "send", "-p", FilePath(from), FilePath(to))
	return shell.CallCont(c, cont)
}

func Recv(volume string, data io.ReadCloser) error {
	c := exec.Command("btrfs", "receive", FilePath(volume))
	log.Print(c)
	stdin, err := c.StdinPipe()
	if err != nil {
		return err
	}
	stderr, err := c.StderrPipe()
	if err != nil {
		return err
	}
	err = c.Start()
	if err != nil {
		return err
	}
	n, err := io.Copy(stdin, data)
	if err != nil {
		return err
	}
	log.Print("Copied bytes:", n)
	err = stdin.Close()
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(stderr)
	log.Print("Stderr:", buf)

	return c.Wait()
}

// Init initializes an empty repo.
func Init(repo string) error {
	if err := SubvolumeCreate(repo); err != nil {
		return err
	}
	if err := SubvolumeCreate(path.Join(repo, "master")); err != nil {
		return err
	}
	if err := Commit(repo, "t0", "master"); err != nil {
		return err
	}
	return nil
}

// Ensure is like Init but won't error if the repo is already present. It will
// error if the repo is not present and we fail to make it.
func Ensure(repo string) error {
	exists, err := FileExists(path.Join(repo, "master"))
	if err != nil {
		return err
	}
	if exists {
		return nil
	} else {
		return Init(repo)
	}
}

// Commit creates a new commit for a branch.
func Commit(repo, commit, branch string) error {
	// First we check to make sure that the branch actually exists
	exists, err := FileExists(path.Join(repo, branch))
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("Branch %s not found.", branch)
	}
	// First we make HEAD readonly
	if err := SetReadOnly(path.Join(repo, branch)); err != nil {
		return err
	}
	// Next move it to being a commit
	if err := Rename(path.Join(repo, branch), path.Join(repo, commit)); err != nil {
		return err
	}
	// Recreate the branch subvolume with a writeable subvolume
	if err := Snapshot(path.Join(repo, commit), path.Join(repo, branch), false); err != nil {
		return err
	}

	return nil
}

// Hold creates a temporary snapshot of a commit that no one else knows about.
// It's your responsibility to release the snapshot with Release
func Hold(repo, commit string) (string, error) {
	MkdirAll("tmp")
	name := path.Join("tmp", uuid.New())
	if err := Snapshot(path.Join(repo, commit), name, false); err != nil {
		return "", err
	}
	return name, nil
}

// Release releases commit snapshots held by Hold.
func Release(name string) {
	SubvolumeDelete(name)
}

func Branch(repo, commit, branch string) error {
	if err := Snapshot(path.Join(repo, commit), path.Join(repo, branch), false); err != nil {
		return err
	}
	return nil
}

//Log returns all of the commits the repo which have generation >= from.
func Log(repo, from string, cont func(io.ReadCloser) error) error {
	if from == "" {
		c := exec.Command("btrfs", "subvolume", "list", "-o", "-c", "-u", "-q", "--sort", "-ogen", FilePath(path.Join(repo)))
		return shell.CallCont(c, cont)
	} else {
		c := exec.Command("btrfs", "subvolume", "list", "-o", "-c", "-u", "-q", "-C", "+"+from, "--sort", "-ogen", FilePath(path.Join(repo)))
		return shell.CallCont(c, cont)
	}
}

type CommitInfo struct {
	gen, id, parent, path string
}

func Commits(repo, from string, cont func(CommitInfo) error) error {
	return Log(repo, from, func(r io.ReadCloser) error {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			// scanner.Text() looks like:
			// ID 299 gen 67 cgen 66 top level 292 parent_uuid 7a4a824b-7b78-d144-a956-eb0229616d21 uuid c1cd770c-600b-a744-940c-835bf73b5fa9 path repo/25853824-60a8-4d32-9168-adfce78a6c91
			// 0  1   2   3  4    5  6   7     8   9           10                                   11   12                                   13   14
			tokens := strings.Split(scanner.Text(), " ")
			if len(tokens) != 15 {
				return fmt.Errorf("Malformed commit line: %s.", scanner.Text())
			}
			_, p := path.Split(tokens[14]) // we want to returns paths without the repo/ before them
			if err := cont(CommitInfo{tokens[3], tokens[12], tokens[10], p}); err != nil {
				return err
			}
		}
		return nil
	})
}

type Diff struct {
	parent, child *CommitInfo
}

func Pull(repo, from string, cont func(io.ReadCloser) error) error {
	// commits indexed by their parents
	parentMap := make(map[string][]CommitInfo)
	var diffs []Diff
	// the body below gets called once per commit
	err := Commits(repo, from, func(c CommitInfo) error {
		// first we check if it's above the cutoff
		// We don't do this if the parent is null (represented by "-")
		if c.gen > from && c.parent != "-" {
			// this commit is part of the pull so we put it in the parentMap
			parentMap[c.parent] = append(parentMap[c.parent], c)
		}
		// Now we pop all of the commits for which this was the parent out of
		// the map
		for _, child := range parentMap[c.id] {
			diffs = append(diffs, Diff{&c, &child})
		}
		delete(parentMap, c.id)

		if len(parentMap) == 0 {
			if c.parent == "-" {
				diffs = append(diffs, Diff{nil, &c})
			}
			return fmt.Errorf("COMPLETE")
		} else {
			return nil
		}
	})
	if err != nil && err.Error() != "COMPLETE" {
		return err
	}

	for _, diff := range diffs {
		if diff.parent == nil {
			if err := SendBase(path.Join(repo, diff.child.path), cont); err != nil {
				return err
			}
		}
		if err := Send(path.Join(repo, diff.parent.path), path.Join(repo, diff.child.path), cont); err != nil {
			return err
		}
	}

	return nil
}

// Transid returns transid of a path in a repo. This value is useful for
// passing to FindNew.
func Transid(repo, commit string) (string, error) {
	//  "9223372036854775810" == 2 ** 63 we use a very big number there so that
	//  we get the transid of the from path. According to the internet this is
	//  the nicest way to get it from btrfs.
	var transid string
	c := exec.Command("btrfs", "subvolume", "find-new", FilePath(path.Join(repo, commit)), "9223372036854775808")
	err := shell.CallCont(c, func(r io.ReadCloser) error {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			// scanner.Text() looks like this:
			// transid marker was 907
			// 0       1      2   3
			tokens := strings.Split(scanner.Text(), " ")
			if len(tokens) != 4 {
				return fmt.Errorf("Failed to parse find-new output.")
			}
			// We want to increment the transid because it's inclusive, if we
			// don't increment we'll get things from the previous commit as
			// well.
			_transid, err := strconv.Atoi(tokens[3])
			if err != nil {
				return err
			}
			_transid++
			transid = strconv.Itoa(_transid)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return transid, err
}

// FindNew returns an array of filenames that have been created since transid.
// transid should come from Transid.
func FindNew(repo, commit, transid string) ([]string, error) {
	var files []string
	c := exec.Command("btrfs", "subvolume", "find-new", FilePath(path.Join(repo, commit)), transid)
	err := shell.CallCont(c, func(r io.ReadCloser) error {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			// scanner.Text() looks like this:
			// inode 6683 file offset 0 len 107 disk start 0 offset 0 gen 909 flags INLINE jobs/rPqZxsaspy
			// 0     1    2    3      4 5   6   7    8     9 10     11 12 13 14     15     16
			tokens := strings.Split(scanner.Text(), " ")
			if len(tokens) == 17 {
				files = append(files, tokens[16])
			} else if len(tokens) == 4 {
				continue //skip transid messages
			} else {
				return fmt.Errorf("Failed to parse find-new output.")
			}
		}
		return nil
	})
	return files, err
}

// NewFiles returns the new files in
func NewFiles(repo, commit string) ([]string, error) {
	var parentId, parent string
	err := Commits(repo, "", func(c CommitInfo) error {
		if c.path == commit {
			parentId = c.parent
			if parentId == "-" {
				// This case indicates no parent, we handle it below.
				return fmt.Errorf("COMPLETE")
			} else {
				return nil
			}
		}
		// When this function is first called parentId == "" this changes only
		// when we find the commit above and learn what the parent is. Commits
		// orders by generation so the parent shows up after the commit which
		// means parentId should be by the time we see it
		if parentId != "" && c.id == parentId {
			parent = c.path
			return fmt.Errorf("COMPLETE")
		}
		return nil
	})
	if err != nil && err.Error() != "COMPLETE" {
		return []string{}, nil
	}

	if parent == "" {
		return []string{}, fmt.Errorf("Failed to find parent for commit.")
	}

	if parent == "-" {
		// No parent was found, everything in the subvolume is new.
		return FindNew(repo, commit, "0")
	} else {
		transid, err := Transid(repo, parent)
		if err != nil {
			return []string{}, err
		}
		return FindNew(repo, commit, transid)
	}
}
