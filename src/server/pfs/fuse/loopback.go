// Copyright 2019 the Go-FUSE Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fuse

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"go.uber.org/zap"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
)

type fileState int32

const (
	_     fileState = iota // we don't know about this file (was "none" but linter complained)
	meta                   // we have meta information (but not content for this file)
	full                   // we have full content for this file
	dirty                  // we have full content for this file and the user has written to it
)

func (l *loopbackRoot) getState(mountName string) string {
	l.mu.Lock()
	defer l.mu.Unlock()
	s, ok := l.stateMap[mountName]
	if !ok {
		return ""
	}
	return s
}

// func (l *loopbackRoot) deleteState(mountName string) {
// 	l.mu.Lock()
// 	defer l.mu.Unlock()
// 	_, ok := l.stateMap[mountName]
// 	if !ok {
// 		return
// 	}
// 	delete(l.stateMap, mountName)
// }

type loopbackRoot struct {
	loopbackNode

	rootPath string
	rootDev  uint64

	targetPath string

	write bool

	c *client.APIClient

	stateMap map[string]string       // key is mount name, value is 'mounted', etc
	repoOpts map[string]*RepoOptions // key is mount name
	branches map[string]string       // key is mount name
	commits  map[string]string       // key is mount name
	files    map[string]fileState    // key is {mount_name}/{path}
	mu       sync.Mutex
}

type loopbackNode struct {
	fs.Inode
}

var _ = (fs.NodeStatfser)((*loopbackNode)(nil))
var _ = (fs.NodeStatfser)((*loopbackNode)(nil))
var _ = (fs.NodeGetattrer)((*loopbackNode)(nil))
var _ = (fs.NodeGetxattrer)((*loopbackNode)(nil))
var _ = (fs.NodeSetxattrer)((*loopbackNode)(nil))
var _ = (fs.NodeRemovexattrer)((*loopbackNode)(nil))
var _ = (fs.NodeListxattrer)((*loopbackNode)(nil))
var _ = (fs.NodeReadlinker)((*loopbackNode)(nil))
var _ = (fs.NodeOpener)((*loopbackNode)(nil))
var _ = (fs.NodeCopyFileRanger)((*loopbackNode)(nil))
var _ = (fs.NodeLookuper)((*loopbackNode)(nil))
var _ = (fs.NodeOpendirer)((*loopbackNode)(nil))
var _ = (fs.NodeReaddirer)((*loopbackNode)(nil))
var _ = (fs.NodeMkdirer)((*loopbackNode)(nil))
var _ = (fs.NodeMknoder)((*loopbackNode)(nil))
var _ = (fs.NodeLinker)((*loopbackNode)(nil))
var _ = (fs.NodeSymlinker)((*loopbackNode)(nil))
var _ = (fs.NodeUnlinker)((*loopbackNode)(nil))
var _ = (fs.NodeRmdirer)((*loopbackNode)(nil))
var _ = (fs.NodeRenamer)((*loopbackNode)(nil))

func (n *loopbackNode) Statfs(ctx context.Context, out *fuse.StatfsOut) syscall.Errno {
	s := syscall.Statfs_t{}
	err := syscall.Statfs(n.path(), &s)
	if err != nil {
		return fs.ToErrno(err)
	}
	out.FromStatfsT(&s)
	return fs.OK
}

func (r *loopbackRoot) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	st := syscall.Stat_t{}
	err := syscall.Stat(r.rootPath, &st)
	if err != nil {
		return fs.ToErrno(err)
	}
	out.FromStat(&st)
	return fs.OK
}

func (r *loopbackRoot) getRepoOptionCommit(repoName string) string {
	return r.repoOpts[repoName].File.Commit.Id
}

func (n *loopbackNode) root() *loopbackRoot {
	return n.Root().Operations().(*loopbackRoot)
}

func (n *loopbackNode) c() *client.APIClient {
	return n.root().c
}

func (n *loopbackNode) path() string {
	path := n.Path(nil)
	return filepath.Join(n.root().rootPath, path)
}

func (n *loopbackNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	p := filepath.Join(n.path(), name)
	if err := n.download(ctx, p, meta); err != nil {
		return nil, fs.ToErrno(err)
	}

	st := syscall.Stat_t{}
	err := syscall.Lstat(p, &st)
	if err != nil {
		return nil, fs.ToErrno(err)
	}

	out.Attr.FromStat(&st)
	node := &loopbackNode{}
	ch := n.NewInode(ctx, node, n.root().idFromStat(&st))
	return ch, 0
}

func (n *loopbackNode) Mknod(ctx context.Context, name string, mode, rdev uint32, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	p := filepath.Join(n.path(), name)
	if errno := n.checkWrite(p); errno != 0 {
		return nil, errno
	}
	err := syscall.Mknod(p, mode, int(rdev))
	if err != nil {
		return nil, fs.ToErrno(err)
	}
	st := syscall.Stat_t{}
	if err := syscall.Lstat(p, &st); err != nil {
		// TODO multierr
		syscall.Rmdir(p) //nolint:errcheck
		return nil, fs.ToErrno(err)
	}

	out.Attr.FromStat(&st)

	node := &loopbackNode{}
	ch := n.NewInode(ctx, node, n.root().idFromStat(&st))

	return ch, 0
}

func (n *loopbackNode) Mkdir(ctx context.Context, name string, mode uint32, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	p := filepath.Join(n.path(), name)
	if errno := n.checkWrite(p); errno != 0 {
		return nil, errno
	}
	if err := n.download(ctx, p, meta); err != nil {
		return nil, fs.ToErrno(err)
	}
	err := os.Mkdir(p, os.FileMode(mode))
	if err != nil {
		return nil, fs.ToErrno(err)
	}
	st := syscall.Stat_t{}
	if err := syscall.Lstat(p, &st); err != nil {
		// TODO multierr
		syscall.Rmdir(p) //nolint:errcheck // favour outer error instead
		return nil, fs.ToErrno(err)
	}

	out.Attr.FromStat(&st)

	node := &loopbackNode{}
	ch := n.NewInode(ctx, node, n.root().idFromStat(&st))

	return ch, 0
}

func (n *loopbackNode) Rmdir(ctx context.Context, name string) syscall.Errno {
	p := filepath.Join(n.path(), name)
	if errno := n.checkWrite(p); errno != 0 {
		return errno
	}
	if err := n.download(ctx, p, meta); err != nil {
		return fs.ToErrno(err)
	}
	err := syscall.Rmdir(p)
	return fs.ToErrno(err)
}

func (n *loopbackNode) Unlink(ctx context.Context, name string) (errno syscall.Errno) {
	p := filepath.Join(n.path(), name)
	if errno := n.checkWrite(p); errno != 0 {
		return errno
	}
	if err := n.download(ctx, p, meta); err != nil {
		return fs.ToErrno(err)
	}
	defer func() {
		if errno == 0 {
			n.setFileState(p, dirty)
		}
	}()
	err := syscall.Unlink(p)
	return fs.ToErrno(err)
}

func toLoopbackNode(op fs.InodeEmbedder) *loopbackNode {
	if r, ok := op.(*loopbackRoot); ok {
		return &r.loopbackNode
	}
	return op.(*loopbackNode)
}

func (n *loopbackNode) Rename(ctx context.Context, name string, newParent fs.InodeEmbedder, newName string, flags uint32) syscall.Errno {
	newParentLoopback := toLoopbackNode(newParent)
	if flags&fs.RENAME_EXCHANGE != 0 {
		return n.renameExchange(name, newParentLoopback, newName)
	}

	p1 := filepath.Join(n.path(), name)

	p2 := filepath.Join(newParentLoopback.path(), newName)
	if errno := n.checkWrite(p1); errno != 0 {
		return errno
	}
	if errno := n.checkWrite(p2); errno != 0 {
		return errno
	}
	err := os.Rename(p1, p2)
	return fs.ToErrno(err)
}

func (r *loopbackRoot) idFromStat(st *syscall.Stat_t) fs.StableAttr {
	return fs.StableAttr{
		Mode: uint32(st.Mode),
		Gen:  1,
		Ino:  0, // let fuse generate this automatically
	}
}

var _ = (fs.NodeCreater)((*loopbackNode)(nil))

func (n *loopbackNode) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (inode *fs.Inode, fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	p := filepath.Join(n.path(), name)
	if errno := n.checkWrite(p); errno != 0 {
		return nil, nil, 0, errno
	}
	if err := n.download(ctx, p, full); err != nil {
		return nil, nil, 0, fs.ToErrno(err)
	}
	defer func() {
		if errno == 0 {
			n.setFileState(p, dirty)
		}
	}()

	fd, err := syscall.Open(p, int(flags)|os.O_CREATE, mode)
	if err != nil {
		return nil, nil, 0, fs.ToErrno(err)
	}

	st := syscall.Stat_t{}
	if err := syscall.Fstat(fd, &st); err != nil {
		syscall.Close(fd)
		return nil, nil, 0, fs.ToErrno(err)
	}

	node := &loopbackNode{}
	ch := n.NewInode(ctx, node, n.root().idFromStat(&st))
	lf := NewLoopbackFile(fd)

	out.FromStat(&st)
	return ch, lf, 0, 0
}

func (n *loopbackNode) Symlink(ctx context.Context, target, name string, out *fuse.EntryOut) (_ *fs.Inode, errno syscall.Errno) {
	p := filepath.Join(n.path(), name)
	if errno := n.checkWrite(p); errno != 0 {
		return nil, errno
	}
	if err := n.download(ctx, p, full); err != nil {
		return nil, fs.ToErrno(err)
	}
	target = filepath.Join(n.root().rootPath, n.trimTargetPath(target))
	if err := n.download(ctx, target, full); err != nil {
		return nil, fs.ToErrno(err)
	}
	defer func() {
		if errno == 0 {
			n.setFileState(p, dirty)
		}
	}()
	err := syscall.Symlink(target, p)
	if err != nil {
		return nil, fs.ToErrno(err)
	}
	st := syscall.Stat_t{}
	if err := syscall.Lstat(p, &st); err != nil {
		// TODO multierr
		syscall.Unlink(p) //nolint:errcheck // favour outer error instead
		return nil, fs.ToErrno(err)
	}
	node := &loopbackNode{}
	ch := n.NewInode(ctx, node, n.root().idFromStat(&st))

	out.Attr.FromStat(&st)
	return ch, 0
}

func (n *loopbackNode) Link(ctx context.Context, target fs.InodeEmbedder, name string, out *fuse.EntryOut) (_ *fs.Inode, errno syscall.Errno) {
	p := filepath.Join(n.path(), name)
	if errno := n.checkWrite(p); errno != 0 {
		return nil, errno
	}
	if err := n.download(ctx, p, full); err != nil {
		return nil, fs.ToErrno(err)
	}
	targetNode := toLoopbackNode(target)
	if err := n.download(ctx, targetNode.path(), full); err != nil {
		return nil, fs.ToErrno(err)
	}
	err := syscall.Link(targetNode.path(), p)
	if err != nil {
		return nil, fs.ToErrno(err)
	}
	defer func() {
		if errno == 0 {
			n.setFileState(p, dirty)
		}
	}()
	st := syscall.Stat_t{}
	if err := syscall.Lstat(p, &st); err != nil {
		// TODO multierr
		syscall.Unlink(p) //nolint:errcheck
		return nil, fs.ToErrno(err)
	}
	node := &loopbackNode{}
	ch := n.NewInode(ctx, node, n.root().idFromStat(&st))

	out.Attr.FromStat(&st)
	return ch, 0
}

func (n *loopbackNode) Readlink(ctx context.Context) ([]byte, syscall.Errno) {
	p := n.path()

	for l := 256; ; l *= 2 {
		buf := make([]byte, l)
		sz, err := syscall.Readlink(p, buf)
		if err != nil {
			return nil, fs.ToErrno(err)
		}

		if sz < len(buf) {
			return buf[:sz], 0
		}
	}
}

func (n *loopbackNode) Open(ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	p := n.path()
	state := full
	if isWrite(flags) {
		if errno := n.checkWrite(p); errno != 0 {
			return nil, 0, errno
		}
		state = dirty
	}
	if err := n.download(ctx, p, state); err != nil {
		return nil, 0, fs.ToErrno(err)
	}
	if isCreate(flags) {
		defer func() {
			if errno == 0 {
				n.setFileState(p, dirty)
			}
		}()
	}
	f, err := syscall.Open(p, int(flags), 0)
	if err != nil {
		return nil, 0, fs.ToErrno(err)
	}
	lf := NewLoopbackFile(f)
	return lf, 0, 0
}

func (n *loopbackNode) Opendir(ctx context.Context) syscall.Errno {
	if err := n.download(ctx, n.path(), meta); err != nil {
		return fs.ToErrno(err)
	}
	fd, err := syscall.Open(n.path(), syscall.O_DIRECTORY, 0755)
	if err != nil {
		return fs.ToErrno(err)
	}
	syscall.Close(fd)
	return fs.OK
}

func (n *loopbackNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	if err := n.download(ctx, n.path(), meta); err != nil {
		return nil, fs.ToErrno(err)
	}
	return fs.NewLoopbackDirStream(n.path())
}

func (n *loopbackNode) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	if f != nil {
		return f.(fs.FileGetattrer).Getattr(ctx, out)
	}
	p := n.path()
	if err := n.download(ctx, p, meta); err != nil {
		return fs.ToErrno(err)
	}

	var err error
	st := syscall.Stat_t{}
	err = syscall.Lstat(p, &st)
	if err != nil {
		return fs.ToErrno(err)
	}
	out.FromStat(&st)
	return fs.OK
}

var _ = (fs.NodeSetattrer)((*loopbackNode)(nil))

func (n *loopbackNode) Setattr(ctx context.Context, f fs.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
	p := n.path()
	fsa, ok := f.(fs.FileSetattrer)
	if ok && fsa != nil {
		if errno := fsa.Setattr(ctx, in, out); errno != 0 {
			return errno
		}
	} else {
		if m, ok := in.GetMode(); ok {
			if err := syscall.Chmod(p, m); err != nil {
				return fs.ToErrno(err)
			}
		}

		uid, uok := in.GetUID()
		gid, gok := in.GetGID()
		if uok || gok {
			suid := -1
			sgid := -1
			if uok {
				suid = int(uid)
			}
			if gok {
				sgid = int(gid)
			}
			if err := syscall.Chown(p, suid, sgid); err != nil {
				return fs.ToErrno(err)
			}
		}

		mtime, mok := in.GetMTime()
		atime, aok := in.GetATime()

		if mok || aok {

			ap := &atime
			mp := &mtime
			if !aok {
				ap = nil
			}
			if !mok {
				mp = nil
			}
			var ts [2]syscall.Timespec
			ts[0] = fuse.UtimeToTimespec(ap)
			ts[1] = fuse.UtimeToTimespec(mp)

			if err := syscall.UtimesNano(p, ts[:]); err != nil {
				return fs.ToErrno(err)
			}
		}

		if sz, ok := in.GetSize(); ok {
			if err := syscall.Truncate(p, int64(sz)); err != nil {
				return fs.ToErrno(err)
			}
		}
	}

	fga, ok := f.(fs.FileGetattrer)
	if ok && fga != nil {
		if errno := fga.Getattr(ctx, out); errno != 0 {
			return errno
		}
	} else {
		st := syscall.Stat_t{}
		err := syscall.Lstat(p, &st)
		if err != nil {
			return fs.ToErrno(err)
		}
		out.FromStat(&st)
	}
	return fs.OK
}

// newLoopbackRoot returns a root node for a loopback file system whose
// root is at the given root. This node implements all NodeXxxxer
// operations available.
func newLoopbackRoot(root, target string, c *client.APIClient, opts *Options) (*loopbackRoot, error) {
	var st syscall.Stat_t
	err := syscall.Stat(root, &st)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	n := &loopbackRoot{
		rootPath:   root,
		rootDev:    uint64(st.Dev),
		targetPath: target,
		write:      opts.getWrite(),
		c:          c,
		repoOpts:   opts.getRepoOpts(),
		branches:   opts.getBranches(),
		commits:    make(map[string]string),
		files:      make(map[string]fileState),
		stateMap:   make(map[string]string),
	}
	return n, nil
}

func (n *loopbackNode) mkdirMountNames() (retErr error) {
	defer func() {
		if retErr == nil {
			n.setFileState("", meta)
		}
	}()
	ros := n.root().repoOpts
	// we only mount explicitly named repos now
	for name := range ros {
		p := n.namePath(name)
		if err := os.MkdirAll(p, 0777); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

// download files into the loopback filesystem, if meta is true then only the
// directory structure will be created, no actual data will be downloaded,
// files will be truncated to their actual sizes (but will be all zeros).
func (n *loopbackNode) download(ctx context.Context, origPath string, state fileState) (retErr error) {
	if n.getFileState(origPath) >= state {
		// Already got this file, so we can just return
		return nil
	}
	if err := n.mkdirMountNames(); err != nil {
		return err
	}
	path := n.trimPath(origPath)
	parts := strings.Split(path, "/")
	defer func() {
		if retErr == nil {
			n.setFileState(path, state)
		}
	}()
	// Note, len(parts) < 1 should not actually be possible, but just in case
	// no need to panic.
	if len(parts) < 1 || parts[0] == "" {
		return nil // already downloaded in downloadRepos
	}
	name := parts[0]
	st := n.root().getState(name)
	// don't download while we're anything other than mounted
	// TODO: we probably want some more locking/coordination (in the other
	// direction) to stop the state machine changing state _during_ a download()
	// NB: empty string case is to support pachctl mount
	if !(st == "" || st == "mounted") {
		log.Info(pctx.TODO(), "Skipping download because of state", zap.String("origPath", origPath), zap.String("name", name), zap.String("state", st), zap.Int32("getFileState(origPath)", int32(n.getFileState(origPath))), zap.Int32("state", int32(state)))
		// return an error to stop an empty directory listing being cached by
		// the OS
		return errors.WithStack(fmt.Errorf("repo at %s is not mounted", name))
	}
	branch := n.root().branch(name)
	commitID, err := n.commit(name)
	if err != nil {
		return err
	}
	if commitID == "" {
		return nil
	}
	// log the commit
	log.Info(pctx.TODO(), "Downloading", zap.String("path", origPath), zap.String("from", fmt.Sprintf("%s@%s", name, commitID)))
	ro, ok := n.root().repoOpts[name]
	if !ok {
		return errors.WithStack(fmt.Errorf("[download] can't find mount named %s", name))
	}
	projectName := ro.File.Commit.Repo.Project.GetName()
	repoName := ro.File.Commit.Repo.Name
	commit := client.NewCommit(projectName, repoName, branch, commitID)
	filePath := filepath.Join(parts[1:]...)
	// ListFile callback function
	createFile := func(fi *pfs.FileInfo) (retErr error) {
		if fi.FileType == pfs.FileType_DIR {
			return errors.EnsureStack(os.MkdirAll(n.filePath(name, fi), 0777))
		}
		p := n.filePath(name, fi)
		// If file manifestation is successful to the given state level, cache
		// that we know about it, to avoid a subsequent stat() going back to
		// Pachyderm over gRPC
		defer func() {
			if retErr == nil {
				n.setFileState(p, state)
			}
		}()
		// Make sure the directory exists
		// I think this may be unnecessary based on the constraints the
		// OS imposes, but don't want to rely on that, especially
		// because Mkdir should be pretty cheap.
		if err := os.MkdirAll(filepath.Dir(p), 0777); err != nil {
			return errors.WithStack(err)
		}
		f, err := os.Create(p)
		if err != nil {
			return errors.WithStack(err)
		}
		defer func() {
			if err := f.Close(); err != nil && retErr == nil {
				retErr = errors.WithStack(err)
			}
		}()
		if state < full {
			return errors.EnsureStack(f.Truncate(int64(fi.SizeBytes)))
		}
		if err := n.c().GetFile(fi.File.Commit, fi.File.Path, f); err != nil {
			return err
		}
		return nil
	}
	// Calling ListFile on a repo with many files at the top level when we only care about a single
	// file, for example, is very expensive. Hence the below optimizations.
	// We only create descendants and ancestors of mounted file ro.File
	// ro.File.Path is the path of a mounted file, filePath is OS requested path
	// Example: ro.File.Path = /file, filePath = /files <--- DON'T CREATE
	// Example: ro.File.Path = /file, filePath = /file/path <--- CREATE
	// Example: ro.File.Path = /dir/file, filePath = /dir <--- CREATE
	ancestor, intermediates := isGrandparentOf(filePath, ro.File.Path)
	switch {
	case ancestor:
		path := filepath.Join(n.namePath(name), filePath, intermediates)
		if err := errors.EnsureStack(os.MkdirAll(path, 0777)); err != nil {
			return err
		}
	case isParentOf(filePath, ro.File.Path):
		fi, err := n.c().InspectFile(commit, ro.File.Path)
		if err != nil {
			return err
		}
		if fi.FileType == pfs.FileType_DIR {
			if err := errors.EnsureStack(os.MkdirAll(n.filePath(name, fi), 0777)); err != nil {
				return err
			}
		} else {
			if err := n.createFiles(commit, ro.File.Path, createFile); err != nil {
				return err
			}
		}
	case isDescendantOf(filePath, ro.File.Path):
		if err := n.createFiles(commit, filePath, createFile); err != nil {
			return err
		}
	default:
		// If a user creates folders in the mount, they will enter a path that fails the
		// previous cases. In this case, we ignore, since there's nothing to mount.
	}
	return nil
}

func (n *loopbackNode) trimPath(path string) string {
	path = strings.TrimPrefix(path, n.root().rootPath)
	return strings.TrimPrefix(path, "/")
}

func (n *loopbackNode) trimTargetPath(path string) string {
	path = strings.TrimPrefix(path, n.root().targetPath)
	return strings.TrimPrefix(path, "/")
}

func (n *loopbackNode) branch(name string) string {
	// no need to lock mu for branches since we only ever read from it.
	if branch, ok := n.root().branches[name]; ok {
		return branch
	}
	return ""
}

func (n *loopbackNode) commit(name string) (string, error) {
	if commit, ok := func() (string, bool) {
		n.root().mu.Lock()
		defer n.root().mu.Unlock()
		commit, ok := n.root().commits[name]
		return commit, ok
	}(); ok {
		return commit, nil
	}
	ro, ok := n.root().repoOpts[name]
	if !ok {
		// often happens that something tries to access e.g.
		// /pfs/.python_version or some file that can't exist at that level. not
		// worth spamming the logs with this
		return "", nil
	}
	projectName := ro.File.Commit.Repo.Project.GetName()
	repoName := ro.File.Commit.Repo.Name
	branch := n.root().branch(name)
	if branch == "" {
		commitId := n.root().getRepoOptionCommit(repoName)
		if commitId == "" {
			return "", errors.New("cannot resolve which commit to mount: not found in branch or repoOptions")
		}
		return commitId, nil
	}
	bi, err := n.root().c.InspectBranch(projectName, repoName, branch)
	if err != nil && !errutil.IsNotFoundError(err) {
		return "", err
	}
	// Lock mu to assign commits
	n.root().mu.Lock()
	defer n.root().mu.Unlock()
	// You can access branches that don't exist, which allows you to create
	// branches through the fuse mount.
	if errutil.IsNotFoundError(err) {
		n.root().commits[name] = ""
		return "", nil
	}
	n.root().commits[name] = bi.Head.Id
	return bi.Head.Id, nil
}

func (n *loopbackNode) namePath(name string) string {
	return filepath.Join(n.root().rootPath, name)
}

func (n *loopbackNode) filePath(name string, fi *pfs.FileInfo) string {
	return filepath.Join(n.root().rootPath, name, fi.File.Path)
}

func (n *loopbackNode) getFileState(path string) fileState {
	n.root().mu.Lock()
	defer n.root().mu.Unlock()
	return n.root().files[n.trimPath(path)]
}

func (n *loopbackNode) setFileState(path string, state fileState) {
	n.root().mu.Lock()
	defer n.root().mu.Unlock()
	n.root().files[n.trimPath(path)] = state
}

func (n *loopbackNode) checkWrite(path string) syscall.Errno {
	name := strings.Split(n.trimPath(path), "/")[0]
	ros := n.root().repoOpts
	if len(ros) > 0 {
		ro, ok := ros[name]
		if !ok || !ro.Write {
			return syscall.EROFS
		}
		return 0
	}
	if !n.root().write {
		return syscall.EROFS
	}
	return 0
}

func (n *loopbackNode) createFiles(commit *pfs.Commit, path string, cb func(fi *pfs.FileInfo) error) error {
	if err := n.c().ListFile(commit, path, cb); err != nil && !errutil.IsNotFoundError(err) &&
		!pfsserver.IsOutputCommitNotFinishedErr(err) {
		return err
	}
	return nil
}

func isWrite(flags uint32) bool {
	return (int(flags) & (os.O_WRONLY | os.O_RDWR)) != 0
}

func isCreate(flags uint32) bool {
	return int(flags)&os.O_CREATE != 0
}
