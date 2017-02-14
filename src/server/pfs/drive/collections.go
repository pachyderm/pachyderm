package drive

import (
	"context"
	"path"

	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
)

const (
	reposPrefix         = "/repos"
	repoRefCountsPrefix = "/repoRefCounts"
	commitsPrefix       = "/commits"
	branchesPrefix      = "/branches"
)

// repos returns a collection of repos
// Example etcd structure, assuming we have two repos "foo" and "bar":
//   /repos
//     /foo
//     /bar
func (d *driver) repos(stm col.STM) *col.Collection {
	return col.NewCollection(
		d.etcdClient,
		path.Join(d.prefix, reposPrefix),
		stm,
	)
}

// repoRefCounts returns a collection of repo reference counters
// Example etcd structure, assuming we have two repos "foo" and "bar":
//   /repoRefCounts
//     /foo
//     /bar
func (d *driver) repoRefCounts(stm col.STM) *col.IntCollection {
	return col.NewIntCollection(
		d.etcdClient,
		path.Join(d.prefix, repoRefCountsPrefix),
		stm,
	)
}

// commits returns a collection of commits
// Example etcd structure, assuming we have two repos "foo" and "bar":
//   /commits
//     /foo
//       /UUID1
//       /UUID2
//     /bar
//       /UUID3
//       /UUID4
func (d *driver) commits(stm col.STM) col.CollectionFactory {
	return func(repo string) *col.Collection {
		return col.NewCollection(
			d.etcdClient,
			path.Join(d.prefix, commitsPrefix, repo),
			stm,
		)
	}
}

// branches returns a collection of branches
// Example etcd structure, assuming we have two repos "foo" and "bar",
// each of which has two branches:
//   /branches
//     /foo
//       /master
//       /test
//     /bar
//       /master
//       /test
func (d *driver) branches(stm col.STM) col.CollectionFactory {
	return func(repo string) *col.Collection {
		return col.NewCollection(
			d.etcdClient,
			path.Join(d.prefix, branchesPrefix, repo),
			stm,
		)
	}
}

func (d *driver) reposReadonly(ctx context.Context) *col.ReadonlyCollection {
	return col.NewReadonlyCollection(
		ctx,
		d.etcdClient,
		path.Join(d.prefix, reposPrefix),
	)
}

func (d *driver) commitsReadonly(ctx context.Context) col.ReadonlyCollectionFactory {
	return func(repo string) *col.ReadonlyCollection {
		return col.NewReadonlyCollection(
			ctx,
			d.etcdClient,
			path.Join(d.prefix, commitsPrefix, repo),
		)
	}
}

func (d *driver) branchesReadonly(ctx context.Context) col.ReadonlyCollectionFactory {
	return func(repo string) *col.ReadonlyCollection {
		return col.NewReadonlyCollection(
			ctx,
			d.etcdClient,
			path.Join(d.prefix, branchesPrefix, repo),
		)
	}
}
