package metrics

import (
	"fmt"
	"time"

	"github.com/segmentio/analytics-go"
	log "github.com/sirupsen/logrus"
)

const reportingInterval time.Duration = 5 * time.Minute

func newPersistentClient() *analytics.Client {
	c := newSegmentClient()
	c.Interval = reportingInterval
	c.Size = 100
	return c
}

func newSegmentClient() *analytics.Client {
	return analytics.New("hhxbyr7x50w3jtgcwcZUyOFrTf4VNMrD")
}

func reportClusterMetricsToSegment(client *analytics.Client, metrics *Metrics) {
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
		log.Errorf("error reporting cluster metrics to Segment: %s", err.Error())
	}
}

/*
Segment needs us to identify a user before we report any events for that user.
We have no way of knowing if a user has previously been identified, so we call this
before every `Track()` call containing user data.
*/
func identifyUser(client *analytics.Client, userID string) {
	err := client.Identify(&analytics.Identify{
		UserId: userID,
	})
	if err != nil {
		log.Errorf("error reporting user identity to Segment: %s", err.Error())
	}
}

func reportUserMetricsToSegment(client *analytics.Client, userID string, prefix string, action string, value interface{}, clusterID string) {
	identifyUser(client, userID)
	properties := map[string]interface{}{
		"ClusterID": clusterID,
	}
	properties[action] = value
	err := client.Track(&analytics.Track{
		Event:      fmt.Sprintf("%v.usage", prefix),
		UserId:     userID,
		Properties: properties,
	})
	if err != nil {
		log.Errorf("error reporting user action to Segment: %s", err.Error())
	}
}
