package server

import (
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/renew"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	"golang.org/x/net/context"
)

func (d *driver) modifyFile(ctx context.Context, commit *pfs.Commit, cb func(*fileset.UnorderedWriter) error) error {
	// Store the originally-requested parameters because they will be overwritten by inspectCommit
	repo := commit.Branch.Repo.Name
	branch := commit.Branch.Name
	commitID := commit.ID
	if branch == "" && !uuid.IsUUIDWithoutDashes(commitID) {
		branch = commitID
		commitID = ""
	}
	commitInfo, err := d.inspectCommit(ctx, commit, pfs.CommitState_STARTED)
	if err != nil {
		if (!isNotFoundErr(err) && !isNoHeadErr(err)) || branch == "" {
			return err
		}
		return d.oneOffModifyFile(ctx, repo, branch, cb)
	}
	if commitInfo.Finished != nil {
		// The commit is already finished - if the commit was explicitly specified,
		// error out, otherwise we can make a child commit since this is the branch head.
		if commitID != "" {
			return pfsserver.ErrCommitFinished{commitInfo.Commit}
		}
		var opts []fileset.UnorderedWriterOption
		if commitInfo.ParentCommit != nil {
			parentFilesetID, err := d.getFileset(ctx, commitInfo.ParentCommit)
			if err != nil {
				return err
			}
			opts = append(opts, fileset.WithParentID(parentFilesetID))
		}
		return d.oneOffModifyFile(ctx, repo, branch, cb, opts...)
	}
	filesetID, err := d.getFileset(ctx, commitInfo.Commit)
	if err != nil {
		return err
	}
	return d.withCommitUnorderedWriter(ctx, commitInfo.Commit, cb, fileset.WithParentID(filesetID))
}

// TODO: Cleanup after failure?
func (d *driver) oneOffModifyFile(ctx context.Context, repo, branch string, cb func(*fileset.UnorderedWriter) error, opts ...fileset.UnorderedWriterOption) error {
	return d.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) (retErr error) {
		commit, err := d.startCommit(txnCtx, nil, client.NewBranch(repo, branch), nil, "")
		if err != nil {
			return err
		}
		if err := d.withCommitUnorderedWriter(ctx, commit, cb, opts...); err != nil {
			return err
		}
		return d.finishCommit(txnCtx, commit, "")
	})
}

// withCommitWriter calls cb with an unordered writer. All data written to cb is added to the commit, or an error is returned.
func (d *driver) withCommitUnorderedWriter(ctx context.Context, commit *pfs.Commit, cb func(*fileset.UnorderedWriter) error, opts ...fileset.UnorderedWriterOption) (retErr error) {
	return d.storage.WithRenewer(ctx, defaultTTL, func(ctx context.Context, renewer *renew.StringSet) error {
		id, err := d.withUnorderedWriter(ctx, renewer, false, cb, opts...)
		if err != nil {
			return err
		}
		return d.commitStore.AddFileset(ctx, commit, *id)
	})
}

func (d *driver) withUnorderedWriter(ctx context.Context, renewer *renew.StringSet, compact bool, cb func(*fileset.UnorderedWriter) error, opts ...fileset.UnorderedWriterOption) (*fileset.ID, error) {
	opts = append([]fileset.UnorderedWriterOption{fileset.WithRenewal(defaultTTL, renewer)}, opts...)
	uw, err := d.storage.NewUnorderedWriter(ctx, opts...)
	if err != nil {
		return nil, err
	}
	if err := cb(uw); err != nil {
		return nil, err
	}
	id, err := uw.Close()
	if err != nil {
		return nil, err
	}
	if !compact {
		renewer.Add(id.HexString())
		return id, nil
	}
	compactedID, err := d.storage.Compact(ctx, []fileset.ID{*id}, defaultTTL)
	if err != nil {
		return nil, err
	}
	renewer.Add(compactedID.HexString())
	return compactedID, nil
}

func (d *driver) openCommit(ctx context.Context, commit *pfs.Commit, opts ...index.Option) (*pfs.CommitInfo, fileset.FileSet, error) {
	if commit.Branch.Repo.Name == fileSetsRepo {
		fsid, err := fileset.ParseID(commit.ID)
		if err != nil {
			return nil, nil, err
		}
		fs, err := d.storage.Open(ctx, []fileset.ID{*fsid}, opts...)
		if err != nil {
			return nil, nil, err
		}
		return &pfs.CommitInfo{Commit: commit}, fs, nil
	}
	if err := d.env.AuthServer().CheckRepoIsAuthorized(ctx, commit.Branch.Repo.Name, auth.Permission_REPO_READ); err != nil {
		return nil, nil, err
	}
	commitInfo, err := d.inspectCommit(ctx, commit, pfs.CommitState_STARTED)
	if err != nil {
		return nil, nil, err
	}
	id, err := d.getFileset(ctx, commitInfo.Commit)
	if err != nil {
		return nil, nil, err
	}
	fs, err := d.storage.Open(ctx, []fileset.ID{*id}, opts...)
	if err != nil {
		return nil, nil, err
	}
	return commitInfo, fs, nil
}

func (d *driver) copyFile(ctx context.Context, uw *fileset.UnorderedWriter, dst string, src *pfs.File, appendFile bool, tag string) (retErr error) {
	srcCommitInfo, err := d.inspectCommit(ctx, src.Commit, pfs.CommitState_STARTED)
	if err != nil {
		return err
	}
	srcCommit := srcCommitInfo.Commit
	srcPath := cleanPath(src.Path)
	dstPath := cleanPath(dst)
	pathTransform := func(x string) string {
		relPath, err := filepath.Rel(srcPath, x)
		if err != nil {
			panic("cannot apply path transform")
		}
		return path.Join(dstPath, relPath)
	}
	_, fs, err := d.openCommit(ctx, srcCommit, index.WithPrefix(srcPath), index.WithTag(src.Tag))
	if err != nil {
		return err
	}
	fs = fileset.NewIndexFilter(fs, func(idx *index.Index) bool {
		return idx.Path == srcPath || strings.HasPrefix(idx.Path, srcPath+"/")
	})
	fs = fileset.NewIndexMapper(fs, func(idx *index.Index) *index.Index {
		idx2 := *idx
		idx2.Path = pathTransform(idx2.Path)
		return &idx2
	})
	return uw.Copy(ctx, fs, tag, appendFile)
}

func (d *driver) getFile(ctx context.Context, file *pfs.File) (Source, error) {
	commit := file.Commit
	glob := cleanPath(file.Path)
	commitInfo, fs, err := d.openCommit(ctx, commit, index.WithPrefix(globLiteralPrefix(glob)), index.WithTag(file.Tag))
	if err != nil {
		return nil, err
	}
	mf, err := globMatchFunction(glob)
	if err != nil {
		return nil, err
	}
	opts := []SourceOption{
		WithFilter(func(fs fileset.FileSet) fileset.FileSet {
			return fileset.NewIndexFilter(fs, func(idx *index.Index) bool {
				return mf(idx.Path)
			}, true)
		}),
	}
	s := NewSource(d.storage, commitInfo, fs, opts...)
	return NewErrOnEmpty(s, &pfsserver.ErrFileNotFound{File: file}), nil
}

func (d *driver) inspectFile(ctx context.Context, file *pfs.File) (*pfs.FileInfo, error) {
	p := cleanPath(file.Path)
	if p == "/" {
		p = ""
	}
	commitInfo, fs, err := d.openCommit(ctx, file.Commit, index.WithPrefix(p), index.WithTag(file.Tag))
	if err != nil {
		return nil, err
	}
	opts := []SourceOption{
		WithFull(),
		WithFilter(func(fs fileset.FileSet) fileset.FileSet {
			return fileset.NewIndexFilter(fs, func(idx *index.Index) bool {
				return idx.Path == p || strings.HasPrefix(idx.Path, p+"/")
			})
		}),
	}
	s := NewSource(d.storage, commitInfo, fs, opts...)
	var ret *pfs.FileInfo
	s = NewErrOnEmpty(s, &pfsserver.ErrFileNotFound{File: file})
	if err := s.Iterate(ctx, func(fi *pfs.FileInfo, f fileset.File) error {
		p2 := fi.File.Path
		if p2 == p || p2 == p+"/" {
			ret = fi
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return ret, nil
}

func (d *driver) listFile(ctx context.Context, file *pfs.File, full bool, cb func(*pfs.FileInfo) error) error {
	name := cleanPath(file.Path)
	commitInfo, fs, err := d.openCommit(ctx, file.Commit, index.WithPrefix(name), index.WithTag(file.Tag))
	if err != nil {
		return err
	}
	opts := []SourceOption{
		WithFull(),
		WithFilter(func(fs fileset.FileSet) fileset.FileSet {
			return fileset.NewIndexFilter(fs, func(idx *index.Index) bool {
				// Check for directory match (don't return directory in list)
				if idx.Path == fileset.Clean(name, true) {
					return false
				}
				// Check for file match.
				if idx.Path == name {
					return true
				}
				// Check for sub directory / file match.
				return strings.HasPrefix(idx.Path, fileset.Clean(name, true))
			})
		}),
	}
	s := NewSource(d.storage, commitInfo, fs, opts...)
	return s.Iterate(ctx, func(fi *pfs.FileInfo, _ fileset.File) error {
		if pathIsChild(name, cleanPath(fi.File.Path)) {
			return cb(fi)
		}
		return nil
	})
}

func (d *driver) walkFile(ctx context.Context, file *pfs.File, cb func(*pfs.FileInfo) error) (retErr error) {
	p := cleanPath(file.Path)
	if p == "/" {
		p = ""
	}
	commitInfo, fs, err := d.openCommit(ctx, file.Commit, index.WithPrefix(p), index.WithTag(file.Tag))
	if err != nil {
		return err
	}
	opts := []SourceOption{
		WithFilter(func(fs fileset.FileSet) fileset.FileSet {
			return fileset.NewIndexFilter(fs, func(idx *index.Index) bool {
				return idx.Path == p || strings.HasPrefix(idx.Path, p+"/")
			})
		}),
	}
	s := NewSource(d.storage, commitInfo, fs, opts...)
	s = NewErrOnEmpty(s, &pfsserver.ErrFileNotFound{File: file})
	return s.Iterate(ctx, func(fi *pfs.FileInfo, f fileset.File) error {
		return cb(fi)
	})
}

func (d *driver) globFile(ctx context.Context, commit *pfs.Commit, glob string, cb func(*pfs.FileInfo) error) error {
	glob = cleanPath(glob)
	commitInfo, fs, err := d.openCommit(ctx, commit, index.WithPrefix(globLiteralPrefix(glob)))
	if err != nil {
		return err
	}
	mf, err := globMatchFunction(glob)
	if err != nil {
		return err
	}
	opts := []SourceOption{
		WithFull(),
		WithFilter(func(fs fileset.FileSet) fileset.FileSet {
			return fileset.NewIndexFilter(fs, func(idx *index.Index) bool {
				return mf(idx.Path)
			}, true)
		}),
	}
	s := NewSource(d.storage, commitInfo, fs, opts...)
	return s.Iterate(ctx, func(fi *pfs.FileInfo, _ fileset.File) error {
		if mf(fi.File.Path) {
			return cb(fi)
		}
		return nil
	})
}

func (d *driver) diffFile(ctx context.Context, oldFile, newFile *pfs.File, cb func(oldFi, newFi *pfs.FileInfo) error) error {
	// TODO: move validation to the Validating API Server
	// Validation
	if newFile == nil {
		return errors.New("file cannot be nil")
	}
	if newFile.Commit == nil {
		return errors.New("file commit cannot be nil")
	}
	if newFile.Commit.Branch == nil {
		return errors.New("file commit branch cannot be nil")
	}
	if newFile.Commit.Branch.Repo == nil {
		return errors.New("file commit repo cannot be nil")
	}
	// Do READER authorization check for both newFile and oldFile
	if oldFile != nil && oldFile.Commit != nil {
		if err := d.env.AuthServer().CheckRepoIsAuthorized(ctx, oldFile.Commit.Branch.Repo.Name, auth.Permission_REPO_READ); err != nil {
			return err
		}
	}
	if newFile != nil && newFile.Commit != nil {
		if err := d.env.AuthServer().CheckRepoIsAuthorized(ctx, newFile.Commit.Branch.Repo.Name, auth.Permission_REPO_READ); err != nil {
			return err
		}
	}
	newCommitInfo, err := d.inspectCommit(ctx, newFile.Commit, pfs.CommitState_STARTED)
	if err != nil {
		return err
	}
	if oldFile == nil {
		oldFile = &pfs.File{
			Commit: newCommitInfo.ParentCommit,
			Path:   newFile.Path,
		}
	}
	oldCommit := oldFile.Commit
	newCommit := newFile.Commit
	oldName := cleanPath(oldFile.Path)
	if oldName == "/" {
		oldName = ""
	}
	newName := cleanPath(newFile.Path)
	if newName == "/" {
		newName = ""
	}
	var old Source = emptySource{}
	if oldCommit != nil {
		oldCommitInfo, fs, err := d.openCommit(ctx, oldCommit, index.WithPrefix(oldName), index.WithTag(oldFile.Tag))
		if err != nil {
			return err
		}
		opts := []SourceOption{
			WithFull(),
			WithFilter(func(fs fileset.FileSet) fileset.FileSet {
				return fileset.NewIndexFilter(fs, func(idx *index.Index) bool {
					return idx.Path == oldName || strings.HasPrefix(idx.Path, oldName+"/")
				})
			}),
		}
		old = NewSource(d.storage, oldCommitInfo, fs, opts...)
	}
	newCommitInfo, fs, err := d.openCommit(ctx, newCommit, index.WithPrefix(newName), index.WithTag(newFile.Tag))
	if err != nil {
		return err
	}
	opts := []SourceOption{
		WithFull(),
		WithFilter(func(fs fileset.FileSet) fileset.FileSet {
			return fileset.NewIndexFilter(fs, func(idx *index.Index) bool {
				return idx.Path == newName || strings.HasPrefix(idx.Path, newName+"/")
			})
		}),
	}
	new := NewSource(d.storage, newCommitInfo, fs, opts...)
	diff := NewDiffer(old, new)
	return diff.Iterate(ctx, cb)
}

// createFileset creates a new temporary fileset and returns it.
func (d *driver) createFileset(ctx context.Context, cb func(*fileset.UnorderedWriter) error) (*fileset.ID, error) {
	var id *fileset.ID
	if err := d.storage.WithRenewer(ctx, defaultTTL, func(ctx context.Context, renewer *renew.StringSet) error {
		var err error
		id, err = d.withUnorderedWriter(ctx, renewer, false, cb)
		return err
	}); err != nil {
		return nil, err
	}
	return id, nil
}

func (d *driver) renewFileset(ctx context.Context, id fileset.ID, ttl time.Duration) error {
	if ttl < time.Second {
		return errors.Errorf("ttl (%d) must be at least one second", ttl)
	}
	if ttl > maxTTL {
		return errors.Errorf("ttl (%d) exceeds max ttl (%d)", ttl, maxTTL)
	}
	_, err := d.storage.SetTTL(ctx, id, ttl)
	return err
}

func (d *driver) addFileset(txnCtx *txncontext.TransactionContext, commit *pfs.Commit, filesetID fileset.ID) error {
	commitInfo, err := d.resolveCommit(txnCtx.SqlTx, commit)
	if err != nil {
		return err
	}
	if commitInfo.Finished != nil {
		return pfsserver.ErrCommitFinished{commitInfo.Commit}
	}
	return d.commitStore.AddFilesetTx(txnCtx.SqlTx, commitInfo.Commit, filesetID)
}

func (d *driver) getFileset(ctx context.Context, commit *pfs.Commit) (*fileset.ID, error) {
	commitInfo, err := d.getCommit(ctx, commit)
	if err != nil {
		return nil, err
	}
	if commitInfo.Finished != nil {
		return d.getOrComputeTotal(ctx, commitInfo.Commit)
	}
	var ids []fileset.ID
	if commitInfo.ParentCommit != nil {
		// ¯\_(ツ)_/¯
		parentId, err := d.getFileset(ctx, commitInfo.ParentCommit)
		if err != nil {
			return nil, err
		}
		ids = append(ids, *parentId)
	}
	id, err := d.commitStore.GetDiffFileset(ctx, commitInfo.Commit)
	if err != nil {
		return nil, err
	}
	ids = append(ids, *id)
	return d.storage.Compose(ctx, ids, defaultTTL)
}

func (d *driver) getOrComputeTotal(ctx context.Context, commit *pfs.Commit) (*fileset.ID, error) {
	commitInfo, err := d.getCommit(ctx, commit)
	if err != nil {
		return nil, err
	}
	if commitInfo.Finished == nil {
		return nil, errors.Errorf("attempted to compute total of unfinished commit")
	}
	commit = commitInfo.Commit
	id, err := d.commitStore.GetTotalFileset(ctx, commit)
	if err != nil && err != errNoTotalFileset {
		return nil, err
	}
	if err == nil {
		return id, nil
	}
	id, err = d.commitStore.GetDiffFileset(ctx, commit)
	if err != nil {
		return nil, err
	}
	var inputs []fileset.ID
	if commitInfo.ParentCommit != nil {
		parentDiff, err := d.getOrComputeTotal(ctx, commitInfo.ParentCommit)
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, *parentDiff)
	}
	inputs = append(inputs, *id)
	output, err := d.compactor.Compact(ctx, inputs, defaultTTL)
	if err != nil {
		return nil, err
	}
	if err := d.commitStore.SetTotalFileset(ctx, commit, *output); err != nil {
		return nil, err
	}
	return d.commitStore.GetTotalFileset(ctx, commit)
}

// sizeOfCommit gets the size of a commit.
func (d *driver) sizeOfCommit(ctx context.Context, commit *pfs.Commit) (int64, error) {
	fsid, err := d.getFileset(ctx, commit)
	if err != nil {
		return 0, err
	}
	return d.storage.SizeOf(ctx, *fsid)
}
