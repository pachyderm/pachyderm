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
	log.Printf("%s -> %s\n", FilePath(oldname), FilePath(newname))
	return os.Symlink(FilePath(oldname), FilePath(newname))
}

func ReadDir(name string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(FilePath(name))
}

func SubvolumeCreate(name string) error {
	return RunStderr(exec.Command("btrfs", "subvolume", "create", FilePath(name)))
}

func SubvolumeDelete(name string) error {
	return RunStderr(exec.Command("btrfs", "subvolume", "delete", FilePath(name)))
}

func Snapshot(volume string, dest string, readonly bool) error {
	if readonly {
		return RunStderr(exec.Command("btrfs", "subvolume", "snapshot", "-r",
			FilePath(volume), FilePath(dest)))
	} else {
		return RunStderr(exec.Command("btrfs", "subvolume", "snapshot",
			FilePath(volume), FilePath(dest)))
	}
}

func SetReadOnly(volume string) error {
	return RunStderr(exec.Command("btrfs", "property", "set", FilePath(volume), "ro", "true"))
}

func UnsetReadOnly(volume string) error {
	return RunStderr(exec.Command("btrfs", "property", "set", FilePath(volume), "ro", "false"))
}

func CallCont(cmd *exec.Cmd, cont func(io.ReadCloser) error) error {
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

func SendBase(to string, cont func(io.ReadCloser) error) error {
	cmd := exec.Command("btrfs", "send", FilePath(to))
	return CallCont(cmd, cont)
}

func Send(from string, to string, cont func(io.ReadCloser) error) error {
	cmd := exec.Command("btrfs", "send", "-p", FilePath(from), FilePath(to))
	return CallCont(cmd, cont)
}

func Recv(volume string, data io.ReadCloser) error {
	cmd := exec.Command("btrfs", "receive", FilePath(volume))
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

func Branch(repo, commit, branch string) error {
	if err := Snapshot(path.Join(repo, commit), path.Join(repo, branch), false); err != nil {
		return err
	}
	return nil
}

//Log returns all of the commits the repo which have generation >= from.
func Log(repo, from string, cont func(io.ReadCloser) error) error {
	cmd := exec.Command("btrfs", "subvolume", "list", "-o", "-c", "-C", "+"+from, "--sort", "-ogen", FilePath(path.Join(repo)))
	return CallCont(cmd, cont)
}

type CommitInfo struct {
	gen, id, parent, path string
}

func Commits(repo, from string, cont func(CommitInfo) error) error {
	return Log(repo, from, func(r io.ReadCloser) error {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			// scanner.Text() looks like:
			// ID 299 gen 67 cgen 66 top level 292 parent_uuid 7a4a824b-7b78-d144-a956-eb0229616d21 uuid c1cd770c-600b-a744-940c-835bf73b5fa9 path fs/repo/2014-11-20T07:25:01.853165+00:00
			// 0  1   2   3  4    5  6   7     8   9           10                                   11   12                                   13   14
			tokens := strings.Split(scanner.Text(), " ")
			if len(tokens) != 15 {
				return fmt.Errorf("Malformed commit line: %s.", scanner.Text())
			}
			if err := cont(CommitInfo{tokens[3], tokens[12], tokens[10], tokens[14]}); err != nil {
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
	var parentMap map[string][]CommitInfo
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
	if err == nil || err.Error() == "COMPLETE" {
		for _, diff := range diffs {
			if diff.parent == nil {
				if err := SendBase(diff.child.path, cont); err != nil {
					return err
				}
			}
			if err := Send(diff.parent.path, diff.child.path, cont); err != nil {
				return err
			}
		}
	} else {
		return err
	}
	return nil
}
