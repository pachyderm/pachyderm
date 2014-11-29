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
	"strings"
	"sync"
	"time"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var once sync.Once
var volume = "/var/lib/pfs/vol"

// Generates a random sequence of letters. Useful for making filesystems that won't interfere with each other.
func RandSeq(n int) string {
	once.Do(func() { rand.Seed(time.Now().UTC().UnixNano()) })
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// FS represents a btrfs filesystem. Underneath it's a subvolume of a larger filesystem.
type FS struct {
	namespace string
}

// NewFS creates a new filesystem.
func NewFS(namespace string) *FS {
	return &FS{namespace}
}

// NewFSWithRandSeq creates a new filesystem with a random sequence appended to the end.
func NewFSWithRandSeq(namespace string) *FS {
	return &FS{namespace + RandSeq(10)}
}

func RunStderr(c *exec.Cmd) error {
	log.Println(c)
	stderr, err := c.StderrPipe()
	if err != nil {
		return err
	}
	err = c.Start()
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(stderr)
	log.Println(buf)
	return c.Wait()
}

func LogErrors(c *exec.Cmd) {
	stderr, err := c.StderrPipe()
	if err != nil {
		log.Println(err)
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(stderr)
	log.Println(buf)
}

func Sync() error {
	return RunStderr(exec.Command("sync"))
}

func (fs *FS) BasePath(name string) string {
	return path.Join(volume, fs.namespace, name)
}

func (fs *FS) FilePath(name string) string {
	return path.Join(volume, fs.namespace, name)
}

func (fs *FS) TrimFilePath(name string) string {
	return strings.TrimPrefix(name, path.Join(volume, fs.namespace))
}

func (fs *FS) Create(name string) (*os.File, error) {
	return os.Create(fs.FilePath(name))
}

func (fs *FS) CreateFromReader(name string, r io.Reader) (int64, error) {
	f, err := fs.Create(name)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return io.Copy(f, r)
}

func (fs *FS) Open(name string) (*os.File, error) {
	return os.Open(fs.FilePath(name))
}

func (fs *FS) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(fs.FilePath(name), flag, perm)
}

func (fs *FS) WriteFile(name string, r io.Reader) (int64, error) {
	f, err := fs.Open(name)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return io.Copy(f, r)
}

func (fs *FS) Remove(name string) error {
	return os.Remove(fs.FilePath(name))
}

func (fs *FS) Rename(oldname, newname string) error {
	return os.Rename(fs.FilePath(oldname), fs.FilePath(newname))
}

func (fs *FS) FileExists(name string) (bool, error) {
	_, err := os.Stat(fs.FilePath(name))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (fs *FS) Mkdir(name string) error {
	return os.Mkdir(fs.FilePath(name), 0777)
}

func (fs *FS) MkdirAll(name string) error {
	return os.MkdirAll(fs.FilePath(name), 0777)
}

func (fs *FS) Link(oldname, newname string) error {
	return os.Link(fs.FilePath(oldname), fs.FilePath(newname))
}

func (fs *FS) Readlink(name string) (string, error) {
	p, err := os.Readlink(fs.FilePath(name))
	if err != nil {
		return "", err
	}
	return fs.TrimFilePath(p), nil
}

func (fs *FS) Symlink(oldname, newname string) error {
	log.Printf("%s -> %s\n", fs.FilePath(oldname), fs.FilePath(newname))
	return os.Symlink(fs.FilePath(oldname), fs.FilePath(newname))
}

func (fs *FS) ReadDir(name string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(fs.FilePath(name))
}

func (fs *FS) EnsureNamespace() error {
	exists, err := fs.FileExists("")
	if err != nil {
		return err
	}
	if !exists {
		return fs.SubvolumeCreate("")
	}
	return nil
}

func (fs *FS) SubvolumeCreate(name string) error {
	return RunStderr(exec.Command("btrfs", "subvolume", "create", fs.FilePath(name)))
}

func (fs *FS) SubvolumeDelete(name string) error {
	return RunStderr(exec.Command("btrfs", "subvolume", "delete", fs.FilePath(name)))
}

func (fs *FS) Snapshot(volume string, dest string, readonly bool) error {
	if readonly {
		return RunStderr(exec.Command("btrfs", "subvolume", "snapshot", "-r",
			fs.FilePath(volume), fs.FilePath(dest)))
	} else {
		return RunStderr(exec.Command("btrfs", "subvolume", "snapshot",
			fs.FilePath(volume), fs.FilePath(dest)))
	}
}

func (fs *FS) SetReadOnly(volume string) error {
	return RunStderr(exec.Command("btrfs", "property", "set", fs.FilePath(volume), "ro", "true"))
}

func (fs *FS) UnsetReadOnly(volume string) error {
	return RunStderr(exec.Command("btrfs", "property", "set", fs.FilePath(volume), "ro", "false"))
}

func (fs *FS) CallCont(cmd *exec.Cmd, cont func(io.ReadCloser) error) error {
	log.Println("CallCont: ", cmd)
	reader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}
	err = cont(reader)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(stderr)
	log.Print("Stderr:", buf)

	return cmd.Wait()
}

func (fs *FS) SendBase(to string, cont func(io.ReadCloser) error) error {
	cmd := exec.Command("btrfs", "send", fs.FilePath(to))
	return fs.CallCont(cmd, cont)
}

func (fs *FS) Send(from string, to string, cont func(io.ReadCloser) error) error {
	cmd := exec.Command("btrfs", "send", "-p", fs.FilePath(from), fs.FilePath(to))
	return fs.CallCont(cmd, cont)
}

func (fs *FS) Recv(volume string, data io.ReadCloser) error {
	cmd := exec.Command("btrfs", "receive", fs.FilePath(volume))
	log.Println(cmd)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}
	n, err := io.Copy(stdin, data)
	if err != nil {
		return err
	}
	log.Println("Copied bytes:", n)
	err = stdin.Close()
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(stderr)
	log.Print("Stderr:", buf)

	return cmd.Wait()
}

func (fs *FS) Init(repo string) error {
	if err := fs.SubvolumeCreate(repo); err != nil {
		return err
	}
	if err := fs.SubvolumeCreate(path.Join(repo, "master")); err != nil {
		return err
	}
	return nil
}

// Commit creates a new commit for a branch.
func (fs *FS) Commit(repo, commit, branch string) (string, error) {
	// First we check to make sure that the branch actually exists
	exists, err := fs.FileExists(path.Join(repo, branch))
	if err != nil {
		return "", err
	}
	if !exists {
		return "", fmt.Errorf("Branch %s not found.", branch)
	}
	// First we make HEAD readonly
	if err := fs.SetReadOnly(path.Join(repo, branch)); err != nil {
		return "", err
	}
	// Next move it to being a commit
	if err := fs.Rename(path.Join(repo, branch), path.Join(repo, commit)); err != nil {
		return "", err
	}
	// Recreate the branch subvolume with a writeable subvolume
	if err := fs.Snapshot(path.Join(repo, commit), path.Join(repo, branch), false); err != nil {
		return "", err
	}
	return commit, nil
}

func (fs *FS) Branch(repo, commit, branch string) error {
	if err := fs.Snapshot(path.Join(repo, commit), path.Join(repo, branch), false); err != nil {
		return err
	}
	return nil
}

//Log returns all of the commits the repo which have generation >= from.
func (fs *FS) Log(repo, from string, cont func(io.ReadCloser) error) error {
	cmd := exec.Command("btrfs", "subvolume", "list", "-o", "-c", "-C", "+"+from, "--sort", "-ogen", fs.FilePath(path.Join(repo)))
	return fs.CallCont(cmd, cont)
}

type Commit struct {
	gen, id, parent, path string
}

func (fs *FS) Commits(repo, from string, cont func(Commit) error) error {
	return fs.Log(repo, from, func(r io.ReadCloser) error {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			// scanner.Text() looks like:
			// ID 299 gen 67 cgen 66 top level 292 parent_uuid 7a4a824b-7b78-d144-a956-eb0229616d21 uuid c1cd770c-600b-a744-940c-835bf73b5fa9 path fs/repo/2014-11-20T07:25:01.853165+00:00
			// 0  1   2   3  4    5  6   7     8   9           10                                   11   12                                   13   14
			tokens := strings.Split(scanner.Text(), " ")
			if len(tokens) != 15 {
				return fmt.Errorf("Malformed commit line: %s.", scanner.Text())
			}
			if err := cont(Commit{tokens[3], tokens[12], tokens[10], tokens[14]}); err != nil {
				return err
			}
		}
		return nil
	})
}

type Diff struct {
	parent, child *Commit
}

func (fs *FS) Pull(repo, from string, cont func(io.ReadCloser) error) error {
	// commits indexed by their parents
	var parentMap map[string][]Commit
	var diffs []Diff
	// the body below gets called once per commit
	err := fs.Commits(repo, from, func(c Commit) error {
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
	if err == nil || err.Error() == "COMPLETE" {
		for _, diff := range diffs {
			if diff.parent == nil {
				if err := fs.SendBase(diff.child.path, cont); err != nil {
					return err
				}
			}
			if err := fs.Send(diff.parent.path, diff.child.path, cont); err != nil {
				return err
			}
		}
	} else {
		return err
	}
	return nil
}
