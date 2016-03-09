package metrics

import (
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/pachyderm/pachyderm/src/pkg/obj"
	"github.com/pachyderm/pachyderm/src/pkg/uuid"
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

func ReportMetrics(client obj.Client) {
	for {
		protolion.Info(metrics)
		if err := reportObj(client); err != nil {
			protolion.Errorf("Error writing to object store: %s", err.Error())
		}
		<-time.After(15 * time.Second)
	}
}

func reportObj(client obj.Client) (retErr error) {
	marshaller := &jsonpb.Marshaler{Indent: "  "}
	writer, err := client.Writer(uuid.NewWithoutDashes())
	if err != nil {
		return err
	}
	defer func() {
		if err := writer.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	if err := marshaller.Marshal(writer, metrics); err != nil {
		return err
	}
	return nil
}
