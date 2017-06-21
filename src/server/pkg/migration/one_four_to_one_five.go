package migration

import (
	"context"
	"fmt"
	"path"

	col "migration/onefoureight/collection"
	"migration/onefoureight/db/pfs"
	"migration/onefoureight/db/pps"

	"github.com/pachyderm/pachyderm/src/client"

	etcd "github.com/coreos/etcd/clientv3"
	"go.pedge.io/lion/proto"
)

const (
	reposPrefix         = "/repos"
	repoRefCountsPrefix = "/repoRefCounts"
	commitsPrefix       = "/commits"
	branchesPrefix      = "/branches"
)

func repos(etcdClient *etcd.Client, etcdPrefix string) col.Collection {
	return col.NewCollection(
		etcdClient,
		path.Join(etcdPrefix, reposPrefix),
		nil,
		&pfs.RepoInfo{},
	)
}

func repoRefCounts(etcdClient *etcd.Client, etcdPrefix string) col.Collection {
	return col.NewCollection(
		etcdClient,
		path.Join(etcdPrefix, repoRefCountsPrefix),
		nil,
		nil,
	)
}

func commits(etcdClient *etcd.Client, etcdPrefix string, repo string) col.Collection {
	return col.NewCollection(
		etcdClient,
		path.Join(etcdPrefix, commitsPrefix, repo),
		nil,
		&pfs.CommitInfo{},
	)
}

func branches(etcdClient *etcd.Client, etcdPrefix string, repo string) col.Collection {
	return col.NewCollection(
		etcdClient,
		path.Join(etcdPrefix, branchesPrefix, repo),
		nil,
		&pfs.Commit{},
	)
}

const (
	pipelinesPrefix = "/pipelines"
	jobsPrefix      = "/jobs"
)

func pipelines(etcdClient *etcd.Client, etcdPrefix string) col.Collection {
	return col.NewCollection(
		etcdClient,
		path.Join(etcdPrefix, pipelinesPrefix),
		nil,
		&pps.PipelineInfo{},
	)
}

func jobs(etcdClient *etcd.Client, etcdPrefix string) col.Collection {
	return col.NewCollection(
		etcdClient,
		path.Join(etcdPrefix, jobsPrefix),
		nil,
		&pps.JobInfo{},
	)
}

func oneFourToOneFive(etcdAddress, pfsPrefix, ppsPrefix string) error {
	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:   []string{fmt.Sprintf("%s:2379", etcdAddress)},
		DialOptions: client.EtcdDialOptions(),
	})
	if err != nil {
		return fmt.Errorf("error constructing etcdClient: %v", err)
	}

	repos := repos(etcdClient, pfsPrefix)
	iter, err := repos.ReadOnly(context.Background()).List()
	if err != nil {
		protolion.Errorf("error obtaining an iterator of repos: %v", err)
	} else {
		var key string
		var repoInfo pfs.RepoInfo
		for {
			ok, err := iter.Next(&key, &repoInfo)
			if err != nil {
				protolion.Errorf("error deserializing repo: %v", err)
			}
			if !ok {
				break
			}
			if _, err := col.NewSTM(context.Background(), etcdClient, func(stm col.STM) error {
				reposWriter := repos.ReadWrite(stm)
				// The vendored collection library has been modified such that
				// Put serializes the object to bytes
				reposWriter.Put(key, &repoInfo)
				return nil
			}); err != nil {
				protolion.Errorf("error updating object %v: %v", key, err)
			}
		}
	}
	return nil
}
