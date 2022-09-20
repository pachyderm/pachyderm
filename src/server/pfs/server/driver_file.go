package server

import (
	"context"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pacherr"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsfile"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	log "github.com/sirupsen/logrus"
)

func (d *driver) modifyFile(ctx context.Context, commit *pfs.Commit, cb func(*fileset.UnorderedWriter) error) error {
	return d.storage.WithRenewer(ctx, defaultTTL, func(ctx context.Context, renewer *fileset.Renewer) error {
		// Store the originally-requested parameters because they will be overwritten by inspectCommit
		branch := proto.Clone(commit.Branch).(*pfs.Branch)
		commitID := commit.ID
		if branch.Name == "" && !uuid.IsUUIDWithoutDashes(commitID) {
			branch.Name = commitID
			commitID = ""
		}
		commitInfo, err := d.inspectCommit(ctx, commit, pfs.CommitState_STARTED)
		if err != nil {
			if !errutil.IsNotFoundError(err) || branch.Name == "" {
				return err
			}
			return d.oneOffModifyFile(ctx, renewer, branch, cb)
		}
		if commitInfo.Finishing != nil {
			// The commit is already finished - if the commit was explicitly specified,
			// error out, otherwise we can make a child commit since this is the branch head.
			if commitID != "" {
				return pfsserver.ErrCommitFinished{Commit: commitInfo.Commit}
			}
			return d.oneOffModifyFile(ctx, renewer, branch, cb, fileset.WithParentID(func() (*fileset.ID, error) {
				parentID, err := d.getFileSet(ctx, commitInfo.Commit)
				if err != nil {
					return nil, err
				}
				if err := renewer.Add(ctx, *parentID); err != nil {
					return nil, err
				}
				return parentID, nil
			}))
		}
		return d.withCommitUnorderedWriter(ctx, renewer, commitInfo.Commit, cb)
	})
}

func (d *driver) oneOffModifyFile(ctx context.Context, renewer *fileset.Renewer, branch *pfs.Branch, cb func(*fileset.UnorderedWriter) error, opts ...fileset.UnorderedWriterOption) error {
	id, err := d.withUnorderedWriter(ctx, renewer, cb, opts...)
	if err != nil {
		return err
	}
	return d.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		commit, err := d.startCommit(txnCtx, nil, branch, "")
		if err != nil {
			return err
		}
		if err := d.commitStore.AddFileSetTx(txnCtx.SqlTx, commit, *id); err != nil {
			return errors.EnsureStack(err)
		}
		return d.finishCommit(txnCtx, commit, "", "", false)
	})
}

// withCommitWriter calls cb with an unordered writer. All data written to cb is added to the commit, or an error is returned.
func (d *driver) withCommitUnorderedWriter(ctx context.Context, renewer *fileset.Renewer, commit *pfs.Commit, cb func(*fileset.UnorderedWriter) error) error {
	id, err := d.withUnorderedWriter(ctx, renewer, cb, fileset.WithParentID(func() (*fileset.ID, error) {
		parentID, err := d.getFileSet(ctx, commit)
		if err != nil {
			return nil, err
		}
		if err := renewer.Add(ctx, *parentID); err != nil {
			return nil, err
		}
		return parentID, nil
	}))
	if err != nil {
		return err
	}
	return errors.EnsureStack(d.commitStore.AddFileSet(ctx, commit, *id))
}

func (d *driver) withUnorderedWriter(ctx context.Context, renewer *fileset.Renewer, cb func(*fileset.UnorderedWriter) error, opts ...fileset.UnorderedWriterOption) (*fileset.ID, error) {
	opts = append([]fileset.UnorderedWriterOption{fileset.WithRenewal(defaultTTL, renewer), fileset.WithValidator(validate)}, opts...)
	uw, err := d.storage.NewUnorderedWriter(ctx, opts...)
	if err != nil {
		return nil, err
	}
	if err := cb(uw); err != nil {
		return nil, err
	}
	id, err := uw.Close(ctx)
	if err != nil {
		return nil, err
	}
	if err := renewer.Add(ctx, *id); err != nil {
		return nil, err
	}
	return id, nil
}

func (d *driver) openCommit(ctx context.Context, commit *pfs.Commit) (*pfs.CommitInfo, fileset.FileSet, error) {
	if commit.Branch.Repo.Name == fileSetsRepo {
		fsid, err := fileset.ParseID(commit.ID)
		if err != nil {
			return nil, nil, err
		}
		fs, err := d.storage.Open(ctx, []fileset.ID{*fsid})
		if err != nil {
			return nil, nil, err
		}
		return &pfs.CommitInfo{Commit: commit}, fs, nil
	}
	if err := d.env.AuthServer.CheckRepoIsAuthorized(ctx, commit.Branch.Repo, auth.Permission_REPO_READ); err != nil {
		return nil, nil, errors.EnsureStack(err)
	}
	commitInfo, err := d.inspectCommit(ctx, commit, pfs.CommitState_STARTED)
	if err != nil {
		return nil, nil, err
	}
	if commitInfo.Finishing != nil && commitInfo.Finished == nil {
		_, err := d.inspectCommit(ctx, commit, pfs.CommitState_FINISHED)
		if err != nil {
			return nil, nil, err
		}
	}
	id, err := d.getFileSet(ctx, commitInfo.Commit)
	if err != nil {
		return nil, nil, err
	}
	fs, err := d.storage.Open(ctx, []fileset.ID{*id})
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
	srcPath := pfsfile.CleanPath(src.Path)
	dstPath := pfsfile.CleanPath(dst)
	pathTransform := func(x string) string {
		relPath, err := filepath.Rel(srcPath, x)
		if err != nil {
			panic("cannot apply path transform")
		}
		return path.Join(dstPath, relPath)
	}
	_, fs, err := d.openCommit(ctx, srcCommit)
	if err != nil {
		return err
	}
	fs = fileset.NewIndexFilter(fs, func(idx *index.Index) bool {
		return idx.Path == srcPath || strings.HasPrefix(idx.Path, fileset.Clean(srcPath, true))
	})
	fs = fileset.NewIndexMapper(fs, func(idx *index.Index) *index.Index {
		idx2 := *idx
		idx2.Path = pathTransform(idx2.Path)
		return &idx2
	})
	return uw.Copy(ctx, fs, tag, appendFile, index.WithPrefix(srcPath), index.WithDatum(src.Datum))
}

func (d *driver) getFile(ctx context.Context, file *pfs.File, pathRange *pfs.PathRange) (Source, error) {
	commitInfo, fs, err := d.openCommit(ctx, file.Commit)
	if err != nil {
		return nil, err
	}
	glob := pfsfile.CleanPath(file.Path)
	opts := []SourceOption{
		WithPrefix(globLiteralPrefix(glob)),
		WithDatum(file.Datum),
	}
	if pathRange != nil {
		opts = append(opts, WithPathRange(pathRange))
	}
	mf, err := globMatchFunction(glob)
	if err != nil {
		return nil, err
	}
	opts = append(opts, WithFilter(func(fs fileset.FileSet) fileset.FileSet {
		fs = fileset.NewIndexFilter(fs, func(idx *index.Index) bool {
			return mf(idx.Path)
		}, true)
		return fileset.NewPrefetcher(d.storage, fs)
	}))
	s := NewSource(commitInfo, fs, opts...)
	return NewErrOnEmpty(s, &pfsserver.ErrFileNotFound{File: file}), nil
}

func (d *driver) inspectFile(ctx context.Context, file *pfs.File) (*pfs.FileInfo, error) {
	commitInfo, fs, err := d.openCommit(ctx, file.Commit)
	if err != nil {
		return nil, err
	}
	p := pfsfile.CleanPath(file.Path)
	if p == "/" {
		p = ""
	}
	opts := []SourceOption{
		WithPrefix(p),
		WithDatum(file.Datum),
		WithFilter(func(fs fileset.FileSet) fileset.FileSet {
			return fileset.NewIndexFilter(fs, func(idx *index.Index) bool {
				return idx.Path == p || strings.HasPrefix(idx.Path, p+"/")
			})
		}),
	}
	s := NewSource(commitInfo, fs, opts...)
	var ret *pfs.FileInfo
	s = NewErrOnEmpty(s, &pfsserver.ErrFileNotFound{File: file})
	if err := s.Iterate(ctx, func(fi *pfs.FileInfo, f fileset.File) error {
		p2 := fi.File.Path
		if p2 == p || p2 == p+"/" {
			ret = fi
		}
		return nil
	}); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return ret, nil
}

func (d *driver) listFile(ctx context.Context, file *pfs.File, cb func(*pfs.FileInfo) error) error {
	commitInfo, fs, err := d.openCommit(ctx, file.Commit)
	if err != nil {
		return err
	}
	name := pfsfile.CleanPath(file.Path)
	opts := []SourceOption{
		WithPrefix(name),
		WithDatum(file.Datum),
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
	s := NewSource(commitInfo, fs, opts...)
	err = s.Iterate(ctx, func(fi *pfs.FileInfo, _ fileset.File) error {
		if pathIsChild(name, pfsfile.CleanPath(fi.File.Path)) {
			return cb(fi)
		}
		return nil
	})
	return errors.EnsureStack(err)
}

func (d *driver) walkFile(ctx context.Context, file *pfs.File, cb func(*pfs.FileInfo) error) (retErr error) {
	commitInfo, fs, err := d.openCommit(ctx, file.Commit)
	if err != nil {
		return err
	}
	p := pfsfile.CleanPath(file.Path)
	if p == "/" {
		p = ""
	}
	opts := []SourceOption{
		WithPrefix(p),
		WithDatum(file.Datum),
		WithFilter(func(fs fileset.FileSet) fileset.FileSet {
			return fileset.NewIndexFilter(fs, func(idx *index.Index) bool {
				return idx.Path == p || strings.HasPrefix(idx.Path, p+"/")
			})
		}),
	}
	s := NewSource(commitInfo, fs, opts...)
	s = NewErrOnEmpty(s, newFileNotFound(commitInfo.Commit.ID, p))
	err = s.Iterate(ctx, func(fi *pfs.FileInfo, f fileset.File) error {
		return cb(fi)
	})
	if p == "" && pacherr.IsNotExist(err) {
		err = nil
	}
	return err
}

func (d *driver) globFile(ctx context.Context, commit *pfs.Commit, glob string, pathRange *pfs.PathRange, cb func(*pfs.FileInfo) error) error {
	commitInfo, fs, err := d.openCommit(ctx, commit)
	if err != nil {
		return err
	}
	glob = pfsfile.CleanPath(glob)
	opts := []SourceOption{
		WithPrefix(globLiteralPrefix(glob)),
	}
	if pathRange != nil {
		opts = append(opts, WithPathRange(pathRange))
	}
	mf, err := globMatchFunction(glob)
	if err != nil {
		return err
	}
	opts = append(opts, WithFilter(func(fs fileset.FileSet) fileset.FileSet {
		return fileset.NewIndexFilter(fs, func(idx *index.Index) bool {
			return mf(idx.Path)
		}, true)
	}))
	s := NewSource(commitInfo, fs, opts...)
	err = s.Iterate(ctx, func(fi *pfs.FileInfo, _ fileset.File) error {
		if mf(fi.File.Path) {
			return cb(fi)
		}
		return nil
	})
	return errors.EnsureStack(err)
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
		if err := d.env.AuthServer.CheckRepoIsAuthorized(ctx, oldFile.Commit.Branch.Repo, auth.Permission_REPO_READ); err != nil {
			return errors.EnsureStack(err)
		}
	}
	if newFile != nil && newFile.Commit != nil {
		if err := d.env.AuthServer.CheckRepoIsAuthorized(ctx, newFile.Commit.Branch.Repo, auth.Permission_REPO_READ); err != nil {
			return errors.EnsureStack(err)
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
	oldName := pfsfile.CleanPath(oldFile.Path)
	if oldName == "/" {
		oldName = ""
	}
	newName := pfsfile.CleanPath(newFile.Path)
	if newName == "/" {
		newName = ""
	}
	var old Source = emptySource{}
	if oldCommit != nil {
		oldCommitInfo, fs, err := d.openCommit(ctx, oldCommit)
		if err != nil {
			return err
		}
		opts := []SourceOption{
			WithPrefix(oldName),
			WithDatum(oldFile.Datum),
			WithFilter(func(fs fileset.FileSet) fileset.FileSet {
				return fileset.NewIndexFilter(fs, func(idx *index.Index) bool {
					return idx.Path == oldName || strings.HasPrefix(idx.Path, oldName+"/")
				})
			}),
		}
		old = NewSource(oldCommitInfo, fs, opts...)
	}
	newCommitInfo, fs, err := d.openCommit(ctx, newCommit)
	if err != nil {
		return err
	}
	opts := []SourceOption{
		WithPrefix(newName),
		WithDatum(newFile.Datum),
		WithFilter(func(fs fileset.FileSet) fileset.FileSet {
			return fileset.NewIndexFilter(fs, func(idx *index.Index) bool {
				return idx.Path == newName || strings.HasPrefix(idx.Path, newName+"/")
			})
		}),
	}
	new := NewSource(newCommitInfo, fs, opts...)
	diff := NewDiffer(old, new)
	return diff.Iterate(ctx, cb)
}

// createFileSet creates a new temporary fileset and returns it.
func (d *driver) createFileSet(ctx context.Context, cb func(*fileset.UnorderedWriter) error) (*fileset.ID, error) {
	var id *fileset.ID
	if err := d.storage.WithRenewer(ctx, defaultTTL, func(ctx context.Context, renewer *fileset.Renewer) error {
		var err error
		id, err = d.withUnorderedWriter(ctx, renewer, cb, fileset.WithCompact(d.env.StorageConfig.StorageCompactionMaxFanIn))
		return err
	}); err != nil {
		return nil, err
	}
	return id, nil
}

func (d *driver) getFileSet(ctx context.Context, commit *pfs.Commit) (*fileset.ID, error) {
	commitInfo, err := d.getCommit(ctx, commit)
	if err != nil {
		return nil, err
	}
	// Get the total file set if the commit has been finished.
	if commitInfo.Finished != nil && commitInfo.Error == "" {
		id, err := d.commitStore.GetTotalFileSet(ctx, commitInfo.Commit)
		if err != nil {
			// TODO: Need to handle this differently if we want to delete total
			// file sets after a commit is finished (to save space for old commits).
			if errors.Is(err, errNoTotalFileSet) {
				return d.storage.Compose(ctx, nil, defaultTTL)
			}
			return nil, errors.EnsureStack(err)
		}
		return id, nil
	}
	// Compose the base file set with the diffs.
	var ids []fileset.ID
	baseCommit := commitInfo.ParentCommit
	for baseCommit != nil {
		baseCommitInfo, err := d.getCommit(ctx, baseCommit)
		if err != nil {
			return nil, err
		}
		if baseCommitInfo.Error == "" {
			if baseCommitInfo.Finished == nil {
				return nil, pfsserver.ErrBaseCommitNotFinished{
					BaseCommit: baseCommit,
					Commit:     commit,
				}
			}
			// ¯\_(ツ)_/¯
			baseId, err := d.getFileSet(ctx, baseCommit)
			if err != nil {
				return nil, err
			}
			ids = append(ids, *baseId)
			break
		}
		baseCommit = baseCommitInfo.ParentCommit
	}
	id, err := d.commitStore.GetDiffFileSet(ctx, commitInfo.Commit)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	ids = append(ids, *id)
	return d.storage.Compose(ctx, ids, defaultTTL)
}

func (d *driver) shardFileSet(ctx context.Context, fsid fileset.ID) ([]*pfs.PathRange, error) {
	fs, err := d.storage.Open(ctx, []fileset.ID{fsid})
	if err != nil {
		return nil, err
	}
	shards, err := fs.Shards(ctx)
	if err != nil {
		return nil, err
	}
	var pathRanges []*pfs.PathRange
	for _, shard := range shards {
		pathRanges = append(pathRanges, &pfs.PathRange{
			Lower: shard.Lower,
			Upper: shard.Upper,
		})
	}
	return pathRanges, nil
}

func (d *driver) addFileSet(txnCtx *txncontext.TransactionContext, commit *pfs.Commit, filesetID fileset.ID) error {
	commitInfo, err := d.resolveCommit(txnCtx.SqlTx, commit)
	if err != nil {
		return err
	}
	// TODO: This check needs to be in the add transaction.
	if commitInfo.Finishing != nil {
		return pfsserver.ErrCommitFinished{Commit: commitInfo.Commit}
	}
	return errors.EnsureStack(d.commitStore.AddFileSetTx(txnCtx.SqlTx, commitInfo.Commit, filesetID))
}

func (d *driver) renewFileSet(ctx context.Context, id fileset.ID, ttl time.Duration) error {
	if ttl < time.Second {
		return errors.Errorf("ttl (%d) must be at least one second", ttl)
	}
	if ttl > maxTTL {
		return errors.Errorf("ttl (%d) exceeds max ttl (%d)", ttl, maxTTL)
	}
	_, err := d.storage.SetTTL(ctx, id, ttl)
	return err
}

func (d *driver) composeFileSet(ctx context.Context, ids []fileset.ID, ttl time.Duration, compact bool) (*fileset.ID, error) {
	if compact {
		compactor := newCompactor(d.storage, log.NewEntry(log.StandardLogger()), d.env.StorageConfig.StorageCompactionMaxFanIn)
		taskDoer := d.env.TaskService.NewDoer(storageTaskNamespace, uuid.NewWithoutDashes(), nil)
		return compactor.Compact(ctx, taskDoer, ids, ttl)
	}
	return d.storage.Compose(ctx, ids, ttl)
}

func (d *driver) commitSizeUpperBound(ctx context.Context, commit *pfs.Commit) (int64, error) {
	fsid, err := d.getFileSet(ctx, commit)
	if err != nil {
		return 0, err
	}
	return d.storage.SizeUpperBound(ctx, *fsid)
}

func newFileNotFound(commitID string, path string) pacherr.ErrNotExist {
	return pacherr.ErrNotExist{
		Collection: "commit/" + commitID,
		ID:         path,
	}
}
