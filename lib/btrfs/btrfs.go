package btrfs

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
    "strings"
    "errors"
)

type FS struct {
	file, dev, mnt string
	size           int     // Size in GB
    formattable    bool    // Whether the filesystem is something we can formattable
}

func NewFS(file, dev, mnt string, size int) *FS {
    return &FS{file, dev, mnt, size, true}
}

func ExistingFS(mnt string) *FS {
    return &FS{"", "", mnt, 0, false}
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

func (fs *FS) DevPath() string {
	return path.Join("/dev", fs.dev)
}

func (fs *FS) MntPath() string {
	return path.Join("/mnt", fs.mnt)
}

func (fs *FS) Filepath(name string) string {
	return path.Join("/mnt", fs.mnt, name)
}

func (fs *FS) StripFilepath(name string) (string, error) {
	log.Print(name)
    stripped_name := strings.TrimPrefix(name, "/mnt/")
    log.Print(stripped_name)
    stripped_name = strings.TrimPrefix(stripped_name,  fs.mnt)
    log.Print(stripped_name)
    return stripped_name, nil
}

func (fs *FS) Exists() bool {
    if _, err := os.Stat(fs.file); err == nil {
        return true
	} else if !fs.formattable {
		// If we can't format the disk then we're not responsible for making
		// sure the filesystem exists so we assume it already does.
		return true
    } else {
        return false
    }
}

func (fs *FS) Format() error {
	if !fs.formattable {
		return errors.New(`This FS was created pointing to an already formatted"
				and mounted filesystem. Thus it cannot perform more
				formatting.`)
	}
	size_flag := fmt.Sprintf("-s %d", 1024*1024*1000*fs.size)
	err := RunStderr(exec.Command("truncate", size_flag, fs.file))
	if err != nil {
		return err
	}
	err = RunStderr(exec.Command("losetup", fs.DevPath(), fs.file))
	if err != nil {
		return err
	}
	err = RunStderr(exec.Command("mkfs.btrfs", fs.DevPath()))
	if err != nil {
		return err
	}
	err = RunStderr(exec.Command("mkdir", "-p", fs.MntPath()))
	if err != nil {
		return err
	}
	return RunStderr(exec.Command("mount", fs.DevPath(), fs.MntPath()))
}

func (fs *FS) Destroy() error {
	if !fs.formattable {
		return errors.New(`This FS was created pointing to an already formatted
				and mounted filesystem. Thus it cannot destroy the filesystem.`)
	}
	err := fs.Unmount()
	if err != nil {
		return err
	}
	return RunStderr(exec.Command("rm", fs.file))
}

func (fs *FS) Mount() error {
	if !fs.formattable {
		return errors.New(`This FS was created pointing to an already formatted
				and mounted filesystem. Thus it cannot be mounted (it's already
					mounted).`)
	}
	err := RunStderr(exec.Command("losetup", fs.DevPath(), fs.file))
	if err != nil {
		return err
	}
	err = RunStderr(exec.Command("mkdir", "-p", fs.MntPath()))
	if err != nil {
		return err
	}
	return RunStderr(exec.Command("mount", fs.DevPath(), fs.MntPath()))
}

func (fs *FS) Unmount() error {
	if !fs.formattable {
		return errors.New(`This FS was created pointing to an already formatted
				and mounted filesystem. Thus it cannot be unmounted.`)
	}
	err := RunStderr(exec.Command("umount", "-l", fs.MntPath()))
	if err != nil {
		return err
	}
	return RunStderr(exec.Command("losetup", "-d", fs.DevPath()))
}

func (fs *FS) EnsureExists() error {
    if fs.Exists() {
        return fs.Mount()
    } else {
		return fs.Format()
    }
}

func (fs *FS) Create(name string) (*os.File, error) {
	return os.Create(fs.Filepath(name))
}

func (fs *FS) Open(name string) (*os.File, error) {
	return os.Open(fs.Filepath(name))
}

func (fs *FS) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(fs.Filepath(name), flag, perm)
}

func (fs *FS) Remove(name string) error {
	return os.Remove(fs.Filepath(name))
}

func (fs *FS) FileExists(name string) (bool, error) {
    _, err := os.Stat(fs.Filepath(name))
    if err == nil { return true, nil }
    if os.IsNotExist(err) { return false, nil }
    return false, err
}

func (fs *FS) Mkdir(name string, prem os.FileMode) error {
    return os.Mkdir(fs.Filepath(name), prem)
}

func (fs *FS) MkdirAll(name string, prem os.FileMode) error {
    return os.MkdirAll(fs.Filepath(name), prem)
}

func (fs *FS) Link(oldname, newname string) error {
    return os.Link(fs.Filepath(oldname), fs.Filepath(newname))
}

func (fs *FS) Readlink(name string) (string, error) {
    p, err := os.Readlink(fs.Filepath(name))
    if err != nil { return "", err }
    return fs.StripFilepath(p)
}

func (fs *FS) Symlink(oldname, newname string) error {
    log.Printf("%s -> %s\n", fs.Filepath(oldname), fs.Filepath(newname))
    return os.Symlink(fs.Filepath(oldname), fs.Filepath(newname))
}

func (fs *FS) ReadDir(name string) ([]os.FileInfo, error) {
    return ioutil.ReadDir(fs.Filepath(name))
}

func (fs *FS) SubvolumeCreate(name string) error {
	return RunStderr(exec.Command("btrfs", "subvolume", "create", fs.Filepath(name)))
}

func (fs *FS) SubvolumeDelete(name string) error {
    return RunStderr(exec.Command("btrfs", "subvolume", "delete", fs.Filepath(name)))
}

func (fs *FS) Snapshot(volume string, dest string, readonly bool) error {
	if readonly {
		return RunStderr(exec.Command("btrfs", "subvolume", "snapshot", "-r",
			fs.Filepath(volume), fs.Filepath(dest)))
	} else {
		return RunStderr(exec.Command("btrfs", "subvolume", "snapshot",
			fs.Filepath(volume), fs.Filepath(dest)))
	}
}

func (fs *FS) SendBase(to string, cont func(io.ReadCloser) error) error {
	cmd := exec.Command("btrfs", "send", fs.Filepath(to))
	log.Println(cmd)
	reader, err := cmd.StdoutPipe()
	if err != nil { return err }
	stderr, err := cmd.StderrPipe()
	if err != nil { return err }
	err = cmd.Start()
	if err != nil { return err }
	err = cont(reader)
	if err != nil { return err }

	buf := new(bytes.Buffer)
	buf.ReadFrom(stderr)
	log.Print("Stderr:", buf)

	return cmd.Wait()
}

func (fs *FS) Send(from string, to string, cont func(io.ReadCloser) error) error {
	cmd := exec.Command("btrfs", "send", "-p", fs.Filepath(from), fs.Filepath(to))
	log.Println(cmd)
	reader, err := cmd.StdoutPipe()
	if err != nil { return err }
	stderr, err := cmd.StderrPipe()
	if err != nil { return err }
	err = cmd.Start()
	if err != nil { return err }
	err = cont(reader)
	if err != nil { return err }

	buf := new(bytes.Buffer)
	buf.ReadFrom(stderr)
	log.Print("Stderr:", buf)

	return cmd.Wait()
}

func (fs *FS) Recv(volume string, data io.ReadCloser) error {
	cmd := exec.Command("btrfs", "receive", fs.Filepath(volume))
	log.Println(cmd)
	stdin, err := cmd.StdinPipe()
	if err != nil { return err }
	stderr, err := cmd.StderrPipe()
	if err != nil { return err }
	err = cmd.Start()
	if err != nil { return err }
	n, err := io.Copy(stdin, data)
	if err != nil { return err }
	log.Println("Copied bytes:", n)
	err = stdin.Close()
	if err != nil { return err }

	buf := new(bytes.Buffer)
	buf.ReadFrom(stderr)
	log.Print("Stderr:", buf)

	return cmd.Wait()
}
