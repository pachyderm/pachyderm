package server

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/dlock"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	_ "github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/middleware/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/internal/watch"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"

	"github.com/gogo/protobuf/types"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

const (
	masterLockPath = "pfs-master-lock"
)

func (d *driver) master(ctx context.Context) {
	ctx = auth.AsInternalUser(ctx, "pfs-master")
	masterLock := dlock.NewDLock(d.etcdClient, path.Join(d.prefix, masterLockPath))
	backoff.RetryUntilCancel(ctx, func() error {
		masterCtx, err := masterLock.Lock(ctx)
		if err != nil {
			return errors.EnsureStack(err)
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
			backoff.RetryUntilCancel(ctx, func() error {
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
		return miscutil.LogStep(fmt.Sprintf("finishing commit %v", commit.ID), func() error {
			// TODO: This retry might not be getting us much if the outer watch still errors due to a transient error.
			return backoff.RetryUntilCancel(ctx, func() error {
				id, err := d.getFileSet(ctx, commit)
				if err != nil {
					if pfsserver.IsCommitNotFoundErr(err) {
						return nil
					}
					return err
				}
				// Compact the commit.
				taskDoer := d.env.TaskService.NewDoer(storageTaskNamespace, commit.ID)
				var totalId *fileset.ID
				start := time.Now()
				if err := miscutil.LogStep(fmt.Sprintf("compacting commit %v", commit), func() error {
					var err error
					totalId, err = compactor.Compact(ctx, taskDoer, []fileset.ID{*id}, defaultTTL)
					if err != nil {
						return err
					}
					return errors.EnsureStack(d.commitStore.SetTotalFileSet(ctx, commit, *totalId))
				}); err != nil {
					return err
				}
				compactingDuration := time.Since(start)
				// Validate the commit.
				start = time.Now()
				var size int64
				var validationError string
				if err := miscutil.LogStep(fmt.Sprintf("validating commit %v", commit), func() error {
					var err error
					size, validationError, err = d.validate(ctx, totalId)
					return err
				}); err != nil {
					return err
				}
				validatingDuration := time.Since(start)
				// Finish the commit.
				return miscutil.LogStep(fmt.Sprintf("finish commit %v", commit), func() error {
					return d.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
						commitInfo := &pfs.CommitInfo{}
						if err := d.commits.ReadWrite(txnCtx.SqlTx).Update(pfsdb.CommitKey(commit), commitInfo, func() error {
							commitInfo.Finished = txnCtx.Timestamp
							commitInfo.SizeBytesUpperBound = size
							if commitInfo.Details == nil {
								commitInfo.Details = &pfs.CommitInfo_Details{}
							}
							commitInfo.Details.SizeBytes = size
							if commitInfo.Error == "" {
								commitInfo.Error = validationError
							}
							commitInfo.Details.CompactingTime = types.DurationProto(compactingDuration)
							commitInfo.Details.ValidatingTime = types.DurationProto(validatingDuration)
							return nil
						}); err != nil {
							return errors.EnsureStack(err)
						}
						if commitInfo.Commit.Branch.Repo.Type == pfs.UserRepoType {
							txnCtx.FinishJob(commitInfo)
						}
						if err := d.finishAliasDescendents(txnCtx, commitInfo, *totalId); err != nil {
							return err
						}
						// TODO(2.0 optional): This is a hack to ensure that commits created by triggers have the same ID as the finished commit.
						// Creating an alias in the branch of the finished commit, then having the trigger alias that commit seems
						// like a better model.
						txnCtx.CommitSetID = commitInfo.Commit.ID
						return d.triggerCommit(txnCtx, commitInfo.Commit)
					})
				})
			}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
				log.Errorf("error finishing commit %v: %v, retrying in %v", commit, err, d)
				return nil
			})
		})
	}, watch.IgnoreDelete)
	return errors.EnsureStack(err)
}

// TODO(2.0 optional): Improve the performance of this by doing a logarithmic lookup per new file,
// rather than a linear scan through all of the files.
func (d *driver) validate(ctx context.Context, id *fileset.ID) (int64, string, error) {
	fs, err := d.storage.Open(ctx, []fileset.ID{*id})
	if err != nil {
		return 0, "", err
	}
	var prev *index.Index
	var size int64
	var validationError string
	if err := fs.Iterate(ctx, func(f fileset.File) error {
		idx := f.Index()
		if prev != nil && validationError == "" {
			if idx.Path == prev.Path {
				validationError = fmt.Sprintf("duplicate path output by different datums (%v from %v and %v from %v)", prev.Path, prev.File.Datum, idx.Path, idx.File.Datum)
			} else if strings.HasPrefix(idx.Path, prev.Path+"/") {
				validationError = fmt.Sprintf("file / directory path collision (%v)", idx.Path)
			}
		}
		prev = idx
		size += index.SizeBytes(idx)
		return nil
	}); err != nil {
		return 0, "", errors.EnsureStack(err)
	}
	return size, validationError, nil
}

// finishAliasDescendents will traverse the given commit's descendents, finding all
// contiguous aliases and finishing them.
func (d *driver) finishAliasDescendents(txnCtx *txncontext.TransactionContext, parentCommitInfo *pfs.CommitInfo, id fileset.ID) error {
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
			if err := d.commitStore.SetTotalFileSetTx(txnCtx.SqlTx, commitInfo.Commit, id); err != nil {
				return errors.EnsureStack(err)
			}

			descendents = append(descendents, commitInfo.ChildCommits...)
		}
	}
	return nil
}
