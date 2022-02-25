package minikubetestenv

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	terraTest "github.com/gruntwork-io/terratest/modules/testing"
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
	localImage             = "local"
	licenseKeySecretName   = "enterprise-license-key-secret"
)

type DeployOpts struct {
	Version      string
	Enterprise   bool
	AuthUser     string
	CleanupAfter bool
	PortOffset   uint16 // might make more sense to declare port instead of offset
}

type helmPutE func(t terraTest.TestingT, options *helm.Options, chart string, releaseName string) error

func getPachAddress(t testing.TB) *grpcutil.PachdAddress {
	cfg, err := config.Read(true, true)
	require.NoError(t, err)
	_, context, err := cfg.ActiveContext(true)
	require.NoError(t, err)
	address, err := client.GetUserMachineAddr(context)
	require.NoError(t, err)
	if address == nil {
		address = &grpcutil.DefaultPachdAddress
	}
	return address
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
			"pachd.enterpriseLicenseKeySecretName": licenseKeySecretName,
			"pachd.rootToken":                      testutil.RootToken,
			"pachd.oauthClientSecret":              "oidc-client-secret",
			"pachd.enterpriseSecret":               "enterprise-secret",
			// TODO: make these ports configurable to support IDP Login in parallel deployments
			"oidc.userAccessibleOauthIssuerHost": fmt.Sprintf("%s:30658", addr.Host),
			"ingress.host":                       fmt.Sprintf("%s:30657", addr.Host),
		},
	}
}

func withPort(t testing.TB, namespace string, port uint16) *helm.Options {
	return &helm.Options{
		KubectlOptions: &k8s.KubectlOptions{Namespace: namespace},
		SetValues: map[string]string{
			"pachd.service.port": fmt.Sprintf("%v", port),
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

func putRelease(t testing.TB, ctx context.Context, namespace string, kubeClient *kube.Clientset, f helmPutE, opts *DeployOpts) *client.APIClient {
	if opts.CleanupAfter {
		t.Cleanup(func() {
			deleteRelease(t, context.Background(), namespace, kubeClient)
		})
	}
	version := localImage
	chartPath := helmChartLocalPath
	if opts.Version != "" {
		version = opts.Version
		chartPath = helmChartPublishedPath
	}
	helmOpts := localDeploymentWithMinioOptions(namespace, version)
	if opts.Enterprise {
		createSecretEnterpriseKeySecret(t, ctx, kubeClient, namespace)
		helmOpts = union(helmOpts, withEnterprise(t, namespace))
	}
	pachAddress := getPachAddress(t)
	if opts.PortOffset != 0 {
		pachAddress.Port = pachAddress.Port + opts.PortOffset
		helmOpts = union(helmOpts, withPort(t, namespace, pachAddress.Port))
	}
	require.NoError(t, f(t, helmOpts, chartPath, namespace))
	waitForPachd(t, ctx, kubeClient, namespace, version)
	c, err := client.NewFromPachdAddress(pachAddress)
	require.NoError(t, err)
	if opts.AuthUser != "" {
		c = testutil.AuthenticateClient(t, c, opts.AuthUser)
	}
	return c
}

// Deploy pachyderm using a `helm upgrade ...`
// returns an API Client corresponding to the deployment
func UpgradeRelease(t testing.TB, ctx context.Context, namespace string, kubeClient *kube.Clientset, opts *DeployOpts) *client.APIClient {
	return putRelease(t, ctx, namespace, kubeClient, helm.UpgradeE, opts)
}

// Deploy pachyderm using a `helm install ...`
// returns an API Client corresponding to the deployment
func InstallRelease(t testing.TB, ctx context.Context, namespace string, kubeClient *kube.Clientset, opts *DeployOpts) *client.APIClient {
	return putRelease(t, ctx, namespace, kubeClient, helm.InstallE, opts)
}

func deleteRelease(t testing.TB, ctx context.Context, namespace string, kubeClient *kube.Clientset) {
	options := &helm.Options{
		KubectlOptions: &k8s.KubectlOptions{Namespace: namespace},
	}
	err := helm.DeleteE(t, options, namespace, true)
	require.True(t, err == nil || strings.Contains(err.Error(), "not found"))
	require.NoError(t, kubeClient.CoreV1().PersistentVolumeClaims(namespace).DeleteCollection(ctx, *metav1.NewDeleteOptions(0), metav1.ListOptions{LabelSelector: "suite=pachyderm"}))
	require.NoError(t, backoff.Retry(func() error {
		pvcs, err := kubeClient.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{LabelSelector: "suite=pachyderm"})
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
