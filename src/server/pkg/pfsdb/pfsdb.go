// Package ppsdb contains the database schema that PFS uses.
package pfsdb

import (
	"path"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
)

const (
	reposPrefix         = "/repos"
	repoRefCountsPrefix = "/repoRefCounts"
	commitsPrefix       = "/commits"
	branchesPrefix      = "/branches"
)

var (
	ProvenanceIndex = col.Index{"Provenance", true}
)

func Repos(etcdClient *etcd.Client, etcdPrefix string) col.Collection {
	return col.NewCollection(
		etcdClient,
		path.Join(etcdPrefix, reposPrefix),
		[]col.Index{ProvenanceIndex},
		&pfs.RepoInfo{},
	)
}

func RepoRefCounts(etcdClient *etcd.Client, etcdPrefix string) col.Collection {
	return col.NewCollection(
		etcdClient,
		path.Join(etcdPrefix, repoRefCountsPrefix),
		nil,
		nil,
	)
}

func Commits(etcdClient *etcd.Client, etcdPrefix string, repo string) col.Collection {
	return col.NewCollection(
		etcdClient,
		path.Join(etcdPrefix, commitsPrefix, repo),
		[]col.Index{ProvenanceIndex},
		&pfs.CommitInfo{},
	)
}

func Branches(etcdClient *etcd.Client, etcdPrefix string, repo string) col.Collection {
	return col.NewCollection(
		etcdClient,
		path.Join(etcdPrefix, branchesPrefix, repo),
		nil,
		&pfs.Commit{},
	)
}
