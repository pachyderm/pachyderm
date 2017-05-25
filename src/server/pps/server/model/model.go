// Package model provides helper functions for operating on the data that
// PPS stores.
package model

import (
	"path"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/pachyderm/pachyderm/src/client/pps"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
)

var (
	// Index mapping pipeline to jobs started by the pipeline
	JobsPipelineIndex = col.Index{"Pipeline", false}

	// Index mapping job inputs (repos + pipeline version) to output commit.
	// This is how we know if we need to start a job.
	JobsInputIndex = col.Index{"Input", false}

	// Index mapping 1.4.5 and earlier style job inputs (repos + pipeline
	// version) to output commit. This is how we know if we need to start a job
	// Needed for legacy compatibility.
	JobsInputsIndex = col.Index{"Inputs", false}

	// PPSEtcdPipelinesPrefix is the etcd prefix under which we store
	// pipeline information.
	pipelinesPrefix = "/pipelines"

	// PPSEtcdJobsPrefix is the etcd prefix under which we store job
	// information.
	jobsPrefix = "/jobs"
)

func Pipelines(client *etcd.Client, etcdPrefix string) col.Collection {
	return col.NewCollection(
		client,
		path.Join(etcdPrefix, pipelinesPrefix),
		nil,
		&pps.PipelineInfo{},
	)
}

func Jobs(client *etcd.Client, etcdPrefix string) col.Collection {
	return col.NewCollection(
		client,
		path.Join(etcdPrefix, jobsPrefix),
		[]col.Index{JobsPipelineIndex, JobsInputIndex},
		&pps.JobInfo{},
	)
}
