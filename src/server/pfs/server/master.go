package server

import (
	"context"
	"fmt"
	"path"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"go.etcd.io/etcd/client/v3/concurrency"

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
	backoff.RetryUntilCancel(ctx, func() error { //nolint:errcheck
		eg, ctx := errgroup.WithContext(ctx)
		trackerPeriod := time.Second * time.Duration(d.env.StorageConfig.StorageGCPeriod)
		if trackerPeriod <= 0 {
			d.log.Info("Skipping Storage GC")
		} else {
			d.log.Infof("Starting Storage GC with period=%v", trackerPeriod)
			eg.Go(func() error {
				lockPath := path.Join(d.prefix, masterLockPath+"-gc-fileset")
				gcLock := dlock.NewDLock(d.etcdClient, lockPath)
				lockCtx, err := gcLock.Lock(ctx)
				if err != nil {
					return errors.EnsureStack(err)
				}
				defer func() {
					if err := gcLock.Unlock(lockCtx); err != nil {
						log.Errorf("error unlocking fileset gc lock: %v", err)
					}
				}()
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
				lockPath := path.Join(d.prefix, masterLockPath+"-gc-chunk")
				gcLock := dlock.NewDLock(d.etcdClient, lockPath)
				lockCtx, err := gcLock.Lock(ctx)
				if err != nil {
					return errors.EnsureStack(err)
				}
				defer func() {
					if err := gcLock.Unlock(lockCtx); err != nil {
						log.Errorf("error unlocking chunkset gc lock: %v", err)
					}
				}()
				gc := chunk.NewGC(d.storage.ChunkStorage(), chunkPeriod, d.log)
				return gc.RunForever(ctx)
			})
		}
		eg.Go(func() error {
			return d.finishCommits(ctx)
		})
		eg.Go(func() error {
			return d.rebalanceLocks(ctx)
		})
		return errors.EnsureStack(eg.Wait())
	}, backoff.NewInfiniteBackOff(), func(err error, _ time.Duration) error {
		log.Errorf("error in pfs master: %v", err)
		return nil
	})
}

func (d *driver) lenRepoMap() int {
	var length int
	d.repoLocks.Range(func(key, val interface{}) bool {
		length++
		return true
	})
	return length
}

func (d *driver) rebalanceLocks(ctx context.Context) error {
	// @Bryce @Sean How do we know when it is safe to release a lock?
	ticker := time.NewTicker(time.Second * 60) // Assuming we go with this approach, this should probably be configurable
	defer ticker.Stop()

	defer func() { // Getting a kube client panics in tests that use real environments that don't have kubernetes
		if r := recover(); r != nil {
			log.Errorf("no kube client was defined. assuming rebalance is not possible")
		}
	}()
	kubeClient := d.env.GetKubeClient()

	for {
		deployment, err := kubeClient.AppsV1().Deployments(d.env.PachNamespace).
			Get(ctx, "pachd", v1.GetOptions{})
		if err != nil {
			log.Errorf("error getting deployment: %v", err)
			break
		}

		buffer := 1.50 // should this be exposed as config?
		numWorkers := int64(*deployment.Spec.Replicas)
		numRepos, err := d.repos.ReadOnly(ctx).Count()
		if err != nil {
			log.Errorf("error getting num repos: %v", err)
			break
		}
		maxLocksShouldOwn := int(float64(numRepos/numWorkers) * buffer)

		var errs []error
		log.Infof("own %d/%d max locks", d.lenRepoMap(), maxLocksShouldOwn)
		d.repoLocks.Range(func(key, mlock interface{}) bool {
			lock := mlock.(dlock.DLock)
			if d.lenRepoMap() <= maxLocksShouldOwn {
				return false
			}
			if !lock.IsLocked() {
				return true
			}
			if err := lock.Unlock(ctx); err != nil {
				errs = append(errs, err)
				return true
			}
			d.repoLocks.Delete(key)
			log.Infof("released lock: %s", key)
			return true
		})
		if len(errs) != 0 {
			log.Errorf("error releasing locks: %v", errs)
			errs = nil
		}

		repos := map[string]string{}
		repoInfo := &pfs.RepoInfo{}
		d.repos.ReadOnly(ctx).List(repoInfo, col.DefaultOptions(), func(string) error { //nolint:errcheck
			repoKey := fmt.Sprintf("%s.%s", repoInfo.Repo.Name, repoInfo.Repo.Type)
			lockPath := path.Join(d.prefix, masterLockPath, repoKey)
			repos[repoKey] = lockPath
			return nil
		})

		for _, lockPath := range repos {
			if d.lenRepoMap() >= maxLocksShouldOwn {
				break
			}
			if _, ok := d.repoLocks.Load(lockPath); ok {
				continue
			}
			lock := dlock.NewDLock(d.etcdClient, lockPath)
			if _, err := lock.TryLock(ctx); err != nil {
				if err == concurrency.ErrLocked {
					continue
				}
				errs = append(errs, err)
			}
			log.Infof("acquired lock: %s", lockPath)
			d.repoLocks.Store(lockPath, lock)
		}
		if len(errs) != 0 {
			log.Errorf("error acquiring locks: %v", errs)
			errs = nil
		}
		log.Infof("currently owns %d locks", d.lenRepoMap())

		select {
		case <-ctx.Done():
			return errors.EnsureStack(ctx.Err())
		case <-ticker.C:
		}
	}
	return nil
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
		lockPath := path.Join(d.prefix, masterLockPath, key)
		var lock dlock.DLock
		mlock, ok := d.repoLocks.Load(lockPath)
		if ok {
			lock = mlock.(dlock.DLock)
		} else {
			lock = dlock.NewDLock(d.etcdClient, lockPath)
		}
		if ev.Type == watch.EventDelete {
			if lock.IsLocked() {
				if err := lock.Unlock(ctx); err != nil && err != context.Canceled { // Do we want to ignore context.Cancelled?
					log.Errorf("error unlocking lock %s before delete: %v", lockPath, err)
				} else {
					log.Infof("unlocked and deleted lock %s", lockPath)
				}
				d.repoLocks.Delete(lockPath)
			}
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
				lockCtx, err := lock.Lock(ctx)
				if err != nil {
					return errors.EnsureStack(err)
				}
				if _, ok := repos[key]; !ok { // repo was deleted while the lock was blocking.
					return nil
				}
				d.repoLocks.Store(lockPath, lock)
				log.Infof("locked repo lock for %s", lockPath)
				defer func() {
					mlock, ok = d.repoLocks.Load(key)
					if !ok { // lock was safely deleted already.
						return
					}
					lock = mlock.(dlock.DLock)
					if !lock.IsLocked() { // lock is not locked.
						d.repoLocks.Delete(lockPath)
						return
					}
					// Do we want to ignore context.Cancelled?
					if err := lock.Unlock(lockCtx); err != nil && err != context.Canceled {
						log.Errorf("error unlocking lock for %s: %v", lockPath, err)
						return
					}
					d.repoLocks.Delete(lockPath)
					log.Infof("defer call: unlocked lock for %s", lockPath)
				}()
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
