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

func ReadFile(name string) ([]byte, error) {
	return ioutil.ReadFile(FilePath(name))
}

func WriteFile(name string, data []byte) error {
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
	log.Printf("btrfs.WaitForFile(%s)", name)
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
	log.Printf("btrfs.SubvolumeCreate(%s)", name)
	return shell.RunStderr(exec.Command("btrfs", "subvolume", "create", FilePath(name)))
}

func SubvolumeDelete(name string) error {
	log.Printf("btrfs.SubvolumeDelete(%s)", name)
	return shell.RunStderr(exec.Command("btrfs", "subvolume", "delete", FilePath(name)))
}

func SubvolumeDeleteAll(name string) error {
	log.Printf("btrfs.SubvolumeDeleteAll(%s)", name)
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
	log.Printf("btrfs.Snapshot(%s, %s, %t)", volume, dest, readonly)
	if readonly {
		return shell.RunStderr(exec.Command("btrfs", "subvolume", "snapshot", "-r",
			FilePath(volume), FilePath(dest)))
	} else {
		return shell.RunStderr(exec.Command("btrfs", "subvolume", "snapshot",
			FilePath(volume), FilePath(dest)))
	}
}

func SetReadOnly(volume string) error {
	log.Printf("btrfs.SetReadOnly(%s)", volume)
	return shell.RunStderr(exec.Command("btrfs", "property", "set", FilePath(volume), "ro", "true"))
}

func UnsetReadOnly(volume string) error {
	log.Printf("btrfs.UnsetReadOnly(%s)", volume)
	return shell.RunStderr(exec.Command("btrfs", "property", "set", FilePath(volume), "ro", "false"))
}

func IsReadOnly(volume string) (bool, error) {
	log.Printf("btrfs.IsReadOnly(%s)", volume)
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

func SendBase(to string, cont func(io.Reader) error) error {
	log.Printf("btrfs.SendBase(%s, <function>)", to)
	c := exec.Command("btrfs", "send", FilePath(to))
	return shell.CallCont(c, cont)
}

func Send(from, to string, cont func(io.Reader) error) error {
	log.Printf("btrfs.Send(%s, %s, <function>)", from, to)
	c := exec.Command("btrfs", "send", "-p", FilePath(from), FilePath(to))
	return shell.CallCont(c, cont)
}

func Send2(repo, commit string, cont func(io.Reader) error) error {
	log.Printf("btrfs.Send(%s, %s, <function>)", repo, commit)
	parent := GetMeta(path.Join(repo, commit), "parent")
	if parent == "" {
		return shell.CallCont(exec.Command("btrfs", "send", FilePath(path.Join(repo, commit))), cont)
	} else {
		return shell.CallCont(exec.Command("btrfs", "send", "-p",
			FilePath(path.Join(repo, parent)), FilePath(path.Join(repo, commit))), cont)
	}
}

func Recv(volume string, data io.Reader) error {
	log.Printf("btrfs.Recv(%s, %#v)", volume, data)
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
	log.Printf("btrfs.Init(%s)", repo)
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
	log.Printf("btrfs.Ensure(%s)", repo)
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

func InitReplica(repo string) error {
	log.Printf("btrfs.InitReplica(%s)", repo)
	if err := SubvolumeCreate(repo); err != nil {
		return err
	}
	return nil
}

func EnsureReplica(repo string) error {
	log.Printf("btrfs.EnsureReplica(%s)", repo)
	exists, err := FileExists(repo)
	if err != nil {
		return err
	}
	if exists {
		return nil
	} else {
		return InitReplica(repo)
	}
}

// Commit creates a new commit for a branch.
func Commit(repo, commit, branch string) error {
	log.Printf("btrfs.Commit(%s, %s, %s)", repo, commit, branch)
	// check to make sure that the branch actually exists
	exists, err := FileExists(path.Join(repo, branch))
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("Branch %s not found.", branch)
	}
	// Record which branch this commit comes from
	if err := SetMeta(path.Join(repo, branch), "branch", branch); err != nil {
		return err
	}
	// make branch readonly
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

	// Record the new commit as the parent of this branch
	if err := SetMeta(path.Join(repo, branch), "parent", commit); err != nil {
		return err
	}

	return nil
}

// Hold creates a temporary snapshot of a commit that no one else knows about.
// It's your responsibility to release the snapshot with Release
func Hold(repo, commit string) (string, error) {
	log.Printf("btrfs.Hold(%s, %s)", repo, commit)
	MkdirAll("tmp")
	name := path.Join("tmp", uuid.New())
	if err := Snapshot(path.Join(repo, commit), name, false); err != nil {
		return "", err
	}
	return name, nil
}

// Release releases commit snapshots held by Hold.
func Release(name string) {
	log.Printf("btrfs.Release(%s)", name)
	SubvolumeDelete(name)
}

func Branch(repo, commit, branch string) error {
	log.Printf("btrfs.Branch(%s, %s, %s)", repo, commit, branch)
	isReadOnly, err := IsReadOnly(path.Join(repo, commit))
	if err != nil {
		return err
	}
	if !isReadOnly {
		return fmt.Errorf("Illegal branch from branch: \"%s\", can only branch from commits.", commit)
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
	if err := Snapshot(path.Join(repo, commit), path.Join(repo, branch), false); err != nil {
		return err
	}

	// Record commit as the parent of this branch
	if err := SetMeta(path.Join(repo, branch), "parent", commit); err != nil {
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
	log.Printf("btrfs.Log(%s, %s, <function>)", repo, from)
	if from == "" {
		c := exec.Command("btrfs", "subvolume", "list", "-o", "-c", "-u", "-q", "--sort", sort, FilePath(path.Join(repo)))
		return shell.CallCont(c, cont)
	} else {
		t, err := transid(repo, from)
		if err != nil {
			return err
		}
		c := exec.Command("btrfs", "subvolume", "list", "-o", "-c", "-u", "-q", "-G", "+"+t, "--sort", sort, FilePath(path.Join(repo)))
		return shell.CallCont(c, cont)
	}
}

type CommitInfo struct {
	gen, id, parent, Path string
}

// Commits is a wrapper around `Log` which parses the output in to a convenient
// struct
func Commits(repo, from string, order int, cont func(CommitInfo) error) error {
	log.Printf("btrfs.Commits(%s, %s, <function>)", repo, from)
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

type diff struct {
	parent, child *CommitInfo
}

// Pull gets all the commits in `repo` that happened after `from`.
// and passes them to cb
// Pull returns the next value that should passed as `from` to pick up where
// this function left off.
func Pull(repo, from string, cb CommitBrancher) error {
	log.Printf("btrfs.Pull(%s, %s, %#v)", repo, from, cb)
	// First check that `from` is actually a valid commit
	if from != "" {
		exists, err := FileExists(path.Join(repo, from))
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("`from` commit %s does not exists", from)
		}
	}
	// commits indexed by their parents
	parentMap := make(map[string][]*CommitInfo)
	var diffs []diff
	pullerHasCommit := false // true once we've hit the `from` commit

	// the body below gets called once per commit
	// Notice that we're passing `""` instead of `from` that's not a mistake.
	// This is because we need to see commits that came before `from` so we can
	// use them as parents in constructing the pull.
	err := Commits(repo, "", Desc, func(c CommitInfo) error {
		log.Printf("Commit: %#v", c)
		if c.Path == from {
			// We've hit the commit the puller specified as `from`, that means
			// this commit and every commit after it are commits the puller
			// already has.
			pullerHasCommit = true
		}
		// We don't do this if the parent is null (represented by "-")
		if c.parent != "-" && !pullerHasCommit {
			// this commit is part of the pull so we put it in the parentMap
			parentMap[c.parent] = append(parentMap[c.parent], &c)
		}
		// Now we pop all of the commits for which this was the parent out of
		// the map
		for _, child := range parentMap[c.id] {
			log.Printf("Append from id: %s\n %#v\n", c.id, child)
			diffs = append(diffs, diff{&c, child})
		}
		delete(parentMap, c.id)

		if len(parentMap) == 0 {
			if c.parent == "-" && !pullerHasCommit {
				diffs = append(diffs, diff{nil, &c})
			}
			return fmt.Errorf("COMPLETE")
		} else {
			return nil
		}
	})
	if err != nil && err.Error() != "COMPLETE" {
		return err
	}

	for i := 0; i < len(diffs); i++ {
		// The diffs are in reverse chronological order and we want to traverse
		// them in chronological order, so we need to traverse in reverse
		diff := diffs[len(diffs)-(i+1)]
		log.Printf("Sending: %#v", diff.child)

		// Check to make sure that what we have is a commit and not a branch
		isCommit, err := IsReadOnly(path.Join(repo, diff.child.Path))
		if err != nil {
			return err
		}
		if isCommit {
			if diff.parent == nil {
				// No Parent, use SendBase
				if err := Send2(repo, diff.child.Path, cb.Commit); err != nil {
					return err
				}
			} else {
				// We have a parent, use normal Send
				if err := Send2(repo, diff.child.Path, cb.Commit); err != nil {
					return err
				}
			}
		} else {
			if diff.parent == nil {
				return fmt.Errorf("It shouldn't be possible to have a branch with a nil parent.")
			}
			if err := cb.Branch(diff.parent.Path, diff.child.Path); err != nil {
				return err
			}
		}
	}

	return nil
}

func Pull2(repo, from string, cb CommitBrancher) error {
	log.Printf("btrfs.Pull(%s, %s, %#v)", repo, from, cb)
	// First check that `from` is actually a valid commit
	if from != "" {
		exists, err := FileExists(path.Join(repo, from))
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("`from` commit %s does not exists", from)
		}
	}

	err := Commits(repo, from, Asc, func(c CommitInfo) error {
		// In some case we need to send the parent of a commit even though that
		// parent came before the `from` commit in btrfs' ordering.
		// This is because when we sent everything up to the `from` commit the
		// parent commit might have been a branch and thus wouldn't have been sent.
		parent := GetMeta(path.Join(repo, c.Path), "parent")
		if parent != "" {
			less, err := Less(repo, parent, from)
			if err != nil {
				return err
			}
			if less {
				// The parent came before `from` that means we're not going to see
				// it else where in this pull so we need to send it
				err := Send2(repo, parent, cb.Commit)
				if err != nil && err != CommitExists {
					return err
				}
			}
		}

		// Send this commit
		isCommit, err := IsReadOnly(path.Join(repo, c.Path))
		if err != nil {
			return err
		}
		if isCommit {
			err := Send2(repo, c.Path, cb.Commit)
			if err != nil && err != CommitExists {
				return err
			}
		} else {
			if err := cb.Branch(parent, c.Path); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil && err.Error() != "COMPLETE" {
		return err
	}
	return nil
}

// Transid returns transid of a path in a repo. This value is useful for
// passing to FindNew.
func transid(repo, commit string) (string, error) {
	log.Printf("btrfs.transid(%s, %s)", repo, commit)
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

func Less(repo, commit1, commit2 string) (bool, error) {
	_c1, err := transid(repo, commit1)
	if err != nil {
		return false, err
	}
	c1, err := strconv.Atoi(_c1)
	if err != nil {
		return false, err
	}

	_c2, err := transid(repo, commit2)
	if err != nil {
		return false, err
	}
	c2, err := strconv.Atoi(_c2)
	if err != nil {
		return false, err
	}

	return c1 < c2, nil
}

// FindNew returns an array of filenames that were created between `from` and `to`
func FindNew(repo, from, to string) ([]string, error) {
	log.Printf("btrfs.FindNew(%s, %s, %s)", repo, from, to)
	var files []string
	t, err := transid(repo, from)
	if err != nil {
		return files, err
	}
	c := exec.Command("btrfs", "subvolume", "find-new", FilePath(path.Join(repo, to)), t)
	err = shell.CallCont(c, func(r io.Reader) error {
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
		return scanner.Err()
	})
	return files, err
}
