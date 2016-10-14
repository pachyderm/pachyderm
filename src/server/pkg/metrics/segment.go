package metrics

import (
	"github.com/segmentio/analytics-go"
	"go.pedge.io/lion"
)

var client = analytics.New("hhxbyr7x50w3jtgcwcZUyOFrTf4VNMrD")

func reportClusterMetricsToSegment(metrics *Metrics) {
	err := client.Track(&analytics.Track{
		Event:       "metrics",
		AnonymousId: metrics.ID,
		Properties: map[string]interface{}{
			"PodID":     metrics.PodID,
			"nodes":     metrics.Nodes,
			"version":   metrics.Version,
			"repos":     metrics.Repos,
			"commits":   metrics.Commits,
			"files":     metrics.Files,
			"bytes":     metrics.Bytes,
			"jobs":      metrics.Jobs,
			"pipelines": metrics.Pipelines,
		},
	})
	if err != nil {
		lion.Errorf("error reporting cluster metrics to Segment: %s", err.Error())
	}
}

func identifyNewUser(userID string) {
	err := client.Identify(&analytics.Identify{
		UserId: userID,
	})
	if err != nil {
		lion.Errorf("error reporting user identity to Segment: %s", err.Error())
	}
}

func reportUserMetricsToSegment(userActions *UserActions) {
	for userID, actions := range userActions {
		err := client.Track(&analytics.Track{
			Event:      "Usage",
			UserId:     userID,
			Properties: actions,
		})
		if err != nil {
			lion.Errorf("error reporting user action to Segment: %s", err.Error())
		}
	}
}
