package server

import (
	"context"
	"fmt"
	"path"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/consistenthashing"
	"github.com/pachyderm/pachyderm/v2/src/internal/cronutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/dlock"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/middleware/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/protoutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/randutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch/postgres"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
)

const (
	masterLockPath = "pfs-master-lock"
)

type Master struct {
	env Env

	driver *driver
}

func NewMaster(env Env) (*Master, error) {
	d, err := newDriver(env)
	if err != nil {
		return nil, err
	}
	return &Master{
		env:    env,
		driver: d,
	}, nil
}

func (m *Master) Run(ctx context.Context) error {
	ctx = auth.AsInternalUser(ctx, "pfs-master")
	return backoff.RetryUntilCancel(ctx, func() error {
		eg, ctx := errgroup.WithContext(ctx)
		trackerPeriod := time.Second * time.Duration(m.env.StorageConfig.StorageGCPeriod)
		if trackerPeriod <= 0 {
			log.Info(ctx, "Skipping Storage GC")
		} else {
			eg.Go(func() error {
				lock := dlock.NewDLock(m.driver.etcdClient, path.Join(m.driver.prefix, masterLockPath, "storage-gc"))
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
				gc := m.driver.storage.Filesets.NewGC(trackerPeriod)
				return gc.RunForever(pctx.Child(ctx, "storage-gc"))
			})
		}
		chunkPeriod := time.Second * time.Duration(m.env.StorageConfig.StorageChunkGCPeriod)
		if chunkPeriod <= 0 {
			log.Info(ctx, "Skipping Chunk Storage GC")
		} else {
			eg.Go(func() error {
				lock := dlock.NewDLock(m.driver.etcdClient, path.Join(m.driver.prefix, masterLockPath, "chunk-gc"))
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
				gc := chunk.NewGC(m.driver.storage.Chunks, chunkPeriod)
				return gc.RunForever(pctx.Child(ctx, "chunk-gc"))
			})
		}
		eg.Go(func() error {
			return m.watchRepos(ctx)
		})
		return errors.EnsureStack(eg.Wait())
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		log.Error(ctx, "error in pfs master; restarting", zap.Error(err), zap.Duration("retryAfter", d))
		return nil
	})
}

func (m *Master) watchRepos(ctx context.Context) error {
	ctx, cancel := pctx.WithCancel(ctx)
	defer cancel()
	repos := make(map[pfsdb.RepoID]context.CancelFunc)
	defer func() {
		for _, cancel := range repos {
			cancel()
		}
	}()
	ringPrefix := path.Join(randutil.UniqueString(m.driver.prefix), masterLockPath, "ring")
	return consistenthashing.WithRing(ctx, m.driver.etcdClient, ringPrefix,
		func(ctx context.Context, ring *consistenthashing.Ring) error {
			// Watch for repo events.
			watcher, err := postgres.NewWatcher(m.env.DB, m.driver.env.Listener, ringPrefix, pfsdb.ReposChannelName)
			if err != nil {
				return errors.Wrap(err, "new watcher")
			}
			defer watcher.Close()
			// Get existing entries.
			if err := dbutil.WithTx(ctx, m.env.DB, func(cbCtx context.Context, tx *pachsql.Tx) error {
				iter, err := pfsdb.ListRepo(cbCtx, tx, nil)
				if err != nil {
					return errors.Wrap(err, "create list repo iterator")
				}
				return errors.Wrap(stream.ForEach[pfsdb.RepoPair](cbCtx, iter, func(repoPair pfsdb.RepoPair) error {
					lockPrefix := path.Join("repos", fmt.Sprintf("%d", repoPair.ID))
					// cbCtx cannot be used here because it expires at the end of the callback and the
					// goroutines spawned by manageRepo need to live until the main master routine is cancelled.
					ctx, cancel := pctx.WithCancel(ctx)
					repos[repoPair.ID] = cancel
					go m.driver.manageRepo(ctx, ring, repoPair, lockPrefix)
					return nil
				}), "for each repo")
			}, dbutil.WithReadOnly()); err != nil {
				return errors.Wrap(err, "list repos")
			}
			// Process new repo events.
			return m.driver.handleRepoEvents(ctx, ring, repos, watcher.Watch())
		})
}

func (d *driver) handleRepoEvents(ctx context.Context, ring *consistenthashing.Ring, repos map[pfsdb.RepoID]context.CancelFunc, watcherChan <-chan *postgres.Event) error {
	for {
		select {
		case event, ok := <-watcherChan:
			if !ok {
				return errors.Wrap(fmt.Errorf("unexpected close on events channel"), "watch repo events")
			}
			if event.Err != nil {
				log.Error(ctx, event.Err.Error())
				continue
			}
			repoID := pfsdb.RepoID(event.Id)
			lockPrefix := path.Join("repos", fmt.Sprintf("%d", event.Id))
			if event.Type == postgres.EventDelete {
				if cancel, ok := repos[repoID]; ok {
					if err := ring.Unlock(lockPrefix); err != nil {
						return err
					}
					cancel()
					delete(repos, repoID)
				}
				continue
			}
			if _, ok := repos[repoID]; ok {
				continue // the master has already called manageRepo for this repo.
			}
			ctx, cancel := pctx.WithCancel(ctx)
			repos[repoID] = cancel
			var repo *pfs.RepoInfo
			var err error
			if err := dbutil.WithTx(ctx, d.env.DB, func(ctx context.Context, tx *pachsql.Tx) error {
				repo, err = pfsdb.GetRepo(ctx, tx, repoID)
				return errors.Wrap(err, "get repo from event id")
			}, dbutil.WithReadOnly()); err != nil {
				return errors.Wrap(err, "get repo")
			}
			go d.manageRepo(ctx, ring, pfsdb.RepoPair{ID: repoID, RepoInfo: repo}, lockPrefix)
		case <-ctx.Done():
			return nil
		}
	}
}

func (d *driver) manageRepo(ctx context.Context, ring *consistenthashing.Ring, repoPair pfsdb.RepoPair, lockPrefix string) {
	key := pfsdb.RepoKey(repoPair.RepoInfo.Repo)
	backoff.RetryUntilCancel(ctx, func() (retErr error) { //nolint:errcheck
		ctx, cancel := pctx.WithCancel(ctx)
		defer cancel()
		var err error
		ctx, err = ring.Lock(ctx, lockPrefix)
		if err != nil {
			return errors.Wrap(err, "locking repo lock")
		}
		defer errors.Invoke1(&retErr, ring.Unlock, lockPrefix, "unlocking repo lock")
		var eg errgroup.Group
		eg.Go(func() error {
			return backoff.RetryUntilCancel(ctx, func() error {
				return d.manageBranches(ctx, repoPair)
			}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
				log.Error(ctx, "managing branches", zap.String("repo", key), zap.Error(err), zap.Duration("retryAfter", d))
				return nil
			})
		})
		eg.Go(func() error {
			return backoff.RetryUntilCancel(ctx, func() error {
				return d.finishRepoCommits(ctx, repoPair)
			}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
				log.Error(ctx, "finishing repo commits", zap.Uint64("repo id", uint64(repoPair.ID)), zap.String("repo", key), zap.Error(err), zap.Duration("retryAfter", d))
				return nil
			})
		})
		return errors.EnsureStack(eg.Wait())
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		log.Error(ctx, "managing repo", zap.String("repo", key), zap.Error(err), zap.Duration("retryAfter", d))
		return nil
	})
}

type cronTrigger struct {
	cancel context.CancelFunc
	spec   string
}

func (d *driver) manageBranches(ctx context.Context, repoPair pfsdb.RepoPair) error {
	repoKey := repoPair.RepoInfo.Repo.Key()
	ctx, cancel := pctx.WithCancel(ctx)
	defer cancel()
	cronTriggers := make(map[pfsdb.BranchID]*cronTrigger)
	defer func() {
		for _, ct := range cronTriggers {
			ct.cancel()
		}
	}()
	watcher, err := postgres.NewWatcher(d.env.DB, d.env.Listener, path.Join(randutil.UniqueString(d.prefix), "manageBranches", repoKey),
		fmt.Sprintf("%s%d", pfsdb.BranchesRepoChannelName, repoPair.ID))
	if err != nil {
		return errors.Wrap(err, "manage branches")
	}
	if err := dbutil.WithTx(ctx, d.env.DB, func(cbCtx context.Context, tx *pachsql.Tx) error {
		iter, err := pfsdb.NewBranchIterator(ctx, tx, 0, 100, &pfs.Branch{
			Repo: repoPair.RepoInfo.Repo})
		if err != nil {
			return errors.Wrap(err, "manage branches")
		}
		return stream.ForEach[pfsdb.BranchInfoWithID](cbCtx, iter, func(branchInfoWithID pfsdb.BranchInfoWithID) error {
			return d.manageBranch(ctx, branchInfoWithID, cronTriggers)
		})
	}); err != nil {
		return err
	}
	for {
		event, ok := <-watcher.Watch()
		if !ok {
			return errors.Errorf("watcher for branch repo %d %s closed channel", repoPair.ID, repoKey)
		}
		key := pfsdb.BranchID(event.Id)
		if event.Type == postgres.EventDelete {
			if ct, ok := cronTriggers[key]; ok {
				ct.cancel()
				delete(cronTriggers, key)
			}
			return nil
		}
		if event.Err != nil {
			return event.Err
		}
		var branchInfo *pfs.BranchInfo
		if err := dbutil.WithTx(ctx, d.env.DB, func(cbCtx context.Context, tx *pachsql.Tx) error {
			branchInfo, err = pfsdb.GetBranchInfo(ctx, tx, key)
			return err
		}, dbutil.WithReadOnly()); err != nil {
			return errors.Wrap(err, "getting commit from event")
		}
		branchInfoWithID := pfsdb.BranchInfoWithID{
			ID:         key,
			BranchInfo: branchInfo,
		}
		if err := d.manageBranch(ctx, branchInfoWithID, cronTriggers); err != nil {
			return errors.Wrap(err, "manage branch")
		}
	}
}

func (d *driver) manageBranch(ctx context.Context, branchInfoWithID pfsdb.BranchInfoWithID, cronTriggers map[pfsdb.BranchID]*cronTrigger) error {
	branchInfo := branchInfoWithID.BranchInfo
	key := branchInfoWithID.ID
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
			log.Error(ctx, "error running cron trigger", zap.Uint64("branch id", uint64(branchInfoWithID.ID)), zap.String("branch", branchInfo.Branch.Key()), zap.Error(err), zap.Duration("retryAfter", d))
			return nil
		})
	}()
	return nil
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
			trigBI, err := d.inspectBranchInTransaction(ctx, txnCtx, branchInfo.Branch.Repo.NewBranch(branchInfo.Trigger.Branch))
			if err != nil {
				return err
			}
			branchInfo.Head = trigBI.Head
			if _, err := pfsdb.UpsertBranch(ctx, txnCtx.SqlTx, branchInfo); err != nil {
				return err
			}
			return txnCtx.PropagateBranch(branchInfo.Branch)
		}); err != nil {
			return err
		}
		log.Info(ctx, "cron tick completed", zap.Time("next", next))
		prev = next
	}
}

func (d *driver) finishRepoCommits(ctx context.Context, repoPair pfsdb.RepoPair) error {
	chanName := fmt.Sprintf("%s%d", pfsdb.CommitsRepoChannelName, repoPair.ID)
	repoKey := repoPair.RepoInfo.Repo.Key()
	watcher, err := postgres.NewWatcher(d.env.DB, d.env.Listener, path.Join(randutil.UniqueString(d.prefix), "finishRepoCommits"), chanName)
	if err != nil {
		return errors.Wrap(err, "new watcher")
	}
	defer watcher.Close()
	// Get existing entries.
	iter, err := pfsdb.ListCommit(ctx, d.env.DB, pfsdb.CommitListFilter{pfsdb.CommitRepos: []string{repoKey}}, false, true)
	if err != nil {
		return errors.Wrap(err, "create list commits iterator")
	}
	if err := stream.ForEach[pfsdb.CommitWithID](ctx, iter, func(commitWithID pfsdb.CommitWithID) error {
		return d.finishRepoCommit(ctx, repoPair, &commitWithID)
	}); err != nil {
		return errors.Wrap(err, "list commits")
	}
	for {
		event, ok := <-watcher.Watch()
		if !ok {
			return errors.Errorf("watcher for repo %d %s closed channel", repoPair.ID, repoKey)
		}
		if event.Type == postgres.EventDelete {
			continue
		}
		if event.Err != nil {
			return event.Err
		}
		var commit *pfs.CommitInfo
		if err := dbutil.WithTx(ctx, d.env.DB, func(cbCtx context.Context, tx *pachsql.Tx) error {
			commit, err = pfsdb.GetCommit(ctx, tx, pfsdb.CommitID(event.Id))
			return err
		}); err != nil {
			return errors.Wrap(err, "getting commit from event")
		}
		if err := d.finishRepoCommit(ctx, repoPair, &pfsdb.CommitWithID{ID: pfsdb.CommitID(event.Id), CommitInfo: commit}); err != nil {
			return errors.Wrap(err, "finishing repo commit")
		}
	}
}

func (d *driver) finishRepoCommit(ctx context.Context, repoPair pfsdb.RepoPair, commitWithID *pfsdb.CommitWithID) error {
	commitInfo := commitWithID.CommitInfo
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
				return d.finalizeCommit(ctx, commitWithID, "", nil, nil)
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
				return d.finalizeCommit(ctx, commitWithID, validationError, details, totalId)
			}))
		}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
			log.Error(ctx, "error finishing commit", zap.Error(err), zap.Duration("retryAfter", d))
			return nil
		})
	}, zap.Bool("finishing", true), log.Proto("commit", commitInfo.Commit), zap.Uint64("repo id", uint64(repoPair.ID)), zap.String("repo", repoPair.RepoInfo.Repo.Key()))
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

func (d *driver) finalizeCommit(ctx context.Context, commitWithID *pfsdb.CommitWithID, validationError string, details *pfs.CommitInfo_Details, totalId *fileset.ID) error {
	return log.LogStep(ctx, "finalizeCommit", func(ctx context.Context) error {
		return d.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
			commitInfo := commitWithID.CommitInfo
			commitInfo.Finished = txnCtx.Timestamp
			if details != nil {
				commitInfo.SizeBytesUpperBound = details.SizeBytes
			}
			if commitInfo.Error == "" {
				commitInfo.Error = validationError
			}
			commitInfo.Details = details
			if err := pfsdb.UpdateCommit(ctx, txnCtx.SqlTx, commitWithID.ID, commitInfo, pfsdb.AncestryOpt{SkipParent: true, SkipChildren: true}); err != nil {
				return errors.Wrap(err, "finalize commit")
			}
			if commitInfo.Commit.Repo.Type == pfs.UserRepoType {
				txnCtx.FinishJob(commitInfo)
			}
			if commitInfo.Error == "" {
				return d.triggerCommit(ctx, txnCtx, commitInfo)
			}
			return nil
		})
	})
}
