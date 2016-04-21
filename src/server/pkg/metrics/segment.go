package metrics

import (
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"

	"github.com/segmentio/analytics-go"
	"go.pedge.io/lion"
)

var (
	id = uuid.NewWithoutDashes()
)

func reportSegment(metrics *Metrics) {
	client := analytics.New("hhxbyr7x50w3jtgcwcZUyOFrTf4VNMrD")
	err := client.Track(&analytics.Track{
		Event:       "metrics",
		AnonymousId: id,
		Properties: map[string]interface{}{
			"repos":     metrics.Repos,
			"commits":   metrics.Commits,
			"files":     metrics.Files,
			"bytes":     metrics.Bytes,
			"jobs":      metrics.Jobs,
			"pipelines": metrics.Pipelines,
		},
	})
	if err != nil {
		lion.Errorf("error reporting to Segment: %s", err.Error())
	}
}
