package metrics

import (
	"sync/atomic"
	"time"

	"go.pedge.io/lion/proto"
)

var metrics *Metrics = &Metrics{}

func AddRepos(num int64) {
	atomic.AddInt64(&metrics.Repos, num)
}

func AddCommits(num int64) {
	atomic.AddInt64(&metrics.Commits, num)
}

func AddFiles(num int64) {
	atomic.AddInt64(&metrics.Files, num)
}

func AddBytes(num int64) {
	atomic.AddInt64(&metrics.Bytes, num)
}

func AddJobs(num int64) {
	atomic.AddInt64(&metrics.Jobs, num)
}

func AddPipelines(num int64) {
	atomic.AddInt64(&metrics.Pipelines, num)
}

func ReportMetrics() {
	for {
		protolion.Info(metrics)
		<-time.After(15 * time.Second)
	}
}
