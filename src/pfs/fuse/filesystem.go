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
	apiClient pfs.APIClient
	shard     uint64
	modulus   uint64
}

func newFilesystem(
	apiClient pfs.APIClient,
	shard uint64,
	modulus uint64,
) *filesystem {
	return &filesystem{
		apiClient,
		shard,
		modulus,
	}
}

func (f *filesystem) Root() (fs.Node, error) {
	return &directory{f, "", "", "", true}, nil
}

type directory struct {
	fs       *filesystem
	repoName string
	commitID string
	path     string
	write    bool
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
	case pfs.FileType_FILE_TYPE_REGULAR:
		directory := *d
		directory.path = path.Join(d.path, fileInfo.File.Path)
		result := file{
			directory,
			0,
			int64(fileInfo.SizeBytes),
		}
		return &result,
			nil
	case pfs.FileType_FILE_TYPE_DIR:
		return &directory{d.fs, d.repoName, d.commitID, fileInfo.File.Path, d.write}, nil
	default:
		return nil, fmt.Errorf("Unrecognized FileType.")
	}
}

func (d *directory) lookUpRepo(ctx context.Context, name string) (fs.Node, error) {
	repoInfo, err := pfsutil.InspectRepo(d.fs.apiClient, name)
	if err != nil {
		return nil, err
	}
	if repoInfo == nil {
		return nil, fuse.ENOENT
	}
	result := *d
	result.repoName = name
	return &result, nil
}

func (d *directory) lookUpCommit(ctx context.Context, name string) (fs.Node, error) {
	commitInfo, err := pfsutil.InspectCommit(
		d.fs.apiClient,
		d.repoName,
		name,
	)
	if err != nil {
		return nil, err
	}
	if commitInfo == nil {
		return nil, fuse.ENOENT
	}
	result := *d
	result.commitID = name
	return &result, nil
}

func (d *directory) lookUpFile(ctx context.Context, name string) (fs.Node, error) {
	fileInfo, err := pfsutil.InspectFile(
		d.fs.apiClient,
		d.repoName,
		d.commitID,
		path.Join(d.path, name),
	)
	if err != nil {
		return nil, err
	}
	if fileInfo == nil {
		return nil, fuse.ENOENT
	}
	directory := *d
	directory.path = fileInfo.File.Path
	switch fileInfo.FileType {
	case pfs.FileType_FILE_TYPE_REGULAR:
		directory.path = fileInfo.File.Path
		return &file{
			directory,
			0,
			int64(fileInfo.SizeBytes),
		}, nil
	case pfs.FileType_FILE_TYPE_DIR:
		return &directory, nil
	default:
		return nil, fmt.Errorf("Unrecognized FileType.")
	}
}

func (d *directory) Lookup(ctx context.Context, name string) (fs.Node, error) {
	if d.repoName == "" {
		return d.lookUpRepo(ctx, name)
	}
	if d.commitID == "" {
		return d.lookUpCommit(ctx, name)
	}
	return d.lookUpFile(ctx, name)
}

func (d *directory) readCommits(ctx context.Context) ([]fuse.Dirent, error) {
	commitInfos, err := pfsutil.ListCommit(d.fs.apiClient, d.repoName)
	if err != nil {
		return nil, err
	}
	result := make([]fuse.Dirent, 0, len(commitInfos))
	for _, commitInfo := range commitInfos {
		result = append(result, fuse.Dirent{Name: commitInfo.Commit.Id, Type: fuse.DT_Dir})
	}
	return result, nil
}

func (d *directory) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	if d.commitID == "" {
		return d.readCommits(ctx)
	}
	fileInfos, err := pfsutil.ListFile(d.fs.apiClient, d.repoName, d.commitID, d.path, d.fs.shard, d.fs.modulus)
	if err != nil {
		return nil, err
	}
	result := make([]fuse.Dirent, 0, len(fileInfos))
	for _, fileInfo := range fileInfos {
		shortPath := strings.TrimPrefix(fileInfo.File.Path, d.path)
		switch fileInfo.FileType {
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
	directory := *d
	directory.path = path.Join(directory.path, request.Name)
	result := &file{directory, 0, 0}
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
	if err := pfsutil.MakeDirectory(d.fs.apiClient, d.repoName, d.commitID, path.Join(d.path, request.Name)); err != nil {
		return nil, err
	}
	result := *d
	result.path = path.Join(result.path, request.Name)
	return &result, nil
}

type file struct {
	directory
	handles int32
	size    int64
}

func (f *file) Attr(ctx context.Context, a *fuse.Attr) error {
	fileInfo, err := pfsutil.InspectFile(
		f.fs.apiClient,
		f.repoName,
		f.commitID,
		f.path,
	)
	if err != nil {
		return err
	}
	if fileInfo != nil {
		a.Size = fileInfo.SizeBytes
	}
	a.Mode = 0666
	return nil
}

func (f *file) Read(ctx context.Context, request *fuse.ReadRequest, response *fuse.ReadResponse) error {
	buffer := bytes.NewBuffer(make([]byte, 0, request.Size))
	if err := pfsutil.GetFile(f.fs.apiClient, f.repoName, f.commitID, f.path, request.Offset, int64(request.Size), buffer); err != nil {
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
	written, err := pfsutil.PutFile(f.fs.apiClient, f.repoName, f.commitID, f.path, request.Offset, bytes.NewReader(request.Data))
	if err != nil {
		return err
	}
	response.Size = written
	if f.size < request.Offset+int64(written) {
		f.size = request.Offset + int64(written)
	}
	return nil
}
