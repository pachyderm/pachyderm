package minikubetestenv

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube "k8s.io/client-go/kubernetes"
)

const (
	helmChartLocalPath     = "../../../etc/helm/pachyderm"
	helmChartPublishedPath = "pach/pachyderm"
	helmRelease            = "test-release"
	ns                     = "default"
	localImage             = "local"
)

func localDeploymentWithMinioOptions(namespace, image string) *helm.Options {
	return &helm.Options{
		KubectlOptions: &k8s.KubectlOptions{Namespace: namespace},
		SetValues: map[string]string{
			"deployTarget": "custom",

			"pachd.service.type":           "NodePort",
			"pachd.image.tag":              image,
			"pachd.storage.backend":        "MINIO",
			"pachd.storage.minio.bucket":   "pachyderm-test",
			"pachd.storage.minio.endpoint": "minio.default.svc.cluster.local:9000",
			"pachd.storage.minio.id":       "minioadmin",
			"pachd.storage.minio.secret":   "minioadmin",

			"global.postgresql.postgresqlPassword":         "pachyderm",
			"global.postgresql.postgresqlPostgresPassword": "pachyderm",
		},
		SetStrValues: map[string]string{
			"pachd.storage.minio.signature": "",
			"pachd.storage.minio.secure":    "false",
		},
	}
}

func waitForPachd(t *testing.T, ctx context.Context, kubeClient *kube.Clientset, namespace, version string) {
	require.NoError(t, backoff.Retry(func() error {
		pachds, err := kubeClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: "app=pachd"})
		if err != nil {
			return errors.Wrap(err, "error on pod list")
		}
		for _, p := range pachds.Items {
			if p.Status.Phase == v1.PodRunning && strings.HasSuffix(p.Spec.Containers[0].Image, ":"+version) && p.Status.ContainerStatuses[0].Ready && len(pachds.Items) == 1 {
				return nil
			}
		}
		return errors.Errorf("deployment in progress")
	}, backoff.RetryEvery(5*time.Second).For(5*time.Minute)))
}

// Deploy pachyderm using a `helm install ...`, then run a test with an API Client corresponding to the deployment
func InstallPublishedRelease(t *testing.T, ctx context.Context, kubeClient *kube.Clientset, version, user string) *client.APIClient {
	require.NoError(t, helm.InstallE(t, localDeploymentWithMinioOptions(ns, version), helmChartPublishedPath, helmRelease))
	waitForPachd(t, ctx, kubeClient, ns, version)
	return testutil.AuthenticatedPachClient(t, testutil.NewPachClient(t), user)
}

func UpgradeRelease(t *testing.T, ctx context.Context, kubeClient *kube.Clientset, user string) *client.APIClient {
	require.NoError(t, helm.UpgradeE(t, localDeploymentWithMinioOptions(ns, localImage), helmChartLocalPath, helmRelease))
	waitForPachd(t, ctx, kubeClient, ns, localImage)
	return testutil.AuthenticatedPachClient(t, testutil.NewPachClient(t), user)
}

func DeleteRelease(t *testing.T, ctx context.Context, kubeClient *kube.Clientset) {
	options := &helm.Options{
		KubectlOptions: &k8s.KubectlOptions{Namespace: ns},
	}
	err := helm.DeleteE(t, options, helmRelease, true)
	require.True(t, err == nil || strings.Contains(err.Error(), "not found"))
	require.NoError(t, kubeClient.CoreV1().PersistentVolumeClaims(ns).DeleteCollection(ctx, *metav1.NewDeleteOptions(0), metav1.ListOptions{LabelSelector: "suite=pachyderm"}))
	require.NoError(t, backoff.Retry(func() error {
		pvcs, err := kubeClient.CoreV1().PersistentVolumeClaims(ns).List(ctx, metav1.ListOptions{LabelSelector: "suite=pachyderm"})
		if err != nil {
			return errors.Wrap(err, "error on pod list")
		}
		if len(pvcs.Items) == 0 {
			return nil
		}
		return errors.Errorf("pvcs have yet to be deleted")
	}, backoff.RetryEvery(5*time.Second).For(2*time.Minute)))
}
