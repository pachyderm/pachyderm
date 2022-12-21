package server

import (
	"context"
	"path"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/hashicorp/go-multierror"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/consistenthashing"
	"github.com/pachyderm/pachyderm/v2/src/internal/dlock"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/middleware/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	masterLockPath = "pfs-master-lock"
)

func (d *driver) master(ctx context.Context) {
	ctx = auth.AsInternalUser(ctx, "pfs-master")
	backoff.RetryUntilCancel(ctx, func() error { //nolint:errcheck
		eg, ctx := errgroup.WithContext(ctx)
		trackerPeriod := time.Second * time.Duration(d.env.StorageConfig.StorageGCPeriod)
		if trackerPeriod <= 0 {
			log.Info(ctx, "Skipping Storage GC")
		} else {
			eg.Go(func() error {
				lock := dlock.NewDLock(d.etcdClient, path.Join(d.prefix, masterLockPath, "storage-gc"))
				log.Info(ctx, "Starting Storage GC", zap.Duration("period", trackerPeriod))
				ctx, err := lock.Lock(ctx)
				if err != nil {
					return errors.EnsureStack(err)
				}
				defer func() {
					if err := lock.Unlock(ctx); err != nil {
						log.Error(ctx, "error unlocking in pfs master (storage gc)", zap.Error(err))
					}
				}()
				gc := d.storage.NewGC(trackerPeriod)
				return gc.RunForever(pctx.Child(ctx, "storage-gc"))
			})
		}
		chunkPeriod := time.Second * time.Duration(d.env.StorageConfig.StorageChunkGCPeriod)
		if chunkPeriod <= 0 {
			log.Info(ctx, "Skipping Chunk Storage GC")
		} else {
			eg.Go(func() error {
				lock := dlock.NewDLock(d.etcdClient, path.Join(d.prefix, masterLockPath, "chunk-gc"))
				log.Info(ctx, "Starting Chunk Storage GC", zap.Duration("period", chunkPeriod))
				ctx, err := lock.Lock(ctx)
				if err != nil {
					return errors.EnsureStack(err)
				}
				defer func() {
					if err := lock.Unlock(ctx); err != nil {
						log.Error(ctx, "error unlocking in pfs master (chunk gc)", zap.Error(err))
					}
				}()
				gc := chunk.NewGC(d.storage.ChunkStorage(), chunkPeriod)
				return gc.RunForever(pctx.Child(ctx, "chunk-gc"))
			})
		}
		eg.Go(func() error {
			return d.finishCommits(ctx)
		})
		return errors.EnsureStack(eg.Wait())
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		log.Error(ctx, "error in pfs master; restarting", zap.Error(err), zap.Duration("retryAfter", d))
		return nil
	})
}

func (d *driver) finishCommits(ctx context.Context) (retErr error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	repos := make(map[string]context.CancelFunc)
	defer func() {
		for _, cancel := range repos {
			cancel()
		}
	}()
	return consistenthashing.WithRing(ctx, d.etcdClient, path.Join(d.prefix, masterLockPath, "ring"),
		func(ctx context.Context, ring *consistenthashing.Ring) error {
			err := d.repos.ReadOnly(ctx).WatchF(func(ev *watch.Event) error {
				if ev.Type == watch.EventError {
					return ev.Err
				}
				key := string(ev.Key)
				lockPrefix := path.Join("repos", key)
				if ev.Type == watch.EventDelete {
					if cancel, ok := repos[key]; ok {
						if err := ring.Unlock(lockPrefix); err != nil {
							return err
						}
						cancel()
					}
					delete(repos, key)
					return nil
				}
				if _, ok := repos[key]; ok {
					return nil
				}
				ctx, cancel := context.WithCancel(ctx)
				repos[key] = cancel
				go func() {
					backoff.RetryUntilCancel(ctx, func() error { //nolint:errcheck
						ctx, cancel := context.WithCancel(ctx)
						defer cancel()
						lockCtx, err := ring.Lock(ctx, lockPrefix)
						if err != nil {
							return errors.Wrap(err, "error locking repo lock")
						}
						defer func() {
							if err := ring.Unlock(lockPrefix); err != nil {
								retErr = multierror.Append(retErr, errors.Wrap(err, "error unlocking"))
							}
						}()
						return d.finishRepoCommits(lockCtx, key)
					}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
						log.Error(ctx, "error finishing commits", zap.String("repo", key), zap.Error(err), zap.Duration("retryAfter", d))
						return nil
					})
				}()
				return nil
			})
			return errors.EnsureStack(err)
		},
	)
}

func (d *driver) finishRepoCommits(ctx context.Context, repoKey string) error {
	err := d.commits.ReadOnly(ctx).WatchByIndexF(pfsdb.CommitsRepoIndex, repoKey, func(ev *watch.Event) error {
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
		commit := commitInfo.Commit
		cache := d.newCache(pfsdb.CommitKey(commit))
		defer func() {
			if err := cache.clear(ctx); err != nil {
				log.Error(ctx, "errored clearing compaction cache", zap.Error(err))
			}
		}()
		return log.LogStep(ctx, "finishCommit", func(ctx context.Context) error {
			// TODO: This retry might not be getting us much if the outer watch still errors due to a transient error.
			return backoff.RetryUntilCancel(ctx, func() error {
				// Skip compaction / validation for errored commits.
				if commitInfo.Error != "" {
					return d.finalizeCommit(ctx, commit, "", nil, nil)
				}
				id, err := d.getFileSet(ctx, commit)
				if err != nil {
					if pfsserver.IsCommitNotFoundErr(err) {
						return nil
					}
					return err
				}
				compactor := newCompactor(d.storage, d.env.StorageConfig.StorageCompactionMaxFanIn)
				taskDoer := d.env.TaskService.NewDoer(StorageTaskNamespace, commit.ID, cache)
				var totalId *fileset.ID
				start := time.Now()
				if err := d.storage.WithRenewer(ctx, defaultTTL, func(ctx context.Context, renewer *fileset.Renewer) error {
					if err := renewer.Add(ctx, *id); err != nil {
						return err
					}
					// Compact the commit.
					return log.LogStep(ctx, "compactCommit", func(ctx context.Context) error {
						var err error
						totalId, err = compactor.Compact(ctx, taskDoer, []fileset.ID{*id}, defaultTTL)
						if err != nil {
							return err
						}
						return errors.EnsureStack(d.commitStore.SetTotalFileSet(ctx, commit, *totalId))
					})
				}); err != nil {
					return err
				}
				details := &pfs.CommitInfo_Details{
					CompactingTime: types.DurationProto(time.Since(start)),
				}
				// Validate the commit.
				start = time.Now()
				var validationError string
				if err := log.LogStep(ctx, "validateCommit", func(ctx context.Context) error {
					var err error
					validationError, details.SizeBytes, err = compactor.Validate(ctx, taskDoer, *totalId)
					return err
				}); err != nil {
					return err
				}
				details.ValidatingTime = types.DurationProto(time.Since(start))
				// Finish the commit.
				return d.finalizeCommit(ctx, commit, validationError, details, totalId)
			}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
				log.Error(ctx, "error finishing commit", zap.Error(err), zap.Duration("retryAfter", d))
				return nil
			})
		}, zap.Bool("finishing", true), log.Proto("commit", commit), zap.String("repo", key))
	}, watch.WithSort(col.SortByCreateRevision, col.SortAscend), watch.IgnoreDelete)
	return errors.EnsureStack(err)
}

func (d *driver) finalizeCommit(ctx context.Context, commit *pfs.Commit, validationError string, details *pfs.CommitInfo_Details, totalId *fileset.ID) error {
	return log.LogStep(ctx, "finalizeCommit", func(ctx context.Context) error {
		return d.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
			commitInfo := &pfs.CommitInfo{}
			if err := d.commits.ReadWrite(txnCtx.SqlTx).Update(pfsdb.CommitKey(commit), commitInfo, func() error {
				commitInfo.Finished = txnCtx.Timestamp
				if details != nil {
					commitInfo.SizeBytesUpperBound = details.SizeBytes
				}
				if commitInfo.Error == "" {
					commitInfo.Error = validationError
				}
				commitInfo.Details = details
				return nil
			}); err != nil {
				return errors.EnsureStack(err)
			}
			if commitInfo.Commit.Branch.Repo.Type == pfs.UserRepoType {
				txnCtx.FinishJob(commitInfo)
			}
			if err := d.finishAliasDescendents(txnCtx, commitInfo, totalId); err != nil {
				return err
			}
			// TODO(2.0 optional): This is a hack to ensure that commits created by triggers have the same ID as the finished commit.
			// Creating an alias in the branch of the finished commit, then having the trigger alias that commit seems
			// like a better model.
			txnCtx.CommitSetID = commitInfo.Commit.ID
			return d.triggerCommit(txnCtx, commitInfo.Commit)
		})
	})
}

// finishAliasDescendents will traverse the given commit's descendents, finding all
// contiguous aliases and finishing them.
func (d *driver) finishAliasDescendents(txnCtx *txncontext.TransactionContext, parentCommitInfo *pfs.CommitInfo, id *fileset.ID) error {
	// Build the starting set of commits to consider
	descendents := append([]*pfs.Commit{}, parentCommitInfo.ChildCommits...)

	// A commit cannot have more than one parent, so no need to track visited nodes
	for len(descendents) > 0 {
		commit := descendents[0]
		descendents = descendents[1:]
		commitInfo := &pfs.CommitInfo{}
		if err := d.commits.ReadWrite(txnCtx.SqlTx).Get(pfsdb.CommitKey(commit), commitInfo); err != nil {
			return errors.EnsureStack(err)
		}

		if commitInfo.Origin.Kind == pfs.OriginKind_ALIAS {
			if commitInfo.Finishing == nil {
				commitInfo.Finishing = txnCtx.Timestamp
			}
			commitInfo.Finished = txnCtx.Timestamp
			commitInfo.Details = parentCommitInfo.Details
			commitInfo.Error = parentCommitInfo.Error
			if err := d.commits.ReadWrite(txnCtx.SqlTx).Put(pfsdb.CommitKey(commit), commitInfo); err != nil {
				return errors.EnsureStack(err)
			}
			if id != nil {
				if err := d.commitStore.SetTotalFileSetTx(txnCtx.SqlTx, commitInfo.Commit, *id); err != nil {
					return errors.EnsureStack(err)
				}
			}

			descendents = append(descendents, commitInfo.ChildCommits...)
		}
	}
	return nil
}
