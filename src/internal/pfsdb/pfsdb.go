// Package pfsdb contains the database schema that PFS uses.
package pfsdb

import (
	"path"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

const (
	reposPrefix       = "/repos"
	commitsPrefix     = "/commits"
	branchesPrefix    = "/branches"
	openCommitsPrefix = "/openCommits"
	mergesPrefix      = "/merges"
	shardsPrefix      = "/shards"
)

// Repos returns a collection of repos
func Repos(etcdClient *etcd.Client, etcdPrefix string) col.EtcdCollection {
	return col.NewEtcdCollection(
		etcdClient,
		path.Join(etcdPrefix, reposPrefix),
		nil,
		&pfs.RepoInfo{},
		nil,
		nil,
	)
}

var CommitsRepoIndex = &col.Index{
	Name: "repo",
	Extract: func(val proto.Message) string {
		return val.(*pfs.CommitInfo).Commit.Branch.Repo.Name
	},
}

func CommitKey(commit *pfs.Commit) string {
	return commit.Branch.Repo.Name + "@" + commit.Branch.Name + "=" + commit.ID
}

// Commits returns a collection of commits
func Commits(etcdClient *etcd.Client, etcdPrefix string, repo string) col.EtcdCollection {
	return col.NewEtcdCollection(
		etcdClient,
		path.Join(etcdPrefix, commitsPrefix, repo),
		nil,
		&pfs.CommitInfo{},
		nil,
		nil,
	)
}

var BranchesRepoIndex = &col.Index{
	Name: "repo",
	Extract: func(val proto.Message) string {
		return val.(*pfs.BranchInfo).Branch.Repo.Name
	},
}

func BranchKey(branch *pfs.Branch) string {
	return branch.Repo.Name + "@" + branch.Name
}

// Branches returns a collection of branches
func Branches(etcdClient *etcd.Client, etcdPrefix string, repo string) col.EtcdCollection {
	return col.NewEtcdCollection(
		etcdClient,
		path.Join(etcdPrefix, branchesPrefix, repo),
		nil,
		&pfs.BranchInfo{},
		func(key string) error {
			if uuid.IsUUIDWithoutDashes(key) {
				return errors.Errorf("branch name cannot be a UUID V4")
			}
			return nil
		},
		nil,
	)
}

// OpenCommits returns a collection of open commits
func OpenCommits(etcdClient *etcd.Client, etcdPrefix string) col.EtcdCollection {
	return col.NewEtcdCollection(
		etcdClient,
		path.Join(etcdPrefix, openCommitsPrefix),
		nil,
		&pfs.Commit{},
		nil,
		nil,
	)
}
