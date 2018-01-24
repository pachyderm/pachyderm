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

func oneSixToOneSeven(etcdAddress string, etcdPrefix string) error {
	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:   []string{fmt.Sprintf("%s:2379", etcdAddress)},
		DialOptions: client.EtcdDialOptions(),
	})
	if err != nil {
		return err
	}
	ctx := context.Background()
	c16 := one_six.NewClient(etcdClient, etcdPrefix)
	c17 := one_seven.NewClient(etcdClient, etcdPrefix)
	if err := c16.Migrate(ctx,
		nil, // Repos are unchanged
		func(commitInfo *one_six.CommitInfo, stm col.STM) error {
			commit := &one_seven.Commit{}
			if err := pbutil.Convert(commitInfo.Commit, commit); err != nil {
				return err
			}
			var provenance []*one_seven.Commit
			for _, provCommit := range commitInfo.Provenance {
				newProvCommit := &one_seven.Commit{}
				if err := pbutil.Convert(provCommit, newProvCommit); err != nil {
					return err
				}
				provenance = append(provenance, newProvCommit)
			}
			tree := &one_seven.Object{}
			if err := pbutil.Convert(commitInfo.Tree, tree); err != nil {
				return err
			}
			if err := c17.Commits(commitInfo.Commit.Repo.Name).ReadWrite(stm).Put(commitInfo.Commit.ID, &one_seven.CommitInfo{
				Commit:      commit,
				Description: commitInfo.Description,
				Started:     commitInfo.Started,
				Finished:    commitInfo.Finished,
				SizeBytes:   commitInfo.SizeBytes,
				Provenance:  provenance,
				Tree:        tree,
			}); err != nil {
				return err
			}
			return nil
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
