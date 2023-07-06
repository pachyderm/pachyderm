package server

import (
	"context"
	"path"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/consistenthashing"
	"github.com/pachyderm/pachyderm/v2/src/internal/cronutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/dlock"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/middleware/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/protoutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
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
				gc := d.storage.Filesets.NewGC(trackerPeriod)
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
				gc := chunk.NewGC(d.storage.Chunks, chunkPeriod)
				return gc.RunForever(pctx.Child(ctx, "chunk-gc"))
			})
		}
		eg.Go(func() error {
			return d.manageRepos(ctx)
		})
		return errors.EnsureStack(eg.Wait())
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		log.Error(ctx, "error in pfs master; restarting", zap.Error(err), zap.Duration("retryAfter", d))
		return nil
	})
}

func (d *driver) manageRepos(ctx context.Context) error {
	ctx, cancel := pctx.WithCancel(ctx)
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
						delete(repos, key)
					}
					return nil
				}
				if _, ok := repos[key]; ok {
					return nil
				}
				ctx, cancel := pctx.WithCancel(ctx)
				repos[key] = cancel
				go func() {
					backoff.RetryUntilCancel(ctx, func() (retErr error) { //nolint:errcheck
						ctx, cancel := pctx.WithCancel(ctx)
						defer cancel()
						var err error
						ctx, err = ring.Lock(ctx, lockPrefix)
						if err != nil {
							return errors.Wrap(err, "error locking repo lock")
						}
						defer errors.Invoke1(&retErr, ring.Unlock, lockPrefix, "unlocking repo lock")
						var eg errgroup.Group
						eg.Go(func() error {
							return backoff.RetryUntilCancel(ctx, func() error {
								return d.manageBranches(ctx, key)
							}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
								log.Error(ctx, "error managing branches", zap.String("repo", key), zap.Error(err), zap.Duration("retryAfter", d))
								return nil
							})
						})
						eg.Go(func() error {
							return backoff.RetryUntilCancel(ctx, func() error {
								return d.finishRepoCommits(ctx, key)
							}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
								log.Error(ctx, "error finishing repo commits", zap.String("repo", key), zap.Error(err), zap.Duration("retryAfter", d))
								return nil
							})
						})
						return errors.EnsureStack(eg.Wait())
					}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
						log.Error(ctx, "error managing repo", zap.String("repo", key), zap.Error(err), zap.Duration("retryAfter", d))
						return nil
					})
				}()
				return nil
			})
			return errors.EnsureStack(err)
		},
	)
}

func (d *driver) manageBranches(ctx context.Context, repoKey string) error {
	ctx, cancel := pctx.WithCancel(ctx)
	defer cancel()
	type cronTrigger struct {
		cancel context.CancelFunc
		spec   string
	}
	cronTriggers := make(map[string]*cronTrigger)
	defer func() {
		for _, ct := range cronTriggers {
			ct.cancel()
		}
	}()
	return d.branches.ReadOnly(ctx).WatchByIndexF(pfsdb.BranchesRepoIndex, repoKey, func(ev *watch.Event) error {
		if ev.Type == watch.EventError {
			return ev.Err
		}
		key := string(ev.Key)
		if ev.Type == watch.EventDelete {
			if ct, ok := cronTriggers[key]; ok {
				ct.cancel()
				delete(cronTriggers, key)
			}
			return nil
		}
		branchInfo, err := d.inspectBranch(ctx, pfsdb.ParseBranch(key))
		if err != nil {
			return err
		}
		// Only create a new goroutine if one doesn't already exist or the spec changed.
		if ct, ok := cronTriggers[key]; ok {
			if branchInfo.Trigger != nil && ct.spec == branchInfo.Trigger.CronSpec {
				return nil
			}
			ct.cancel()
			delete(cronTriggers, key)
		}
		if branchInfo.Trigger == nil || branchInfo.Trigger.CronSpec == "" {
			return nil
		}
		ctx, cancel := pctx.WithCancel(ctx)
		cronTriggers[key] = &cronTrigger{
			cancel: cancel,
			spec:   branchInfo.Trigger.CronSpec,
		}
		go func() {
			backoff.RetryUntilCancel(ctx, func() error { //nolint:errcheck
				return d.runCronTrigger(ctx, branchInfo.Branch)
			}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
				log.Error(ctx, "error running cron trigger", zap.String("branch", key), zap.Error(err), zap.Duration("retryAfter", d))
				return nil
			})
		}()
		return nil
	})
}

func (d *driver) runCronTrigger(ctx context.Context, branch *pfs.Branch) error {
	branchInfo, err := d.inspectBranch(ctx, branch)
	if err != nil {
		return err
	}
	schedule, err := cronutil.ParseCronExpression(branchInfo.Trigger.CronSpec)
	if err != nil {
		return err
	}
	// Use the current head commit start time as the previous tick.
	// This prevents the timer from restarting if the master restarts.
	ci, err := d.inspectCommit(ctx, branchInfo.Head, pfs.CommitState_STARTED)
	if err != nil {
		return err
	}
	prev := protoutil.MustTime(ci.Started)
	for {
		next := schedule.Next(prev)
		if next.IsZero() {
			log.Debug(ctx, "no more scheduled ticks; exiting loop")
			return nil
		}
		log.Info(ctx, "waiting for next cron tick", zap.Time("next", next), zap.Time("prev", prev))
		select {
		case <-time.After(time.Until(next)):
		case <-ctx.Done():
			return errors.EnsureStack(context.Cause(ctx))
		}
		if err := d.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
			trigBI, err := d.inspectBranchInTransaction(txnCtx, branchInfo.Branch.Repo.NewBranch(branchInfo.Trigger.Branch))
			if err != nil {
				return err
			}
			var bi pfs.BranchInfo
			if err := d.branches.ReadWrite(txnCtx.SqlTx).Update(branchInfo.Branch, &bi, func() error {
				bi.Head = trigBI.Head
				return nil
			}); err != nil {
				return errors.EnsureStack(err)
			}
			return txnCtx.PropagateBranch(bi.Branch)
		}); err != nil {
			return err
		}
		log.Info(ctx, "cron tick completed", zap.Time("next", next))
		prev = next
	}
}

func (d *driver) finishRepoCommits(ctx context.Context, repoKey string) error {
	return errors.EnsureStack(d.commits.ReadOnly(ctx).WatchByIndexF(pfsdb.CommitsRepoIndex, repoKey, func(ev *watch.Event) error {
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
			return backoff.RetryUntilCancel(ctx, func() error {
				// In the case where a commit is squashed between retries (commit not found), the master can return.
				if _, err := d.getCommit(ctx, commit); err != nil {
					if pfsserver.IsCommitNotFoundErr(err) {
						return nil
					}
					return err
				}
				// Skip compaction / validation for errored commits.
				if commitInfo.Error != "" {
					return d.finalizeCommit(ctx, commit, "", nil, nil)
				}
				compactor := newCompactor(d.storage.Filesets, d.env.StorageConfig.StorageCompactionMaxFanIn)
				taskDoer := d.env.TaskService.NewDoer(StorageTaskNamespace, commit.Id, cache)
				return errors.EnsureStack(d.storage.Filesets.WithRenewer(ctx, defaultTTL, func(ctx context.Context, renewer *fileset.Renewer) error {
					start := time.Now()
					// Compacting the diff before getting the total allows us to compose the
					// total file set so that it includes the compacted diff.
					if err := log.LogStep(ctx, "compactDiffFileSet", func(ctx context.Context) error {
						_, err := d.compactDiffFileSet(ctx, compactor, taskDoer, renewer, commit)
						return err
					}); err != nil {
						return err
					}
					var totalId *fileset.ID
					var err error
					if err := log.LogStep(ctx, "compactTotalFileSet", func(ctx context.Context) error {
						totalId, err = d.compactTotalFileSet(ctx, compactor, taskDoer, renewer, commit)
						return err
					}); err != nil {
						return err
					}
					details := &pfs.CommitInfo_Details{
						CompactingTime: durationpb.New(time.Since(start)),
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
					details.ValidatingTime = durationpb.New(time.Since(start))
					// Finish the commit.
					return d.finalizeCommit(ctx, commit, validationError, details, totalId)
				}))
			}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
				log.Error(ctx, "error finishing commit", zap.Error(err), zap.Duration("retryAfter", d))
				return nil
			})
		}, zap.Bool("finishing", true), log.Proto("commit", commitInfo.Commit), zap.String("repo", key))
	}, watch.WithSort(col.SortByCreateRevision, col.SortAscend), watch.IgnoreDelete))
}

func (d *driver) compactDiffFileSet(ctx context.Context, compactor *compactor, doer task.Doer, renewer *fileset.Renewer, commit *pfs.Commit) (*fileset.ID, error) {
	id, err := d.commitStore.GetDiffFileSet(ctx, commit)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	if err := renewer.Add(ctx, *id); err != nil {
		return nil, err
	}
	diffId, err := compactor.Compact(ctx, doer, []fileset.ID{*id}, defaultTTL)
	if err != nil {
		return nil, err
	}
	return diffId, errors.EnsureStack(d.commitStore.SetDiffFileSet(ctx, commit, *diffId))
}

func (d *driver) compactTotalFileSet(ctx context.Context, compactor *compactor, doer task.Doer, renewer *fileset.Renewer, commit *pfs.Commit) (*fileset.ID, error) {
	id, err := d.getFileSet(ctx, commit)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	if err := renewer.Add(ctx, *id); err != nil {
		return nil, err
	}
	totalId, err := compactor.Compact(ctx, doer, []fileset.ID{*id}, defaultTTL)
	if err != nil {
		return nil, err
	}
	if err := errors.EnsureStack(d.commitStore.SetTotalFileSet(ctx, commit, *totalId)); err != nil {
		return nil, err
	}
	return totalId, nil
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
			if commitInfo.Commit.Repo.Type == pfs.UserRepoType {
				txnCtx.FinishJob(commitInfo)
			}
			if commitInfo.Error == "" {
				return d.triggerCommit(txnCtx, commitInfo.Commit)
			}
			return nil
		})
	})
}
