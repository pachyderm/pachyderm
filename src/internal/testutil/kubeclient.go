package testutil

import (
	"context"
	"os"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"

	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	kube "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	zero int64
)

// GetKubeClient connects to the Kubernetes API server either from inside the
// cluster or from a test binary running on a machine with kubectl (it will
// connect to the same cluster as kubectl)
func GetKubeClient(t testing.TB) *kube.Clientset {
	var config *rest.Config
	var err error
	host := os.Getenv("KUBERNETES_SERVICE_HOST")
	if host != "" {
		config, err = rest.InClusterConfig()
	} else {
		rules := clientcmd.NewDefaultClientConfigLoadingRules()
		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules,
			&clientcmd.ConfigOverrides{})
		config, err = kubeConfig.ClientConfig()
	}
	require.NoError(t, err)
	k, err := kube.NewForConfig(config)
	require.NoError(t, err)
	return k
}

// PachdDeployment finds the corresponding deployment for pachd in the
// kubernetes namespace and returns it.
func PachdDeployment(t testing.TB, namespace string) *apps.Deployment {
	k := GetKubeClient(t)
	result, err := k.AppsV1().Deployments(namespace).Get(context.Background(), "pachd", metav1.GetOptions{})
	require.NoError(t, err)
	return result
}

func podRunningAndReady(e watch.Event) (bool, error) {
	if e.Type == watch.Deleted {
		return false, errors.New("received DELETE while watching pods")
	}
	pod, ok := e.Object.(*v1.Pod)
	if !ok {
		return false, errors.Errorf("unexpected object type in watch.Event")
	}
	return pod.Status.Phase == v1.PodRunning, nil
}
