package server

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dlock"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/middleware/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
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
	ctx = auth.AsInternalUser(ctx, "pfs-master")
	masterLock := dlock.NewDLock(d.etcdClient, path.Join(d.prefix, masterLockPath))
	backoff.RetryUntilCancel(ctx, func() error { //nolint:errcheck
		masterCtx, err := masterLock.Lock(ctx)
		if err != nil {
			return errors.EnsureStack(err)
		}
		defer masterLock.Unlock(masterCtx) //nolint:errcheck
		eg, ctx := errgroup.WithContext(masterCtx)
		trackerPeriod := time.Second * time.Duration(d.env.StorageConfig.StorageGCPeriod)
		if trackerPeriod <= 0 {
			d.log.Info("Skipping Storage GC")
		} else {
			d.log.Infof("Starting Storage GC with period=%v", trackerPeriod)
			eg.Go(func() error {
				gc := d.storage.NewGC(trackerPeriod)
				return gc.RunForever(ctx)
			})
		}
		chunkPeriod := time.Second * time.Duration(d.env.StorageConfig.StorageChunkGCPeriod)
		if chunkPeriod <= 0 {
			d.log.Info("Skipping Chunk Storage GC")
		} else {
			d.log.Infof("Starting Chunk Storage GC with period=%v", chunkPeriod)
			eg.Go(func() error {
				gc := chunk.NewGC(d.storage.ChunkStorage(), chunkPeriod, d.log)
				return gc.RunForever(ctx)
			})
		}
		eg.Go(func() error {
			return d.finishCommits(ctx)
		})
		return errors.EnsureStack(eg.Wait())
	}, backoff.NewInfiniteBackOff(), func(err error, _ time.Duration) error {
		log.Errorf("error in pfs master: %v", err)
		return nil
	})
}

func (d *driver) finishCommits(ctx context.Context) error {
	repos := make(map[string]context.CancelFunc)
	defer func() {
		for _, cancel := range repos {
			cancel()
		}
	}()
	compactor := newCompactor(d.storage, d.env.StorageConfig.StorageCompactionMaxFanIn)
	err := d.repos.ReadOnly(ctx).WatchF(func(ev *watch.Event) error {
		if ev.Type == watch.EventError {
			return ev.Err
		}
		key := string(ev.Key)
		if ev.Type == watch.EventDelete {
			if cancel, ok := repos[key]; ok {
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
				return d.finishRepoCommits(ctx, compactor, key)
			}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
				log.Errorf("error finishing commits for repo %v: %v, retrying in %v", key, err, d)
				return nil
			})
		}()
		return nil
	})
	return errors.EnsureStack(err)
}

func (d *driver) finishRepoCommits(ctx context.Context, compactor *compactor, repoKey string) error {
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
				log.Errorf("errored clearing compaction cache: %s", err)
			}
		}()
		return miscutil.LogStep(ctx, fmt.Sprintf("finishing commit %v", commit), func() error {
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
				details := &pfs.CommitInfo_Details{}
				// Compact the commit.
				taskDoer := d.env.TaskService.NewDoer(storageTaskNamespace, commit.ID, cache)
				var totalId *fileset.ID
				start := time.Now()
				if err := miscutil.LogStep(ctx, fmt.Sprintf("compacting commit %v", commit), func() error {
					var err error
					totalId, err = compactor.Compact(ctx, taskDoer, []fileset.ID{*id}, defaultTTL)
					if err != nil {
						return err
					}
					return errors.EnsureStack(d.commitStore.SetTotalFileSet(ctx, commit, *totalId))
				}); err != nil {
					return err
				}
				details.CompactingTime = types.DurationProto(time.Since(start))
				// Validate the commit.
				start = time.Now()
				var validationError string
				if err := miscutil.LogStep(ctx, fmt.Sprintf("validating commit %v", commit), func() error {
					var err error
					details.SizeBytes, validationError, err = compactor.Validate(ctx, taskDoer, *totalId)
					return err
				}); err != nil {
					return err
				}
				details.ValidatingTime = types.DurationProto(time.Since(start))
				// Finish the commit.
				return d.finalizeCommit(ctx, commit, validationError, details, totalId)
			}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
				log.Errorf("error finishing commit %v: %v, retrying in %v", commit, err, d)
				return nil
			})
		})
	}, watch.WithSort(col.SortByCreateRevision, col.SortAscend), watch.IgnoreDelete)
	return errors.EnsureStack(err)
}

func (d *driver) finalizeCommit(ctx context.Context, commit *pfs.Commit, validationError string, details *pfs.CommitInfo_Details, totalId *fileset.ID) error {
	return miscutil.LogStep(ctx, fmt.Sprintf("finalizing commit %v", commit), func() error {
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
