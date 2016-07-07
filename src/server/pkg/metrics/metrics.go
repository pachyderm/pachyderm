package metrics

import (
	"sync/atomic"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/client/version"

	"go.pedge.io/lion/proto"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
)

var metrics = &Metrics{}
var modified int64

// AddRepos atomically adds a number of repos to metrics.
func AddRepos(num int64) {
	atomic.AddInt64(&metrics.Repos, num)
	atomic.SwapInt64(&modified, 1)
}

// AddCommits atomically adds a number of commits to metrics.
func AddCommits(num int64) {
	atomic.AddInt64(&metrics.Commits, num)
	atomic.SwapInt64(&modified, 1)
}

// AddFiles atomically adds a number of files to metrics.
func AddFiles(num int64) {
	atomic.AddInt64(&metrics.Files, num)
	atomic.SwapInt64(&modified, 1)
}

// AddBytes atomically adds a number of bytes to metrics.
func AddBytes(num int64) {
	atomic.AddInt64(&metrics.Bytes, num)
	atomic.SwapInt64(&modified, 1)
}

// AddJobs atomically adds a number of jobs to metrics.
func AddJobs(num int64) {
	atomic.AddInt64(&metrics.Jobs, num)
	atomic.SwapInt64(&modified, 1)
}

// AddPipelines atomically adds a number of pipelines to metrics.
func AddPipelines(num int64) {
	atomic.AddInt64(&metrics.Pipelines, num)
	atomic.SwapInt64(&modified, 1)
}

// ReportMetrics blocks and reports metrics, if modified, to the
// given kubernetes client every 15 seconds.
func ReportMetrics(clusterID string, kubeClient *kube.Client) {
	metrics.ID = clusterID
	metrics.PodID = uuid.NewWithoutDashes()
	metrics.Version = version.PrettyPrintVersion(version.Version)
	for {
		write := atomic.SwapInt64(&modified, 0)
		if write == 1 {
			externalMetrics(kubeClient, metrics)
			protolion.Info(metrics)
			reportSegment(metrics)
		}
		<-time.After(15 * time.Second)
	}
}
