package server

import (
	"context"
	"path"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/dlock"
	_ "github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

const (
	masterLockPath = "pfs-master-lock"
)

func (d *driver) master(ctx context.Context) {
	masterLock := dlock.NewDLock(d.etcdClient, path.Join(d.prefix, masterLockPath))
	backoff.RetryUntilCancel(ctx, func() error {
		masterCtx, err := masterLock.Lock(ctx)
		if err != nil {
			return err
		}
		defer masterLock.Unlock(masterCtx)
		eg, ctx := errgroup.WithContext(masterCtx)
		eg.Go(func() error {
			return d.storage.GC(ctx)
		})
		eg.Go(func() error {
			gc := chunk.NewGC(d.storage.ChunkStorage())
			return gc.RunForever(ctx)
		})
		eg.Go(func() error {
			return d.finishCommits(ctx)
		})
		return eg.Wait()
	}, backoff.NewInfiniteBackOff(), func(err error, _ time.Duration) error {
		log.Errorf("error in pfs master: %v", err)
		return nil
	})
}

func (d *driver) finishCommits(ctx context.Context) error {
	return d.commits.ReadOnly(ctx).WatchF(func(ev *watch.Event) error {
		if ev.Type == watch.EventError {
			return ev.Err
		}
		var key string
		commitInfo := &pfs.CommitInfo{}
		if err := ev.Unmarshal(&key, commitInfo); err != nil {
			return err
		}
		if commitInfo.Finishing == nil || commitInfo.Finished != nil {
			return nil
		}
		return backoff.RetryUntilCancel(ctx, func() error {
			id, err := d.getFileSet(ctx, commitInfo.Commit)
			if err != nil {
				if pfsserver.IsCommitNotFoundErr(err) {
					return nil
				}
				return err
			}
			// Compact the commit.
			compactedId, err := d.compactor.Compact(ctx, []fileset.ID{*id}, defaultTTL)
			if err != nil {
				return err
			}
			if err := d.commitStore.SetTotalFileSet(ctx, commitInfo.Commit, *compactedId); err != nil {
				return err
			}
			// Validate the commit.
			// TODO(2.0 optional): Improve the performance of this by doing a logarithmic lookup per new file,
			// rather than a linear scan through all of the files.
			fs, err := d.storage.Open(ctx, []fileset.ID{*compactedId})
			if err != nil {
				return err
			}
			var prev string
			var size int64
			if err := fs.Iterate(ctx, func(f fileset.File) error {
				idx := f.Index()
				if prev != "" && (idx.Path == prev || strings.HasPrefix(idx.Path, prev+"/")) {
					commitInfo.Error = true
				}
				prev = idx.Path
				size += index.SizeBytes(idx)
				return nil
			}); err != nil {
				return err
			}
			// Finish the commit.
			return d.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
				commitInfo.Finished = txnCtx.Timestamp
				commitInfo.SizeBytesUpperBound = size
				if commitInfo.Details == nil {
					commitInfo.Details = &pfs.CommitInfo_Details{}
				}
				commitInfo.Details.SizeBytes = size
				if err := d.commits.ReadWrite(txnCtx.SqlTx).Put(pfsdb.CommitKey(commitInfo.Commit), commitInfo); err != nil {
					return err
				}
				if commitInfo.Commit.Branch.Repo.Type == pfs.UserRepoType {
					txnCtx.FinishJob(commitInfo)
				}
				if err := d.finishAliasDescendents(txnCtx, commitInfo); err != nil {
					return err
				}
				// TODO(2.0 optional): This is a hack to ensure that commits created by triggers have the same ID as the finished commit.
				// Creating an alias in the branch of the finished commit, then having the trigger alias that commit seems
				// like a better model.
				txnCtx.CommitSetID = commitInfo.Commit.ID
				return d.triggerCommit(txnCtx, commitInfo.Commit)
			})
		}, backoff.NewInfiniteBackOff(), func(err error, _ time.Duration) error {
			log.Errorf("error finishing commit %v: %v", commitInfo.Commit.ID, err)
			return nil
		})
	}, watch.IgnoreDelete)
}

// finishAliasChildren will traverse the given commit's children, finding all
// continguous aliases and finishing them.
func (d *driver) finishAliasDescendents(txnCtx *txncontext.TransactionContext, parentCommitInfo *pfs.CommitInfo) error {
	// Build the starting set of commits to consider
	descendents := append([]*pfs.Commit{}, parentCommitInfo.ChildCommits...)

	// A commit cannot have more than one parent, so no need to track visited nodes
	for len(descendents) > 0 {
		commit := descendents[0]
		descendents = descendents[1:]
		commitInfo := &pfs.CommitInfo{}
		if err := d.commits.ReadWrite(txnCtx.SqlTx).Get(pfsdb.CommitKey(commit), commitInfo); err != nil {
			return err
		}

		if commitInfo.Origin.Kind == pfs.OriginKind_ALIAS {
			if commitInfo.Finishing == nil {
				commitInfo.Finishing = txnCtx.Timestamp
			}
			commitInfo.Finished = txnCtx.Timestamp
			commitInfo.Error = parentCommitInfo.Error
			if err := d.commits.ReadWrite(txnCtx.SqlTx).Put(pfsdb.CommitKey(commit), commitInfo); err != nil {
				return err
			}

			descendents = append(descendents, commitInfo.ChildCommits...)
		}
	}
	return nil
}
