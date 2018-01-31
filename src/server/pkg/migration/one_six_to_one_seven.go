package migration

import (
	"context"
	"fmt"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/pachyderm/pachyderm/src/client"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/migration/one_seven"
	"github.com/pachyderm/pachyderm/src/server/pkg/migration/one_six"
	"github.com/pachyderm/pachyderm/src/server/pkg/pbutil"
)

func OneSixToOneSeven(etcdAddress, pfsEtcdPrefix, ppsEtcdPrefix string) error {
	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:   []string{fmt.Sprintf("%s:2379", etcdAddress)},
		DialOptions: client.EtcdDialOptions(),
	})
	if err != nil {
		return err
	}
	ctx := context.Background()
	c16 := one_six.NewClient(etcdClient, pfsEtcdPrefix, ppsEtcdPrefix)
	c17 := one_seven.NewClient(etcdClient, pfsEtcdPrefix, ppsEtcdPrefix)
	if err := c16.Migrate(ctx,
		func(repoInfo *one_six.RepoInfo, stm col.STM) error {
			newRepoInfo := &one_seven.RepoInfo{}
			if err := pbutil.Convert(repoInfo, newRepoInfo); err != nil {
				return err
			}
			return c17.Repos.ReadWrite(stm).Put(repoInfo.Repo.Name, newRepoInfo)
		},
		func(commitInfo *one_six.CommitInfo, stm col.STM) error {
			newCommitInfo := &one_seven.CommitInfo{}
			if err := pbutil.Convert(commitInfo, newCommitInfo); err != nil {
				return err
			}
			return c17.Commits(commitInfo.Commit.Repo.Name).ReadWrite(stm).Put(commitInfo.Commit.ID, newCommitInfo)
		},
		func(branchInfo *one_six.BranchInfo, stm col.STM) error {
			newBranchInfo := &one_seven.BranchInfo{}
			if err := pbutil.Convert(branchInfo, newBranchInfo); err != nil {
				return err
			}
			return c17.Branches(branchInfo.Head.Repo.Name).ReadWrite(stm).Put(branchInfo.Name, newBranchInfo)
		},
		func(pipelineInfo *one_six.PipelineInfo, stm col.STM) error {
			return nil
		},
		func(jobInfo *one_six.JobInfo, stm col.STM) error {
			return nil
		},
	); err != nil {
		return err
	}
	return nil
}
