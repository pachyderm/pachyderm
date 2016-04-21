package metrics

import (
	"fmt"

	kube_api "k8s.io/kubernetes/pkg/api"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
)

func externalMetrics(kubeClient *kube.Client, metrics *Metrics) error {
	nodeList, err := kubeClient.Nodes().List(kube_api.ListOptions{})
	if err != nil {
		return fmt.Errorf("externalMetrics: unable to retrieve node list from k8s")
	}
	metrics.Nodes = int64(len(nodeList.Items))
	return nil
}
