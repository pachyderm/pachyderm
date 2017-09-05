// Package ppsdb contains the database schema that PPS uses.
package ppsdb

import (
	"path"

	etcd "github.com/coreos/etcd/clientv3"

	"github.com/pachyderm/pachyderm/src/client/pps"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
)

const (
	pipelinesPrefix = "/pipelines"
	jobsPrefix      = "/jobs"
)

var (
	// JobsPipelineIndex maps pipeline to jobs started by the pipeline
	JobsPipelineIndex = col.Index{"Pipeline", false}

	// JobsInputIndex maps job inputs (repos + pipeline version) to output
	// commit. This is how we know if we need to start a job.
	JobsInputIndex = col.Index{"Input", false}

	// JobsInputsIndex maps 1.4.5 and earlier style job inputs (repos +
	// pipeline version) to output commit. This is how we know if we need
	// to start a job Needed for legacy compatibility.
	JobsInputsIndex = col.Index{"Inputs", false}

	// JobsOutputIndex maps job outputs to the job that create them.
	JobsOutputIndex = col.Index{"OutputCommit", false}
)

// Pipelines returns a Collection of pipelines
func Pipelines(etcdClient *etcd.Client, etcdPrefix string) col.Collection {
	return col.NewCollection(
		etcdClient,
		path.Join(etcdPrefix, pipelinesPrefix),
		[]col.Index{},
		&pps.PipelineInfo{},
		nil,
	)
}

// Jobs returns a Collection of jobs
func Jobs(etcdClient *etcd.Client, etcdPrefix string) col.Collection {
	return col.NewCollection(
		etcdClient,
		path.Join(etcdPrefix, jobsPrefix),
		[]col.Index{JobsPipelineIndex, JobsInputIndex, JobsOutputIndex},
		&pps.JobInfo{},
		nil,
	)
}
