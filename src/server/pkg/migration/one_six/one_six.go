package one_six

import (
	"context"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/pfsdb"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsdb"
)

type Client struct {
	etcdClient *etcd.Client
	Repos      col.Collection
	Commits    func(string) col.Collection
	Branches   func(string) col.Collection
	Pipelines  col.Collection
	Jobs       col.Collection
}

func NewClient(etcdClient *etcd.Client, etcdPrefix string) *Client {
	return &Client{
		etcdClient: etcdClient,
		Repos:      pfsdb.Repos(etcdClient, etcdPrefix),
		Commits:    func(repo string) col.Collection { return pfsdb.Commits(etcdClient, etcdPrefix, repo) },
		Branches:   func(repo string) col.Collection { return pfsdb.Branches(etcdClient, etcdPrefix, repo) },
		Pipelines:  ppsdb.Pipelines(etcdClient, etcdPrefix),
		Jobs:       ppsdb.Jobs(etcdClient, etcdPrefix),
	}
}

type Repo = pfs.Repo
type RepoInfo = pfs.RepoInfo
type Commit = pfs.Commit
type CommitInfo = pfs.CommitInfo
type BranchInfo = pfs.BranchInfo
type PipelineInfo = pps.PipelineInfo
type JobInfo = pps.JobInfo
type Object = pfs.Object

// Migrate extracts pachyderm structures and calls the callback function with
// each one.
func (c *Client) Migrate(ctx context.Context,
	repoF func(*RepoInfo, col.STM) error, branchF func(*BranchInfo, col.STM) error, commitF func(*CommitInfo, col.STM) error,
	pipelineF func(*PipelineInfo, col.STM) error, jobF func(*JobInfo, col.STM) error) error {
	rIt, err := c.Repos.ReadOnly(ctx).List()
	if err != nil {
		return err
	}
	for {
		repoName, repoInfo := "", &RepoInfo{}
		ok, err := rIt.Next(&repoName, repoInfo)
		if err != nil {
			return err
		}
		if !ok {
			break
		}
		if repoF != nil {
			if _, err := col.NewSTM(ctx, c.etcdClient, func(stm col.STM) error {
				if err := c.Repos.ReadWrite(stm).Delete(repoName); err != nil {
					return err
				}
				return repoF(repoInfo, stm)
			}); err != nil {
				return err
			}
		}
		if commitF != nil {
			commits := c.Commits(repoName)
			cIt, err := commits.ReadOnly(ctx).List()
			if err != nil {
				return err
			}
			for {
				commitID, commitInfo := "", &CommitInfo{}
				ok, err := cIt.Next(&commitID, commitInfo)
				if err != nil {
					return err
				}
				if !ok {
					break
				}
				if _, err := col.NewSTM(ctx, c.etcdClient, func(stm col.STM) error {
					if err := commits.ReadWrite(stm).Delete(commitID); err != nil {
						return err
					}
					return commitF(commitInfo, stm)
				}); err != nil {
					return err
				}
			}
		}
		if branchF != nil {
			branches := c.Branches(repoName)
			bIt, err := branches.ReadOnly(ctx).List()
			if err != nil {
				return err
			}
			for {
				branchName, commit := "", &Commit{}
				ok, err := bIt.Next(&branchName, commit)
				if err != nil {
					return err
				}
				if !ok {
					break
				}
				if _, err := col.NewSTM(ctx, c.etcdClient, func(stm col.STM) error {
					if err := branches.ReadWrite(stm).Delete(branchName); err != nil {
						return err
					}
					return branchF(&BranchInfo{
						Head: commit,
						Name: branchName,
					}, stm)
				}); err != nil {
					return err
				}
			}
		}
	}
	if pipelineF != nil {
		pIt, err := c.Pipelines.ReadOnly(ctx).List()
		if err != nil {
			return err
		}
		for {
			pipelineName, pipelineInfo := "", &PipelineInfo{}
			ok, err := pIt.Next(&pipelineName, pipelineInfo)
			if err != nil {
				return err
			}
			if !ok {
				break
			}
			if _, err := col.NewSTM(ctx, c.etcdClient, func(stm col.STM) error {
				if err := c.Pipelines.ReadWrite(stm).Delete(pipelineName); err != nil {
					return err
				}
				return pipelineF(pipelineInfo, stm)
			}); err != nil {
				return err
			}
		}
	}
	if jobF != nil {
		jIt, err := c.Jobs.ReadOnly(ctx).List()
		if err != nil {
			return err
		}
		for {
			jobID, jobInfo := "", &JobInfo{}
			ok, err := jIt.Next(&jobID, jobInfo)
			if err != nil {
				return err
			}
			if !ok {
				break
			}
			if _, err := col.NewSTM(ctx, c.etcdClient, func(stm col.STM) error {
				if err := c.Jobs.ReadWrite(stm).Delete(jobID); err != nil {
					return err
				}
				return jobF(jobInfo, stm)
			}); err != nil {
				return err
			}
		}
	}
	return nil
}
