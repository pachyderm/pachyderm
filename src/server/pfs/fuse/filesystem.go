package fuse

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	pfs_server "github.com/pachyderm/pachyderm/src/server/pfs/server"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/gogo/protobuf/types"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type filesystem struct {
	apiClient *client.APIClient
	Filesystem
	inodes map[string]uint64
	lock   sync.RWMutex
}

func newFilesystem(
	apiClient *client.APIClient,
	commitMounts []*CommitMount,
) *filesystem {
	return &filesystem{
		apiClient: apiClient,
		Filesystem: Filesystem{
			commitMounts,
		},
		inodes: make(map[string]uint64),
	}
}

// repoFilesystem is the same as filesystem except that it only presents the content
// of one specific repo.
type repoFilesystem struct {
	*filesystem
}

func newRepoFilesystem(
	apiClient *client.APIClient,
	commitMount *CommitMount,
) *repoFilesystem {
	return &repoFilesystem{&filesystem{
		apiClient: apiClient,
		Filesystem: Filesystem{
			[]*CommitMount{commitMount},
		},
		inodes: make(map[string]uint64),
	}}
}

func (f *repoFilesystem) Root() (result fs.Node, retErr error) {
	defer func() {
		if retErr == nil {
			log.Debug(&Root{&f.Filesystem, getNode(result), errorToString(retErr)})
		} else {
			log.Error(&Root{&f.Filesystem, getNode(result), errorToString(retErr)})
		}
	}()
	return &directory{
		f.filesystem,
		Node{
			File: &pfsclient.File{
				Commit: f.filesystem.CommitMounts[0].Commit,
			},
		},
	}, nil
}

func (f *filesystem) Root() (result fs.Node, retErr error) {
	defer func() {
		if retErr == nil {
			log.Debug(&Root{&f.Filesystem, getNode(result), errorToString(retErr)})
		} else {
			log.Error(&Root{&f.Filesystem, getNode(result), errorToString(retErr)})
		}
	}()
	return &directory{
		f,
		Node{
			File: &pfsclient.File{
				Commit: &pfsclient.Commit{
					Repo: &pfsclient.Repo{},
				},
			},
		},
	}, nil
}

type directory struct {
	fs *filesystem
	Node
}

func (d *directory) Attr(ctx context.Context, a *fuse.Attr) (retErr error) {
	defer func() {
		if retErr == nil {
			log.Debug(&DirectoryAttr{&d.Node, &Attr{uint32(a.Mode)}, errorToString(retErr)})
		} else {
			log.Error(&DirectoryAttr{&d.Node, &Attr{uint32(a.Mode)}, errorToString(retErr)})
		}
	}()

	a.Valid = time.Nanosecond
	if d.Write {
		a.Mode = os.ModeDir | 0775
	} else {
		a.Mode = os.ModeDir | 0555
	}
	a.Inode = d.fs.inode(d.File)
	a.Mtime, _ = types.TimestampFromProto(d.Modified)
	return nil
}

func (d *directory) Lookup(ctx context.Context, name string) (result fs.Node, retErr error) {
	defer func() {
		if retErr == nil || isNotFound(retErr) || retErr == fuse.ENOENT {
			log.Debug(&DirectoryLookup{&d.Node, name, getNode(result), errorToString(retErr)})
		} else {
			log.Error(&DirectoryLookup{&d.Node, name, getNode(result), errorToString(retErr)})
		}
	}()
	if d.File.Commit.Repo.Name == "" {
		return d.lookUpRepo(ctx, name)
	}
	if d.File.Commit.ID == "" {
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
		if retErr == nil {
			log.Debug(&DirectoryReadDirAll{&d.Node, dirents, errorToString(retErr)})
		} else {
			log.Error(&DirectoryReadDirAll{&d.Node, dirents, errorToString(retErr)})
		}
	}()
	if d.File.Commit.Repo.Name == "" {
		return d.readRepos(ctx)
	}
	if d.File.Commit.ID == "" {
		commitMount := d.fs.getCommitMount(d.getRepoOrAliasName())
		if commitMount != nil && commitMount.Commit.ID != "" {
			d.File.Commit.ID = commitMount.Commit.ID
			return d.readFiles(ctx)
		}
		return d.readCommits(ctx)
	}
	return d.readFiles(ctx)
}

func (d *directory) Create(ctx context.Context, request *fuse.CreateRequest, response *fuse.CreateResponse) (result fs.Node, _ fs.Handle, retErr error) {
	defer func() {
		if retErr == nil {
			log.Debug(&DirectoryCreate{&d.Node, getNode(result), errorToString(retErr)})
		} else {
			log.Error(&DirectoryCreate{&d.Node, getNode(result), errorToString(retErr)})
		}
	}()
	if d.File.Commit.ID == "" {
		return nil, 0, fuse.EPERM
	}
	directory := d.copy()
	directory.File.Path = path.Join(directory.File.Path, request.Name)
	localResult := &file{
		directory: *directory,
		size:      0,
	}
	if err := localResult.touch(); err != nil {
		// Check if its a write on a finished commit:
		if pfs_server.IsPermissionError(err) {
			err = fuse.EPERM
		}
		return nil, 0, err
	}
	response.Flags |= fuse.OpenDirectIO | fuse.OpenNonSeekable
	handle := localResult.newHandle(0)
	return localResult, handle, nil
}

func (d *directory) Mkdir(ctx context.Context, request *fuse.MkdirRequest) (result fs.Node, retErr error) {
	defer func() {
		if retErr == nil {
			log.Debug(&DirectoryMkdir{&d.Node, getNode(result), errorToString(retErr)})
		} else {
			log.Error(&DirectoryMkdir{&d.Node, getNode(result), errorToString(retErr)})
		}
	}()
	if d.File.Commit.ID == "" {
		return nil, fuse.EPERM
	}
	// Server has no concept of empty directories.
	localResult := d.copy()
	localResult.File.Path = path.Join(localResult.File.Path, request.Name)
	return localResult, nil
}

func (d *directory) Remove(ctx context.Context, req *fuse.RemoveRequest) (retErr error) {
	defer func() {
		if retErr == nil {
			log.Debug(&FileRemove{&d.Node, req.Name, req.Dir, errorToString(retErr)})
		} else {
			log.Error(&FileRemove{&d.Node, req.Name, req.Dir, errorToString(retErr)})
		}
	}()
	return d.fs.apiClient.DeleteFile(d.Node.File.Commit.Repo.Name,
		d.Node.File.Commit.ID, filepath.Join(d.Node.File.Path, req.Name))
}

type file struct {
	directory
	size    int64
	handles []*handle
	lock    sync.Mutex
}

func (f *file) Attr(ctx context.Context, a *fuse.Attr) (retErr error) {
	defer func() {
		if retErr == nil {
			log.Debug(&FileAttr{&f.Node, &Attr{uint32(a.Mode)}, errorToString(retErr)})
		} else {
			log.Error(&FileAttr{&f.Node, &Attr{uint32(a.Mode)}, errorToString(retErr)})
		}
	}()
	fileInfo, err := f.fs.apiClient.InspectFile(
		f.File.Commit.Repo.Name,
		f.File.Commit.ID,
		f.File.Path,
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

func (f *file) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) (retErr error) {
	defer func() {
		if retErr == nil {
			log.Debug(&FileSetAttr{&f.Node, errorToString(retErr)})
		} else {
			log.Error(&FileSetAttr{&f.Node, errorToString(retErr)})
		}
	}()
	if req.Size == 0 && (req.Valid&fuse.SetattrSize) > 0 {
		err := f.fs.apiClient.DeleteFile(f.Node.File.Commit.Repo.Name,
			f.Node.File.Commit.ID, f.Node.File.Path)
		if err != nil {
			return err
		}
		if err := f.touch(); err != nil {
			return err
		}
		for _, handle := range f.handles {
			handle.lock.Lock()
			handle.cursor = 0
			handle.lock.Unlock()
		}
	}
	return nil
}

func (f *file) Open(ctx context.Context, request *fuse.OpenRequest, response *fuse.OpenResponse) (_ fs.Handle, retErr error) {
	defer func() {
		if retErr == nil {
			log.Debug(&FileOpen{&f.Node, errorToString(retErr)})
		} else {
			log.Error(&FileOpen{&f.Node, errorToString(retErr)})
		}
	}()
	response.Flags |= fuse.OpenDirectIO | fuse.OpenNonSeekable
	fileInfo, err := f.fs.apiClient.InspectFile(
		f.File.Commit.Repo.Name,
		f.File.Commit.ID,
		f.File.Path,
	)
	if err != nil {
		return nil, err
	}
	return f.newHandle(int(fileInfo.SizeBytes)), nil
}

func (f *file) Fsync(ctx context.Context, req *fuse.FsyncRequest) error {
	f.lock.Lock()
	defer f.lock.Unlock()
	for _, h := range f.handles {
		if h.w != nil {
			w := h.w
			h.w = nil
			if err := w.Close(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (f *file) delimiter() pfsclient.Delimiter {
	if strings.HasSuffix(f.File.Path, ".json") {
		return pfsclient.Delimiter_JSON
	}
	if strings.HasSuffix(f.File.Path, ".bin") {
		return pfsclient.Delimiter_NONE
	}
	return pfsclient.Delimiter_LINE
}

func (f *file) touch() error {
	w, err := f.fs.apiClient.PutFileWriter(
		f.File.Commit.Repo.Name,
		f.File.Commit.ID,
		f.File.Path,
	)
	if err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return nil
}

func (f *filesystem) inode(file *pfsclient.File) uint64 {
	f.lock.RLock()
	inode, ok := f.inodes[key(file)]
	f.lock.RUnlock()
	if ok {
		return inode
	}
	f.lock.Lock()
	defer f.lock.Unlock()
	if inode, ok := f.inodes[key(file)]; ok {
		return inode
	}
	newInode := uint64(len(f.inodes))
	f.inodes[key(file)] = newInode
	return newInode
}

func (f *file) newHandle(cursor int) *handle {
	f.lock.Lock()
	defer f.lock.Unlock()

	h := &handle{
		f:      f,
		cursor: cursor,
	}

	f.handles = append(f.handles, h)

	return h
}

type handle struct {
	f      *file
	w      io.WriteCloser
	cursor int
	lock   sync.Mutex
}

func (h *handle) Read(ctx context.Context, request *fuse.ReadRequest, response *fuse.ReadResponse) (retErr error) {
	defer func() {
		if retErr == nil {
			log.Debug(&FileRead{&h.f.Node, string(response.Data), errorToString(retErr)})
		} else {
			log.Error(&FileRead{&h.f.Node, string(response.Data), errorToString(retErr)})
		}
	}()
	var buffer bytes.Buffer
	if err := h.f.fs.apiClient.GetFile(
		h.f.File.Commit.Repo.Name,
		h.f.File.Commit.ID,
		h.f.File.Path,
		request.Offset,
		int64(request.Size),
		&buffer,
	); err != nil {
		if grpc.Code(err) == codes.NotFound {
			// ENOENT from read(2) is weird, let's call this EINVAL
			// instead.
			return fuse.Errno(syscall.EINVAL)
		}
		return err
	}
	response.Data = buffer.Bytes()
	return nil
}

func (h *handle) Write(ctx context.Context, request *fuse.WriteRequest, response *fuse.WriteResponse) (retErr error) {
	defer func() {
		if retErr == nil {
			log.Debug(&FileWrite{&h.f.Node, string(request.Data), request.Offset, errorToString(retErr)})
		} else {
			log.Error(&FileWrite{&h.f.Node, string(request.Data), request.Offset, errorToString(retErr)})
		}
	}()
	h.lock.Lock()
	defer h.lock.Unlock()
	if h.w == nil {
		w, err := h.f.fs.apiClient.PutFileWriter(
			h.f.File.Commit.Repo.Name, h.f.File.Commit.ID, h.f.File.Path)
		if err != nil {
			return err
		}
		h.w = w
	}
	// repeated is how many bytes in this write have already been sent in
	// previous call to Write. Why does the OS send us the same data twice in
	// different calls? Good question, this is a behavior that's only been
	// observed on osx, not on linux.
	repeated := h.cursor - int(request.Offset)
	if repeated < 0 {
		return fmt.Errorf("gap in bytes written, (OpenNonSeekable should make this impossible)")
	}
	if repeated > len(request.Data) {
		// it's currently unclear under which conditions fuse sends us a
		// request with only repeated data. This prevents us from getting an
		// array bounds error when it does though.
		return nil
	}
	written, err := h.w.Write(request.Data[repeated:])
	if err != nil {
		return err
	}
	response.Size = written + repeated
	h.cursor += written
	if h.f.size < request.Offset+int64(written) {
		h.f.size = request.Offset + int64(written)
	}
	return nil
}

func (h *handle) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	h.lock.Lock()
	defer h.lock.Unlock()
	if h.w != nil {
		w := h.w
		h.w = nil
		if err := w.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (h *handle) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	return nil
}

func (d *directory) copy() *directory {
	return &directory{
		fs: d.fs,
		Node: Node{
			File: &pfsclient.File{
				Commit: &pfsclient.Commit{
					Repo: &pfsclient.Repo{
						Name: d.File.Commit.Repo.Name,
					},
					ID: d.File.Commit.ID,
				},
				Path: d.File.Path,
			},
			Write:     d.Write,
			RepoAlias: d.RepoAlias,
		},
	}
}

func (d *directory) getRepoOrAliasName() string {
	if d.RepoAlias != "" {
		return d.RepoAlias
	}
	return d.File.Commit.Repo.Name
}

func (f *filesystem) getCommitMount(nameOrAlias string) *CommitMount {
	if len(f.CommitMounts) == 0 {
		return &CommitMount{
			Commit: client.NewCommit(nameOrAlias, ""),
		}
	}

	// We prefer alias matching over repo name matching, since there can be
	// two commit mounts with the same repo but different aliases, such as
	// "out" and "prev"
	for _, commitMount := range f.CommitMounts {
		if commitMount.Alias == nameOrAlias {
			return commitMount
		}
	}
	for _, commitMount := range f.CommitMounts {
		if commitMount.Commit.Repo.Name == nameOrAlias {
			return commitMount
		}
	}

	return nil
}

func (d *directory) lookUpRepo(ctx context.Context, name string) (fs.Node, error) {
	commitMount := d.fs.getCommitMount(name)
	if commitMount == nil {
		return nil, fuse.EPERM
	}
	repoInfo, err := d.fs.apiClient.InspectRepo(commitMount.Commit.Repo.Name)
	if err != nil {
		return nil, err
	}
	if repoInfo == nil {
		return nil, fuse.ENOENT
	}
	result := d.copy()
	result.File.Commit.Repo.Name = commitMount.Commit.Repo.Name
	result.File.Commit.ID = commitMount.Commit.ID
	result.RepoAlias = commitMount.Alias

	if commitMount.Commit.ID == "" {
		// We don't have a commit mount
		result.Write = false
		result.Modified = repoInfo.Created
		return result, nil
	}
	commitInfo, err := d.fs.apiClient.InspectCommit(
		commitMount.Commit.Repo.Name,
		commitMount.Commit.ID,
	)
	if err != nil {
		return nil, err
	}
	if commitInfo.Finished != nil {
		result.Write = false
	} else {
		result.Write = true
	}
	result.Modified = commitInfo.Finished

	return result, nil
}

func (d *directory) lookUpCommit(ctx context.Context, name string) (fs.Node, error) {
	commitID := commitPathToID(name)
	commitInfo, err := d.fs.apiClient.InspectCommit(
		d.File.Commit.Repo.Name,
		commitID,
	)
	if err != nil {
		return nil, err
	}
	if commitInfo == nil {
		return nil, fuse.ENOENT
	}
	result := d.copy()
	result.File.Commit.ID = commitID
	if commitInfo.Finished != nil {
		result.Write = false
	} else {
		result.Write = true
	}
	result.Modified = commitInfo.Finished
	return result, nil
}

func (d *directory) lookUpFile(ctx context.Context, name string) (fs.Node, error) {
	var fileInfo *pfsclient.FileInfo
	var err error

	fileInfo, err = d.fs.apiClient.InspectFile(
		d.File.Commit.Repo.Name,
		d.File.Commit.ID,
		path.Join(d.File.Path, name),
	)
	if err != nil {
		return nil, fuse.ENOENT
	}
	if d.Node.Write {
		fileInfo.SizeBytes = 0
	}

	// We want to inherit the metadata other than the path, which should be the
	// path currently being looked up
	directory := d.copy()
	directory.File.Path = fileInfo.File.Path

	switch fileInfo.FileType {
	case pfsclient.FileType_FILE:
		return &file{
			directory: *directory,
			size:      int64(fileInfo.SizeBytes),
		}, nil
	case pfsclient.FileType_DIR:
		return directory, nil
	default:
		return nil, fmt.Errorf("unrecognized file type")
	}
}

func (d *directory) readRepos(ctx context.Context) ([]fuse.Dirent, error) {
	var result []fuse.Dirent
	if len(d.fs.CommitMounts) == 0 {
		repoInfos, err := d.fs.apiClient.ListRepo(nil)
		if err != nil {
			return nil, err
		}
		for _, repoInfo := range repoInfos {
			result = append(result, fuse.Dirent{Name: repoInfo.Repo.Name, Type: fuse.DT_Dir})
		}
	} else {
		for _, mount := range d.fs.CommitMounts {
			name := mount.Commit.Repo.Name
			if mount.Alias != "" {
				name = mount.Alias
			}
			result = append(result, fuse.Dirent{Name: name, Type: fuse.DT_Dir})
		}
	}
	return result, nil
}

func (d *directory) readCommits(ctx context.Context) ([]fuse.Dirent, error) {
	commitInfos, err := d.fs.apiClient.ListCommit(d.File.Commit.Repo.Name, "", "", 0)
	if err != nil {
		return nil, err
	}
	var result []fuse.Dirent
	for _, commitInfo := range commitInfos {
		commitPath := commitIDToPath(commitInfo.Commit.ID)
		result = append(result, fuse.Dirent{Name: commitPath, Type: fuse.DT_Dir})
	}
	branches, err := d.fs.apiClient.ListBranch(d.File.Commit.Repo.Name)
	if err != nil {
		return nil, err
	}
	for _, branch := range branches {
		result = append(result, fuse.Dirent{Name: branch.Name, Type: fuse.DT_Dir})
	}
	return result, nil
}

func (d *directory) readFiles(ctx context.Context) ([]fuse.Dirent, error) {
	fileInfos, err := d.fs.apiClient.ListFile(
		d.File.Commit.Repo.Name,
		d.File.Commit.ID,
		d.File.Path,
	)
	if err != nil {
		return nil, err
	}
	var result []fuse.Dirent
	for _, fileInfo := range fileInfos {
		shortPath := strings.TrimPrefix(fileInfo.File.Path, d.File.Path)
		if shortPath[0] == '/' {
			shortPath = shortPath[1:]
		}
		switch fileInfo.FileType {
		case pfsclient.FileType_FILE:
			result = append(result, fuse.Dirent{Name: shortPath, Type: fuse.DT_File})
		case pfsclient.FileType_DIR:
			result = append(result, fuse.Dirent{Name: shortPath, Type: fuse.DT_Dir})
		default:
			continue
		}
	}
	return result, nil
}

// Since commit IDs look like "master/2", we can't directly use them as filenames
// due to the slash.  So we convert them to something like "master-2"
func commitIDToPath(commitID string) string {
	return strings.Join(strings.Split(commitID, "/"), "-")
}

// the opposite of commitIDToPath
func commitPathToID(path string) string {
	parts := strings.Split(path, "-")
	if len(parts) < 2 {
		// In this case, path is just a branch name
		return path
	}
	return strings.Join(parts[:len(parts)-1], "-") + "/" + parts[len(parts)-1]
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

func key(file *pfsclient.File) string {
	return fmt.Sprintf("%s/%s/%s", file.Commit.Repo.Name, file.Commit.ID, file.Path)
}

func isNotFound(err error) bool {
	return strings.Contains(err.Error(), "not found")
}
