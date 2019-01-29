package fuse

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
)

const (
	modeFile = fuse.S_IFREG | 0666 // everyone can read and write
	modeDir  = fuse.S_IFDIR | 0777 // everyone can read and execute and write (execute permission is required to list a dir)
)

// Mount pfs to mountPoint, opts may be left nil.
func Mount(c *client.APIClient, mountPoint string, opts *Options) error {
	fs, err := newFileSystem(c, opts.getCommits())
	if err != nil {
		return err
	}
	nfs := pathfs.NewPathNodeFs(fs, nil)
	server, _, err := nodefs.MountRoot(mountPoint, nfs.Root(), opts.getFuse())
	if err != nil {
		return fmt.Errorf("nodefs.MountRoot: %v", err)
	}
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		select {
		case <-sigChan:
		case <-opts.getUnmount():
		}
		server.Unmount()
	}()
	server.Serve()
	return nil
}

type filesystem struct {
	pathfs.FileSystem
	c          *client.APIClient
	commits    map[string]string
	commitsMu  sync.RWMutex
	files      *sync.Map
	dir        string
	loopBackFS pathfs.FileSystem
}

func newFileSystem(c *client.APIClient, commits map[string]string) (pathfs.FileSystem, error) {
	if commits == nil {
		commits = make(map[string]string)
	}
	dir, err := ioutil.TempDir("", "pfs-fuse")
	if err != nil {
		return nil, err
	}
	return &filesystem{
		FileSystem: pathfs.NewDefaultFileSystem(),
		c:          c,
		commits:    commits,
		files:      &sync.Map{},
		dir:        dir,
		loopBackFS: pathfs.NewLoopbackFileSystem(dir),
	}, nil
}

func (fs *filesystem) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	return fs.getAttr(name, context)
}

func (fs *filesystem) listFile(repo, commit, path string, context *fuse.Context) ([]fuse.DirEntry, fuse.Status) {
	result, status := fs.loopBackFS.OpenDir(filepath.Join(repo, path), context)
	if status != fuse.OK {
		return nil, status
	}
	// Check which names we have from loopbackfs, those take priority since they may have been modified.
	names := make(map[string]bool)
	for _, dirent := range result {
		names[dirent.Name] = true
	}
	if commit != "" { // Don't check pfs if the branch has no head
		if err := fs.c.ListFileF(repo, commit, path, 0, func(fi *pfs.FileInfo) error {
			dirent := fileDirEntry(fi)
			if !names[dirent.Name] {
				result = append(result, fileDirEntry(fi))
			}
			return nil
		}); err != nil {
			return nil, toStatus(err)
		}
	}
	return result, fuse.OK
}

func (fs *filesystem) OpenDir(name string, context *fuse.Context) ([]fuse.DirEntry, fuse.Status) {
	var result []fuse.DirEntry
	r, f, err := fs.parsePath(name)
	if err != nil {
		return nil, toStatus(err)
	}
	switch {
	case r != nil:
		commit, err := fs.commit(r.Name)
		if err != nil {
			return nil, toStatus(err)
		}
		return fs.listFile(r.Name, commit, "", context)
	case f != nil:
		return fs.listFile(f.Commit.Repo.Name, f.Commit.ID, f.Path, context)
	default:
		ris, err := fs.c.ListRepo()
		if err != nil {
			return nil, toStatus(err)
		}
		for _, ri := range ris {
			result = append(result, repoDirEntry(ri))
		}
	}
	return result, fuse.OK
}

func (fs *filesystem) Open(name string, flags uint32, context *fuse.Context) (nodefs.File, fuse.Status) {
	return newFile(fs, name, int(flags))
}

func (fs *filesystem) Create(name string, flags uint32, mode uint32, context *fuse.Context) (file nodefs.File, code fuse.Status) {
	return newFile(fs, name, int(flags))
}

func (fs *filesystem) commit(repo string) (string, error) {
	commitOrBranch := func() string {
		fs.commitsMu.RLock()
		defer fs.commitsMu.RUnlock()
		return fs.commits[repo]
	}()
	if uuid.IsUUIDWithoutDashes(commitOrBranch) {
		// it's a commit, return it
		return commitOrBranch, nil
	}
	// it's a branch, resolve the head and return that
	branch := commitOrBranch
	if branch == "" {
		branch = "master"
	}
	bi, err := fs.c.InspectBranch(repo, branch)
	if err != nil {
		return "", err
	}
	fs.commitsMu.Lock()
	defer fs.commitsMu.Unlock()
	if bi.Head != nil {
		fs.commits[repo] = bi.Head.ID
	} else {
		fs.commits[repo] = ""
	}
	return fs.commits[repo], nil
}

func (fs *filesystem) parsePath(name string) (*pfs.Repo, *pfs.File, error) {
	components := strings.Split(name, "/")
	switch {
	case name == "":
		return nil, nil, nil
	case len(components) == 1:
		return client.NewRepo(components[0]), nil, nil
	default:
		commit, err := fs.commit(components[0])
		if err != nil {
			return nil, nil, err
		}
		return nil, client.NewFile(components[0], commit, path.Join(components[1:]...)), nil
	}
}

func (fs *filesystem) getAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	r, f, err := fs.parsePath(name)
	if err != nil {
		return nil, toStatus(err)
	}
	switch {
	case r != nil:
		return fs.repoAttr(r)
	case f != nil:
		return fs.fileAttr(f, context)
	default:
		return &fuse.Attr{
			Mode: modeDir,
		}, fuse.OK
	}
}

func (fs *filesystem) repoAttr(r *pfs.Repo) (*fuse.Attr, fuse.Status) {
	ri, err := fs.c.InspectRepo(r.Name)
	if err != nil {
		return nil, toStatus(err)
	}
	return &fuse.Attr{
		Mode:      modeDir,
		Ctime:     uint64(ri.Created.Seconds),
		Ctimensec: uint32(ri.Created.Nanos),
		Mtime:     uint64(ri.Created.Seconds),
		Mtimensec: uint32(ri.Created.Nanos),
	}, fuse.OK
}

func repoDirEntry(ri *pfs.RepoInfo) fuse.DirEntry {
	return fuse.DirEntry{
		Name: ri.Repo.Name,
		Mode: modeDir,
	}
}

func fileMode(fi *pfs.FileInfo) uint32 {
	switch fi.FileType {
	case pfs.FileType_FILE:
		return modeFile
	case pfs.FileType_DIR:
		return modeDir
	default:
		return 0
	}
}

func (fs *filesystem) fileAttr(f *pfs.File, context *fuse.Context) (_ *fuse.Attr, retStatus fuse.Status) {
	fileIf, ok := fs.files.Load(fileString(f))
	if ok {
		<-fileIf.(*file).ready
		return fs.loopBackFS.GetAttr(fileString(f), context)
	}
	fi, err := fs.c.InspectFile(f.Commit.Repo.Name, f.Commit.ID, f.Path)
	if err != nil {
		return nil, toStatus(err)
	}
	return &fuse.Attr{
		Mode:      fileMode(fi),
		Size:      fi.SizeBytes,
		Mtime:     uint64(fi.Committed.Seconds),
		Mtimensec: uint32(fi.Committed.Nanos),
	}, fuse.OK
}

func fileDirEntry(fi *pfs.FileInfo) fuse.DirEntry {
	return fuse.DirEntry{
		Mode: fileMode(fi),
		Name: path.Base(fi.File.Path),
	}
}

func toStatus(err error) fuse.Status {
	if strings.Contains(err.Error(), "not found") {
		return fuse.ENOENT
	}
	return fuse.EIO
}
