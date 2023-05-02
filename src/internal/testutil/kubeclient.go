package testutil

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	k, _, err := GetKubeClientAndConfig()
	require.NoError(t, err, "could not get k8s REST config and client")
	return k
}

// GetKubeClientAndConfig returns a Kubernetes client and its REST config.
func GetKubeClientAndConfig() (*kube.Clientset, *rest.Config, error) {
	var (
		config *rest.Config
		err    error
		host   = os.Getenv("KUBERNETES_SERVICE_HOST")
	)
	if host != "" {
		config, err = rest.InClusterConfig()
	} else {
		rules := clientcmd.NewDefaultClientConfigLoadingRules()
		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules,
			&clientcmd.ConfigOverrides{})
		config, err = kubeConfig.ClientConfig()
	}
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not get k8s REST config")
	}
	k, err := kube.NewForConfig(config)
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not get k8s client")
	}
	return k, config, nil
}

// DeletePipelineRC deletes the RC belonging to the pipeline 'pipeline'. This
// can be used to test PPS's robustness
func DeletePipelineRC(t testing.TB, pipeline, namespace string) {
	kubeClient := GetKubeClient(t)
	rcs, err := kubeClient.CoreV1().ReplicationControllers(namespace).List(
		context.Background(),
		metav1.ListOptions{
			LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(
				map[string]string{"pipelineName": pipeline},
			)),
		})
	require.NoError(t, err)
	require.Equal(t, 1, len(rcs.Items))
	require.NoError(t, kubeClient.CoreV1().ReplicationControllers(namespace).Delete(
		context.Background(),
		rcs.Items[0].ObjectMeta.Name, metav1.DeleteOptions{
			GracePeriodSeconds: &zero,
		}))
	require.NoErrorWithinTRetry(t, 30*time.Second, func() error {
		rcs, err := kubeClient.CoreV1().ReplicationControllers(namespace).List(
			context.Background(),
			metav1.ListOptions{
				LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(
					map[string]string{"pipelineName": pipeline},
				)),
			})
		if err != nil {
			return errors.EnsureStack(err)
		}
		if len(rcs.Items) != 0 {
			return errors.Errorf("RC %q not deleted yet", pipeline)
		}
		return nil
	})
}
