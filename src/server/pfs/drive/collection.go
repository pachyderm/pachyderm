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
func (d *driver) repos(stm STM) *col.Collection {
	return &collection{
		prefix:     path.Join(d.prefix, reposPrefix),
		etcdClient: d.etcdClient,
		stm:        stm,
	}
}

// repoRefCounts returns a collection of repo reference counters
// Example etcd structure, assuming we have two repos "foo" and "bar":
//   /repoRefCounts
//     /foo
//     /bar
func (d *driver) repoRefCounts(stm STM) *col.IntCollection {
	return &intCollection{
		prefix:     path.Join(d.prefix, repoRefCountsPrefix),
		etcdClient: d.etcdClient,
		stm:        stm,
	}
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
func (d *driver) commits(stm STM) col.CollectionFactory {
	return func(repo string) *collection {
		return &collection{
			prefix:     path.Join(d.prefix, commitsPrefix, repo),
			etcdClient: d.etcdClient,
			stm:        stm,
		}
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
func (d *driver) branches(stm STM) col.CollectionFactory {
	return func(repo string) *collection {
		return &collection{
			prefix:     path.Join(d.prefix, branchesPrefix, repo),
			etcdClient: d.etcdClient,
			stm:        stm,
		}
	}
}

func (d *driver) reposReadonly(ctx context.Context) *col.ReadonlyCollection {
	return &readonlyCollection{
		ctx:        ctx,
		prefix:     path.Join(d.prefix, reposPrefix),
		etcdClient: d.etcdClient,
	}
}

func (d *driver) commitsReadonly(ctx context.Context) *col.ReadonlyCollectionFactory {
	return func(repo string) *readonlyCollection {
		return &readonlyCollection{
			ctx:        ctx,
			prefix:     path.Join(d.prefix, commitsPrefix, repo),
			etcdClient: d.etcdClient,
		}
	}
}

func (d *driver) branchesReadonly(ctx context.Context) col.ReadonlyCollectionFactory {
	return func(repo string) *readonlyCollection {
		return &readonlyCollection{
			ctx:        ctx,
			prefix:     path.Join(d.prefix, branchesPrefix, repo),
			etcdClient: d.etcdClient,
		}
	}
}
