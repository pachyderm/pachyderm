package minikubetestenv

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
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
	licenseKeySecretName   = "enterprise-license-key-secret"
)

func getPachAddress(t testing.TB) *grpcutil.PachdAddress {
	cfg, err := config.Read(false, false)
	require.NoError(t, err)

	_, context, err := cfg.ActiveContext(true)
	require.NoError(t, err)

	pa, err := grpcutil.ParsePachdAddress(context.PachdAddress)
	require.NoError(t, err)

	return pa
}

func localDeploymentWithMinioOptions(namespace, image string) *helm.Options {
	return &helm.Options{
		KubectlOptions: &k8s.KubectlOptions{Namespace: namespace},
		SetValues: map[string]string{
			"deployTarget": "custom",

			"pachd.service.type":        "NodePort",
			"pachd.image.tag":           image,
			"pachd.clusterDeploymentID": "dev",

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

func withEnterprise(t testing.TB, namespace string) *helm.Options {
	addr := getPachAddress(t)
	return &helm.Options{
		KubectlOptions: &k8s.KubectlOptions{Namespace: namespace},
		SetValues: map[string]string{
			"pachd.enterpriseLicenseKeySecretName":               licenseKeySecretName,
			"pachd.rootToken":                                    testutil.RootToken,
			"pachd.oauthClientSecret":                            "oidc-client-secret",
			"pachd.enterpriseSecret":                             "enterprise-secret",
			"oidc.userAccessibleOauthIssuerHost":                 fmt.Sprintf("%s:30658", addr.Host),
			"ingress.host":                                       fmt.Sprintf("%s:30657", addr.Host),
			"global.postgresql.identityDatabaseFullNameOverride": "dexdb",
		},
	}
}

func union(a, b *helm.Options) *helm.Options {
	c := &helm.Options{
		KubectlOptions: &k8s.KubectlOptions{Namespace: b.KubectlOptions.Namespace},
		SetValues:      make(map[string]string),
		SetStrValues:   make(map[string]string),
	}
	copy := func(src, dst *helm.Options) {
		for k, v := range src.SetValues {
			dst.SetValues[k] = v
		}
		for k, v := range src.SetStrValues {
			dst.SetStrValues[k] = v
		}
	}
	copy(a, c)
	copy(b, c)
	return c
}

func waitForPachd(t testing.TB, ctx context.Context, kubeClient *kube.Clientset, namespace, version string) {
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

func waitForPgbouncer(t testing.TB, ctx context.Context, kubeClient *kube.Clientset, namespace string) {
	require.NoError(t, backoff.Retry(func() error {
		pbs, err := kubeClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: "app=pg-bouncer"})
		if err != nil {
			return errors.Wrap(err, "error on pod list")
		}
		for _, p := range pbs.Items {
			if p.Status.Phase == v1.PodRunning && p.Status.ContainerStatuses[0].Ready && len(pbs.Items) == 1 {
				return nil
			}
		}
		return errors.Errorf("deployment in progress")
	}, backoff.RetryEvery(5*time.Second).For(5*time.Minute)))
}

// Deploy pachyderm using a `helm install ...`, then run a test with an API Client corresponding to the deployment
func InstallPublishedRelease(t testing.TB, ctx context.Context, kubeClient *kube.Clientset, version, user string, cleanup bool) *client.APIClient {
	maybeCleanup(t, cleanup, kubeClient)
	opts := localDeploymentWithMinioOptions(ns, version)
	opts.Version = version
	require.NoError(t, helm.InstallE(t, opts, helmChartPublishedPath, helmRelease))
	waitForPachd(t, ctx, kubeClient, ns, version)
	return testutil.AuthenticatedPachClient(t, testutil.NewPachClient(t), user)
}

func UpgradeRelease(t testing.TB, ctx context.Context, kubeClient *kube.Clientset, user string, cleanup bool) *client.APIClient {
	maybeCleanup(t, cleanup, kubeClient)
	require.NoError(t, helm.UpgradeE(t, localDeploymentWithMinioOptions(ns, localImage), helmChartLocalPath, helmRelease))
	waitForPachd(t, ctx, kubeClient, ns, localImage)
	waitForPgbouncer(t, ctx, kubeClient, ns)
	// use a timeout in case neighboring services such as pg-bouncer also
	// restart, and prolongue the stabilization period
	time.Sleep(time.Duration(10) * time.Second)
	return testutil.AuthenticatedPachClient(t, testutil.NewPachClient(t), user)
}

func InstallReleaseEnterprise(t testing.TB, ctx context.Context, kubeClient *kube.Clientset, user string, cleanup bool) *client.APIClient {
	maybeCleanup(t, cleanup, kubeClient)
	createSecretEnterpriseKeySecret(t, ctx, kubeClient, ns)
	opts := union(localDeploymentWithMinioOptions(ns, localImage), withEnterprise(t, ns))
	require.NoError(t, helm.InstallE(t, opts, helmChartLocalPath, helmRelease))
	waitForPachd(t, ctx, kubeClient, ns, localImage)
	return testutil.AuthenticatedPachClientPostActivate(t, testutil.NewPachClient(t), user)
}
func maybeCleanup(t testing.TB, cleanup bool, kubeClient *kube.Clientset) {
	if cleanup {
		t.Cleanup(func() {
			deleteRelease(t, context.Background(), kubeClient)
		})
	}
}

func deleteRelease(t testing.TB, ctx context.Context, kubeClient *kube.Clientset) {
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

func createSecretEnterpriseKeySecret(t testing.TB, ctx context.Context, kubeClient *kube.Clientset, ns string) {
	_, err := kubeClient.CoreV1().Secrets(ns).Create(ctx, &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: licenseKeySecretName},
		StringData: map[string]string{
			"enterprise-license-key": testutil.GetTestEnterpriseCode(t),
		},
	}, metav1.CreateOptions{})
	require.True(t, err == nil || strings.Contains(err.Error(), "already exists"))
}
