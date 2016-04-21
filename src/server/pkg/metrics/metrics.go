package metrics

import (
	"sync/atomic"
	"time"

	"go.pedge.io/lion/proto"
)

var metrics = &Metrics{}
var modified int64

func AddRepos(num int64) {
	atomic.AddInt64(&metrics.Repos, num)
	atomic.SwapInt64(&modified, 1)
}

func AddCommits(num int64) {
	atomic.AddInt64(&metrics.Commits, num)
	atomic.SwapInt64(&modified, 1)
}

func AddFiles(num int64) {
	atomic.AddInt64(&metrics.Files, num)
	atomic.SwapInt64(&modified, 1)
}

func AddBytes(num int64) {
	atomic.AddInt64(&metrics.Bytes, num)
	atomic.SwapInt64(&modified, 1)
}

func AddJobs(num int64) {
	atomic.AddInt64(&metrics.Jobs, num)
	atomic.SwapInt64(&modified, 1)
}

func AddPipelines(num int64) {
	atomic.AddInt64(&metrics.Pipelines, num)
	atomic.SwapInt64(&modified, 1)
}

func ReportMetrics(clusterID string) {
	metrics.ID = id
	for {
		write := atomic.SwapInt64(&modified, 0)
		if write == 1 {
			protolion.Info(metrics)
			reportSegment(metrics)
		}
		<-time.After(15 * time.Second)
	}
}
