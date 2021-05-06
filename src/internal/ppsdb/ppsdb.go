// Package ppsdb contains the database schema that PPS uses.
package ppsdb

import (
	"path"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

const (
	pipelinesPrefix = "/pipelines"
	jobsPrefix      = "/jobs"
)

var (
	// JobsPipelineIndex maps pipeline to jobs started by the pipeline
	JobsPipelineIndex = &col.Index{
		Name: "Pipeline",
		Extract: func(val proto.Message) string {
			return val.(*pps.StoredPipelineJobInfo).Pipeline.Name
		},
	}

	// JobsOutputIndex maps job outputs to the job that create them.
	JobsOutputIndex = &col.Index{
		Name: "OutputCommit",
		Extract: func(val proto.Message) string {
			return pfsdb.CommitKey(val.(*pps.StoredPipelineJobInfo).OutputCommit)
		},
	}
)

// Pipelines returns an EtcdCollection of pipelines
func Pipelines(etcdClient *etcd.Client, etcdPrefix string) col.EtcdCollection {
	return col.NewEtcdCollection(
		etcdClient,
		path.Join(etcdPrefix, pipelinesPrefix),
		nil,
		&pps.StoredPipelineInfo{},
		nil,
		nil,
	)
}

// Jobs returns an EtcdCollection of jobs
func Jobs(etcdClient *etcd.Client, etcdPrefix string) col.EtcdCollection {
	return col.NewEtcdCollection(
		etcdClient,
		path.Join(etcdPrefix, jobsPrefix),
		[]*col.Index{JobsPipelineIndex, JobsOutputIndex},
		&pps.StoredPipelineJobInfo{},
		nil,
		nil,
	)
}
