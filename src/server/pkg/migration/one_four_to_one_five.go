package migration

import (
	"context"
	"fmt"
	"path"

	"migration/onefoureight/db/pfs"
	"migration/onefoureight/db/pps"

	"github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/src/client"

	etcd "github.com/coreos/etcd/clientv3"
	"go.pedge.io/lion/proto"
)

const (
	reposPrefix     = "/repos"
	commitsPrefix   = "/commits"
	branchesPrefix  = "/branches"
	pipelinesPrefix = "/pipelines"
	jobsPrefix      = "/jobs"
)

func oneFourToOneFive(etcdAddress, pfsPrefix, ppsPrefix string) error {
	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:   []string{fmt.Sprintf("%s:2379", etcdAddress)},
		DialOptions: client.EtcdDialOptions(),
	})
	if err != nil {
		return fmt.Errorf("error constructing etcdClient: %v", err)
	}

	// This function migrates objects under a specific prefix
	migrate := func(prefix string, template proto.Message) {
		resp, err := etcdClient.Get(context.Background(), prefix, etcd.WithPrefix())
		if err != nil {
			protolion.Errorf("error getting %v: %v", prefix, err)
			return
		}
		for _, kv := range resp.Kvs {
			key := string(kv.Key)
			if err := proto.UnmarshalText(string(kv.Value), template); err != nil {
				protolion.Errorf("error unmarshalling object %v: %v", key, err)
				continue
			}
			bytes, err := proto.Marshal(template)
			if err != nil {
				protolion.Errorf("error marshalling object %v: %v", key, err)
				continue
			}
			if _, err := etcdClient.Put(context.Background(), key, string(bytes)); err != nil {
				protolion.Errorf("error putting object %v: %v", key, err)
				continue
			}
		}
	}

	var repoInfo pfs.RepoInfo
	migrate(path.Join(pfsPrefix, reposPrefix), &repoInfo)
	protolion.Infof("finished migrating repos")

	var commitInfo pfs.CommitInfo
	migrate(path.Join(pfsPrefix, commitsPrefix), &commitInfo)
	protolion.Infof("finished migrating commits")

	var head pfs.Commit
	migrate(path.Join(pfsPrefix, branchesPrefix), &head)
	protolion.Infof("finished migrating branches")

	var pipelineInfo pps.PipelineInfo
	migrate(path.Join(ppsPrefix, pipelinesPrefix), &pipelineInfo)
	protolion.Infof("finished migrating pipelines")

	var jobInfo pps.JobInfo
	migrate(path.Join(ppsPrefix, jobsPrefix), &jobInfo)
	protolion.Infof("finished migrating jobs")

	return nil
}
