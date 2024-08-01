package driver

import (
	"context"
	"math"
	"path"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pacherr"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsfile"
	"github.com/pachyderm/pachyderm/v2/src/internal/protoutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	"github.com/pachyderm/pachyderm/v2/src/server/pfs/server"
)

func (d *Driver) getCompactedDiffFileSet(ctx context.Context, commit *pfsdb.Commit) (*fileset.ID, error) {
	diff, err := d.CommitStore.GetDiffFileSet(ctx, commit)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	isCompacted, err := d.Storage.Filesets.IsCompacted(ctx, *diff)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	if isCompacted {
		return diff, nil
	}
	if err := d.Storage.Filesets.WithRenewer(ctx, server.DefaultTTL, func(ctx context.Context, renewer *fileset.Renewer) error {
		compactor := server.NewCompactor(d.Storage.Filesets, d.Env.StorageConfig.StorageCompactionMaxFanIn)
		taskDoer := d.Env.TaskService.NewDoer(server.StorageTaskNamespace, commit.Commit.Id, nil)
		diff, err = d.CompactDiffFileset(ctx, compactor, taskDoer, renewer, commit)
		return err
	}); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return diff, nil
}

// TODO(acohen4): signature should accept a branch seperate from the commit
func (d *Driver) ModifyFile(ctx context.Context, commitHandle *pfs.Commit, cb func(*fileset.UnorderedWriter) error) error {
	return d.Storage.Filesets.WithRenewer(ctx, server.DefaultTTL, func(ctx context.Context, renewer *fileset.Renewer) error {
		// Store the originally-requested parameters because they will be overwritten by InspectCommitInfo
		branch := proto.Clone(commitHandle.Branch).(*pfs.Branch)
		commitID := commitHandle.Id
		if branch != nil && branch.Name == "" && !uuid.IsUUIDWithoutDashes(commitID) {
			branch.Name = commitID
			commitID = ""
		}
		commit, err := d.inspectCommit(ctx, commitHandle, pfs.CommitState_STARTED)
		if err != nil {
			if !errutil.IsNotFoundError(err) || branch == nil || branch.Name == "" {
				return err
			}
			return d.oneOffModifyFile(ctx, renewer, branch, cb)
		}
		if commit.Finishing != nil {
			// The commit is already finished - if the commit was explicitly specified,
			// error out, otherwise we can make a child commit since this is the branch head.
			if commitID != "" {
				return pfsserver.ErrCommitFinished{Commit: commit.Commit}
			}
			return d.oneOffModifyFile(ctx, renewer, branch, cb, fileset.WithParentID(func() (*fileset.ID, error) {
				parentID, err := d.GetFileset(ctx, commit)
				if err != nil {
					return nil, err
				}
				if err := renewer.Add(ctx, *parentID); err != nil {
					return nil, err
				}
				return parentID, nil
			}))
		}
		return d.withCommitUnorderedWriter(ctx, renewer, commit, cb)
	})
}

func (d *Driver) oneOffModifyFile(ctx context.Context, renewer *fileset.Renewer, branch *pfs.Branch, cb func(*fileset.UnorderedWriter) error, opts ...fileset.UnorderedWriterOption) error {
	id, err := WithUnorderedWriter(ctx, d.Storage, renewer, cb, opts...)
	if err != nil {
		return err
	}
	return d.TxnEnv.WithWriteContext(ctx, func(ctx context.Context, txnCtx *txncontext.TransactionContext) error {
		commit, err := d.StartCommit(ctx, txnCtx, nil, branch, "")
		if err != nil {
			return err
		}
		if err := d.CommitStore.AddFileSetTx(txnCtx.SqlTx, commit, *id); err != nil {
			return errors.EnsureStack(err)
		}
		return d.FinishCommit(ctx, txnCtx, commit, "", "", false)
	})
}

// withCommitWriter calls cb with an unordered writer. All data written to cb is added to the commit, or an error is returned.
func (d *Driver) withCommitUnorderedWriter(ctx context.Context, renewer *fileset.Renewer, commit *pfsdb.Commit, cb func(*fileset.UnorderedWriter) error) error {
	id, err := WithUnorderedWriter(ctx, d.Storage, renewer, cb, fileset.WithParentID(func() (*fileset.ID, error) {
		parentID, err := d.GetFileset(ctx, commit)
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
	return errors.Wrap(d.CommitStore.AddFileSet(ctx, commit, *id), "with commit unordered writer")
}

func WithUnorderedWriter(ctx context.Context, storage *storage.Server, renewer *fileset.Renewer, cb func(*fileset.UnorderedWriter) error, opts ...fileset.UnorderedWriterOption) (*fileset.ID, error) {
	opts = append([]fileset.UnorderedWriterOption{fileset.WithRenewal(server.DefaultTTL, renewer), fileset.WithValidator(server.ValidateFilename)}, opts...)
	uw, err := storage.Filesets.NewUnorderedWriter(ctx, opts...)
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

func (d *Driver) CopyFile(ctx context.Context, uw *fileset.UnorderedWriter, dst string, src *pfs.File, appendFile bool, tag string) (retErr error) {
	srcCommitInfo, err := d.InspectCommitInfo(ctx, src.Commit, pfs.CommitState_STARTED)
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
		idx2 := protoutil.Clone(idx)
		idx2.Path = pathTransform(idx2.Path)
		return idx2
	})
	return uw.Copy(ctx, fs, tag, appendFile, index.WithPrefix(srcPath), index.WithDatum(src.Datum))
}

func (d *Driver) GetFile(ctx context.Context, file *pfs.File, pathRange *pfs.PathRange) (server.Source, error) {
	commitInfo, fs, err := d.openCommit(ctx, file.Commit)
	if err != nil {
		return nil, err
	}
	glob := pfsfile.CleanPath(file.Path)
	opts := []server.SourceOption{
		server.WithPrefix(storage.GlobLiteralPrefix(glob)),
		server.WithDatum(file.Datum),
	}
	var upper string
	if pathRange != nil {
		opts = append(opts, server.WithPathRange(pathRange))
		upper = pathRange.Upper
	}
	mf, err := server.GlobMatchFunction(glob)
	if err != nil {
		return nil, err
	}
	opts = append(opts, server.WithFilter(func(fs fileset.FileSet) fileset.FileSet {
		fs = fileset.NewIndexFilter(fs, func(idx *index.Index) bool {
			return mf(idx.Path)
		}, true)
		return fileset.NewPrefetcher(d.Storage.Filesets, fs, upper)
	}))
	s := server.NewSource(commitInfo, fs, opts...)
	return server.NewErrOnEmpty(s, &pfsserver.ErrFileNotFound{File: file}), nil
}

func (d *Driver) InspectFile(ctx context.Context, file *pfs.File) (*pfs.FileInfo, error) {
	commitInfo, fs, err := d.openCommit(ctx, file.Commit)
	if err != nil {
		return nil, err
	}
	p := pfsfile.CleanPath(file.Path)
	if p == "/" {
		p = ""
	}
	opts := []server.SourceOption{
		server.WithPrefix(p),
		server.WithDatum(file.Datum),
		server.WithFilter(func(fs fileset.FileSet) fileset.FileSet {
			return fileset.NewIndexFilter(fs, func(idx *index.Index) bool {
				return idx.Path == p || strings.HasPrefix(idx.Path, p+"/")
			})
		}),
	}
	s := server.NewSource(commitInfo, fs, opts...)
	var ret *pfs.FileInfo
	s = server.NewErrOnEmpty(s, &pfsserver.ErrFileNotFound{File: file})
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

func validatePagination(number int64, reverse bool) error {
	if number == 0 && reverse {
		return errors.Errorf("number must be > 0 when reverse is true")
	}
	if number > 100000 {
		return errors.Errorf("cannot return more than 100000 files at a time")
	}
	return nil
}

func (d *Driver) ListFile(ctx context.Context, file *pfs.File, paginationMarker *pfs.File, number int64, reverse bool, cb func(*pfs.FileInfo) error) error {
	if err := validatePagination(number, reverse); err != nil {
		return err
	}
	commitInfo, fs, err := d.openCommit(ctx, file.Commit)
	if err != nil {
		return err
	}
	name := pfsfile.CleanPath(file.Path)
	opts := []server.SourceOption{
		server.WithPrefix(name),
		server.WithDatum(file.Datum),
		server.WithFilter(func(fs fileset.FileSet) fileset.FileSet {
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
	if paginationMarker != nil {
		pathRange := &pfs.PathRange{}
		if reverse {
			pathRange.Upper = paginationMarker.Path
		} else {
			pathRange.Lower = paginationMarker.Path
		}
		opts = append(opts, server.WithPathRange(pathRange))
	}
	s := server.NewSource(commitInfo, fs, opts...)
	if number == 0 {
		number = math.MaxInt64
	}
	if reverse {
		fis := newCircularList(number)
		if err := s.Iterate(ctx, func(fi *pfs.FileInfo, _ fileset.File) error {
			if isPaginationMarker(paginationMarker, fi) {
				return nil
			}
			if server.PathIsChild(name, pfsfile.CleanPath(fi.File.Path)) {
				fis.add(fi)
			}
			return nil
		}); err != nil {
			return errors.EnsureStack(err)
		}
		return fis.iterateReverse(cb)
	}
	if err = s.Iterate(ctx, func(fi *pfs.FileInfo, _ fileset.File) error {
		if number == 0 {
			return errutil.ErrBreak
		}
		if isPaginationMarker(paginationMarker, fi) {
			return nil
		}
		if server.PathIsChild(name, pfsfile.CleanPath(fi.File.Path)) {
			number--
			return cb(fi)
		}
		return nil
	}); err != nil {
		return errors.EnsureStack(err)
	}
	return errors.EnsureStack(err)
}

// isPaginationMarker returns true if the file info has the same path as the pagination marker
func isPaginationMarker(marker *pfs.File, fi *pfs.FileInfo) bool {
	return marker != nil && pfsfile.CleanPath(marker.Path) == pfsfile.CleanPath(fi.File.Path)
}

type circularList struct {
	// a circular buffer of file infos
	buffer []*pfs.FileInfo
	// next index to be populated
	index int
	// size is the number of file infos in the ring
	size int
}

func newCircularList(size int64) *circularList {
	return &circularList{
		buffer: make([]*pfs.FileInfo, size),
	}
}

func (r *circularList) add(fi *pfs.FileInfo) {
	r.buffer[r.index] = fi
	r.index++
	// if we are at the end of the buffer, wrap around
	if r.index == len(r.buffer) {
		r.index = 0
	}
	if r.size < len(r.buffer) {
		r.size++
	}
}

func (r *circularList) iterateReverse(cb func(*pfs.FileInfo) error) error {
	// last element to be inserted
	idx := r.index

	for i := 0; i < r.size; i++ {
		idx--
		// if we are at the beginning of the buffer, wrap around
		if idx < 0 {
			idx = len(r.buffer) - 1
		}
		if err := cb(r.buffer[idx]); err != nil {
			return err
		}
	}
	return nil
}

func (d *Driver) WalkFile(ctx context.Context, file *pfs.File, paginationMarker *pfs.File, number int64, reverse bool, cb func(*pfs.FileInfo) error) (retErr error) {
	if err := validatePagination(number, reverse); err != nil {
		return err
	}
	commitInfo, fs, err := d.openCommit(ctx, file.Commit)
	if err != nil {
		return err
	}
	p := pfsfile.CleanPath(file.Path)
	if p == "/" {
		p = ""
	}
	opts := []server.SourceOption{
		server.WithPrefix(p),
		server.WithDatum(file.Datum),
		server.WithFilter(func(fs fileset.FileSet) fileset.FileSet {
			return fileset.NewIndexFilter(fs, func(idx *index.Index) bool {
				return idx.Path == p || strings.HasPrefix(idx.Path, p+"/")
			})
		}),
	}
	if paginationMarker != nil {
		pathRange := &pfs.PathRange{}
		if reverse {
			pathRange.Upper = paginationMarker.Path
		} else {
			pathRange.Lower = paginationMarker.Path
		}
		opts = append(opts, server.WithPathRange(pathRange))
	}
	s := server.NewSource(commitInfo, fs, opts...)
	s = server.NewErrOnEmpty(s, newFileNotFound(commitInfo.Commit.Id, p))
	if number == 0 {
		number = math.MaxInt64
	}
	if reverse {
		fis := newCircularList(number)
		if err := s.Iterate(ctx, func(fi *pfs.FileInfo, _ fileset.File) error {
			if isPaginationMarker(paginationMarker, fi) {
				return nil
			}
			fis.add(fi)
			return nil
		}); err != nil {
			return errors.EnsureStack(err)
		}
		return fis.iterateReverse(cb)
	}
	err = s.Iterate(ctx, func(fi *pfs.FileInfo, f fileset.File) error {
		if number == 0 {
			return errutil.ErrBreak
		}
		if isPaginationMarker(paginationMarker, fi) {
			return nil
		}
		number--
		return cb(fi)
	})
	if (p == "" && pacherr.IsNotExist(err)) || errors.Is(err, errutil.ErrBreak) {
		err = nil
	}
	return err
}

func (d *Driver) GlobFile(ctx context.Context, commit *pfs.Commit, glob string, pathRange *pfs.PathRange, cb func(*pfs.FileInfo) error) error {
	commitInfo, fs, err := d.openCommit(ctx, commit)
	if err != nil {
		return err
	}
	glob = pfsfile.CleanPath(glob)
	opts := []server.SourceOption{
		server.WithPrefix(storage.GlobLiteralPrefix(glob)),
	}
	if pathRange != nil {
		opts = append(opts, server.WithPathRange(pathRange))
	}
	mf, err := server.GlobMatchFunction(glob)
	if err != nil {
		return err
	}
	opts = append(opts, server.WithFilter(func(fs fileset.FileSet) fileset.FileSet {
		return fileset.NewIndexFilter(fs, func(idx *index.Index) bool {
			return mf(idx.Path)
		}, true)
	}))
	s := server.NewSource(commitInfo, fs, opts...)
	err = s.Iterate(ctx, func(fi *pfs.FileInfo, _ fileset.File) error {
		if mf(fi.File.Path) {
			return cb(fi)
		}
		return nil
	})
	return errors.EnsureStack(err)
}

func (d *Driver) DiffFile(ctx context.Context, oldFile, newFile *pfs.File, cb func(oldFi, newFi *pfs.FileInfo) error) error {
	// Do READER authorization check for both newFile and oldFile
	if oldFile != nil && oldFile.Commit != nil {
		if err := d.Env.Auth.CheckRepoIsAuthorized(ctx, oldFile.Commit.Repo, auth.Permission_REPO_READ); err != nil {
			return errors.EnsureStack(err)
		}
	}
	if newFile.Commit != nil {
		if err := d.Env.Auth.CheckRepoIsAuthorized(ctx, newFile.Commit.Repo, auth.Permission_REPO_READ); err != nil {
			return errors.EnsureStack(err)
		}
	}
	newCommitInfo, err := d.InspectCommitInfo(ctx, newFile.Commit, pfs.CommitState_STARTED)
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
	var old server.Source = server.EmptySource{}
	if oldCommit != nil {
		oldCommitInfo, fs, err := d.openCommit(ctx, oldCommit)
		if err != nil {
			return err
		}
		opts := []server.SourceOption{
			server.WithPrefix(oldName),
			server.WithDatum(oldFile.Datum),
			server.WithFilter(func(fs fileset.FileSet) fileset.FileSet {
				return fileset.NewIndexFilter(fs, func(idx *index.Index) bool {
					return idx.Path == oldName || strings.HasPrefix(idx.Path, oldName+"/")
				})
			}),
		}
		old = server.NewSource(oldCommitInfo, fs, opts...)
	}
	newCommitInfo, fs, err := d.openCommit(ctx, newCommit)
	if err != nil {
		return err
	}
	opts := []server.SourceOption{
		server.WithPrefix(newName),
		server.WithDatum(newFile.Datum),
		server.WithFilter(func(fs fileset.FileSet) fileset.FileSet {
			return fileset.NewIndexFilter(fs, func(idx *index.Index) bool {
				return idx.Path == newName || strings.HasPrefix(idx.Path, newName+"/")
			})
		}),
	}
	new := server.NewSource(newCommitInfo, fs, opts...)
	diff := server.NewDiffer(old, new)
	return diff.Iterate(ctx, cb)
}

// CreateFileset creates a new temporary fileset and returns it.
func (d *Driver) CreateFileset(ctx context.Context, cb func(*fileset.UnorderedWriter) error) (*fileset.ID, error) {
	var id *fileset.ID
	if err := d.Storage.Filesets.WithRenewer(ctx, server.DefaultTTL, func(ctx context.Context, renewer *fileset.Renewer) error {
		var err error
		id, err = WithUnorderedWriter(ctx, d.Storage, renewer, cb)
		return err
	}); err != nil {
		return nil, err
	}
	return id, nil
}

func (d *Driver) GetFileset(ctx context.Context, commit *pfsdb.Commit) (*fileset.ID, error) {
	// Get the total file set if the commit has been finished.
	if commit.Finished != nil && commit.Error == "" {
		id, err := d.CommitStore.GetTotalFileSet(ctx, commit)
		if err != nil {
			if errors.Is(err, server.ErrNoTotalFileset) {
				return nil, errors.Errorf("the commit is forgotten")
			}
			return nil, errors.EnsureStack(err)
		}
		return id, nil
	}
	// Compose the base file set with the diffs.
	var ids []fileset.ID
	baseCommitHandle := commit.ParentCommit
	for baseCommitHandle != nil {
		baseCommit, err := d.GetCommit(ctx, baseCommitHandle)
		if err != nil {
			return nil, err
		}
		if baseCommit.Error == "" {
			if baseCommit.Finished == nil {
				return nil, pfsserver.ErrBaseCommitNotFinished{
					BaseCommit: baseCommitHandle,
					Commit:     commit.Commit,
				}
			}
			// ¯\_(ツ)_/¯
			baseId, err := d.GetFileset(ctx, baseCommit)
			if err != nil {
				return nil, err
			}
			ids = append(ids, *baseId)
			break
		}
		baseCommitHandle = baseCommit.ParentCommit
	}
	id, err := d.CommitStore.GetDiffFileSet(ctx, commit)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	ids = append(ids, *id)
	return d.Storage.Filesets.Compose(ctx, ids, server.DefaultTTL)
}

func (d *Driver) ShardFileset(ctx context.Context, fsid fileset.ID, numFiles, sizeBytes int64) ([]*pfs.PathRange, error) {
	fs, err := d.Storage.Filesets.Open(ctx, []fileset.ID{fsid})
	if err != nil {
		return nil, err
	}
	shardConfig := d.Storage.Filesets.ShardConfig()
	if numFiles > 0 {
		shardConfig.NumFiles = numFiles
	}
	if sizeBytes > 0 {
		shardConfig.SizeBytes = sizeBytes
	}
	shards, err := fs.Shards(ctx, index.WithShardConfig(shardConfig))
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

func (d *Driver) AddFileset(ctx context.Context, txnCtx *txncontext.TransactionContext, commitHandle *pfs.Commit, filesetID fileset.ID) error {
	commit, err := d.ResolveCommit(ctx, txnCtx.SqlTx, commitHandle)
	if err != nil {
		return err
	}
	// TODO: This check needs to be in the add transaction.
	if commit.Finishing != nil {
		return pfsserver.ErrCommitFinished{Commit: commit.Commit}
	}
	return errors.Wrap(d.CommitStore.AddFileSetTx(txnCtx.SqlTx, commit, filesetID), "add file set")
}

func (d *Driver) RenewFileset(ctx context.Context, id fileset.ID, ttl time.Duration) error {
	if ttl < time.Second {
		return errors.Errorf("ttl (%d) must be at least one second", ttl)
	}
	if ttl > server.MaxTTL {
		return errors.Errorf("ttl (%d) exceeds max ttl (%d)", ttl, server.MaxTTL)
	}
	_, err := d.Storage.Filesets.SetTTL(ctx, id, ttl)
	return err
}

func (d *Driver) ComposeFileset(ctx context.Context, ids []fileset.ID, ttl time.Duration, compact bool) (*fileset.ID, error) {
	if compact {
		compactor := server.NewCompactor(d.Storage.Filesets, d.Env.StorageConfig.StorageCompactionMaxFanIn)
		taskDoer := d.Env.TaskService.NewDoer(server.StorageTaskNamespace, uuid.NewWithoutDashes(), nil)
		return compactor.Compact(ctx, taskDoer, ids, ttl)
	}
	return d.Storage.Filesets.Compose(ctx, ids, ttl)
}

func (d *Driver) CompactDiffFileset(ctx context.Context, compactor *server.Compactor, doer task.Doer, renewer *fileset.Renewer, commit *pfsdb.Commit) (*fileset.ID, error) {
	id, err := d.CommitStore.GetDiffFileSet(ctx, commit)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	if err := renewer.Add(ctx, *id); err != nil {
		return nil, err
	}
	diffId, err := compactor.Compact(ctx, doer, []fileset.ID{*id}, server.DefaultTTL)
	if err != nil {
		return nil, err
	}
	return diffId, errors.EnsureStack(d.CommitStore.SetDiffFileSet(ctx, commit, *diffId))
}

func (d *Driver) CompactTotalFileset(ctx context.Context, compactor *server.Compactor, doer task.Doer, renewer *fileset.Renewer, commit *pfsdb.Commit) (*fileset.ID, error) {
	id, err := d.GetFileset(ctx, commit)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	if err := renewer.Add(ctx, *id); err != nil {
		return nil, err
	}
	totalId, err := compactor.Compact(ctx, doer, []fileset.ID{*id}, server.DefaultTTL)
	if err != nil {
		return nil, err
	}
	if err := errors.EnsureStack(d.CommitStore.SetTotalFileSet(ctx, commit, *totalId)); err != nil {
		return nil, err
	}
	return totalId, nil
}

func (d *Driver) commitSizeUpperBound(ctx context.Context, commit *pfsdb.Commit) (int64, error) {
	fsid, err := d.GetFileset(ctx, commit)
	if err != nil {
		return 0, err
	}
	return d.Storage.Filesets.SizeUpperBound(ctx, *fsid)
}

func newFileNotFound(commitID string, path string) pacherr.ErrNotExist {
	return pacherr.ErrNotExist{
		Collection: "commit/" + commitID,
		ID:         path,
	}
}
