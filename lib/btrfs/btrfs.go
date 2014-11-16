package btrfs

import (
	"bytes"
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

func RandSeq(n int) string {
	once.Do(func() { rand.Seed(time.Now().UTC().UnixNano()) })
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

type FS struct {
	namespace string
}

func NewFS(namespace string) *FS {
	return &FS{namespace}
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

func (fs *FS) CreateFile(name string, r io.Reader) (int64, error) {
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

	if err := fs.SubvolumeCreate(path.Join(repo, "branches")); err != nil {
		return err
	}

	if err := fs.SubvolumeCreate(path.Join(repo, "commits")); err != nil {
		return err
	}

	return nil
}

func (fs *FS) Commit(repo, branch string) error {
	tstamp := time.Now().Format("2006-01-02T15:04:05.999999-07:00")

	if err := fs.Snapshot(path.Join(repo, branch), path.Join(repo, "commits", branch+"-"+tstamp), true); err != nil {
		return err
	}

	return nil
}

//Log returns all of the commits the repo which are >= a given generation.
func (fs *FS) Log(repo, from string, cont func(io.ReadCloser) error) error {
	cmd := exec.Command("btrfs", "s,ubvolume", "list", "-c", "-g", "-C", "+"+from, fs.FilePath(path.Join(repo, "commits")))
	return fs.CallCont(cmd, cont)
}
