// Package pfsdb contains the database schema that PFS uses.
package pfsdb

import (
	"path"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/jinzhu/gorm"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
)

const (
	reposPrefix          = "/repos"
	putFileRecordsPrefix = "/putFileRecords"
	commitsPrefix        = "/commits"
	branchesPrefix       = "/branches"
	openCommitsPrefix    = "/openCommits"
)

// Repos returns a collection of repos
func Repos(etcdClient *etcd.Client, etcdPrefix string) col.Collection {
	return col.NewEtcdCollection(
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
	return col.NewEtcdCollection(
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
	return col.NewEtcdCollection(
		etcdClient,
		path.Join(etcdPrefix, commitsPrefix, repo),
		[]*col.Index{{Field: "Provenance", Multi: true}},
		&pfs.CommitInfo{},
		nil,
		nil,
	)
}

// Branches returns a collection of branches
func Branches(etcdClient *etcd.Client, etcdPrefix string, repo string) col.Collection {
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
func OpenCommits(etcdClient *etcd.Client, etcdPrefix string) col.Collection {
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
	Name        string `gorm:"primary_key"`
	CreatedAt   *time.Time
	SizeBytes   uint64
	Description string
	Branches    []branchModel
}

func (*repoModel) TableName() string {
	return reposTable
}

type branchModel struct {
	ID               uint   `gorm:"primaryKey,autoIncrement"`
	RepoName         string `gorm:"index:idx_repo_branch,unique"`
	BranchName       string `gorm:"index:idx_repo_branch,unique"`
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
	RepoName string `gorm:"primary_key"`
	CommitID string `gorm:"primary_key"`
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
	Sourcetype string `sql:"type:reftype" gorm:"primary_key"`
	Source     string `gorm:"primary_key"`
	Chunk      string `gorm:"primary_key"`
}

func (*commitModel) TableName() string {
	return commitTable
}
*/

func initializeDb(db *gorm.DB) error {
	if err := db.AutoMigrate(&repoModel{}, &branchModel{}).Error; err != nil {
		return err
	}

	return nil
}
