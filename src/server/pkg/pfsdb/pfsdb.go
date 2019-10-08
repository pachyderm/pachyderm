// Package pfsdb contains the database schema that PFS uses.
package pfsdb

import (
	"fmt"
	"path"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
)

const (
	reposPrefix          = "/repos"
	putFileRecordsPrefix = "/putFileRecords"
	commitsPrefix        = "/commits"
	branchesPrefix       = "/branches"
	openCommitsPrefix    = "/openCommits"
	mergesPrefix         = "/merges"
	shardsPrefix         = "/shards"
)

var (
	// ProvenanceIndex is a secondary index on provenance
	ProvenanceIndex = &col.Index{Field: "Provenance", Multi: true}
)

// Repos returns a collection of repos
func Repos(etcdClient *etcd.Client, etcdPrefix string) col.Collection {
	return col.NewCollection(
		etcdClient,
		path.Join(etcdPrefix, reposPrefix),
		nil,
		&pfs.RepoInfo{},
		nil,
		nil,
	)
}

// PutFileRecords returns a collection of putFileRecords
func PutFileRecords(etcdClient *etcd.Client, etcdPrefix string) col.Collection {
	return col.NewCollection(
		etcdClient,
		path.Join(etcdPrefix, putFileRecordsPrefix),
		nil,
		&pfs.PutFileRecords{},
		nil,
		nil,
	)
}

// Commits returns a collection of commits
func Commits(etcdClient *etcd.Client, etcdPrefix string, repo string) col.Collection {
	return col.NewCollection(
		etcdClient,
		path.Join(etcdPrefix, commitsPrefix, repo),
		[]*col.Index{ProvenanceIndex},
		&pfs.CommitInfo{},
		nil,
		nil,
	)
}

// Branches returns a collection of branches
func Branches(etcdClient *etcd.Client, etcdPrefix string, repo string) col.Collection {
	return col.NewCollection(
		etcdClient,
		path.Join(etcdPrefix, branchesPrefix, repo),
		nil,
		&pfs.BranchInfo{},
		func(key string) error {
			if uuid.IsUUIDWithoutDashes(key) {
				return fmt.Errorf("branch name cannot be a UUID V4")
			}
			return nil
		},
		nil,
	)
}

// OpenCommits returns a collection of open commits
func OpenCommits(etcdClient *etcd.Client, etcdPrefix string) col.Collection {
	return col.NewCollection(
		etcdClient,
		path.Join(etcdPrefix, openCommitsPrefix),
		nil,
		&pfs.Commit{},
		nil,
		nil,
	)
}
