package testutil

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"

	v1 "k8s.io/api/core/v1"
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

// DeletePachdPod deletes the pachd pod in a test cluster (restarting it, e.g.
// to retart the PPS master)
func DeletePachdPod(t testing.TB) {
	kubeClient := GetKubeClient(t)
	podList, err := kubeClient.CoreV1().Pods(v1.NamespaceDefault).List(
		metav1.ListOptions{
			LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(
				map[string]string{"app": "pachd", "suite": "pachyderm"},
			)),
		})
	require.NoError(t, err)
	require.Equal(t, 1, len(podList.Items))
	require.NoError(t, kubeClient.CoreV1().Pods(v1.NamespaceDefault).Delete(
		podList.Items[0].ObjectMeta.Name, &metav1.DeleteOptions{}))

	// Make sure pachd goes down
	startTime := time.Now()
	require.NoError(t, backoff.Retry(func() error {
		podList, err := kubeClient.CoreV1().Pods(v1.NamespaceDefault).List(
			metav1.ListOptions{
				LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(
					map[string]string{"app": "pachd", "suite": "pachyderm"},
				)),
			})
		if err != nil {
			return err
		}
		if len(podList.Items) == 0 {
			return nil
		}
		if time.Since(startTime) > 10*time.Second {
			return nil
		}
		return fmt.Errorf("waiting for old pachd pod to be killed")
	}, backoff.NewTestingBackOff()))

	// Make sure pachd comes back up
	require.NoErrorWithinTRetry(t, 30*time.Second, func() error {
		podList, err := kubeClient.CoreV1().Pods(v1.NamespaceDefault).List(
			metav1.ListOptions{
				LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(
					map[string]string{"app": "pachd", "suite": "pachyderm"},
				)),
			})
		if err != nil {
			return err
		}
		if len(podList.Items) == 0 {
			return fmt.Errorf("no pachd pod up yet")
		}
		return nil
	})

	require.NoErrorWithinTRetry(t, 30*time.Second, func() error {
		podList, err := kubeClient.CoreV1().Pods(v1.NamespaceDefault).List(
			metav1.ListOptions{
				LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(
					map[string]string{"app": "pachd", "suite": "pachyderm"},
				)),
			})
		if err != nil {
			return err
		}
		if len(podList.Items) == 0 {
			return fmt.Errorf("no pachd pod up yet")
		}
		if podList.Items[0].Status.Phase != v1.PodRunning {
			return fmt.Errorf("pachd not running yet")
		}
		return err
	})
}

// DeletePipelineRC deletes the RC belonging to the pipeline 'pipeline'. This
// can be used to test PPS's robustness
func DeletePipelineRC(t testing.TB, pipeline string) {
	kubeClient := GetKubeClient(t)
	rcs, err := kubeClient.CoreV1().ReplicationControllers(v1.NamespaceDefault).List(
		metav1.ListOptions{
			LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(
				map[string]string{"pipelineName": pipeline},
			)),
		})
	require.NoError(t, err)
	require.Equal(t, 1, len(rcs.Items))
	require.NoError(t, kubeClient.CoreV1().ReplicationControllers(v1.NamespaceDefault).Delete(
		rcs.Items[0].ObjectMeta.Name, &metav1.DeleteOptions{
			GracePeriodSeconds: &zero,
		}))
	require.NoErrorWithinTRetry(t, 30*time.Second, func() error {
		rcs, err := kubeClient.CoreV1().ReplicationControllers(v1.NamespaceDefault).List(
			metav1.ListOptions{
				LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(
					map[string]string{"pipelineName": pipeline},
				)),
			})
		if err != nil {
			return err
		}
		if len(rcs.Items) != 0 {
			return fmt.Errorf("RC %q not deleted yet", pipeline)
		}
		return nil
	})
}
