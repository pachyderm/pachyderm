package server

import (
	"context"
	"fmt"
	"path"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/pachyderm/pachyderm/v2/src/pfs"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/consistenthashing"
	"github.com/pachyderm/pachyderm/v2/src/internal/cronutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/dlock"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/middleware/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/protoutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/randutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
)

const (
	masterLockPath = "pfs-master-lock"
)

type Master struct {
	env Env
	*apiServer
}

func NewMaster(ctx context.Context, env Env) (*Master, error) {
	// test object storage.
	if err := func() error {
		ctx, cf := context.WithTimeout(pctx.Child(ctx, "newDriver"), 30*time.Second)
		defer cf()
		return obj.TestStorage(ctx, env.Bucket)
	}(); err != nil {
		return nil, err
	}
	a, err := newAPIServer(ctx, env)
	if err != nil {
		return nil, err
	}
	return &Master{
		env:       env,
		apiServer: a,
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
				lock := dlock.NewDLock(m.etcdClient, path.Join(m.prefix, masterLockPath, "storage-gc"))
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
				gc := m.storage.Filesets.NewGC(trackerPeriod)
				return gc.RunForever(pctx.Child(ctx, "storage-gc"))
			})
		}
		chunkPeriod := time.Second * time.Duration(m.env.StorageConfig.StorageChunkGCPeriod)
		if chunkPeriod <= 0 {
			log.Info(ctx, "Skipping Chunk Storage GC")
		} else {
			eg.Go(func() error {
				lock := dlock.NewDLock(m.etcdClient, path.Join(m.prefix, masterLockPath, "chunk-gc"))
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
				gc := chunk.NewGC(m.storage.Chunks, chunkPeriod)
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
	ringPrefix := path.Join(randutil.UniqueString(m.prefix), masterLockPath, "ring")
	return consistenthashing.WithRing(ctx, m.etcdClient, ringPrefix,
		func(ctx context.Context, ring *consistenthashing.Ring) error {
			return pfsdb.WatchRepos(ctx, m.env.DB, m.env.Listener,
				func(id pfsdb.RepoID, repoInfo *pfs.RepoInfo) error {
					// On Upsert.
					lockPrefix := fmt.Sprintf("repos/%d", id)
					ctx, cancel := pctx.WithCancel(ctx)
					repos[id] = cancel
					go m.manageRepo(ctx, ring, pfsdb.Repo{ID: id, RepoInfo: repoInfo}, lockPrefix)
					return nil
				},
				func(id pfsdb.RepoID) error {
					// On Delete.
					cancel, ok := repos[id]
					if !ok {
						return nil
					}
					defer func() {
						cancel()
						delete(repos, id)
					}()
					return ring.Unlock(fmt.Sprintf("repos/%d", id))
				})
		})
}

func (m *Master) manageRepo(ctx context.Context, ring *consistenthashing.Ring, repo pfsdb.Repo, lockPrefix string) {
	key := pfsdb.RepoKey(repo.RepoInfo.Repo)
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
				return m.manageBranches(ctx, repo)
			}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
				log.Error(ctx, "managing branches", zap.String("repo", key), zap.Error(err), zap.Duration("retryAfter", d))
				return nil
			})
		})
		eg.Go(func() error {
			return backoff.RetryUntilCancel(ctx, func() error {
				return m.watchCommitsByRepo(ctx, repo)
			}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
				log.Error(ctx, "finishing repo commits", zap.Uint64("repo id", uint64(repo.ID)), zap.String("repo", key), zap.Error(err), zap.Duration("retryAfter", d))
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

func (m *Master) manageBranches(ctx context.Context, repo pfsdb.Repo) error {
	ctx, cancel := pctx.WithCancel(ctx)
	defer cancel()
	cronTriggers := make(map[pfsdb.BranchID]*cronTrigger)
	defer func() {
		for _, ct := range cronTriggers {
			ct.cancel()
		}
	}()
	return pfsdb.WatchBranchesInRepo(ctx, m.env.DB, m.env.Listener, repo.ID,
		func(id pfsdb.BranchID, branchInfo *pfs.BranchInfo) error {
			return m.manageBranch(ctx, pfsdb.Branch{ID: id, BranchInfo: branchInfo}, cronTriggers)
		},
		func(id pfsdb.BranchID) error {
			if ct, ok := cronTriggers[id]; ok {
				ct.cancel()
				delete(cronTriggers, id)
			}
			return nil
		})
}

func (m *Master) manageBranch(ctx context.Context, branch pfsdb.Branch, cronTriggers map[pfsdb.BranchID]*cronTrigger) error {
	branchInfo := branch.BranchInfo
	key := branch.ID
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
			return m.runCronTrigger(ctx, branchInfo.Branch)
		}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
			log.Error(ctx, "error running cron trigger", zap.Uint64("branch id", uint64(branch.ID)), zap.String("branch", branchInfo.Branch.Key()), zap.Error(err), zap.Duration("retryAfter", d))
			return nil
		})
	}()
	return nil
}

func (m *Master) runCronTrigger(ctx context.Context, branch *pfs.Branch) error {
	branchInfo, err := m.inspectBranch(ctx, branch)
	if err != nil {
		return err
	}
	schedule, err := cronutil.ParseCronExpression(branchInfo.Trigger.CronSpec)
	if err != nil {
		return err
	}
	// Use the current head commit start time as the previous tick.
	// This prevents the timer from restarting if the master restarts.
	ci, err := m.resolveCommit(ctx, branchInfo.Head)
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
		if err := m.txnEnv.WithWriteContext(ctx, func(ctx context.Context, txnCtx *txncontext.TransactionContext) error {
			trigBI, err := m.inspectBranchInTransaction(ctx, txnCtx, branchInfo.Branch.Repo.NewBranch(branchInfo.Trigger.Branch))
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

func (m *Master) watchCommitsByRepo(ctx context.Context, repo pfsdb.Repo) error {
	return pfsdb.WatchCommitsInRepo(ctx, m.env.DB, m.env.Listener, repo.ID,
		func(commit pfsdb.Commit) error {
			c := commit
			return m.postProcessCommit(ctx, repo, &c)
		},
		func(_ pfsdb.CommitID) error {
			// Don't finish commits that are deleted.
			return nil
		},
	)
}

func (m *Master) postProcessCommit(ctx context.Context, repo pfsdb.Repo, commit *pfsdb.Commit) error {
	if commit.Finishing == nil || commit.Finished != nil {
		return nil
	}
	cache := m.newCache(commit.Commit.Key())
	defer func() {
		if err := cache.clear(ctx); err != nil {
			log.Error(ctx, "errored clearing compaction cache", zap.Error(err))
		}
	}()
	return log.LogStep(ctx, "postProcessCommit", func(ctx context.Context) error {
		return backoff.RetryUntilCancel(ctx, func() error {
			// In the case where a commit is squashed between retries (commit not found), the master can return.
			if _, err := m.resolveCommit(ctx, commit.Commit); err != nil {
				if pfsserver.IsCommitNotFoundErr(err) {
					return nil
				}
				return err
			}
			// Skip compaction / validation for errored commits.
			if commit.Error != "" {
				return m.finishCommit(ctx, commit, "", nil)
			}
			return m.compactAndValidateCommit(ctx, commit, cache)
		}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
			log.Error(ctx, "error post-processing commit", zap.Error(err), zap.Duration("retryAfter", d))
			return nil
		})
	}, zap.Bool("finishing", true), log.Proto("commit", commit.Commit), zap.Uint64("repo id", uint64(repo.ID)), zap.String("repo", repo.RepoInfo.Repo.Key()))
}

func (m *Master) compactAndValidateCommit(ctx context.Context, commit *pfsdb.Commit, cache *cache) error {
	compactor := newCompactor(m.storage.Filesets, m.env.StorageConfig.StorageCompactionMaxFanIn)
	taskDoer := m.env.TaskService.NewDoer(StorageTaskNamespace, commit.Commit.Id, cache)
	return errors.Wrap(m.storage.Filesets.WithRenewer(ctx, defaultTTL, func(ctx context.Context, renewer *fileset.Renewer) error {
		resp, err := m.compactCommit(ctx, compactor, taskDoer, renewer, commit)
		if err != nil {
			return errors.Wrap(err, "post process commit")
		}
		if err = m.validateCommit(ctx, taskDoer, resp.totalId, compactor, resp.CommitInfo_Details); err != nil {
			return m.finishCommit(ctx, commit, err.Error(), resp.CommitInfo_Details)
		}
		return m.finishCommit(ctx, commit, "", resp.CommitInfo_Details)
	}), "finishing commit with renewer")
}

type compactResp struct {
	*pfs.CommitInfo_Details
	diffId, totalId *fileset.ID
}

func (m *Master) compactCommit(ctx context.Context, compactor *compactor, taskDoer task.Doer, renewer *fileset.Renewer, commit *pfsdb.Commit) (*compactResp, error) {
	start := time.Now()
	var (
		diffId  *fileset.ID // currently unused, but may be needed later.
		totalId *fileset.ID
		err     error
	)
	// Compacting the diff before getting the total allows us to compose the
	// total file set so that it includes the compacted diff.
	if err := log.LogStep(ctx, "compactDiffFileSet", func(ctx context.Context) error {
		diffId, err = m.compactDiffFileSet(ctx, compactor, taskDoer, renewer, commit)
		return err
	}); err != nil {
		return nil, errors.Wrap(err, "compact commit")
	}
	if err := log.LogStep(ctx, "compactTotalFileSet", func(ctx context.Context) error {
		totalId, err = m.compactTotalFileSet(ctx, compactor, taskDoer, renewer, commit)
		return err
	}); err != nil {
		return nil, errors.Wrap(err, "compact commit")
	}
	return &compactResp{
		CommitInfo_Details: &pfs.CommitInfo_Details{
			CompactingTime: durationpb.New(time.Since(start)),
		},
		diffId:  diffId,
		totalId: totalId,
	}, nil
}

func (m *Master) validateCommit(ctx context.Context, taskDoer task.Doer, totalId *fileset.ID, compactor *compactor, details *pfs.CommitInfo_Details) error {
	start := time.Now()
	var validationError string
	if err := log.LogStep(ctx, "validateCommit", func(ctx context.Context) error {
		var err error
		validationError, details.SizeBytes, err = compactor.Validate(ctx, taskDoer, *totalId)
		return err
	}); err != nil {
		return errors.Wrap(err, "validate commit")
	}
	details.ValidatingTime = durationpb.New(time.Since(start))
	if validationError != "" {
		return errors.New(validationError)
	}
	return nil
}

func (m *Master) finishCommit(ctx context.Context, commit *pfsdb.Commit, validationError string, details *pfs.CommitInfo_Details) error {
	return log.LogStep(ctx, "finishCommit", func(ctx context.Context) error {
		return m.txnEnv.WithWriteContext(ctx, func(ctx context.Context, txnCtx *txncontext.TransactionContext) error {
			commit.Finished = txnCtx.Timestamp
			if details == nil {
				details = &pfs.CommitInfo_Details{SizeBytes: commit.SizeBytesUpperBound}
			}
			commit.SizeBytesUpperBound = details.SizeBytes
			commit.Details = details
			if commit.Error == "" {
				commit.Error = validationError
			}
			if err := pfsdb.FinishCommit(ctx, txnCtx.SqlTx, commit.ID, commit.Finished, commit.Error, commit.Details); err != nil {
				return errors.Wrap(err, "finalize commit")
			}
			if commit.Commit.Repo.Type == pfs.UserRepoType {
				txnCtx.FinishJob(commit.CommitInfo)
			}
			if commit.Error == "" {
				return m.triggerCommit(ctx, txnCtx, commit.CommitInfo)
			}
			return nil
		})
	})
}
