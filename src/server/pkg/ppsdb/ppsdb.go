// ppsdb contains the database schema that PPS uses.
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
	// Index mapping pipeline to jobs started by the pipeline
	JobsPipelineIndex = col.Index{"Pipeline", false}

	// Index mapping job inputs (repos + pipeline version) to output commit. This
	// is how we know if we need to start a job
	JobsInputIndex = col.Index{"Input", false}

	// Index mapping 1.4.5 and earlier style job inputs (repos + pipeline
	// version) to output commit. This is how we know if we need to start a job
	// Needed for legacy compatibility.
	JobsInputsIndex = col.Index{"Inputs", false}

	// Index of pipelines and jobs that have been stopped (state is "success" or
	// "failure" for jobs, or "stopped" or "failure" for pipelines). See
	// (Job|Pipeline)StateToStopped in s/s/pps/server/api_server.go
	StoppedIndex = col.Index{"Stopped", false}
)

func Pipelines(etcdClient *etcd.Client, etcdPrefix string) col.Collection {
	return col.NewCollection(
		etcdClient,
		path.Join(etcdPrefix, pipelinesPrefix),
		[]col.Index{StoppedIndex},
		&pps.PipelineInfo{},
	)
}

func Jobs(etcdClient *etcd.Client, etcdPrefix string) col.Collection {
	return col.NewCollection(
		etcdClient,
		path.Join(etcdPrefix, jobsPrefix),
		[]col.Index{JobsPipelineIndex, StoppedIndex, JobsInputIndex},
		&pps.JobInfo{},
	)
}
