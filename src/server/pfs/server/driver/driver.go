package driver

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/server/pfs/server"
	etcd "go.etcd.io/etcd/client/v3"
	"google.golang.org/protobuf/types/known/anypb"
)

// IsPermissionError returns true if a given error is a permission error.
func IsPermissionError(err error) bool {
	return strings.Contains(err.Error(), "has already finished")
}

type Driver struct {
	Env server.Env

	// EtcdClient and Prefix write repo and other metadata to etcd
	EtcdClient *etcd.Client
	TxnEnv     *txnenv.TransactionEnv
	Prefix     string

	Storage     *storage.Server
	CommitStore server.CommitStore

	cache *fileset.Cache
}

func NewDriver(ctx context.Context, env server.Env) (*Driver, error) {
	// test object Storage.
	if err := func() error {
		ctx, cf := context.WithTimeout(pctx.Child(ctx, "NewDriver"), 30*time.Second)
		defer cf()
		return obj.TestStorage(ctx, env.Bucket)
	}(); err != nil {
		return nil, err
	}

	// Setup Driver struct.
	d := &Driver{
		Env:        env,
		EtcdClient: env.EtcdClient,
		TxnEnv:     env.TxnEnv,
		Prefix:     env.EtcdPrefix,
	}
	storageEnv := storage.Env{
		DB:     env.DB,
		Bucket: env.Bucket,
		Config: env.StorageConfig,
	}
	storageSrv, err := storage.New(ctx, storageEnv)
	if err != nil {
		return nil, err
	}
	d.Storage = storageSrv
	d.CommitStore = server.NewPostgresCommitStore(env.DB, storageSrv.Tracker, storageSrv.Filesets)
	// TODO: Make the cache max size configurable.
	d.cache = fileset.NewCache(env.DB, storageSrv.Tracker, 10000)
	return d, nil
}

func (d *Driver) getPermissionsInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, repo *pfs.Repo) ([]auth.Permission, []string, error) {
	resp, err := d.Env.Auth.GetPermissionsInTransaction(ctx, txnCtx, &auth.GetPermissionsRequest{Resource: repo.AuthResource()})
	if err != nil {
		return nil, nil, errors.EnsureStack(err)
	}

	return resp.Permissions, resp.Roles, nil
}

func (d *Driver) DeleteAll(ctx context.Context) error {
	return d.TxnEnv.WithWriteContext(ctx, func(ctx context.Context, txnCtx *txncontext.TransactionContext) error {
		if _, err := d.deleteReposInTransaction(ctx, txnCtx, nil /* projects */, true /* force */); err != nil {
			return errors.Wrap(err, "could not delete all repos")
		}
		if err := d.listProjectInTransaction(ctx, txnCtx, func(pi *pfs.ProjectInfo) error {
			return errors.Wrapf(d.DeleteProject(ctx, txnCtx, pi.Project, true /* force */), "delete project %q", pi.Project.String())
		}); err != nil {
			return err
		} // now that the cluster is empty, recreate the default project
		return d.createProjectTx(ctx, txnCtx, &pfs.CreateProjectRequest{Project: &pfs.Project{Name: "default"}})
	})
}

func (d *Driver) PutCache(ctx context.Context, key string, value *anypb.Any, fileSetIds []fileset.ID, tag string) error {
	return d.cache.Put(ctx, key, value, fileSetIds, tag)
}

func (d *Driver) GetCache(ctx context.Context, key string) (*anypb.Any, error) {
	return d.cache.Get(ctx, key)
}

func (d *Driver) ClearCache(ctx context.Context, tagPrefix string) error {
	return d.cache.Clear(ctx, tagPrefix)
}

// TODO: Is this really necessary?
type branchSet []*pfs.Branch

func (b *branchSet) search(branch *pfs.Branch) (int, bool) {
	key := pfsdb.BranchKey(branch)
	i := sort.Search(len(*b), func(i int) bool {
		return pfsdb.BranchKey((*b)[i]) >= key
	})
	if i == len(*b) {
		return i, false
	}
	return i, pfsdb.BranchKey((*b)[i]) == pfsdb.BranchKey(branch)
}

func (b *branchSet) add(branch *pfs.Branch) {
	i, ok := b.search(branch)
	if !ok {
		*b = append(*b, nil)
		copy((*b)[i+1:], (*b)[i:])
		(*b)[i] = branch
	}
}

func add(bs *[]*pfs.Branch, branch *pfs.Branch) {
	(*branchSet)(bs).add(branch)
}

func (b *branchSet) has(branch *pfs.Branch) bool {
	_, ok := b.search(branch)
	return ok
}

func has(bs *[]*pfs.Branch, branch *pfs.Branch) bool {
	return (*branchSet)(bs).has(branch)
}

func same(bs []*pfs.Branch, branches []*pfs.Branch) bool {
	if len(bs) != len(branches) {
		return false
	}
	for _, br := range branches {
		if !has(&bs, br) {
			return false
		}
	}
	return true
}
