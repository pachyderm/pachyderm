package fuse

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

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
	conn, err := fuse.Mount(
		mountPoint,
		fuse.FSName("pfs://"+repositoryName),
		fuse.Subtype("pfs"),
		fuse.VolumeName("pfs://"+repositoryName),
	)
	if err != nil {
		return err
	}
	defer conn.Close()
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
	log.Print("directory.Attr")
	a.Inode = 1
	a.Mode = os.ModeDir | 0555
	return nil
}

func nodeFromFileInfo(fs *filesystem, fileInfo *pfs.FileInfo) (fs.Node, error) {
	if fileInfo == nil {
		return nil, fuse.ENOENT
	}
	switch fileInfo.FileType {
	case pfs.FileType_FILE_TYPE_NONE:
		log.Print("FileType_FILE_TYPE_NONE")
		return nil, fuse.ENOENT
	case pfs.FileType_FILE_TYPE_OTHER:
		log.Print("FileType_FILE_TYPE_OTHER")
		return nil, fuse.ENOENT
	case pfs.FileType_FILE_TYPE_REGULAR:
		return &file{fs, fileInfo.Path.Path, 0, fileInfo.SizeBytes, nil}, nil
	case pfs.FileType_FILE_TYPE_DIR:
		return &directory{fs, fileInfo.Path.Path}, nil
	default:
		return nil, fmt.Errorf("Unrecognized FileType.")
	}
}

func (d *directory) Lookup(ctx context.Context, name string) (fs.Node, error) {
	log.Print("directory.Lookup")
	response, err := pfsutil.GetFileInfo(
		d.fs.apiClient,
		d.fs.repositoryName,
		d.fs.commitID,
		filepath.Join(d.path, name),
	)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return nodeFromFileInfo(d.fs, response.GetFileInfo())
}

func (d *directory) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	log.Print("directory.ReadDirAll")
	response, err := pfsutil.ListFiles(d.fs.apiClient, d.fs.repositoryName, d.fs.commitID, d.path, 0, 1)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	var result []fuse.Dirent
	for _, fileInfo := range response.GetFileInfo() {
		shortPath := strings.TrimPrefix(fileInfo.Path.Path, d.path)
		switch fileInfo.FileType {
		case pfs.FileType_FILE_TYPE_NONE:
			continue
		case pfs.FileType_FILE_TYPE_OTHER:
			continue
		case pfs.FileType_FILE_TYPE_REGULAR:
			result = append(result, fuse.Dirent{Inode: 2, Name: shortPath, Type: fuse.DT_File})
		case pfs.FileType_FILE_TYPE_DIR:
			result = append(result, fuse.Dirent{Inode: 2, Name: shortPath, Type: fuse.DT_Dir})
		default:
			continue
		}
	}
	log.Print(result)
	return result, nil
}

type file struct {
	fs      *filesystem
	path    string
	handles int32
	size    uint64
	reader  io.Reader
}

func (f *file) Attr(ctx context.Context, a *fuse.Attr) error {
	log.Printf("Attr: %#v", f)
	a.Inode = 2
	a.Mode = 0666
	a.Size = f.size
	return nil
}

func (f *file) Read(ctx context.Context) ([]byte, error) {
	log.Printf("Read: %#v", f)
	if f.reader == nil {
		reader, err := pfsutil.GetFile(f.fs.apiClient, f.fs.repositoryName, f.fs.commitID, f.path)
		if err != nil {
			return nil, err
		}
		f.reader = reader
	}
	var result []byte
	if _, err := f.reader.Read(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (f *file) Open(ctx context.Context, request *fuse.OpenRequest, response *fuse.OpenResponse) (fs.Handle, error) {
	log.Printf("Open: %#v", f)
	atomic.AddInt32(&f.handles, 1)
	return f, nil
}

func (f *file) Write(ctx context.Context, request *fuse.WriteRequest, response *fuse.WriteResponse) error {
	log.Printf("Write: %#v", f)
	written, err := pfsutil.PutFile(f.fs.apiClient, f.fs.repositoryName, f.fs.commitID, f.path, bytes.NewReader(request.Data))
	if err != nil {
		return err
	}
	response.Size = written
	return nil
}
