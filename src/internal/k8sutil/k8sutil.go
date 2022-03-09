package k8sutil

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetRCPods(ctx context.Context, clientset *kubernetes.Clientset, namespace, rcName string) ([]v1.Pod, error) {
	podList, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		},
		LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(map[string]string{"app": rcName})),
	})
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return podList.Items, nil
}
