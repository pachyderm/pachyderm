package metrics

type client interface {
	reportClusterMetrics(metrics *Metrics)
	reportUserMetrics(userID string, prefix string, action string, value interface{}, clusterID string)
}
