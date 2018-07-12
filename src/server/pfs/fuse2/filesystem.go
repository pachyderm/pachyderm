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

func (fs *filesystem) GetAttr(name string, context *fuse.Context) (result *fuse.Attr, _ fuse.Status) {
	r, f := parsePath(name)
	switch {
	case r != nil:
		ri, err := fs.c.InspectRepo(r.Name)
		if err != nil {
			return nil, fuse.EIO
		}
		return repoAttr(ri), fuse.OK
	case f != nil:
		fi, err := fs.c.InspectFile(f.Commit.Repo.Name, f.Commit.ID, f.Path)
		if err != nil {
			return nil, fuse.EIO
		}
		return fileAttr(fi), fuse.OK
	default:
		return &fuse.Attr{
			Mode: fuse.S_IFDIR | 0755,
		}, fuse.OK
	}
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
			return nil, fuse.EIO
		}
	case f != nil:
		if err := fs.c.ListFileF(f.Commit.Repo.Name, f.Commit.ID, f.Path, func(fi *pfs.FileInfo) error {
			result = append(result, fileDirEntry(fi))
			return nil
		}); err != nil {
			return nil, fuse.EIO
		}
	default:
		ris, err := fs.c.ListRepo()
		if err != nil {
			return nil, fuse.EIO
		}
		for _, ri := range ris {
			result = append(result, repoDirEntry(ri))
		}
	}
	return result, fuse.OK
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

func repoAttr(ri *pfs.RepoInfo) *fuse.Attr {
	return &fuse.Attr{
		Mode:      fuse.S_IFDIR | 0755,
		Ctime:     uint64(ri.Created.Seconds),
		Ctimensec: uint32(ri.Created.Nanos),
		Mtime:     uint64(ri.Created.Seconds),
		Mtimensec: uint32(ri.Created.Nanos),
	}
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

func fileAttr(fi *pfs.FileInfo) *fuse.Attr {
	return &fuse.Attr{
		Mode: fileMode(fi),
		Size: fi.SizeBytes,
	}
}

func fileDirEntry(fi *pfs.FileInfo) fuse.DirEntry {
	return fuse.DirEntry{
		Mode: fileMode(fi),
		Name: fi.File.Path,
	}
}
