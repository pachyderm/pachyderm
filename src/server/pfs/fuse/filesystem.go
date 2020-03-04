package fuse

import (
	"fmt"
	"os"
	"os/signal"
	"path"
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
	modeFile = fuse.S_IFREG | 0444 // everyone can read, no one can do anything else
	modeDir  = fuse.S_IFDIR | 0555 // everyone can read and execute, no one can do anything else (execute permission is required to list a dir)
)

// Mount pfs to mountPoint, opts may be left nil.
func Mount(c *client.APIClient, mountPoint string, opts *Options) error {
	nfs := pathfs.NewPathNodeFs(newFileSystem(c, opts.getCommits()), nil)
	server, _, err := nodefs.MountRoot(mountPoint, nfs.Root(), opts.getFuse())
	if err != nil {
		return errors.Wrapf(err, "nodefs.MountRoot")
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
	c         *client.APIClient
	commits   map[string]string
	commitsMu sync.RWMutex
}

func newFileSystem(c *client.APIClient, commits map[string]string) pathfs.FileSystem {
	if commits == nil {
		commits = make(map[string]string)
	}
	return &filesystem{
		FileSystem: pathfs.NewDefaultFileSystem(),
		c:          c,
		commits:    commits,
	}
}

func (fs *filesystem) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	return fs.getAttr(name)
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
		if commit == "" {
			// master branch has no head, so we report an empty dir
			return result, fuse.OK
		}
		if err := fs.c.ListFileF(r.Name, commit, "", 0, func(fi *pfs.FileInfo) error {
			result = append(result, fileDirEntry(fi))
			return nil
		}); err != nil {
			return nil, toStatus(err)
		}
	case f != nil:
		if f.Commit.ID == "" {
			// master branch has no head, so we report an empty dir
			return result, fuse.OK
		}
		if err := fs.c.ListFileF(f.Commit.Repo.Name, f.Commit.ID, f.Path, 0, func(fi *pfs.FileInfo) error {
			result = append(result, fileDirEntry(fi))
			return nil
		}); err != nil {
			return nil, toStatus(err)
		}
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
	f := int(flags)
	writeFlags := os.O_WRONLY | os.O_RDWR
	if f&writeFlags != 0 {
		return nil, fuse.EROFS
	}
	return newFile(fs, name)
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

func (fs *filesystem) getAttr(name string) (*fuse.Attr, fuse.Status) {
	r, f, err := fs.parsePath(name)
	if err != nil {
		return nil, toStatus(err)
	}
	switch {
	case r != nil:
		return fs.repoAttr(r)
	case f != nil:
		return fs.fileAttr(f)
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

func (fs *filesystem) fileAttr(f *pfs.File) (*fuse.Attr, fuse.Status) {
	fi, err := fs.c.InspectFile(f.Commit.Repo.Name, f.Commit.ID, f.Path)
	if err != nil {
		return nil, toStatus(err)
	}
	return &fuse.Attr{
		Mode: fileMode(fi),
		Size: fi.SizeBytes,
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
