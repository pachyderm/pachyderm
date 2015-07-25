package fuse

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/pfsutil"
	"golang.org/x/net/context"
)

func Mount(apiClient pfs.ApiClient, repositoryName string, commitID string, mountPoint string) error {
	if err := os.MkdirAll(mountPoint, 0777); err != nil {
		return err
	}
	c, err := fuse.Mount(
		mountPoint,
		fuse.FSName("pfs"),
		fuse.Subtype("pfs"),
		fuse.VolumeName("pfs://"+repositoryName),
	)
	if err != nil {
		return err
	}
	defer c.Close()

	if err := fs.Serve(c, &filesystem{apiClient, repositoryName, commitID}); err != nil {
		return err
	}

	// check if the mount process has an error to report
	<-c.Ready
	return c.MountError
}

func Unmount(mountPoint string) error {
	return nil
}

type filesystem struct {
	apiClient      pfs.ApiClient
	repositoryName string
	commitID       string
}

func (f *filesystem) Root() (fs.Node, error) {
	return &directory{f, "/"}, nil
}

type directory struct {
	fs   *filesystem
	path string
}

func (*directory) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 1
	a.Mode = os.ModeDir | 0555
	return nil
}

func nodeFromFileInfo(fs *filesystem, fileInfo *pfs.FileInfo) (fs.Node, error) {
	switch fileInfo.FileType {
	case pfs.FileType_FILE_TYPE_NONE:
		log.Print("FileType_FILE_TYPE_NONE")
		return nil, fuse.ENOENT
	case pfs.FileType_FILE_TYPE_OTHER:
		log.Print("FileType_FILE_TYPE_OTHER")
		return nil, fuse.ENOENT
	case pfs.FileType_FILE_TYPE_REGULAR:
		return &file{fs, fileInfo.Path.Path, fileInfo.SizeBytes}, nil
	case pfs.FileType_FILE_TYPE_DIR:
		return &directory{fs, fileInfo.Path.Path}, nil
	default:
		return nil, fmt.Errorf("Unrecognized FileType.")
	}
}

func (d *directory) Lookup(ctx context.Context, name string) (fs.Node, error) {
	response, err := pfsutil.GetFileInfo(
		d.fs.apiClient,
		d.fs.repositoryName,
		d.fs.commitID,
		filepath.Join(d.path, name),
	)
	if err != nil {
		return nil, err
	}
	return nodeFromFileInfo(d.fs, response.GetFileInfo())
}

func (d *directory) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	response, err := pfsutil.ListFiles(d.fs.apiClient, d.fs.repositoryName, d.fs.commitID, d.path, 0, 1)
	if err != nil {
		return nil, err
	}
	var result []fuse.Dirent
	for _, fileInfo := range response.GetFileInfo() {
		switch fileInfo.FileType {
		case pfs.FileType_FILE_TYPE_NONE:
			continue
		case pfs.FileType_FILE_TYPE_OTHER:
			continue
		case pfs.FileType_FILE_TYPE_REGULAR:
			result = append(result, fuse.Dirent{Inode: 3, Name: fileInfo.Path.Path, Type: fuse.DT_File})
		case pfs.FileType_FILE_TYPE_DIR:
			result = append(result, fuse.Dirent{Inode: 3, Name: fileInfo.Path.Path, Type: fuse.DT_Dir})
		}
	}
	return result, nil
}

type file struct {
	fs   *filesystem
	path string
	size uint64
}

func (f *file) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 2
	a.Mode = 0444
	a.Size = f.size
	return nil
}
