package fuse

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"strings"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
)

type filesystem struct {
	pathfs.FileSystem
	c *client.APIClient
}

func NewFileSystem(c *client.APIClient) pathfs.FileSystem {
	return &filesystem{
		FileSystem: pathfs.NewDefaultFileSystem(),
		c:          c,
	}
}

func Mount(mountPoint string, fs pathfs.FileSystem) error {
	nfs := pathfs.NewPathNodeFs(fs, nil)
	server, _, err := nodefs.MountRoot(mountPoint, nfs.Root(), &nodefs.Options{Debug: true})
	if err != nil {
		return fmt.Errorf("nodefs.MountRoot: %v", err)
	}
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		server.Unmount()
	}()
	server.Serve()
	return nil
}

func (fs *filesystem) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	return getAttr(fs.c, name)
}

func (fs *filesystem) OpenDir(name string, context *fuse.Context) ([]fuse.DirEntry, fuse.Status) {
	var result []fuse.DirEntry
	r, f := parsePath(name)
	switch {
	case r != nil:
		if err := fs.c.ListFileF(r.Name, "master", "", func(fi *pfs.FileInfo) error {
			result = append(result, fileDirEntry(fi))
			return nil
		}); err != nil {
			return nil, toStatus(err)
		}
	case f != nil:
		if err := fs.c.ListFileF(f.Commit.Repo.Name, f.Commit.ID, f.Path, func(fi *pfs.FileInfo) error {
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
	// TODO use flags
	return newFile(fs.c, name), fuse.OK
}

func parsePath(name string) (*pfs.Repo, *pfs.File) {
	components := strings.Split(name, "/")
	switch {
	case name == "":
		return nil, nil
	case len(components) == 1:
		return client.NewRepo(components[0]), nil
	default:
		return nil, client.NewFile(components[0], "master", path.Join(components[1:]...))
	}
}

func getAttr(c *client.APIClient, name string) (*fuse.Attr, fuse.Status) {
	r, f := parsePath(name)
	switch {
	case r != nil:
		return repoAttr(c, r)
	case f != nil:
		return fileAttr(c, f)
	default:
		return &fuse.Attr{
			Mode: fuse.S_IFDIR | 0755,
		}, fuse.OK
	}
}

func repoAttr(c *client.APIClient, r *pfs.Repo) (*fuse.Attr, fuse.Status) {
	ri, err := c.InspectRepo(r.Name)
	if err != nil {
		return nil, toStatus(err)
	}
	return &fuse.Attr{
		Mode:      fuse.S_IFDIR | 0755,
		Ctime:     uint64(ri.Created.Seconds),
		Ctimensec: uint32(ri.Created.Nanos),
		Mtime:     uint64(ri.Created.Seconds),
		Mtimensec: uint32(ri.Created.Nanos),
	}, fuse.OK
}

func repoDirEntry(ri *pfs.RepoInfo) fuse.DirEntry {
	return fuse.DirEntry{
		Name: ri.Repo.Name,
		Mode: fuse.S_IFDIR | 0755,
	}
}

func fileMode(fi *pfs.FileInfo) uint32 {
	switch fi.FileType {
	case pfs.FileType_FILE:
		return fuse.S_IFREG | 0644
	case pfs.FileType_DIR:
		return fuse.S_IFDIR | 0644
	default:
		return 0
	}
}

func fileAttr(c *client.APIClient, f *pfs.File) (*fuse.Attr, fuse.Status) {
	fi, err := c.InspectFile(f.Commit.Repo.Name, f.Commit.ID, f.Path)
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
		Name: fi.File.Path,
	}
}

func toStatus(err error) fuse.Status {
	if strings.Contains(err.Error(), "not found") {
		return fuse.ENOENT
	}
	return fuse.EIO
}
