package minikubetestenv

import (
	"bytes"
	"context"
	"crypto/x509"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
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

const (
	MinioEndpoint = "minio.default.svc.cluster.local:9000"
	MinioBucket   = "pachyderm-test"
)

var (
	mu           sync.Mutex // defensively lock around helm calls
	hostOverride *string    = flag.String("testenv.host", "", "override the default host used for testenv clusters")
	basePort     *int       = flag.Int("testenv.baseport", 0, "alternative base port for testenv to begin assigning clusters from")
)

type DeployOpts struct {
	Version            string
	Enterprise         bool
	Console            bool
	AuthUser           string
	CleanupAfter       bool
	UseLeftoverCluster bool
	// Because NodePorts are cluster-wide, we use a PortOffset to
	// assign separate ports per deployment.
	// NOTE: it might make more sense to declare port instead of offset
	PortOffset       uint16
	DisableLoki      bool
	WaitSeconds      int
	EnterpriseMember bool
	EnterpriseServer bool
	ValueOverrides   map[string]string
	TLS              bool
	CertPool         *x509.CertPool
	ValuesFiles      []string
}

type helmPutE func(t terraTest.TestingT, options *helm.Options, chart string, releaseName string) error

func getLocalImage() string {
	if sha := os.Getenv("TEST_IMAGE_SHA"); sha != "" {
		return sha
	}
	return localImage
}

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

func GetPachAddress(t testing.TB) *grpcutil.PachdAddress {
	addr := grpcutil.DefaultPachdAddress
	if *hostOverride != "" {
		addr.Host = *hostOverride
	} else if exposedServiceType() == "NodePort" {
		in := strings.NewReader("minikube ip")
		out := new(bytes.Buffer)
		cmd := exec.Command("/bin/bash")
		cmd.Stdin = in
		cmd.Stdout = out
		if err := cmd.Run(); err != nil {
			t.Errorf("'minikube ip': %v. Try running tests with the 'testenv.host' argument", err)
		}
		addr.Host = strings.TrimSpace(out.String())
	}
	if *basePort != 0 {
		addr.Port = uint16(*basePort)
	}
	return &addr
}

func exposedServiceType() string {
	os := runtime.GOOS
	serviceType := ""
	switch os {
	case "darwin", "windows":
		serviceType = "LoadBalancer"
	default:
		serviceType = "NodePort"
	}
	return serviceType
}

func withoutLokiOptions(namespace string, port int) *helm.Options {
	return &helm.Options{
		KubectlOptions: &k8s.KubectlOptions{Namespace: namespace},
		SetValues: map[string]string{
			"pachd.lokiDeploy":  "false",
			"pachd.lokiLogging": "false",
		},
	}
}

func withLokiOptions(namespace string, port int) *helm.Options {
	return &helm.Options{
		KubectlOptions: &k8s.KubectlOptions{Namespace: namespace},
		SetValues: map[string]string{
			"pachd.lokiDeploy":                                                "true",
			"pachd.lokiLogging":                                               "true",
			"loki-stack.loki.service.type":                                    exposedServiceType(),
			"loki-stack.loki.service.port":                                    fmt.Sprintf("%v", port+9),
			"loki-stack.loki.service.nodePort":                                fmt.Sprintf("%v", port+9),
			"loki-stack.loki.config.server.http_listen_port":                  fmt.Sprintf("%v", port+9),
			"loki-stack.promtail.config.serverPort":                           fmt.Sprintf("%v", port+9),
			"loki-stack.promtail.config.clients[0].url":                       fmt.Sprintf("http://%s-loki:%d/loki/api/v1/push", namespace, port+9),
			"loki-stack.promtail.initContainer[0].name":                       "init",
			"loki-stack.promtail.initContainer[0].image":                      "docker.io/busybox:1.33",
			"loki-stack.promtail.initContainer[0].imagePullPolicy":            "IfNotPresent",
			"loki-stack.promtail.initContainer[0].command[0]":                 "sh",
			"loki-stack.promtail.initContainer[0].command[1]":                 "-c",
			"loki-stack.promtail.initContainer[0].command[2]":                 "sysctl -w fs.inotify.max_user_instances=8000",
			"loki-stack.promtail.initContainer[0].securityContext.privileged": "true",
		},
	}
}

func withBase(namespace string) *helm.Options {
	return &helm.Options{
		KubectlOptions: &k8s.KubectlOptions{Namespace: namespace},
		SetValues: map[string]string{
			"pachd.clusterDeploymentID":       "dev",
			"pachd.resources.requests.cpu":    "250m",
			"pachd.resources.requests.memory": "512M",
			"etcd.resources.requests.cpu":     "250m",
			"etcd.resources.requests.memory":  "512M",
			"console.enabled":                 "false",
		},
	}
}

func withPachd(image string) *helm.Options {
	return &helm.Options{
		SetValues: map[string]string{
			"pachd.image.tag":    image,
			"pachd.service.type": "ClusterIP",
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
	}
}

func withMinio() *helm.Options {
	return &helm.Options{
		SetValues: map[string]string{
			"deployTarget":                 "custom",
			"pachd.storage.backend":        "MINIO",
			"pachd.storage.minio.bucket":   MinioBucket,
			"pachd.storage.minio.endpoint": MinioEndpoint,
			"pachd.storage.minio.id":       "minioadmin",
			"pachd.storage.minio.secret":   "minioadmin",
		},
		SetStrValues: map[string]string{
			"pachd.storage.minio.signature": "",
			"pachd.storage.minio.secure":    "false",
		},
	}
}

func withEnterprise(host, rootToken string, issuerPort, clientPort int) *helm.Options {
	return &helm.Options{
		SetValues: map[string]string{
			"pachd.enterpriseLicenseKeySecretName": licenseKeySecretName,
			"pachd.rootToken":                      rootToken,
			// TODO: make these ports configurable to support IDP Login in parallel deployments
			"oidc.userAccessibleOauthIssuerHost": fmt.Sprintf("%s:%v", host, issuerPort),
			"oidc.issuerURI":                     fmt.Sprintf("http://pachd:%v/dex", issuerPort),
			"proxy.host":                         fmt.Sprintf("%s:%v", host, clientPort),
			// to test that the override works
			"global.postgresql.identityDatabaseFullNameOverride": "dexdb",
		},
	}
}

func withEnterpriseServer(image, host string) *helm.Options {
	return &helm.Options{SetValues: map[string]string{
		"pachd.enabled":                      "false",
		"enterpriseServer.enabled":           "true",
		"enterpriseServer.image.tag":         image,
		"oidc.mockIDP":                       "true",
		"oidc.issuerURI":                     "http://pach-enterprise.enterprise.svc.cluster.local:31658/dex",
		"oidc.userAccessibleOauthIssuerHost": fmt.Sprintf("%s:31658", host),
		"pachd.oauthClientID":                "enterprise-pach",
		"pachd.oauthRedirectURI":             fmt.Sprintf("http://%s:31657/authorization-code/callback", host),
		"enterpriseServer.service.type":      "ClusterIP",
		// For tests, traffic from outside the cluster is routed through the proxy,
		// but we bind the internal k8s service ports to the same numbers for
		// in-cluster traffic, like enterprise registration.
		"proxy.enabled":                      "true",
		"proxy.service.type":                 exposedServiceType(),
		"proxy.service.httpPort":             "31650",
		"proxy.service.httpNodePort":         "31650",
		"pachd.service.apiGRPCPort":          "31650",
		"proxy.service.legacyPorts.oidc":     "31657",
		"pachd.service.oidcPort":             "31657",
		"proxy.service.legacyPorts.identity": "31658",
		"pachd.service.identityPort":         "31658",
		"proxy.service.legacyPorts.metrics":  "31656",
		"pachd.service.prometheusPort":       "31656",
	}}
}

func withEnterpriseMember(host string, grpcPort int) *helm.Options {
	return &helm.Options{SetValues: map[string]string{
		"pachd.activateEnterpriseMember":     "true",
		"pachd.enterpriseServerAddress":      "grpc://pach-enterprise.enterprise.svc.cluster.local:31650",
		"pachd.enterpriseCallbackAddress":    fmt.Sprintf("grpc://pachd.default.svc.cluster.local:%v", grpcPort),
		"pachd.enterpriseRootToken":          testutil.RootToken,
		"oidc.issuerURI":                     "http://pach-enterprise.enterprise.svc.cluster.local:31658/dex",
		"oidc.userAccessibleOauthIssuerHost": fmt.Sprintf("%s:31658", host),
	}}
}

func withPort(namespace string, port uint16, tls bool) *helm.Options {
	opts := &helm.Options{
		KubectlOptions: &k8s.KubectlOptions{Namespace: namespace},
		SetValues:      map[string]string{},
	}

	// Allow the internal use to have the same port numbers as the proxy.
	opts.SetValues["pachd.service.apiGRPCPort"] = fmt.Sprintf("%v", port)
	opts.SetValues["pachd.service.oidcPort"] = fmt.Sprintf("%v", port+7)
	opts.SetValues["pachd.service.identityPort"] = fmt.Sprintf("%v", port+8)
	opts.SetValues["pachd.service.s3GatewayPort"] = fmt.Sprintf("%v", port+3)
	opts.SetValues["pachd.service.prometheusPort"] = fmt.Sprintf("%v", port+4)

	if tls {
		// Use the default port for HTTPS.
		opts.SetValues["proxy.service.httpsPort"] = fmt.Sprintf("%v", port)
		opts.SetValues["proxy.service.httpsNodePort"] = fmt.Sprintf("%v", port)

		// Disable HTTP.
		opts.SetValues["proxy.service.httpPort"] = "0"
		opts.SetValues["proxy.service.httpNodePort"] = "0"

		// Disable the legacy services, since anything depending on the TLS argument to this
		// function expects to do everything through the proxy
		opts.SetValues["proxy.service.legacyPorts.console"] = "0"
		opts.SetValues["proxy.service.legacyPorts.oidc"] = "0"
		opts.SetValues["proxy.service.legacyPorts.identity"] = "0"
		opts.SetValues["proxy.service.legacyPorts.s3Gateway"] = "0"
		opts.SetValues["proxy.service.legacyPorts.metrics"] = "0"
	} else {
		// Run gRPC traffic through the full router.
		opts.SetValues["proxy.service.httpPort"] = fmt.Sprintf("%v", port)
		opts.SetValues["proxy.service.httpNodePort"] = fmt.Sprintf("%v", port)

		// Let the legacy ports go through the proxy.
		opts.SetValues["proxy.service.legacyPorts.oidc"] = fmt.Sprintf("%v", port+7)
		opts.SetValues["proxy.service.legacyPorts.identity"] = fmt.Sprintf("%v", port+8)
		opts.SetValues["proxy.service.legacyPorts.s3Gateway"] = fmt.Sprintf("%v", port+3)
		opts.SetValues["proxy.service.legacyPorts.metrics"] = fmt.Sprintf("%v", port+4)
	}
	return opts
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

func waitForPachd(t testing.TB, ctx context.Context, kubeClient *kube.Clientset, namespace, version string, enterpriseServer bool) {
	label := "app=pachd"
	if enterpriseServer {
		label = "app=pach-enterprise"
	}
	require.NoErrorWithinTRetry(t, 5*time.Minute, func() error {
		pachds, err := kubeClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: label})
		if err != nil {
			return errors.Wrap(err, "error on pod list")
		}
		var unacceptablePachds []string
		for _, p := range pachds.Items {
			if p.Status.Phase == v1.PodRunning && strings.HasSuffix(p.Spec.Containers[0].Image, ":"+version) && p.Status.ContainerStatuses[0].Ready && len(pachds.Items) == 1 {
				return nil
			}
			unacceptablePachds = append(unacceptablePachds, fmt.Sprintf("%v: image=%v status=%#v", p.Name, p.Spec.Containers[0].Image, p.Status))
		}
		return errors.Errorf("deployment in progress: pachds: %v", strings.Join(unacceptablePachds, "; "))
	})
}

func waitForLoki(t testing.TB, lokiHost string, lokiPort int) {
	require.NoError(t, backoff.Retry(func() error {
		req, _ := http.NewRequest("GET", fmt.Sprintf("http://%s:%v/ready", lokiHost, lokiPort), nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return errors.Wrap(err, "loki not ready")
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return errors.Errorf("loki not ready")
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

func pachClient(t testing.TB, pachAddress *grpcutil.PachdAddress, authUser, namespace string, certpool *x509.CertPool) *client.APIClient {
	var c *client.APIClient
	// retry connecting if it doesn't immediately work
	require.NoError(t, backoff.Retry(func() error {
		t.Logf("Connecting to pachd on port: %v, in namespace: %s", pachAddress.Port, namespace)
		var err error
		opts := []client.Option{client.WithDialTimeout(10 * time.Second)}
		if certpool != nil {
			opts = append(opts, client.WithCertPool(certpool))
		}
		c, err = client.NewFromPachdAddress(pachAddress, opts...)
		if err != nil {
			t.Logf("retryable: failed to connect to pachd on port %v: %v", pachAddress.Port, err)
			return errors.Wrapf(err, "failed to connect to pachd on port %v", pachAddress.Port)
		}
		// Ensure that pachd is really ready to receive requests.
		if _, err := c.InspectCluster(); err != nil {
			t.Logf("retryable: failed to inspect cluster on port %v: %v", pachAddress.Port, err)
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
	require.True(t, err == nil || strings.Contains(err.Error(), "already exists"), "Error '%v' does not contain 'already exists'", err)
}

func putRelease(t testing.TB, ctx context.Context, namespace string, kubeClient *kube.Clientset, f helmPutE, opts *DeployOpts) *client.APIClient {
	if opts.CleanupAfter {
		t.Cleanup(func() {
			deleteRelease(t, context.Background(), namespace, kubeClient)
		})
	}
	version := getLocalImage()
	chartPath := helmChartLocalPath(t)
	helmOpts := withBase(namespace)
	if opts.Version != "" {
		version = opts.Version
		chartPath = helmChartPublishedPath
		helmOpts.Version = version
		helmOpts.SetValues["pachd.image.tag"] = version
	}
	pachAddress := GetPachAddress(t)
	if opts.Enterprise || opts.EnterpriseServer {
		createSecretEnterpriseKeySecret(t, ctx, kubeClient, namespace)
		issuerPort := int(pachAddress.Port+opts.PortOffset) + 8
		if opts.EnterpriseMember {
			issuerPort = 31658
		}
		helmOpts = union(helmOpts, withEnterprise(pachAddress.Host, testutil.RootToken, issuerPort, int(pachAddress.Port+opts.PortOffset)+7))
	}
	if opts.EnterpriseServer {
		helmOpts = union(helmOpts, withEnterpriseServer(version, pachAddress.Host))
		pachAddress.Port = uint16(31650)
	} else {
		helmOpts = union(helmOpts, withPachd(version))
		// TODO(acohen4): apply minio deployment to this namespace
		helmOpts = union(helmOpts, withMinio())
	}
	if opts.PortOffset != 0 {
		pachAddress.Port += opts.PortOffset
		helmOpts = union(helmOpts, withPort(namespace, pachAddress.Port, opts.TLS))
	}
	if opts.EnterpriseMember {
		helmOpts = union(helmOpts, withEnterpriseMember(pachAddress.Host, int(pachAddress.Port)))
	}
	if opts.Console {
		helmOpts.SetValues["console.enabled"] = "true"
	}
	if opts.DisableLoki {
		helmOpts = union(helmOpts, withoutLokiOptions(namespace, int(pachAddress.Port)))
	} else {
		helmOpts = union(helmOpts, withLokiOptions(namespace, int(pachAddress.Port)))
	}
	if !(opts.Version == "" || strings.HasPrefix(opts.Version, "2.3")) {
		helmOpts = union(helmOpts, withoutProxy(namespace))
	}
	if opts.ValueOverrides != nil {
		helmOpts = union(helmOpts, &helm.Options{SetValues: opts.ValueOverrides})
	}
	if opts.TLS {
		pachAddress.Secured = true
	}
	helmOpts.ValuesFiles = opts.ValuesFiles
	if err := f(t, helmOpts, chartPath, namespace); err != nil {
		if opts.UseLeftoverCluster {
			return pachClient(t, pachAddress, opts.AuthUser, namespace, opts.CertPool)
		}
		deleteRelease(t, context.Background(), namespace, kubeClient)
		// When CircleCI was experiencing network slowness, downloading
		// the Helm chart would sometimes fail.  Retrying it was
		// successful.
		require.NoErrorWithinTRetry(t, time.Minute, func() error { return f(t, helmOpts, chartPath, namespace) })
	}
	waitForPachd(t, ctx, kubeClient, namespace, version, opts.EnterpriseServer)
	if !opts.DisableLoki {
		waitForLoki(t, pachAddress.Host, int(pachAddress.Port)+9)
	}
	waitForPgbouncer(t, ctx, kubeClient, namespace)
	if opts.WaitSeconds > 0 {
		time.Sleep(time.Duration(opts.WaitSeconds) * time.Second)
	}
	return pachClient(t, pachAddress, opts.AuthUser, namespace, opts.CertPool)
}

func PutNamespace(t testing.TB, namespace string) {
	kube := testutil.GetKubeClient(t)
	if _, err := kube.CoreV1().Namespaces().Get(context.Background(), namespace, metav1.GetOptions{}); err != nil {
		_, err := kube.CoreV1().Namespaces().Create(context.Background(),
			&v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
				},
			},
			metav1.CreateOptions{})
		require.NoError(t, err)
	}
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
