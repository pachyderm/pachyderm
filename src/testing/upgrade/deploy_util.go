package main

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
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
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

type deployTest func(*testing.T, *client.APIClient)

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
			return err
		}
		for _, p := range pachds.Items {
			if p.Status.Phase == v1.PodRunning && strings.HasSuffix(p.Spec.Containers[0].Image, ":"+version) && p.Status.ContainerStatuses[0].Ready && len(pachds.Items) == 1 {
				return nil
			}
		}
		return errors.Errorf("deployment in progress")
	}, backoff.RetryEvery(5*time.Second).For(5*time.Minute)))
	//time.Sleep(10 * time.Second)
}

// func runPortForward(t *testing.T, namespace string) *client.PortForwarder {
// 	cfg, err := config.Read(false, true)
// 	require.NoError(t, err)

// 	_, context, err := cfg.ActiveContext(true)
// 	require.NoError(t, err)

// 	pf, err := client.NewPortForwarder(context, namespace)
// 	require.NoError(t, err)

// 	_, err = pf.RunForPachd(30650, 1650)
// 	require.NoError(t, err)

// 	// for now this timeout allows the port-forwarder to
// 	// work following a previously closed forwarder.
// 	// It appears that the Accept() system call does not receive connections
// 	// until a sufficient period has passed in the case that a port-forwarder
// 	// was previously used, such as in the upgrade case.
// 	time.Sleep(30 * time.Second)
// 	return pf
// }

// Deploy pachyderm using a `helm install ...`, then run a test with an API Client corresponding to the deployment
func WithInstallPublishedRelease(t *testing.T, ctx context.Context, kubeClient *kube.Clientset, version, user string, test deployTest) {
	require.NoError(t, helm.InstallE(t, localDeploymentWithMinioOptions(ns, version), helmChartPublishedPath, helmRelease))
	waitForPachd(t, ctx, kubeClient, ns, version)

	// pf := runPortForward(t, ns)
	// defer pf.Close()

	test(t, tu.GetAuthenticatedPachClient(t, user))
}

func WithUpgradeRelease(t *testing.T, ctx context.Context, kubeClient *kube.Clientset, user string, test deployTest) {
	require.NoError(t, helm.UpgradeE(t, localDeploymentWithMinioOptions(ns, localImage), helmChartLocalPath, helmRelease))
	waitForPachd(t, ctx, kubeClient, ns, localImage)

	// pf := runPortForward(t, ns)
	// defer pf.Close()

	test(t, tu.GetAuthenticatedPachClient(t, user))
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
			return err
		}
		if len(pvcs.Items) == 0 {
			return nil
		}
		return errors.Errorf("pvcs have yet to be deleted")
	}, backoff.RetryEvery(5*time.Second).For(2*time.Minute)))
}
