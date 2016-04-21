package metrics

import (
	"github.com/segmentio/analytics-go"
	"go.pedge.io/lion"
)

func reportSegment(metrics *Metrics) {
	client := analytics.New("hhxbyr7x50w3jtgcwcZUyOFrTf4VNMrD")
	err := client.Track(&analytics.Track{
		Event:       "metrics",
		AnonymousId: metrics.ID,
		Properties: map[string]interface{}{
			"nodes":     metrics.Nodes,
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
