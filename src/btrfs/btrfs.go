package btrfs

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-fsnotify/fsnotify"
	"github.com/pachyderm/pachyderm/src/log"
	"github.com/pachyderm/pachyderm/src/pkg/executil"
)

// Constants used for passing to log
const (
	Desc = iota
	Asc

	majorVersion = 3
	minorVersion = 14
)

var (
	ErrComplete  = errors.New("pfs: complete")
	ErrCancelled = errors.New("pfs: cancelled")

	checkVersionOnce sync.Once
	checkVersionErr  error
)

// Pusher is an interface that wraps the Push method.
type Pusher interface {
	// From returns the last commit pushed.
	// This value should be passed to Pull.
	From() (string, error)
	// Push applies diff to an underlying storage layer.
	Push(diff io.Reader) error
}

// Puller is an interface that wraps the Pull method.
type Puller interface {
	// Pull produces binary diffs and passes them to p's Push method.
	Pull(from string, p Pusher) error
}

// Replica is the interface that groups the Puller and Pusher methods.
type Replica interface {
	Pusher
	Puller
}

func CheckVersion() error {
	checkVersionOnce.Do(checkVersion)
	return checkVersionErr
}

func NewLocalReplica(repo string) Replica {
	return newLocalReplica(repo)
}

func NewS3Replica(uri string) Replica {
	return newS3Replica(uri)
}

// FilePath returns an absolute path for a file in the btrfs volume *inside*
// the container.
func FilePath(name string) string {
	return path.Join(localVolume(), name)
}

// HostPath returns an absolute for a file *outside* the container
func HostPath(name string) string {
	return path.Join(hostVolume(), name)
}

// Create creates a new file in the btrfs volume
func Create(name string) (*os.File, error) {
	return os.Create(FilePath(name))
}

// CreateAll is like create but it will create the directory for the file if it
// doesn't already exist.
func CreateAll(name string) (*os.File, error) {
	err := MkdirAll(path.Dir(name))
	if err != nil {
		return nil, err
	}
	return os.Create(FilePath(name))
}

// CreateFromReader is like Create but automatically sets the content of the
// file to the data found in `r`
func CreateFromReader(name string, r io.Reader) (int64, error) {
	f, err := Create(name)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return io.Copy(f, r)
}

// Open opens a file for reading.
func Open(name string) (*os.File, error) {
	return os.Open(FilePath(name))
}

// OpenFile is a generalize form of Open
func OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(FilePath(name), flag, perm)
}

// OpenFd opens a file and gives you the file descriptor.
func OpenFd(name string, mode int, perm uint32) (int, error) {
	return syscall.Open(FilePath(name), mode, perm)
}

// ReadFile returns the contents of a file.
func ReadFile(name string) ([]byte, error) {
	return ioutil.ReadFile(FilePath(name))
}

// Writefile sets the contents of a file to `data`
func WriteFile(name string, data []byte) error {
	err := MkdirAll(path.Dir(name))
	if err != nil {
		return err
	}
	return ioutil.WriteFile(FilePath(name), data, 0666)
}

// CopyFile copies the contents of `r` in the a file
func CopyFile(name string, r io.Reader) (int64, error) {
	f, err := Open(name)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return io.Copy(f, r)
}

// Append reads data out of reader and appends it to the file.
// If the file doesn't exist it creates it.
func Append(name string, r io.Reader) (int64, error) {
	exists, err := FileExists(name)
	if err != nil {
		return 0, err
	}
	var f io.WriteCloser
	if !exists {
		f, err = Create(name)
	} else {
		f, err = OpenFile(name, os.O_APPEND|os.O_WRONLY, 0600)
	}
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return io.Copy(f, r)
}

// Remove removes a file or directory
func Remove(name string) error {
	return os.Remove(FilePath(name))
}

// RemoveAll removes a path and any children it contains
func RemoveAll(name string) error {
	return os.RemoveAll(FilePath(name))
}

// Rename renames a file
func Rename(oldname, newname string) error {
	return os.Rename(FilePath(oldname), FilePath(newname))
}

// Stat returns a FileInfo describing a file.
func Stat(name string) (os.FileInfo, error) {
	return os.Stat(FilePath(name))
}

// Lstat returns FileInfo describing a file, if the file is a symbolic link it
// still works.
func Lstat(name string) (os.FileInfo, error) {
	return os.Lstat(FilePath(name))
}

// Chtimes changes the access and modification times of a file
func Chtimes(name string, atime, mtime time.Time) error {
	return os.Chtimes(FilePath(name), atime, mtime)
}

// Changed returns true if `mtime` is after the filesystem time for name
func Changed(name string, mtime time.Time) (bool, error) {
	info, err := Stat(name)
	if err != nil && os.IsNotExist(err) {
		return true, nil
	} else if err != nil {
		return false, err
	}
	if mtime.After(info.ModTime()) {
		return true, nil
	}
	return false, nil
}

// Before eturn true if name1 was last modified before name2
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

// FileExists returns true if a file exists in the filesystem
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

// Mkdir creates a directory
func Mkdir(name string) error {
	return os.Mkdir(FilePath(name), 0777)
}

// MkdirAll creates a directory and all parent directories
func MkdirAll(name string) error {
	return os.MkdirAll(FilePath(name), 0777)
}

// ReadDir returns a list of files found in the name directory
func ReadDir(name string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(FilePath(name))
}

// Glob returns the names of all files matching pattern or nil if there's no match.
// Glob uses syntax that should be familiar from shell like /foo/bar/*
func Glob(pattern string) ([]string, error) {
	paths, err := filepath.Glob(FilePath(pattern))
	if err != nil {
		return nil, err
	}
	for i, p := range paths {
		paths[i] = trimFilePath(p)
	}
	return paths, nil
}

// WaitFile waits for a file to exist in the filesystem
// NOTE: You NEVER want to pass an unbuffered channel as cancel because
// WaitFile provides no guarantee that it will ever read from cancel.  Thus if
// you passed an unbuffered channel as cancel sending to the channel may block
// forever.
func WaitFile(name string, cancel chan struct{}) error {
	log.Print("WaitFile(", name, ")")
	dir, err := largestExistingPath(name)
	if err != nil {
		return err
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	if err := watcher.Add(FilePath(dir)); err != nil {
		return err
	}

	// Notice that we check to see if the file exists AFTER we create the watcher.
	// That means if we don't see the file with this check we're guaranteed to
	// get a notification for it.
	exists, err := FileExists(name)
	if err != nil {
		return err
	}
	if exists {
		log.Print("Found: ", name)
		return nil
	}

	for {
		select {
		case event := <-watcher.Events:
			if event.Op == fsnotify.Create && event.Name == FilePath(name) {
				return nil
			} else if event.Op == fsnotify.Create && strings.HasPrefix(FilePath(name), event.Name) {
				//fsnotify isn't recursive so we need to recurse for it.
				return WaitFile(name, cancel)
			}
		case err := <-watcher.Errors:
			return err
		case <-cancel:
			return ErrCancelled
		}
	}
}

// WaitAnyFile returns as soon as ANY of the files exists.
// It returns an error if waiting for ANY of the files errors.
func WaitAnyFile(files ...string) (string, error) {
	// Channel for files that appear
	done := make(chan string, len(files))
	// Channel for errors that occur
	errors := make(chan error, len(files))

	cancellers := make([]chan struct{}, len(files))
	// Make sure that everyone gets cancelled after this function exits.
	defer func() {
		for _, c := range cancellers {
			c <- struct{}{}
		}
	}()
	for i, _ := range files {
		file := files[i]
		cancellers[i] = make(chan struct{}, 1)
		go func(i int) {
			err := WaitFile(file, cancellers[i])
			if err != nil {
				// Never blocks due to size of channel's buffer.
				errors <- err
			}
			// Never blocks due to size of channel's buffer.
			done <- file
		}(i)
	}

	select {
	case file := <-done:
		log.Print("Done: ", file)
		return file, nil
	case err := <-errors:
		return "", err
	}
}

// IsCommit returns true if the volume is a commit and false if it's a branch.
func IsCommit(name string) (bool, error) {
	// "-t s" indicates to btrfs that this is a subvolume without the "-t s"
	// btrfs will still output what we want, but it will have a nonzero return
	// code
	reader, err := executil.RunStdout("btrfs", "property", "get", "-t", "s", FilePath(name))
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

// SetMeta sets metadata for a branch
func SetMeta(branch, key, value string) error {
	return WriteFile(path.Join(branch, ".meta", key), []byte(value))
}

// GetMeta gets metadata from a commit or branch
func GetMeta(name, key string) string {
	value, err := ReadFile(path.Join(name, ".meta", key))
	if err != nil {
		return ""
	}
	return string(value)
}

// Init initializes an empty repo.
func Init(repo string) error {
	if err := subvolumeCreate(repo); err != nil {
		return err
	}
	if err := subvolumeCreate(path.Join(repo, "master")); err != nil {
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
	err := MkdirAll(path.Dir(repo))
	if err != nil {
		return err
	}
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

// Commit creates a new commit on a branch.
// The contents of the commit will be the same as the current contents of the branch.
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
	if err := snapshot(path.Join(repo, branch), path.Join(repo, commit), true); err != nil {
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
	err = subvolumeDelete(path.Join(repo, branch))
	if err != nil {
		return err
	}
	err = Branch(repo, parent, branch)
	if err != nil {
		return err
	}
	return nil
}

// Branch creates a new writeable branch from commit.
func Branch(repo, commit, branch string) error {
	// Check that the commit is read only
	if commit != "" {
		isReadOnly, err := IsCommit(path.Join(repo, commit))
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
		if err := subvolumeCreate(path.Join(repo, branch)); err != nil {
			return err
		}
	} else {
		if err := snapshot(path.Join(repo, commit), path.Join(repo, branch), false); err != nil {
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

// Commits is a wrapper around `Log` which parses the output in to a convenient
// struct
func Commits(repo, from string, order int, cont func(string) error) error {
	return _log(repo, from, order, func(r io.Reader) error {
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
			if err := cont(p); err != nil {
				return err
			}
		}
		if scanner.Err() != nil && scanner.Err() != io.EOF {
			return scanner.Err()
		}
		return nil
	})
}

// Pull produces a binary diff stream from repo and passes it to cb
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
		isCommit, err := IsCommit(path.Join(repo, from))
		if err != nil {
			return err
		}
		if !isCommit {
			return fmt.Errorf("`from` %s cannot be a branch", from)
		}
	}

	err := Commits(repo, from, Asc, func(name string) error {
		if name == from {
			// Commits gives us things >= `from` so we explicitly skip `from`
			return nil
		}
		// Send this commit
		isCommit, err := IsCommit(path.Join(repo, name))
		if err != nil {
			return err
		}
		if isCommit {
			err := send(repo, name, cb.Push)
			if err != nil {
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

// FindNew returns an array of filenames that were created or modified between `from` and `to`
func FindNew(repo, from, to string) ([]string, error) {
	var files []string
	t, err := transid(repo, from)
	if err != nil {
		return files, err
	}
	reader, err := executil.RunStdout("btrfs", "subvolume", "find-new", FilePath(path.Join(repo, to)), t)
	if err != nil {
		return files, err
	}
	scanner := bufio.NewScanner(reader)
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
			return files, errors.New("Failed to parse find-new output.")
		}
	}
	return files, scanner.Err()
}

// NewIn returns all of the files that changed in a commit
func NewIn(repo, commit string) ([]string, error) {
	parent := GetMeta(path.Join(repo, commit), "parent")
	return FindNew(repo, parent, commit)
}

func Show(repo string, commit string, out string) error {
	if err := subvolumeCreate(path.Join(repo, out)); err != nil {
		return err
	}
	files, err := NewIn(repo, commit)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	for _, file := range files {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()
			in, err := Open(path.Join(repo, commit, file))
			if err != nil {
				log.Print(err)
				return
			}
			if _, err := CreateFromReader(path.Join(repo, out, file), in); err != nil {
				log.Print(err)
				return
			}
		}(file)
	}
	wg.Wait()
	return nil
}

//_log returns all of the commits the repo which have generation >= from.
func _log(repo, from string, order int, cont func(io.Reader) error) error {
	var sort string
	if order == Desc {
		sort = "-ogen"
	} else {
		sort = "+ogen"
	}

	if from == "" {
		reader, err := executil.RunStdout("btrfs", "subvolume", "list", "-o", "-c", "-u", "-q", "--sort", sort, FilePath(path.Join(repo)))
		if err != nil {
			return err
		}
		return cont(reader)
	}
	t, err := transid(repo, from)
	if err != nil {
		return err
	}
	reader, err := executil.RunStdout("btrfs", "subvolume", "list", "-o", "-c", "-u", "-q", "-C", "+"+t, "--sort", sort, FilePath(path.Join(repo)))
	if err != nil {
		return err
	}
	return cont(reader)
}

func checkVersion() {
	reader, err := executil.RunStdout("btrfs", "--version")
	if err != nil {
		checkVersionErr = err
		return
	}
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		checkVersionErr = err
		return
	}
	versionString := strings.TrimSpace(string(data))
	splittedVersion := strings.Split(versionString, " v")
	if len(splittedVersion) != 2 {
		checkVersionErr = fmt.Errorf("unknown version string: %s", versionString)
		return
	}
	split := strings.Split(splittedVersion[1], ".")
	versionNumbers := len(split)
	if versionNumbers != 2 && versionNumbers != 3 {
		checkVersionErr = fmt.Errorf("unknown version string: %s", versionString)
		return
	}
	major, err := strconv.ParseInt(split[0], 10, 64)
	if err != nil {
		checkVersionErr = err
		return
	}
	if major < majorVersion {
		checkVersionErr = fmt.Errorf("need at least btrfs version %d.%d, got %s", majorVersion, minorVersion, splittedVersion[1])
		return
	} else if major == majorVersion {
		minor, err := strconv.ParseInt(split[1], 10, 64)
		if err != nil {
			checkVersionErr = err
			return
		}
		if minor < minorVersion {
			checkVersionErr = fmt.Errorf("need at least btrfs version %d.%d, got %s", majorVersion, minorVersion, splittedVersion[1])
			return
		}
	}
}

func trimFilePath(name string) string {
	return strings.TrimPrefix(name, localVolume())
}

// localVolume returns the path *inside* the container that we look for the
// btrfs volume at
func localVolume() string {
	if val := os.Getenv("PFS_LOCAL_VOLUME"); val != "" {
		return val
	}
	return "/var/lib/pfs/vol"
}

// hostVolume returns the path *outside* container where you can find the btrfs
// volume
func hostVolume() string {
	if val := os.Getenv("PFS_HOST_VOLUME"); val != "" {
		return val
	}
	return "/var/lib/pfs/vol"
}

// largestExistingPath takes a path and trims it until it gets something that
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

// transid returns transid of a path in a repo. This function is used in
// several other internal functions.
func transid(repo, commit string) (string, error) {
	//  "9223372036854775810" == 2 ** 63 we use a very big number there so that
	//  we get the transid of the from path. According to the internet this is
	//  the nicest way to get it from btrfs.
	var transid string
	reader, err := executil.RunStdout("btrfs", "subvolume", "find-new", FilePath(path.Join(repo, commit)), "9223372036854775808")
	if err != nil {
		return "", err
	}
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		// scanner.Text() looks like this:
		// transid marker was 907
		// 0       1      2   3
		tokens := strings.Split(scanner.Text(), " ")
		if len(tokens) != 4 {
			return "", errors.New("Failed to parse find-new output.")
		}
		// We want to increment the transid because it's inclusive, if we
		// don't increment we'll get things from the previous commit as
		// well.
		transid = tokens[3]
	}
	return transid, scanner.Err()
}

func subvolumeCreate(name string) error {
	return executil.Run("btrfs", "subvolume", "create", FilePath(name))
}

func subvolumeDelete(name string) error {
	return executil.Run("btrfs", "subvolume", "delete", FilePath(name))
}

func subvolumeDeleteAll(name string) error {
	subvolumeExists, err := FileExists(name)
	if err != nil {
		return err
	}
	if subvolumeExists {
		return subvolumeDelete(name)
	} else {
		return nil
	}
}

func snapshot(volume string, dest string, readonly bool) error {
	if readonly {
		return executil.Run("btrfs", "subvolume", "snapshot", "-r", FilePath(volume), FilePath(dest))
	}
	return executil.Run("btrfs", "subvolume", "snapshot", FilePath(volume), FilePath(dest))
}
