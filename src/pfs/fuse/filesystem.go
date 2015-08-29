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

type filesystem struct {
	apiClient      pfs.ApiClient
	repositoryName string
	commitID       string
	shard          uint64
	modulus        uint64
}

func newFilesystem(
	apiClient pfs.ApiClient,
	repositoryName string,
	commitID string,
	shard uint64,
	modulus uint64,
) *filesystem {
	return &filesystem{
		apiClient,
		repositoryName,
		commitID,
		shard,
		modulus,
	}
}

func (f *filesystem) Root() (fs.Node, error) {
	return &directory{f, f.commitID, true, "/"}, nil
}

type directory struct {
	fs       *filesystem
	commitID string
	write    bool
	path     string
}

func (d *directory) Attr(ctx context.Context, a *fuse.Attr) error {
	if d.write {
		a.Mode = os.ModeDir | 0775
	} else {
		a.Mode = os.ModeDir | 0555
	}
	return nil
}

func (d *directory) nodeFromFileInfo(fileInfo *pfs.FileInfo) (fs.Node, error) {
	if fileInfo == nil {
		return nil, fuse.ENOENT
	}
	switch fileInfo.FileType {
	case pfs.FileType_FILE_TYPE_NONE:
		return nil, fuse.ENOENT
	case pfs.FileType_FILE_TYPE_OTHER:
		return nil, fuse.ENOENT
	case pfs.FileType_FILE_TYPE_REGULAR:
		return &file{d.fs, d.commitID, path.Join(d.path, fileInfo.Path.Path), 0,
			int64(fileInfo.SizeBytes)}, nil
	case pfs.FileType_FILE_TYPE_DIR:
		return &directory{d.fs, d.commitID, d.write, fileInfo.Path.Path}, nil
	default:
		return nil, fmt.Errorf("Unrecognized FileType.")
	}
}

func (d *directory) Lookup(ctx context.Context, name string) (fs.Node, error) {
	if d.commitID == "" {
		response, err := pfsutil.GetCommitInfo(
			d.fs.apiClient,
			d.fs.repositoryName,
			name,
		)
		if err != nil {
			return nil, err
		}
		if response.CommitInfo == nil {
			return nil, fuse.ENOENT
		}
		return &directory{
				d.fs,
				name,
				response.CommitInfo.CommitType == pfs.CommitType_COMMIT_TYPE_WRITE,
				"",
			},
			nil
	}
	response, err := pfsutil.GetFileInfo(
		d.fs.apiClient,
		d.fs.repositoryName,
		d.commitID,
		path.Join(d.path, name),
	)
	if err != nil {
		return nil, err
	}
	return d.nodeFromFileInfo(response.FileInfo)
}

func (d *directory) readCommits(ctx context.Context) ([]fuse.Dirent, error) {
	response, err := pfsutil.ListCommits(d.fs.apiClient, d.fs.repositoryName)
	if err != nil {
		return nil, err
	}
	result := make([]fuse.Dirent, 0, len(response.CommitInfo))
	for _, commitInfo := range response.CommitInfo {
		result = append(result, fuse.Dirent{Name: commitInfo.Commit.Id, Type: fuse.DT_Dir})
	}
	return result, nil
}

func (d *directory) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	if d.commitID == "" {
		return d.readCommits(ctx)
	}
	response, err := pfsutil.ListFiles(d.fs.apiClient, d.fs.repositoryName, d.commitID, d.path, d.fs.shard, d.fs.modulus)
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
	if d.commitID == "" {
		return nil, 0, fuse.EPERM
	}
	result := &file{d.fs, d.commitID, path.Join(d.path, request.Name), 0, 0}
	handle, err := result.Open(ctx, nil, nil)
	if err != nil {
		return nil, nil, err
	}
	return result, handle, nil
}

func (d *directory) Mkdir(ctx context.Context, request *fuse.MkdirRequest) (fs.Node, error) {
	if d.commitID == "" {
		return nil, fuse.EPERM
	}
	if err := pfsutil.MakeDirectory(d.fs.apiClient, d.fs.repositoryName, d.commitID, path.Join(d.path, request.Name)); err != nil {
		return nil, err
	}
	return &directory{
		d.fs, d.commitID, d.write, path.Join(d.path, request.Name),
	}, nil
}

type file struct {
	fs       *filesystem
	commitID string
	path     string
	handles  int32
	size     int64
}

func (f *file) Attr(ctx context.Context, a *fuse.Attr) error {
	response, err := pfsutil.GetFileInfo(
		f.fs.apiClient,
		f.fs.repositoryName,
		f.commitID,
		f.path,
	)
	if err != nil {
		return err
	}
	if response.FileInfo != nil {
		a.Size = response.FileInfo.SizeBytes
	}
	a.Mode = 0666
	return nil
}

func (f *file) Read(ctx context.Context, request *fuse.ReadRequest, response *fuse.ReadResponse) error {
	buffer := bytes.NewBuffer(make([]byte, 0, request.Size))
	if err := pfsutil.GetFile(f.fs.apiClient, f.fs.repositoryName, f.commitID, f.path, request.Offset, int64(request.Size), buffer); err != nil {
		return err
	}
	response.Data = buffer.Bytes()
	return nil
}

func (f *file) Open(ctx context.Context, request *fuse.OpenRequest, response *fuse.OpenResponse) (fs.Handle, error) {
	atomic.AddInt32(&f.handles, 1)
	return f, nil
}

func (f *file) Write(ctx context.Context, request *fuse.WriteRequest, response *fuse.WriteResponse) error {
	written, err := pfsutil.PutFile(f.fs.apiClient, f.fs.repositoryName, f.commitID, f.path, request.Offset, bytes.NewReader(request.Data))
	if err != nil {
		return err
	}
	response.Size = int(written)
	if f.size < request.Offset+written {
		f.size = request.Offset + written
	}
	return nil
}
