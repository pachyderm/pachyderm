package btrfs

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"runtime"
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
var hostVolume = "/home/jdoliner/.pfs/vol"

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

func FilePath(name string) string {
	return path.Join(volume, name)
}

func HostPath(name string) string {
	return path.Join(hostVolume, name)
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

func ReadFile(name string) ([]byte, error) {
	return ioutil.ReadFile(FilePath(name))
}

func WriteFile(name string, data []byte) error {
	err := MkdirAll(path.Dir(name))
	if err != nil {
		return err
	}
	return ioutil.WriteFile(FilePath(name), data, 0666)
}

func CopyFile(name string, r io.Reader) (int64, error) {
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

func Stat(name string) (os.FileInfo, error) {
	return os.Stat(FilePath(name))
}

func Lstat(name string) (os.FileInfo, error) {
	return os.Lstat(FilePath(name))
}

// return true if name1 was last modified before name2
func Before(name1, name2 string) (bool, error) {
	info1, err := Stat(name1)
	if err != nil {
		return false, err
	}

	info2, err := Stat(name2)
	if err != nil {
		return false, err
	}

	return info1.ModTime().Before(info2.ModTime()), nil
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

// smallestExistingPath takes a path and trims it until it gets something that
// exists
func largestExistingPath(name string) (string, error) {
	for {
		exists, err := FileExists(name)
		if err != nil {
			return "", err
		}
		if exists {
			return name, nil
		}
		name = path.Dir(name)
	}
}

func WaitForFile(name string) error {
	dir, err := largestExistingPath(name)
	if err != nil {
		log.Print(err)
		return err
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Print(err)
		return err
	}
	defer watcher.Close()

	log.Print("Adding: ", FilePath(dir))
	if err := watcher.Add(FilePath(dir)); err != nil {
		log.Print(err)
		return err
	}

	// Notice that we check to see if the file exists AFTER we create the watcher.
	// That means if we don't see the file with this check we're guaranteed to
	// get a notification for it.
	exists, err := FileExists(name)
	if err != nil {
		log.Print(err)
		return err
	}
	if exists {
		return nil
	}

	for {
		select {
		case event := <-watcher.Events:
			log.Print(event)
			if event.Op == fsnotify.Create && event.Name == FilePath(name) {
				return nil
			} else if event.Op == fsnotify.Create && strings.HasPrefix(FilePath(name), event.Name) {
				//fsnotify isn't recursive so we need to recurse for it.
				return WaitForFile(name)
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

func SubvolumeDeleteAll(name string) error {
	subvolumeExists, err := FileExists(name)
	if err != nil {
		return err
	}
	if subvolumeExists {
		return SubvolumeDelete(name)
	} else {
		return nil
	}
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

func IsReadOnly(volume string) (bool, error) {
	var res bool
	// "-t s" indicates to btrfs that this is a subvolume without the "-t s"
	// btrfs will still output what we want, but it will have a nonzero return
	// code
	err := shell.CallCont(exec.Command("btrfs", "property", "get", "-t", "s", FilePath(volume)),
		func(r io.Reader) error {
			scanner := bufio.NewScanner(r)
			for scanner.Scan() {
				if strings.Contains(scanner.Text(), "ro=true") {
					res = true
					return nil
				}
			}
			return scanner.Err()
		})
	return res, err
}

func ensureMetaDir(branch string) error {
	branchExists, err := FileExists(branch)
	if err != nil {
		return err
	}
	if !branchExists {
		fmt.Errorf("Cannot create meta dir for nonexistant branch %s.", branch)
	}
	return MkdirAll(path.Join(branch, ".meta"))
}

// SetMeta sets metadata for a branch.
func SetMeta(branch, key, value string) error {
	if err := ensureMetaDir(branch); err != nil {
		return err
	}
	return WriteFile(path.Join(branch, ".meta", key), []byte(value))
}

// GetMeta gets metadata from a commit.
func GetMeta(name, key string) string {
	value, err := ReadFile(path.Join(name, ".meta", key))
	if err != nil {
		return ""
	}
	return string(value)
}

func Send(repo, commit string, cont func(io.Reader) error) error {
	parent := GetMeta(path.Join(repo, commit), "parent")
	if parent == "" {
		return shell.CallCont(exec.Command("btrfs", "send", FilePath(path.Join(repo, commit))), cont)
	} else {
		return shell.CallCont(exec.Command("btrfs", "send", "-p",
			FilePath(path.Join(repo, parent)), FilePath(path.Join(repo, commit))), cont)
	}
}

// createNewBranch gets called after a new commit has been `Recv`ed it creates
// the branch that should be pointing to the newly made commit.
func createNewBranch(repo string) error {
	err := Commits(repo, "", Desc, func(c CommitInfo) error {
		branch := GetMeta(path.Join(repo, c.Path), "branch")
		err := SubvolumeDeleteAll(path.Join(repo, branch))
		if err != nil {
			return err
		}
		err = Branch(repo, c.Path, branch)
		if err != nil {
			return err
		}
		return Complete
	})
	if err != nil && err != Complete {
		return err
	}
	return nil
}

func Recv(repo string, data io.Reader) error {
	c := exec.Command("btrfs", "receive", FilePath(repo))
	_, callerFile, callerLine, _ := runtime.Caller(0)
	log.Printf("%15s:%.3d -> %s", path.Base(callerFile), callerLine, strings.Join(c.Args, " "))
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

	err = c.Wait()
	if err != nil {
		return err
	}
	createNewBranch(repo)
	return nil
}

// Init initializes an empty repo.
func Init(repo string) error {
	if err := SubvolumeCreate(repo); err != nil {
		return err
	}
	if err := SubvolumeCreate(path.Join(repo, "master")); err != nil {
		return err
	}
	if err := SetMeta(path.Join(repo, "master"), "branch", "master"); err != nil {
		return err
	}
	return nil
}

// Ensure is like Init but won't error if the repo is already present. It will
// error if the repo is not present and we fail to make it.
func Ensure(repo string) error {
	exists, err := FileExists(repo)
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
	// check to make sure that the branch actually exists
	exists, err := FileExists(path.Join(repo, branch))
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("Branch %s not found.", branch)
	}
	// Snapshot the branch
	if err := Snapshot(path.Join(repo, branch), path.Join(repo, commit), true); err != nil {
		return err
	}

	// Record the new commit as the parent of this branch
	if err := SetMeta(path.Join(repo, branch), "parent", commit); err != nil {
		return err
	}

	return nil
}

// DanglingCommit creates a commit but resets the branch to point to its
// current parent
func DanglingCommit(repo, commit, branch string) error {
	parent := GetMeta(path.Join(repo, branch), "parent")
	err := Commit(repo, commit, branch)
	if err != nil {
		return err
	}
	err = SubvolumeDelete(path.Join(repo, branch))
	if err != nil {
		return err
	}
	err = Branch(repo, parent, branch)
	if err != nil {
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
	// Check that the commit is read only
	if commit != "" {
		isReadOnly, err := IsReadOnly(path.Join(repo, commit))
		if err != nil {
			return err
		}
		if !isReadOnly {
			return fmt.Errorf("Illegal branch from branch: \"%s\", can only branch from commits.", commit)
		}
	}

	// Check that the branch doesn't exist
	exists, err := FileExists(path.Join(repo, branch))
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("Branch \"%s\" already exists.", branch)
	}

	// Create a writeable subvolume for the branch
	if commit == "" {
		if err := SubvolumeCreate(path.Join(repo, branch)); err != nil {
			return err
		}
	} else {
		if err := Snapshot(path.Join(repo, commit), path.Join(repo, branch), false); err != nil {
			return err
		}

		// Record commit as the parent of this branch
		if err := SetMeta(path.Join(repo, branch), "parent", commit); err != nil {
			return err
		}
	}

	// Record the name of the branch
	if err := SetMeta(path.Join(repo, branch), "branch", branch); err != nil {
		return err
	}
	return nil
}

// Constants used for passing to log
const (
	Desc = iota
	Asc  = iota
)

//Log returns all of the commits the repo which have generation >= from.
func Log(repo, from string, order int, cont func(io.Reader) error) error {
	var sort string
	if order == Desc {
		sort = "-ogen"
	} else {
		sort = "+ogen"
	}

	if from == "" {
		c := exec.Command("btrfs", "subvolume", "list", "-o", "-c", "-u", "-q", "--sort", sort, FilePath(path.Join(repo)))
		return shell.CallCont(c, cont)
	} else {
		t, err := transid(repo, from)
		if err != nil {
			return err
		}
		c := exec.Command("btrfs", "subvolume", "list", "-o", "-c", "-u", "-q", "-C", "+"+t, "--sort", sort, FilePath(path.Join(repo)))
		return shell.CallCont(c, cont)
	}
}

type CommitInfo struct {
	gen, id, parent, Path string
}

var Complete = errors.New("Complete")

// Commits is a wrapper around `Log` which parses the output in to a convenient
// struct
func Commits(repo, from string, order int, cont func(CommitInfo) error) error {
	return Log(repo, from, order, func(r io.Reader) error {
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
			if err := cont(CommitInfo{gen: tokens[3], id: tokens[12], parent: tokens[10], Path: p}); err != nil {
				return err
			}
		}
		if scanner.Err() != nil && scanner.Err() != io.EOF {
			return scanner.Err()
		}
		return nil
	})
}

// GetFrom returns the commit that this repo should pass to Pull to get itself up
// to date.
func GetFrom(repo string) (string, error) {
	from := ""
	err := Commits(repo, "", Desc, func(c CommitInfo) error {
		isCommit, err := IsReadOnly(path.Join(repo, c.Path))
		if err != nil {
			return err
		}
		if isCommit {
			from = c.Path
			return Complete
		}
		return nil
	})
	if err != nil && err != Complete {
		return "", err
	}

	return from, nil
}

func Pull(repo, from string, cb Pusher) error {
	// First check that `from` is actually a valid commit
	if from != "" {
		exists, err := FileExists(path.Join(repo, from))
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("`from` commit %s does not exists", from)
		}
		// from should also be a commit not a branch
		isCommit, err := IsReadOnly(path.Join(repo, from))
		if err != nil {
			return err
		}
		if !isCommit {
			return fmt.Errorf("`from` %s cannot be a branch", from)
		}
	}

	err := Commits(repo, from, Asc, func(c CommitInfo) error {
		if c.Path == from {
			// Commits gives us things >= `from` so we explicitly skip `from`
			return nil
		}
		// Send this commit
		isCommit, err := IsReadOnly(path.Join(repo, c.Path))
		if err != nil {
			return err
		}
		if isCommit {
			err := Send(repo, c.Path, cb.Push)
			if err != nil {
				log.Print(err)
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// transid returns transid of a path in a repo. This function is used in
// several other internal functions.
func transid(repo, commit string) (string, error) {
	//  "9223372036854775810" == 2 ** 63 we use a very big number there so that
	//  we get the transid of the from path. According to the internet this is
	//  the nicest way to get it from btrfs.
	var transid string
	c := exec.Command("btrfs", "subvolume", "find-new", FilePath(path.Join(repo, commit)), "9223372036854775808")
	err := shell.CallCont(c, func(r io.Reader) error {
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
			transid = tokens[3]
		}
		return scanner.Err()
	})
	if err != nil {
		return "", err
	}
	return transid, err
}

// FindNew returns an array of filenames that were created or modified between `from` and `to`
func FindNew(repo, from, to string) ([]string, error) {
	var files []string
	t, err := transid(repo, from)
	if err != nil {
		return files, err
	}
	c := exec.Command("btrfs", "subvolume", "find-new", FilePath(path.Join(repo, to)), t)
	err = shell.CallCont(c, func(r io.Reader) error {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			log.Print(scanner.Text())
			// scanner.Text() looks like this:
			// inode 6683 file offset 0 len 107 disk start 0 offset 0 gen 909 flags INLINE jobs/rPqZxsaspy
			// 0     1    2    3      4 5   6   7    8     9 10     11 12 13 14     15     16
			tokens := strings.Split(scanner.Text(), " ")
			// Make sure the line is parseable as a file and the path isn't hidden.
			if len(tokens) == 17 {
				if !strings.HasPrefix(tokens[16], ".") { // check if it's a hidden file
					files = append(files, tokens[16])
				}
			} else if len(tokens) == 4 {
				continue //skip transid messages
			} else {
				return fmt.Errorf("Failed to parse find-new output.")
			}
		}
		return scanner.Err()
	})
	return files, err
}

func NewIn(repo, commit string) ([]string, error) {
	parent := GetMeta(path.Join(repo, commit), "parent")
	return FindNew(repo, parent, commit)
}
