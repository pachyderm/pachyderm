// Package pfsdb contains the database schema that PFS uses.
package pfsdb

import (
	"path"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
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

var (
	reposTable    = "repos"
	branchesTable = "branches"
	commitsTable  = "commits"
)

type repoModel struct {
	Name        string `collection:"primary_key"`
	CreatedAt   *time.Time
	SizeBytes   uint64
	Description string
	Branches    []branchModel
}

func (*repoModel) TableName() string {
	return reposTable
}

type branchModel struct {
	ID               uint   `collection:"primaryKey,autoIncrement"`
	RepoName         string `collection:"index:idx_repo_branch,unique"`
	BranchName       string `collection:"index:idx_repo_branch,unique"`
	HeadCommitID     string
	Provenance       []string
	Subvenance       []string
	DirectProvenance []string
}

func (*branchModel) TableName() string {
	return branchesTable
}

/*
type commitModel struct {
	RepoName string `collection:"primary_key"`
	CommitID string `collection:"primary_key"`
	ParentCommitID string
	BranchName string

	Started *time.Time
	Finished *time.Time
  Description string

	Origin string `sql:"type:originkind"`
	SizeBytes uint64

  repeated Commit child_commits = 11;

  // the commits and their original branches on which this commit is provenant
  repeated CommitProvenance provenance = 16;

  // ReadyProvenance is the number of provenant commits which have been
  // finished, if ReadyProvenance == len(Provenance) then the commit is ready
  // to be processed by pps.
  int64 ready_provenance = 12;

  repeated CommitRange subvenance = 9;
  // this is the block that stores the serialized form of a tree that
  // represents the entire file system hierarchy of the repo at this commit
  // If this is nil, then the commit is either open (in which case 'finished'
  // will also be nil) or is the output commit of a failed job (in which case
  // 'finished' will have a value -- the end time of the job)
  Object tree = 7;
  repeated Object trees = 13;
  Object datums = 14;

  int64 subvenant_commits_success = 18;
  int64 subvenant_commits_failure = 19;
  int64 subvenant_commits_total = 20;
	Sourcetype string `sql:"type:reftype" collection:"primary_key"`
	Source     string `collection:"primary_key"`
	Chunk      string `collection:"primary_key"`
}

func (*commitModel) TableName() string {
	return commitTable
}
*/
