package metrics

import (
	"github.com/segmentio/analytics-go"
	"go.pedge.io/lion"
)

var client = analytics.New("hhxbyr7x50w3jtgcwcZUyOFrTf4VNMrD")

func reportClusterMetricsToSegment(metrics *Metrics) {
	err := client.Track(&analytics.Track{
		Event:       "cluster.metrics",
		AnonymousId: metrics.ClusterID,
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

/*
Segment needs us to identify a user before we report any events for that user.
There seems to be no penalty to re-identifying a user, so we call this aggressively
*/
func identifyUser(userID string) {
	err := client.Identify(&analytics.Identify{
		UserId: userID,
	})
	if err != nil {
		lion.Errorf("error reporting user identity to Segment: %s", err.Error())
	}
}

func reportUserMetricsToSegment(allUsersActions countableUserActions, clusterID string) {
	for userID, actions := range allUsersActions {
		identifyUser(userID)
		actions["ClusterID"] = clusterID
		err := client.Track(&analytics.Track{
			Event:      "user.usage",
			UserId:     userID,
			Properties: actions,
		})
		if err != nil {
			lion.Errorf("error reporting user action to Segment: %s", err.Error())
		}
	}
}
