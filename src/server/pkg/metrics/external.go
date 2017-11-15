package metrics

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube "k8s.io/client-go/kubernetes"
)

func externalMetrics(kubeClient *kube.Clientset, metrics *Metrics) error {
	nodeList, err := kubeClient.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("externalMetrics: unable to retrieve node list from k8s")
	}
	metrics.Nodes = int64(len(nodeList.Items))
	return nil
}
