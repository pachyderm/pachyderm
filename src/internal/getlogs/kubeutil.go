package getlogs

import (
	"context"
  "time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

type KubeLogSource struct {
  Kube kubernetes.Interface
  Namespace string
  FromTime time.Time
  Follow bool
}

func (k *KubeLogSource) podContainers(ctx context.Context, kube kubernetes.Interface, namespace string, labels labels.Set) (map[string][]string, error) {
	pods, err := k.Kube.CoreV1().Pods(k.Namespace).List(ctx, meta.ListOptions{
		TypeMeta: meta.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		},
		LabelSelector: meta.FormatLabelSelector(meta.SetAsLabelSelector(labels)),
	})
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
  podContainers := make(map[string][]string)
  for _, pod := range pods.Items {
    podName := pod.ObjectMeta.Name
    for _, container := range pod.Spec.Containers {
      podContainers[podName] = append(podContainers[podName], container.Name)
    }
  }

  return podContainers, nil
}

func (k *KubeLogSource) GetPodStreamers(ctx context.Context, pod string, container string) ([]LogSource, error) {
  podLabels := map[string]string{}
  if pod == "" {
    podLabels["suite"] = "pachyderm"
  } else {
    podLabels["app"] = pod
    podLabels["app"] = "pachd"
  }

  pods, err := k.podContainers(ctx, k.Kube, k.Namespace, podLabels)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}

  sinceSeconds := new(int64)
  *sinceSeconds = int64(time.Since(k.FromTime) / time.Second)

  podLogOptions := &v1.PodLogOptions{
    Follow: k.Follow,
    SinceSeconds: sinceSeconds,
  }

  podStreamers := []LogSource{}

  if container != "" {
    request := k.Kube.CoreV1().Pods(k.Namespace).GetLogs(pod, podLogOptions).Timeout(10 * time.Second)
    podStreamers = append(podStreamers, request)
    return podStreamers, nil
  }
  for pod, containers := range pods {
    for _, container := range containers {
      podLogOptions.Container = container
      request := k.Kube.CoreV1().Pods(k.Namespace).GetLogs(pod, podLogOptions).Timeout(10 * time.Second)
      podStreamers = append(podStreamers, request)
    }
  }
  return podStreamers, nil
}

