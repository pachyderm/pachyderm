package minikubetestenv

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	coordination "k8s.io/api/coordination/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/net"
	kube "k8s.io/client-go/kubernetes"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

const (
	helmChartPublishedPath = "pachyderm/pachyderm"
	localImage             = "local"
	licenseKeySecretName   = "enterprise-license-key-secret"
	leasePrefix            = "minikubetestenv-"
	defaultLeaseDuration   = 30
)

const (
	MinioEndpoint = "minio.default.svc.cluster.local:9000"
	MinioBucket   = "pachyderm-test"
)

const (
	determinedRegistry       = "registry-1.docker.io/determinedai"
	determinedRegistrySecret = "detregcred"
	determinedLoginSecret    = "detlogin"
	helmOptsConfigMap        = "testhelmopts"
)

var (
	mu           sync.Mutex // defensively lock around helm calls
	hostOverride *string    = flag.String("testenv.host", "", "override the default host used for testenv clusters")
	basePort     *int       = flag.Int("testenv.baseport", 0, "alternative base port for testenv to begin assigning clusters from")
)

var (
	detDockerUser = os.Getenv("DET_DOCKER_USER")
	detDockerPass = os.Getenv("DET_DOCKER_PASS")
)

type DeployOpts struct {
	Version string
	// Necessarry to enable auth and activate enterprise.
	// EnterpriseServer and EnterpriseMember can be used
	// without this flag to test manual enterprise/auth activation.
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
	EnterpriseMember bool
	EnterpriseServer bool
	Determined       bool
	ValueOverrides   map[string]string
	TLS              bool
	CertPool         *x509.CertPool
	ValuesFiles      []string
}

func getLocalImage() string {
	if sha := os.Getenv("TEST_IMAGE_SHA"); sha != "" {
		return sha
	}
	return localImage
}

func helmChartLocalPath(t testing.TB) string {
	return localPath(t, "etc", "helm", "pachyderm")
}
func exampleValuesLocalPath(t testing.TB, fileName string) string {
	return localPath(t, "etc", "helm", "examples", fileName)
}

func localPath(t testing.TB, pathParts ...string) string {
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
	relPathParts = append(relPathParts, pathParts...)
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
			"loki-stack.persistence.size":                                     "5Gi",
		},
	}
}

func withBase(namespace string) *helm.Options {
	return &helm.Options{
		KubectlOptions: &k8s.KubectlOptions{Namespace: namespace},
		SetValues: map[string]string{
			"pachd.clusterDeploymentID":           "dev",
			"pachd.resources.requests.cpu":        "250m",
			"pachd.resources.requests.memory":     "512M",
			"etcd.resources.requests.cpu":         "250m",
			"etcd.resources.requests.memory":      "512M",
			"pachd.defaultPipelineCPURequest":     "100m",
			"pachd.defaultPipelineMemoryRequest":  "64M",
			"pachd.defaultPipelineStorageRequest": "100Mi",
			"pachd.defaultSidecarCPURequest":      "100m",
			"pachd.defaultSidecarMemoryRequest":   "64M",
			"pachd.defaultSidecarStorageRequest":  "100Mi",
			"console.enabled":                     "false",
			"postgresql.persistence.size":         "5Gi",
			"etcd.storageSize":                    "5Gi",
			"pachw.resources.requests.cpu":        "250m",
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
		"pachd.enabled":                           "false",
		"enterpriseServer.enabled":                "true",
		"enterpriseServer.image.tag":              image,
		"oidc.mockIDP":                            "true",
		"oidc.issuerURI":                          "http://pach-enterprise.enterprise.svc.cluster.local:31658/dex",
		"oidc.userAccessibleOauthIssuerHost":      fmt.Sprintf("%s:31658", host),
		"pachd.oauthClientID":                     "enterprise-pach",
		"pachd.oauthRedirectURI":                  fmt.Sprintf("http://%s:31657/authorization-code/callback", host),
		"enterpriseServer.service.type":           "ClusterIP",
		"enterpriseServer.resources.requests.cpu": "250m",
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

func formatSince(t *metav1.Time) string {
	if t == nil {
		return "∅"
	}
	return time.Since(t.Time).Round(time.Second).String()
}

func formatPodStatus(s v1.PodStatus) string {
	return fmt.Sprintf("{phase:%v message:%v reason:%v nominatedNodeName:%v hostIP:%v podIP:%v startTime:%v conditions:%v containers:{init:%v ephemeral:%v normal:%v}}", s.Phase, s.Message, s.Reason, s.NominatedNodeName, s.HostIP, s.PodIP, formatSince(s.StartTime), formatPodConditions(s.Conditions), formatContainerStatuses(s.InitContainerStatuses), formatContainerStatuses(s.EphemeralContainerStatuses), formatContainerStatuses(s.ContainerStatuses))
}

func formatPodConditions(cs []v1.PodCondition) string {
	var result []string
	for _, c := range cs {
		result = append(result, fmt.Sprintf("%v=%v@%v", c.Type, c.Status, formatSince(&c.LastTransitionTime)))
	}
	return "{" + strings.Join(result, " ") + "}"
}

func formatContainerStatuses(ss []v1.ContainerStatus) (result []string) {
	for _, s := range ss {
		started := "∅"
		if s.Started != nil {
			started = strconv.FormatBool(*s.Started)
		}
		var state string
		switch {
		case s.State.Waiting != nil:
			x := s.State.Waiting
			state = fmt.Sprintf("waiting{reason:%v message:%v}", x.Reason, x.Message)
		case s.State.Running != nil:
			x := s.State.Running
			state = fmt.Sprintf("running{started:%v}", formatSince(&x.StartedAt))
		case s.State.Terminated != nil:
			x := s.State.Terminated
			state = fmt.Sprintf("terminated{reason:%v message:%v started:%v finished:%v code:%v}", x.Reason, x.Message, formatSince(&x.StartedAt), formatSince(&x.FinishedAt), x.ExitCode)
		}
		result = append(result, fmt.Sprintf("{name:%v started:%v ready:%v state:%v}", s.Name, started, s.Ready, state))
	}
	return result
}

func waitForPachd(t testing.TB, ctx context.Context, kubeClient *kube.Clientset, namespace, version string, enterpriseServer bool) {
	label := "app=pachd"
	if enterpriseServer {
		label = "app=pach-enterprise"
	}
	require.NoErrorWithinTRetryConstant(t, 5*time.Minute, func() error {
		pachds, err := kubeClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: label})
		if err != nil {
			return errors.Wrap(err, "error on pod list")
		}
		var unacceptablePachds []string
		var acceptablePachds []string
		for _, p := range pachds.Items {
			if p.Status.Phase == v1.PodRunning && strings.HasSuffix(p.Spec.Containers[0].Image, ":"+version) && p.Status.ContainerStatuses[0].Ready {
				acceptablePachds = append(acceptablePachds, fmt.Sprintf("%v: image=%v status=%s", p.Name, p.Spec.Containers[0].Image, formatPodStatus(p.Status)))
			} else {
				unacceptablePachds = append(unacceptablePachds, fmt.Sprintf("%v: image=%v status=%s", p.Name, p.Spec.Containers[0].Image, formatPodStatus(p.Status)))
			}
		}
		if len(acceptablePachds) > 0 && (len(unacceptablePachds) == 0) {
			return nil
		}
		return errors.Errorf("deployment in progress pachds ready:\n %v unacceptable pachds: %v \n %v acceptable pachds: %v",
			len(unacceptablePachds),
			strings.Join(unacceptablePachds, "; "),
			len(acceptablePachds),
			strings.Join(acceptablePachds, "; "),
		)
	}, 5*time.Second)
}

func checkForPachd(t testing.TB, ctx context.Context, kubeClient *kube.Clientset, namespace string, enterpriseServer bool) (bool, error) {
	label := "app=pachd"
	if enterpriseServer {
		label = "app=pach-enterprise"
	}
	pachds, err := kubeClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: label})
	if k8serrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, errors.Wrap(err, "error on pod list to check pachd existence")
	}
	return len(pachds.Items) > 0, nil
}

func waitForLoki(t testing.TB, lokiHost string, lokiPort int) {
	require.NoError(t, backoff.RetryNotify(func() error {
		client := http.Client{
			Timeout: 15 * time.Second,
		}
		req, _ := http.NewRequest("GET", fmt.Sprintf("http://%s:%v/ready", lokiHost, lokiPort), nil)
		t.Logf("Attempting to connect to loki at lokiHost %v and lokiPort %v", lokiHost, lokiPort)
		resp, err := client.Do(req)
		if os.IsTimeout(err) {
			return errors.Wrap(err, "loki attempt to connect timed out")
		}
		if err != nil {
			return errors.Wrap(err, "loki not ready due to error")
		}
		t.Logf("Connected to loki at lokiHost %v and lokiPort %v", lokiHost, lokiPort)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return errors.Errorf("loki not ready. http response code %v", resp.StatusCode)
		}
		return nil
	}, backoff.RetryEvery(5*time.Second).For(5*time.Minute), func(err error, d time.Duration) error {
		t.Logf("Retrying connection to loki at lokiHost %v and lokiPort %v. Error: %v", lokiHost, lokiPort, err)
		return nil
	}))
}

func waitForPgbouncer(t testing.TB, ctx context.Context, kubeClient *kube.Clientset, namespace string) {
	waitForLabeledPod(t, ctx, kubeClient, namespace, "app=pg-bouncer")
}

func waitForPostgres(t testing.TB, ctx context.Context, kubeClient *kube.Clientset, namespace string) {
	waitForLabeledPod(t, ctx, kubeClient, namespace, "app.kubernetes.io/name=postgresql")
}

func waitForDetermined(t testing.TB, ctx context.Context, kubeClient *kube.Clientset, namespace string) {
	waitForLabeledPod(t, ctx, kubeClient, namespace, "determined-system=master")
}

func waitForLabeledPod(t testing.TB, ctx context.Context, kubeClient *kube.Clientset, namespace string, label string) {
	require.NoError(t, backoff.Retry(func() error {
		pbs, err := kubeClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: label})
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
		c, err = client.NewFromPachdAddress(pctx.TODO(), pachAddress, opts...)
		if err != nil {
			t.Logf("retryable: failed to connect to pachd on port %v: %v", pachAddress.Port, err)
			return errors.Wrapf(err, "failed to connect to pachd on port %v", pachAddress.Port)
		}
		// Ensure that pachd is really ready to receive requests.
		if _, err := c.InspectCluster(); err != nil {
			scrubbedErr := grpcutil.ScrubGRPC(err)
			t.Logf("retryable: failed to inspect cluster on port %v: %v", pachAddress.Port, scrubbedErr)
			return errors.Wrapf(scrubbedErr, "failed to inspect cluster on port %v", pachAddress.Port)
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
	t.Logf("Deleting release in namespace %v", namespace)
	options := &helm.Options{
		KubectlOptions: &k8s.KubectlOptions{Namespace: namespace},
	}
	mu.Lock()
	err := helm.DeleteE(t, options, namespace, true)
	mu.Unlock()
	require.True(t, err == nil || strings.Contains(err.Error(), "not found"))

	require.NoError(t, kubeClient.CoreV1().ConfigMaps(namespace).DeleteCollection(ctx, *metav1.NewDeleteOptions(0), metav1.ListOptions{LabelSelector: "suite=pachyderm"}))
	require.NoError(t, kubeClient.CoreV1().PersistentVolumeClaims(namespace).DeleteCollection(ctx, *metav1.NewDeleteOptions(0), metav1.ListOptions{LabelSelector: "suite=pachyderm"}))
	require.NoError(t, kubeClient.CoreV1().PersistentVolumeClaims(namespace).DeleteCollection(ctx, *metav1.NewDeleteOptions(0), metav1.ListOptions{LabelSelector: "app=loki"}))
	require.NoError(t, backoff.Retry(func() error {
		pachPvcs, err := kubeClient.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{LabelSelector: "suite=pachyderm"})
		if err != nil {
			return errors.Wrap(err, "error on pachyderm pvc list")
		}
		lokiPvcs, err := kubeClient.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{LabelSelector: "app=loki"})
		if err != nil {
			return errors.Wrap(err, "error on loki pvc list")
		}
		// All Determined pvcs are removed by the helm uninstall, but
		// we should still wait for them because if they are deleted
		// by the storage provisioner just when a new cluster starts it
		// can prevent the new pvc from being created.
		detPvcs, err := kubeClient.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{LabelSelector: fmt.Sprintf("app=determined-db-%s", namespace)})
		if err != nil {
			return errors.Wrap(err, "error on det pvc list")
		}
		if len(pachPvcs.Items) == 0 && len(lokiPvcs.Items) == 0 && len(detPvcs.Items) == 0 {
			return nil
		}
		return errors.Errorf("pvcs have yet to be deleted. Namespace: %s PVCs left - pach: %d loki: %d det: %d", namespace, len(pachPvcs.Items), len(lokiPvcs.Items), len(detPvcs.Items))
	}, backoff.RetryEvery(5*time.Second).For(2*time.Minute)))
}

// returns the Nodeport url for accessing the determined service via REST/HTTP with an empty Path
func DetNodeportHttpUrl(t testing.TB, namespace string) *url.URL {
	ctx := context.Background()
	kube := testutil.GetKubeClient(t)
	service, err := kube.CoreV1().Services(namespace).Get(ctx, fmt.Sprintf("determined-master-service-%s", namespace), metav1.GetOptions{})
	detPort := service.Spec.Ports[0].NodePort
	require.NoError(t, err, "Fininding Determined service")
	node, err := kube.CoreV1().Nodes().Get(ctx, "minikube", metav1.GetOptions{})
	require.NoError(t, err, "Fininding node for Determined")
	var detHost string
	for _, addr := range node.Status.Addresses {
		if addr.Type == "InternalIP" {
			detHost = addr.Address
		}
	}
	detUrl := net.FormatURL("http", detHost, int(detPort), "")
	return detUrl
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

// Create the secret kubernetes uses to pull the Determined image
func createSecretDeterminedRegcred(t testing.TB, ctx context.Context, kubeClient *kube.Clientset, ns string) {
	require.NotEqual(t, "", detDockerUser, "Missing required user for Determined integration testing")
	require.NotEqual(t, "", detDockerPass, "Missing required password for Determined integration testing")
	dockerConfig, err := json.Marshal(
		map[string]any{
			"auths": map[string]any{
				determinedRegistry: map[string]string{
					"username": detDockerUser,
					"password": detDockerPass,
				},
			},
		},
	)
	require.NoError(t, err, "Marshalling determined registry credentials")
	_, err = kubeClient.CoreV1().Secrets(ns).Create(ctx, &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      determinedRegistrySecret,
			Namespace: ns,
		},
		Type: "kubernetes.io/dockerconfigjson",
		StringData: map[string]string{
			".dockerconfigjson": string(dockerConfig),
		},
	}, metav1.CreateOptions{})
	require.True(t, err == nil || strings.Contains(err.Error(), "already exists"), "Error '%v' does not contain 'already exists' with Determined regcred secret setup", err)
}

// Create the secret that pachd uses to connect
func createSecretDeterminedLogin(t testing.TB, ctx context.Context, kubeClient *kube.Clientset, ns string) {
	_, err := kubeClient.CoreV1().Secrets(ns).Create(ctx, &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      determinedLoginSecret,
			Namespace: ns,
		},
		StringData: map[string]string{
			"determined-username": "admin",
			"determined-password": "",
		},
	}, metav1.CreateOptions{})
	require.True(t, err == nil || strings.Contains(err.Error(), "already exists"), "Error '%v' does not contain 'already exists' with Determined login secret setup", err)
}

func determinedPriorityClassesExist(t testing.TB, ctx context.Context, kubeClient *kube.Clientset) bool {
	pcs, err := kubeClient.SchedulingV1().PriorityClasses().List(ctx, metav1.ListOptions{})
	require.NoError(t, err, "finding determined priorityclasses")
	for _, pc := range pcs.Items {
		if pc.Name == "determined-medium-priority" || pc.Name == "determined-system-priority" {
			return true
		}
	}
	return false
}

// deletes the existing configmap and writes the new hel opts to this one
func createOptsConfigMap(t testing.TB, ctx context.Context, kubeClient *kube.Clientset, ns string, helmOpts *helm.Options, chartPath string) {
	err := kubeClient.CoreV1().ConfigMaps(ns).Delete(ctx, helmOptsConfigMap, *metav1.NewDeleteOptions(0))
	require.True(t, err == nil || k8serrors.IsNotFound(err), "deleting configMap error: %v", err)
	_, err = kubeClient.CoreV1().ConfigMaps(ns).Create(ctx, &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      helmOptsConfigMap,
			Namespace: ns,
			Labels:    map[string]string{"suite": "pachyderm"},
		},
		BinaryData: map[string][]byte{
			"helmopts-hash": hashOpts(t, helmOpts, chartPath),
		},
	}, metav1.CreateOptions{})
	require.NoError(t, err, "creating helmOpts ConfigMap")
}

func getOptsConfigMapData(t testing.TB, ctx context.Context, kubeClient *kube.Clientset, ns string) []byte {
	configMap, err := kubeClient.CoreV1().ConfigMaps(ns).Get(ctx, helmOptsConfigMap, metav1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		return []byte{}
	}
	require.NoError(t, err, "getting helmOpts ConfigMap")
	return configMap.BinaryData["helmopts-hash"]
}

func hashOpts(t testing.TB, helmOpts *helm.Options, chartPath string) []byte { // return string for consistency with configMap data
	// we don't need to deserialize, we just use this to check differences, so just save a hash
	helmBytes, err := json.Marshal(helmOpts) // lists might be ordered differently, but there are not many list fields that we actively use that would be different in practice
	require.NoError(t, err)
	optsHash := sha256.New()
	_, err = optsHash.Write(helmBytes)
	require.NoError(t, err)
	_, err = optsHash.Write([]byte(chartPath))
	require.NoError(t, err)
	return optsHash.Sum(nil)
}

func putLease(t testing.TB, namespace string) error {
	t.Logf("Creating lease on namespace %s for test %v", namespace, t.Name())
	kube := testutil.GetKubeClient(t)
	identity := t.Name()
	leaseDuration := int32(defaultLeaseDuration)
	acquireTime := metav1.NewMicroTime(time.Now())
	lease, err := kube.CoordinationV1().Leases(namespace).Create(context.Background(), &coordination.Lease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s%s", leasePrefix, namespace),
			Namespace: namespace,
		},
		Spec: coordination.LeaseSpec{
			HolderIdentity:       &identity,
			LeaseDurationSeconds: &leaseDuration,
			AcquireTime:          &acquireTime,
			RenewTime:            &acquireTime,
		},
	}, metav1.CreateOptions{})
	if err == nil {
		leaseCtx, leaseCancel := pctx.WithCancel(pctx.Background("minikube test lease renewer"))
		go leaseRenewer(t, leaseCtx, kube, lease, time.Duration(defaultLeaseDuration/2)*time.Second)
		t.Cleanup(func() {
			leaseCancel()
			err := kube.CoordinationV1().Leases(namespace).Delete(context.Background(), lease.Name, metav1.DeleteOptions{})
			require.NoError(t, err)
		})
	}
	return errors.EnsureStack(err)
}

// renews the lease every interval fo continued use, cancelled with the context
func leaseRenewer(t testing.TB, ctx context.Context, kube *kube.Clientset, initialLease *coordination.Lease, interval time.Duration) {
	latestLease := initialLease
	for {
		select {
		case <-time.After(interval):
			t.Logf("Renewing lease on namespace %s for test %v", latestLease.GetNamespace(), t.Name())
			err := backoff.Retry(func() error {
				renewTime := metav1.NewMicroTime(time.Now())
				var err error
				latestLease, err = kube.CoordinationV1().Leases(latestLease.GetNamespace()).Update(ctx, &coordination.Lease{
					ObjectMeta: metav1.ObjectMeta{
						Name:            latestLease.GetName(),
						Namespace:       latestLease.GetNamespace(),
						ResourceVersion: latestLease.GetResourceVersion(),
					},
					Spec: coordination.LeaseSpec{
						HolderIdentity:       latestLease.Spec.HolderIdentity,
						LeaseDurationSeconds: latestLease.Spec.LeaseDurationSeconds,
						AcquireTime:          latestLease.Spec.AcquireTime,
						RenewTime:            &renewTime,
					},
				}, metav1.UpdateOptions{})
				return errors.EnsureStack(err)
			}, backoff.RetryEvery(time.Millisecond*200).For(time.Second))
			if err != nil {
				t.Logf("Could not renew lease on %v for test %v", latestLease.Namespace, t.Name())
				// doesn't exist or it maybe is in progress deleting, or a new one was re-created.
				// lots of reasons it could be =in a weird state, but we don't want to fail the test over that.
				// We retried so it's probably not an ephemeral issue.
				// Just log it, release the lease, and let the main process recover.
				return
			}
		case <-ctx.Done():
			t.Logf("Releasing lease on namespace %s for test %v", latestLease.Namespace, t.Name())
			return
		}
	}
}

func putRelease(t testing.TB, ctx context.Context, namespace string, kubeClient *kube.Clientset, mustUpgrade bool, opts *DeployOpts) *client.APIClient {
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
	if opts.Enterprise {
		createSecretEnterpriseKeySecret(t, ctx, kubeClient, namespace)
		issuerPort := int(pachAddress.Port+opts.PortOffset) + 8
		if opts.EnterpriseMember {
			issuerPort = 31658
		}
		helmOpts = union(helmOpts, withEnterprise(pachAddress.Host, testutil.RootToken, issuerPort, int(pachAddress.Port+opts.PortOffset)+7))
	}
	if opts.EnterpriseServer {
		helmOpts = union(helmOpts, withEnterpriseServer(version, pachAddress.Host))
		helmOpts = union(helmOpts, withMinio())
		pachAddress.Port = uint16(31650)
	} else {
		helmOpts = union(helmOpts, withPachd(version))
		// TODO(acohen4): apply minio deployment to this namespace
		helmOpts = union(helmOpts, withMinio())
	}
	if opts.Determined {
		createSecretDeterminedRegcred(t, ctx, kubeClient, namespace)
		createSecretDeterminedLogin(t, ctx, kubeClient, namespace)
		valuesTemplate, err := template.ParseFiles(exampleValuesLocalPath(t, "int-test-values-with-det.yaml"))
		require.NoError(t, err, "Creating determined values template")
		valuesFile, err := os.CreateTemp("", "detvalues.*.yaml")
		require.NoError(t, err, "Creating determined values temp file")
		defer valuesFile.Close()
		err = valuesTemplate.Execute(valuesFile, struct{ K8sNamespace string }{K8sNamespace: namespace})
		require.NoError(t, err, "Error templating determined values temp file")
		opts.ValuesFiles = append([]string{valuesFile.Name()}, opts.ValuesFiles...) // we want any user specified values files to be applied after
		if !mustUpgrade && determinedPriorityClassesExist(t, ctx, kubeClient) {     // installing on a seperate namespace with determined means we can't make priority classes
			opts.ValueOverrides["determined.createNonNamespacedObjects"] = "false"
		}
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
	waitForInstallFinished := func() {
		createOptsConfigMap(t, ctx, kubeClient, namespace, helmOpts, chartPath)

		waitForPachd(t, ctx, kubeClient, namespace, version, opts.EnterpriseServer)

		if !opts.DisableLoki {
			waitForLoki(t, pachAddress.Host, int(pachAddress.Port)+9)
		}
		waitForPgbouncer(t, ctx, kubeClient, namespace)
		waitForPostgres(t, ctx, kubeClient, namespace)
		if opts.Determined {
			waitForDetermined(t, ctx, kubeClient, namespace)
		}
	}
	previousOptsHash := getOptsConfigMapData(t, ctx, kubeClient, namespace)
	pachdExists, err := checkForPachd(t, ctx, kubeClient, namespace, opts.EnterpriseServer)
	require.NoError(t, err)
	if mustUpgrade {
		t.Logf("Test must upgrade cluster in place, upgrading cluster with new opts in %v", namespace)
		require.NoErrorWithinTRetry(t,
			time.Minute,
			func() error {
				return errors.EnsureStack(helm.UpgradeE(t, helmOpts, chartPath, namespace))
			})
		waitForInstallFinished()
	} else if !bytes.Equal(previousOptsHash, hashOpts(t, helmOpts, chartPath)) ||
		!opts.UseLeftoverCluster ||
		!pachdExists { // In case somehow a config map got left without a corresponding installed release. Cleanup *shouldn't* let this happen, but in case something failed, check pachd for sanity.
		t.Logf("New namespace acquired or helm options don't match, doing a fresh Helm install in %v", namespace)
		if err := helm.InstallE(t, helmOpts, chartPath, namespace); err != nil {
			require.NoErrorWithinTRetry(t,
				time.Minute,
				func() error {
					deleteRelease(t, context.Background(), namespace, kubeClient)
					return errors.EnsureStack(helm.InstallE(t, helmOpts, chartPath, namespace))
				})
		}
		waitForInstallFinished()
	} else { // same config, no need to change anything
		t.Logf("Previous helmOpts matched the previous cluster config, no changes made to cluster in %v", namespace)
	}
	pClient := pachClient(t, pachAddress, opts.AuthUser, namespace, opts.CertPool)
	t.Cleanup(func() {
		collectMinikubeCodeCoverage(t, pClient, opts.ValueOverrides)
	})
	return pClient
}

// LeaseNamespace attempts to lock a namespace for exclusive use by a test. It creates a k8s lease that is cleaned up at the end of the test.
// Calling LeaseNamespace will create the namespace if it doesn't exist. It returns true if Lease was acquired, false if not.
// To block until a namespace lock can be acquired, retry for the desired amount of time. LeaseNamespace will force-acquire
// the namespace lock after the the lease duration has passed.
func LeaseNamespace(t testing.TB, namespace string) bool {
	kube := testutil.GetKubeClient(t)
	lease, err := kube.CoordinationV1().
		Leases(namespace).
		Get(context.Background(), fmt.Sprintf("%s%s", leasePrefix, namespace), metav1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		PutNamespace(t, namespace)
		err = putLease(t, namespace)
		if k8serrors.IsAlreadyExists(err) {
			// if it already exists, but didn't before then we are racing another process.
			// Just try to take the next namespace since there are many available.
			return false
		}
		require.NoError(t, err)
		return true
	} else {
		require.NoError(t, err)
		timeSinceRenewal := time.Since(lease.Spec.RenewTime.Time)
		maxLeaseTime := time.Duration(*lease.Spec.LeaseDurationSeconds) * time.Second
		if timeSinceRenewal > maxLeaseTime {
			t.Logf("Lease expired. Force acquiring lock on namespace %v.", namespace)
			err := kube.CoordinationV1().
				Leases(namespace).
				Delete(context.Background(), lease.Name, metav1.DeleteOptions{})
			if k8serrors.IsNotFound(err) {
				return false // someone else deleted it first and so is likely in the process of leasing this namespace, move on
			}
			require.NoError(t, err)
			PutNamespace(t, namespace)
			err = putLease(t, namespace)
			if k8serrors.IsAlreadyExists(err) {
				// if it already exists, we deleted it, but another runner snuck in and
				// created it somehow. Let them have it and go to the next open namespace.
				return false
			}
			require.NoError(t, err)
			return true
		}
		return false // lease exists and is not expired, can't take it
	}
}

// Creates a Namespace if it doesn't exist. If it does exist, does nothing.
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
		if !k8serrors.IsAlreadyExists(err) { // if it already exists we still have the end result we want and idempotence
			require.NoError(t, err)
		}
	}
}

// Deploy pachyderm using a `helm upgrade ...`
// returns an API Client corresponding to the deployment
func UpgradeRelease(t testing.TB, ctx context.Context, namespace string, kubeClient *kube.Clientset, opts *DeployOpts) *client.APIClient {
	return putRelease(t, ctx, namespace, kubeClient, true, opts)
}

// Deploy pachyderm using a `helm install ...`
// returns an API Client corresponding to the deployment
func InstallRelease(t testing.TB, ctx context.Context, namespace string, kubeClient *kube.Clientset, opts *DeployOpts) *client.APIClient {
	return putRelease(t, ctx, namespace, kubeClient, false, opts)
}
