package metrics

import (
	"github.com/segmentio/analytics-go"
)

// router sends the same metrics to multiple endpoints
type router struct {
	clients []*analytics.Client
}

func newRouter(endpoints ...string) *router {
	clients := make([]*analytics.Client, len(endpoints)+1)
	clients[0] = newPersistentClient()
	for i, e := range endpoints {
		clients[i+1] = newPersistentClient()
		clients[i+1].Endpoint = e
	}
	return &router{clients}
}

func (r *router) reportClusterMetricsToSegment(metrics *Metrics) {
	for _, client := range r.clients {
		reportClusterMetricsToSegment(client, metrics)
	}
}

func (r *router) reportUserMetricsToSegment(userID string, prefix string, action string, value interface{}, clusterID string) {
	for _, client := range r.clients {
		reportUserMetricsToSegment(client, userID, prefix, action, value, clusterID)
	}
}
