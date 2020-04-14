package metrics

import (
	"fmt"
	"time"

	"github.com/segmentio/analytics-go"
)

const reportingInterval time.Duration = 5 * time.Minute

type segmentClient struct {
	client *analytics.Client
}

func newSegmentClient() *segmentClient {
	c := analytics.New("hhxbyr7x50w3jtgcwcZUyOFrTf4VNMrD")
	c.Interval = reportingInterval
	c.Size = 100

	return &segmentClient{
		client: c,
	}
}

func (c *segmentClient) reportClusterMetrics(metrics *Metrics) {
	// We're intentionally ignoring an error here because metrics code is
	// non-critical
	c.client.Track(&analytics.Track{
		Event:       "cluster.metrics",
		AnonymousId: metrics.ClusterID,
		Properties: map[string]interface{}{
			"PodID":          metrics.PodID,
			"nodes":          metrics.Nodes,
			"version":        metrics.Version,
			"repos":          metrics.Repos,
			"commits":        metrics.Commits,
			"files":          metrics.Files,
			"bytes":          metrics.Bytes,
			"jobs":           metrics.Jobs,
			"pipelines":      metrics.Pipelines,
			"ActivationCode": metrics.ActivationCode,
		},
	})
}

/*
Segment needs us to identify a user before we report any events for that user.
We have no way of knowing if a user has previously been identified, so we call this
before every `Track()` call containing user data.
*/
func (c *segmentClient) identifyUser(userID string) {
	// We're intentionally ignoring an error here because metrics code is
	// non-critical
	c.client.Identify(&analytics.Identify{
		UserId: userID,
	})
}

func (c *segmentClient) reportUserMetrics(userID string, prefix string, action string, value interface{}, clusterID string) {
	c.identifyUser(userID)
	properties := map[string]interface{}{
		"ClusterID": clusterID,
	}
	properties[action] = value
	// We're intentionally ignoring an error here because metrics code is
	// non-critical
	c.client.Track(&analytics.Track{
		Event:      fmt.Sprintf("%v.usage", prefix),
		UserId:     userID,
		Properties: properties,
	})
}

func (c *segmentClient) Close() {
	c.client.Close()
}
