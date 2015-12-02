package fuse

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.pedge.io/protolog"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/pfsutil"
	"golang.org/x/net/context"
)

type filesystem struct {
	apiClient pfs.APIClient
	Filesystem
	inodes map[string]uint64
	lock   sync.RWMutex
}

func newFilesystem(
	apiClient pfs.APIClient,
	shard *pfs.Shard,
	commits []*pfs.Commit,
) *filesystem {
	return &filesystem{
		apiClient,
		Filesystem{
			shard,
			commits,
		},
		make(map[string]uint64),
		sync.RWMutex{},
	}
}

func (f *filesystem) Root() (result fs.Node, retErr error) {
	defer func() {
		protolog.Debug(&Root{&f.Filesystem, getNode(result), errorToString(retErr)})
	}()
	return &directory{
		f,
		Node{&pfs.File{
			Commit: &pfs.Commit{
				Repo: &pfs.Repo{},
			},
		},
			true},
	}, nil
}

type directory struct {
	fs *filesystem
	Node
}

func (d *directory) Attr(ctx context.Context, a *fuse.Attr) (retErr error) {
	defer func() {
		protolog.Debug(&DirectoryAttr{&d.Node, &Attr{uint32(a.Mode)}, errorToString(retErr)})
	}()
	a.Valid = time.Nanosecond
	if d.Write {
		a.Mode = os.ModeDir | 0775
	} else {
		a.Mode = os.ModeDir | 0555
	}
	a.Inode = d.fs.inode(d.File)
	return nil
}

func (d *directory) Lookup(ctx context.Context, name string) (result fs.Node, retErr error) {
	defer func() {
		protolog.Debug(&DirectoryLookup{&d.Node, name, getNode(result), errorToString(retErr)})
	}()
	if d.File.Commit.Repo.Name == "" {
		return d.lookUpRepo(ctx, name)
	}
	if d.File.Commit.Id == "" {
		return d.lookUpCommit(ctx, name)
	}
	return d.lookUpFile(ctx, name)
}

func (d *directory) ReadDirAll(ctx context.Context) (result []fuse.Dirent, retErr error) {
	defer func() {
		var dirents []*Dirent
		for _, dirent := range result {
			dirents = append(dirents, &Dirent{dirent.Inode, dirent.Name})
		}
		protolog.Debug(&DirectoryReadDirAll{&d.Node, dirents, errorToString(retErr)})
	}()
	if d.File.Commit.Repo.Name == "" {
		return d.readRepos(ctx)
	}
	if d.File.Commit.Id == "" {
		return d.readCommits(ctx)
	}
	return d.readFiles(ctx)
}

func (d *directory) Create(ctx context.Context, request *fuse.CreateRequest, response *fuse.CreateResponse) (result fs.Node, _ fs.Handle, retErr error) {
	defer func() {
		protolog.Debug(&DirectoryCreate{&d.Node, getNode(result), errorToString(retErr)})
	}()
	if d.File.Commit.Id == "" {
		return nil, 0, fuse.EPERM
	}
	directory := d.copy()
	directory.File.Path = path.Join(directory.File.Path, request.Name)
	localResult := &file{*directory, 0, 0}
	handle, err := localResult.Open(ctx, nil, nil)
	if err != nil {
		return nil, nil, err
	}
	return localResult, handle, nil
}

func (d *directory) Mkdir(ctx context.Context, request *fuse.MkdirRequest) (result fs.Node, retErr error) {
	defer func() {
		protolog.Debug(&DirectoryMkdir{&d.Node, getNode(result), errorToString(retErr)})
	}()
	if d.File.Commit.Id == "" {
		return nil, fuse.EPERM
	}
	if err := pfsutil.MakeDirectory(d.fs.apiClient, d.File.Commit.Repo.Name, d.File.Commit.Id, path.Join(d.File.Path, request.Name)); err != nil {
		return nil, err
	}
	localResult := d.copy()
	localResult.File.Path = path.Join(localResult.File.Path, request.Name)
	return localResult, nil
}

type file struct {
	directory
	handles int32
	size    int64
}

func (f *file) Attr(ctx context.Context, a *fuse.Attr) (retErr error) {
	defer func() {
		protolog.Debug(&FileAttr{&f.Node, &Attr{uint32(a.Mode)}, errorToString(retErr)})
	}()
	fileInfo, err := pfsutil.InspectFile(
		f.fs.apiClient,
		f.File.Commit.Repo.Name,
		f.File.Commit.Id,
		f.File.Path,
		f.fs.Shard,
	)
	if err != nil {
		return err
	}
	if fileInfo != nil {
		a.Size = fileInfo.SizeBytes
	}
	a.Mode = 0666
	a.Inode = f.fs.inode(f.File)
	return nil
}

func (f *file) Read(ctx context.Context, request *fuse.ReadRequest, response *fuse.ReadResponse) (retErr error) {
	defer func() {
		protolog.Debug(&FileRead{&f.Node, errorToString(retErr)})
	}()
	var buffer bytes.Buffer
	if err := pfsutil.GetFile(
		f.fs.apiClient,
		f.File.Commit.Repo.Name,
		f.File.Commit.Id,
		f.File.Path,
		request.Offset,
		int64(request.Size),
		f.fs.Shard,
		&buffer,
	); err != nil {
		return err
	}
	response.Data = buffer.Bytes()
	return nil
}

func (f *file) Open(ctx context.Context, request *fuse.OpenRequest, response *fuse.OpenResponse) (_ fs.Handle, retErr error) {
	defer func() {
		protolog.Debug(&FileRead{&f.Node, errorToString(retErr)})
	}()
	atomic.AddInt32(&f.handles, 1)
	return f, nil
}

func (f *file) Write(ctx context.Context, request *fuse.WriteRequest, response *fuse.WriteResponse) (retErr error) {
	defer func() {
		protolog.Debug(&FileWrite{&f.Node, errorToString(retErr)})
	}()
	written, err := pfsutil.PutFile(f.fs.apiClient, f.File.Commit.Repo.Name, f.File.Commit.Id, f.File.Path, request.Offset, bytes.NewReader(request.Data))
	if err != nil {
		return err
	}
	response.Size = written
	if f.size < request.Offset+int64(written) {
		f.size = request.Offset + int64(written)
	}
	return nil
}

func (f *filesystem) inode(file *pfs.File) uint64 {
	f.lock.RLock()
	inode, ok := f.inodes[key(file)]
	f.lock.RUnlock()
	if ok {
		return inode
	}
	f.lock.Lock()
	if inode, ok := f.inodes[key(file)]; ok {
		return inode
	}
	newInode := uint64(len(f.inodes))
	f.inodes[key(file)] = newInode
	f.lock.Unlock()
	return newInode
}

func (d *directory) copy() *directory {
	return &directory{
		fs: d.fs,
		Node: Node{
			File: &pfs.File{
				Commit: &pfs.Commit{
					Repo: &pfs.Repo{
						Name: d.File.Commit.Repo.Name,
					},
					Id: d.File.Commit.Id,
				},
				Path: d.File.Path,
			},
			Write: d.Write,
		},
	}
}

func (f *filesystem) repoWhitelisted(name string) bool {
	if len(f.Commits) == 0 {
		return true
	}
	found := false
	for _, commit := range f.Commits {
		if commit.Repo.Name == name {
			found = true
			break
		}
	}
	return found
}

func (d *directory) commitWhitelisted(name string) bool {
	if len(d.fs.Commits) == 0 {
		return true
	}
	found := false
	for _, commit := range d.fs.Commits {
		if commit.Repo.Name == d.File.Commit.Repo.Name {
			found = true
			break
		}
	}
	return found
}

func (d *directory) lookUpRepo(ctx context.Context, name string) (fs.Node, error) {
	if !d.fs.repoWhitelisted(name) {
		return nil, fuse.EPERM
	}
	repoInfo, err := pfsutil.InspectRepo(d.fs.apiClient, name)
	if err != nil {
		return nil, err
	}
	if repoInfo == nil {
		return nil, fuse.ENOENT
	}
	result := d.copy()
	result.File.Commit.Repo.Name = name
	if d.fs.Commits != nil {
		for _, commit := range d.fs.Commits {
			if commit.Repo.Name == name {
				return result.lookUpCommit(ctx, commit.Id)
			}
		}
		return nil, fmt.Errorf("unreachable")
	}
	return result, nil
}

func (d *directory) lookUpCommit(ctx context.Context, name string) (fs.Node, error) {
	if !d.commitWhitelisted(name) {
		return nil, fuse.EPERM
	}
	commitInfo, err := pfsutil.InspectCommit(
		d.fs.apiClient,
		d.File.Commit.Repo.Name,
		name,
	)
	if err != nil {
		return nil, err
	}
	if commitInfo == nil {
		return nil, fuse.ENOENT
	}
	result := d.copy()
	result.File.Commit.Id = name
	if commitInfo.CommitType == pfs.CommitType_COMMIT_TYPE_READ {
		result.Write = false
	} else {
		result.Write = true
	}
	return result, nil
}

func (d *directory) lookUpFile(ctx context.Context, name string) (fs.Node, error) {
	fileInfo, err := pfsutil.InspectFile(
		d.fs.apiClient,
		d.File.Commit.Repo.Name,
		d.File.Commit.Id,
		path.Join(d.File.Path, name),
		d.fs.Shard,
	)
	if err != nil {
		return nil, err
	}
	if fileInfo.FileType == pfs.FileType_FILE_TYPE_NONE {
		return nil, fuse.ENOENT
	}
	directory := d.copy()
	directory.File.Path = fileInfo.File.Path
	switch fileInfo.FileType {
	case pfs.FileType_FILE_TYPE_REGULAR:
		directory.File.Path = fileInfo.File.Path
		return &file{
			*directory,
			0,
			int64(fileInfo.SizeBytes),
		}, nil
	case pfs.FileType_FILE_TYPE_DIR:
		return directory, nil
	default:
		return nil, fmt.Errorf("Unrecognized FileType.")
	}
}

func (d *directory) readRepos(ctx context.Context) ([]fuse.Dirent, error) {
	repoInfos, err := pfsutil.ListRepo(d.fs.apiClient)
	if err != nil {
		return nil, err
	}
	var result []fuse.Dirent
	for _, repoInfo := range repoInfos {
		if d.fs.repoWhitelisted(repoInfo.Repo.Name) {
			result = append(result, fuse.Dirent{Name: repoInfo.Repo.Name, Type: fuse.DT_Dir})
		}
	}
	return result, nil
}

func (d *directory) readCommits(ctx context.Context) ([]fuse.Dirent, error) {
	commitInfos, err := pfsutil.ListCommit(d.fs.apiClient, d.File.Commit.Repo.Name)
	if err != nil {
		return nil, err
	}
	result := make([]fuse.Dirent, 0, len(commitInfos))
	for _, commitInfo := range commitInfos {
		if d.commitWhitelisted(commitInfo.Commit.Id) {
			result = append(result, fuse.Dirent{Name: commitInfo.Commit.Id, Type: fuse.DT_Dir})
		}
	}
	return result, nil
}

func (d *directory) readFiles(ctx context.Context) ([]fuse.Dirent, error) {
	fileInfos, err := pfsutil.ListFile(d.fs.apiClient, d.File.Commit.Repo.Name, d.File.Commit.Id, d.File.Path, d.fs.Shard)
	if err != nil {
		return nil, err
	}
	var result []fuse.Dirent
	for _, fileInfo := range fileInfos {
		shortPath := strings.TrimPrefix(fileInfo.File.Path, d.File.Path)
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

// TODO this code is duplicate elsewhere, we should put it somehwere.
func errorToString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func getNode(node fs.Node) *Node {
	switch n := node.(type) {
	default:
		return nil
	case *directory:
		return &n.Node
	case *file:
		return &n.Node
	}
}

func key(file *pfs.File) string {
	return fmt.Sprintf("%s/%s/%s", file.Commit.Repo.Name, file.Commit.Id, file.Path)
}
