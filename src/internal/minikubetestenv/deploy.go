package minikubetestenv

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
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
	helmChartPublishedPath = "pach/pachyderm"
	localImage             = "local"
	licenseKeySecretName   = "enterprise-license-key-secret"
)

var (
	mu sync.Mutex // defensively lock around helm calls

	computedPachAddress *grpcutil.PachdAddress
)

type DeployOpts struct {
	Version            string
	Enterprise         bool
	AuthUser           string
	CleanupAfter       bool
	UseLeftoverCluster bool
	// Because NodePorts are cluster-wide, we use a PortOffset to
	// assign separate ports per deployment.
	// NOTE: it might make more sense to declare port instead of offset
	PortOffset     uint16
	Loki           bool
	WaitSeconds    int
	ValueOverrides map[string]string
}

type helmPutE func(t terraTest.TestingT, options *helm.Options, chart string, releaseName string) error

func helmLock(f helmPutE) helmPutE {
	return func(t terraTest.TestingT, options *helm.Options, chart string, releaseName string) error {
		mu.Lock()
		defer mu.Unlock()
		return f(t, options, chart, releaseName)
	}
}

func helmChartLocalPath(t testing.TB) string {
	dir, err := os.Getwd()
	require.NoError(t, err)
	parts := strings.Split(dir, string(os.PathSeparator))
	var relPathParts []string
	for i := len(parts) - 1; i >= 0; i-- {
		relPathParts = append(relPathParts, "..")
		if parts[i] == "src" {
			break
		}
	}
	relPathParts = append(relPathParts, "etc", "helm", "pachyderm")
	return filepath.Join(relPathParts...)
}

func getPachAddress(t testing.TB) *grpcutil.PachdAddress {
	if computedPachAddress == nil {
		cfg, err := config.Read(true, true)
		require.NoError(t, err)
		_, context, err := cfg.ActiveContext(false)
		require.NoError(t, err)
		computedPachAddress, err = client.GetUserMachineAddr(context)
		require.NoError(t, err)
		if computedPachAddress == nil {
			copy := grpcutil.DefaultPachdAddress
			computedPachAddress = &copy
		}
	}
	copy := *computedPachAddress
	return &copy
}

func exposedServiceType() string {
	os := runtime.GOOS
	serviceType := ""
	switch os {
	case "darwin":
		serviceType = "LoadBalancer"
	default:
		serviceType = "NodePort"
	}
	return serviceType
}

func withLokiOptions(namespace string, port int) *helm.Options {
	return &helm.Options{
		KubectlOptions: &k8s.KubectlOptions{Namespace: namespace},
		SetValues: map[string]string{
			"pachd.lokiDeploy":             "true",
			"loki-stack.loki.service.type": exposedServiceType(),
			"loki-stack.promtail.initContainer.fsInotifyMaxUserInstances": "8000",
			"loki-stack.loki.service.port":                                fmt.Sprintf("%v", port+9),
			"loki-stack.loki.service.nodePort":                            fmt.Sprintf("%v", port+9),
			"loki-stack.loki.config.server.http_listen_port":              fmt.Sprintf("%v", port+9),
			"loki-stack.promtail.loki.servicePort":                        fmt.Sprintf("%v", port+9),
		},
		SetStrValues: map[string]string{
			"loki-stack.promtail.initContainer.enabled": "true",
		},
	}
}

func localDeploymentWithMinioOptions(namespace, image string) *helm.Options {
	return &helm.Options{
		KubectlOptions: &k8s.KubectlOptions{Namespace: namespace},
		SetValues: map[string]string{
			"deployTarget": "custom",

			"pachd.service.type":        "ClusterIP",
			"pachd.image.tag":           image,
			"pachd.clusterDeploymentID": "dev",

			"pachd.storage.backend":        "MINIO",
			"pachd.storage.minio.bucket":   "pachyderm-test",
			"pachd.storage.minio.endpoint": "minio.default.svc.cluster.local:9000",
			"pachd.storage.minio.id":       "minioadmin",
			"pachd.storage.minio.secret":   "minioadmin",

			"pachd.resources.requests.cpu":    "250m",
			"pachd.resources.requests.memory": "512M",
			"etcd.resources.requests.cpu":     "250m",
			"etcd.resources.requests.memory":  "512M",

			"global.postgresql.postgresqlPassword":         "pachyderm",
			"global.postgresql.postgresqlPostgresPassword": "pachyderm",

			// For tests, traffic from outside the cluster is routed through the proxy,
			// but we bind the internal k8s service ports to the same numbers for
			// in-cluster traffic, like enterprise registration.
			"proxy.enabled":                       "true",
			"proxy.service.type":                  exposedServiceType(),
			"proxy.service.httpPort":              "30650",
			"proxy.service.httpNodePort":          "30650",
			"pachd.service.apiGRPCPort":           "30650",
			"proxy.service.legacyPorts.oidc":      "30657",
			"pachd.service.oidcPort":              "30657",
			"proxy.service.legacyPorts.identity":  "30658",
			"pachd.service.identityPort":          "30658",
			"proxy.service.legacyPorts.s3Gateway": "30600",
			"pachd.service.s3GatewayPort":         "30600",
			"proxy.service.legacyPorts.metrics":   "30656",
			"pachd.service.prometheusPort":        "30656",
		},
		SetStrValues: map[string]string{
			"pachd.storage.minio.signature": "",
			"pachd.storage.minio.secure":    "false",
		},
	}
}

func withEnterprise(namespace string, address *grpcutil.PachdAddress) *helm.Options {
	return &helm.Options{
		KubectlOptions: &k8s.KubectlOptions{Namespace: namespace},
		SetValues: map[string]string{
			"pachd.enterpriseLicenseKeySecretName": licenseKeySecretName,
			"pachd.rootToken":                      testutil.RootToken,
			"pachd.oauthClientSecret":              "oidc-client-secret",
			"pachd.enterpriseSecret":               "enterprise-secret",
			// TODO: make these ports configurable to support IDP Login in parallel deployments
			"oidc.userAccessibleOauthIssuerHost": fmt.Sprintf("%s:30658", address.Host),
			"ingress.host":                       fmt.Sprintf("%s:30657", address.Host),
			// to test that the override works
			"global.postgresql.identityDatabaseFullNameOverride": "dexdb",
		},
	}
}

func withPort(namespace string, port uint16) *helm.Options {
	return &helm.Options{
		KubectlOptions: &k8s.KubectlOptions{Namespace: namespace},
		SetValues: map[string]string{
			// Run gRPC traffic through the full router.
			"proxy.service.httpPort":     fmt.Sprintf("%v", port),
			"proxy.service.httpNodePort": fmt.Sprintf("%v", port),

			// Let everything else use the legacy way.  We use the same mapping for
			// internal ports, so in-cluster traffic doesn't go through the proxy, but
			// doesn't need to confusingly use a different port number.
			"pachd.service.apiGRPCPort":           fmt.Sprintf("%v", port),
			"proxy.service.legacyPorts.oidc":      fmt.Sprintf("%v", port+7),
			"pachd.service.oidcPort":              fmt.Sprintf("%v", port+7),
			"proxy.service.legacyPorts.identity":  fmt.Sprintf("%v", port+8),
			"pachd.service.identityPort":          fmt.Sprintf("%v", port+8),
			"proxy.service.legacyPorts.s3Gateway": fmt.Sprintf("%v", port+3),
			"pachd.service.s3GatewayPort":         fmt.Sprintf("%v", port+3),
			"proxy.service.legacyPorts.metrics":   fmt.Sprintf("%v", port+4),
			"pachd.service.prometheusPort":        fmt.Sprintf("%v", port+4),
		},
	}
}

// withoutProxy disables the Pachyderm proxy.  It's used to test upgrades from versions of Pachyderm
// that didn't have the proxy.
func withoutProxy(namespace string) *helm.Options {
	return &helm.Options{
		KubectlOptions: &k8s.KubectlOptions{Namespace: namespace},
		SetValues: map[string]string{
			"proxy.enabled":      "false",
			"pachd.service.type": exposedServiceType(),
		},
	}
}

func union(a, b *helm.Options) *helm.Options {
	c := &helm.Options{
		SetValues:    make(map[string]string),
		SetStrValues: make(map[string]string),
	}
	copy := func(src, dst *helm.Options) {
		if src.KubectlOptions != nil {
			dst.KubectlOptions = &k8s.KubectlOptions{Namespace: src.KubectlOptions.Namespace}
		}
		for k, v := range src.SetValues {
			dst.SetValues[k] = v
		}
		for k, v := range src.SetStrValues {
			dst.SetStrValues[k] = v
		}
		if src.Version != "" {
			dst.Version = src.Version
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

func waitForLoki(t testing.TB, lokiHost string, lokiPort int) {
	require.NoError(t, backoff.Retry(func() error {
		req, _ := http.NewRequest("GET", fmt.Sprintf("http://%s:%v", lokiHost, lokiPort), nil)
		_, err := http.DefaultClient.Do(req)
		if err != nil {
			return errors.Wrap(err, "loki not ready")
		}
		return nil
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

func pachClient(t testing.TB, pachAddress *grpcutil.PachdAddress, authUser, namespace string) *client.APIClient {
	var c *client.APIClient
	// retry connecting if it doesn't immediately work
	require.NoError(t, backoff.Retry(func() error {
		t.Logf("Connecting to pachd on port: %v, in namespace: %s", pachAddress.Port, namespace)
		var err error
		c, err = client.NewFromPachdAddress(pachAddress, client.WithDialTimeout(10*time.Second))
		if err != nil {
			return errors.Wrapf(err, "failed to connect to pachd on port %v", pachAddress.Port)
		}
		// Ensure that pachd is really ready to receive requests.
		if _, err := c.InspectCluster(); err != nil {
			return errors.Wrapf(err, "failed to inspect cluster on port %v", pachAddress.Port)
		}
		return nil
	}, backoff.RetryEvery(time.Second).For(50*time.Second)))
	t.Logf("Success connecting to pachd on port: %v, in namespace: %s", pachAddress.Port, namespace)
	if authUser != "" {
		c = testutil.AuthenticateClient(t, c, authUser)
	}
	return c
}

func deleteRelease(t testing.TB, ctx context.Context, namespace string, kubeClient *kube.Clientset) {
	options := &helm.Options{
		KubectlOptions: &k8s.KubectlOptions{Namespace: namespace},
	}
	mu.Lock()
	err := helm.DeleteE(t, options, namespace, true)
	mu.Unlock()
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

func putRelease(t testing.TB, ctx context.Context, namespace string, kubeClient *kube.Clientset, f helmPutE, opts *DeployOpts) *client.APIClient {
	if opts.CleanupAfter {
		t.Cleanup(func() {
			deleteRelease(t, context.Background(), namespace, kubeClient)
		})
	}
	version := localImage
	chartPath := helmChartLocalPath(t)
	// TODO(acohen4): apply minio deployment to this namespace
	helmOpts := localDeploymentWithMinioOptions(namespace, version)
	if opts.Version != "" {
		version = opts.Version
		chartPath = helmChartPublishedPath
		helmOpts.Version = version
		helmOpts.SetValues["pachd.image.tag"] = version
	}
	pachAddress := getPachAddress(t)
	if opts.PortOffset != 0 {
		pachAddress.Port += opts.PortOffset
		helmOpts = union(helmOpts, withPort(namespace, pachAddress.Port))
	}
	if opts.Enterprise {
		createSecretEnterpriseKeySecret(t, ctx, kubeClient, namespace)
		helmOpts = union(helmOpts, withEnterprise(namespace, pachAddress))
	}
	if opts.Loki {
		helmOpts = union(helmOpts, withLokiOptions(namespace, int(pachAddress.Port)))
	}
	if !(opts.Version == "" || strings.HasPrefix(opts.Version, "2.3")) {
		helmOpts = union(helmOpts, withoutProxy(namespace))
	}
	if opts.ValueOverrides != nil {
		helmOpts = union(helmOpts, &helm.Options{SetValues: opts.ValueOverrides})
	}
	if err := f(t, helmOpts, chartPath, namespace); err != nil {
		if opts.UseLeftoverCluster {
			return pachClient(t, pachAddress, opts.AuthUser, namespace)
		}
		deleteRelease(t, context.Background(), namespace, kubeClient)
		// When CircleCI was experiencing network slowness, downloading
		// the Helm chart would sometimes fail.  Retrying it was
		// successful.
		require.NoErrorWithinTRetry(t, time.Minute, func() error { return f(t, helmOpts, chartPath, namespace) })
	}
	waitForPachd(t, ctx, kubeClient, namespace, version)
	if opts.Loki {
		waitForLoki(t, pachAddress.Host, int(pachAddress.Port)+9)
	}
	waitForPgbouncer(t, ctx, kubeClient, namespace)
	if opts.WaitSeconds > 0 {
		time.Sleep(time.Duration(opts.WaitSeconds) * time.Second)
	}
	return pachClient(t, pachAddress, opts.AuthUser, namespace)
}

// Deploy pachyderm using a `helm upgrade ...`
// returns an API Client corresponding to the deployment
func UpgradeRelease(t testing.TB, ctx context.Context, namespace string, kubeClient *kube.Clientset, opts *DeployOpts) *client.APIClient {
	return putRelease(t, ctx, namespace, kubeClient, helmLock(helm.UpgradeE), opts)
}

// Deploy pachyderm using a `helm install ...`
// returns an API Client corresponding to the deployment
func InstallRelease(t testing.TB, ctx context.Context, namespace string, kubeClient *kube.Clientset, opts *DeployOpts) *client.APIClient {
	return putRelease(t, ctx, namespace, kubeClient, helmLock(helm.InstallE), opts)
}
