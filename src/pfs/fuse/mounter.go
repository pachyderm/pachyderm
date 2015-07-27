package fuse

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
	"sync/atomic"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/pfsutil"
	"golang.org/x/net/context"
)

type mounter struct{}

func newMounter() Mounter {
	return &mounter{}
}

func (*mounter) Mount(apiClient pfs.ApiClient, repositoryName string, commitID string, mountPoint string) (retErr error) {
	if err := os.MkdirAll(mountPoint, 0777); err != nil {
		return err
	}
	conn, err := fuse.Mount(
		mountPoint,
		fuse.FSName("pfs://"+repositoryName),
		fuse.Subtype("pfs"),
		fuse.VolumeName("pfs://"+repositoryName),
	)
	if err != nil {
		return err
	}
	defer func() {
		if err := conn.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	if err := fs.Serve(conn, &filesystem{apiClient, repositoryName, commitID}); err != nil {
		return err
	}

	// check if the mount process has an error to report
	<-conn.Ready
	return conn.MountError
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
	a.Mode = os.ModeDir | 0775
	return nil
}

func nodeFromFileInfo(fs *filesystem, fileInfo *pfs.FileInfo) (fs.Node, error) {
	if fileInfo == nil {
		return nil, fuse.ENOENT
	}
	switch fileInfo.FileType {
	case pfs.FileType_FILE_TYPE_NONE:
		return nil, fuse.ENOENT
	case pfs.FileType_FILE_TYPE_OTHER:
		return nil, fuse.ENOENT
	case pfs.FileType_FILE_TYPE_REGULAR:
		return &file{fs, fileInfo.Path.Path, 0, fileInfo.SizeBytes}, nil
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
		path.Join(d.path, name),
	)
	if err != nil {
		return nil, err
	}
	return nodeFromFileInfo(d.fs, response.FileInfo)
}

func (d *directory) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	response, err := pfsutil.ListFiles(d.fs.apiClient, d.fs.repositoryName, d.fs.commitID, d.path, 0, 1)
	if err != nil {
		return nil, err
	}
	result := make([]fuse.Dirent, 0, len(response.FileInfo))
	for _, fileInfo := range response.FileInfo {
		shortPath := strings.TrimPrefix(fileInfo.Path.Path, d.path)
		switch fileInfo.FileType {
		case pfs.FileType_FILE_TYPE_NONE:
			continue
		case pfs.FileType_FILE_TYPE_OTHER:
			continue
		case pfs.FileType_FILE_TYPE_REGULAR:
			result = append(result, fuse.Dirent{Name: shortPath, Type: fuse.DT_File})
		case pfs.FileType_FILE_TYPE_DIR:
			result = append(result, fuse.Dirent{Name: shortPath, Type: fuse.DT_Dir})
		default:
			continue
		}
	}
	return result, nil
}

func (d *directory) Create(ctx context.Context, request *fuse.CreateRequest, response *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	result := &file{d.fs, path.Join(d.path, request.Name), 0, 0}
	handle, err := result.Open(ctx, nil, nil)
	if err != nil {
		return nil, nil, err
	}
	return result, handle, nil
}

type file struct {
	fs      *filesystem
	path    string
	handles int32
	size    uint64
}

func (f *file) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mode = 0666
	a.Size = f.size
	return nil
}

func (f *file) Read(ctx context.Context, request *fuse.ReadRequest, response *fuse.ReadResponse) error {
	reader, err := pfsutil.GetFile(f.fs.apiClient, f.fs.repositoryName, f.fs.commitID, f.path)
	if err != nil {
		return err
	}
	response.Data = make([]byte, request.Size)
	if _, err := reader.Read(response.Data); err != nil {
		return err
	}
	return nil
}

func (f *file) Open(ctx context.Context, request *fuse.OpenRequest, response *fuse.OpenResponse) (fs.Handle, error) {
	atomic.AddInt32(&f.handles, 1)
	return f, nil
}

func (f *file) Write(ctx context.Context, request *fuse.WriteRequest, response *fuse.WriteResponse) error {
	written, err := pfsutil.PutFile(f.fs.apiClient, f.fs.repositoryName, f.fs.commitID, f.path, bytes.NewReader(request.Data))
	if err != nil {
		return err
	}
	response.Size = written
	return nil
}
