package helmlib

import (
	"context"
	"crypto/x509"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

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

	MinioEndpoint = "minio.default.svc.cluster.local:9000"
	MinioBucket   = "pachyderm-test"

	DefaultPachdNodePort = 30650
)

type DeploymentOpts struct {
	// Namespace is the kubernetes namespace into which resources should be
	// deployed. If no kubernetes namespace with the given name exists, one will
	// be created.
	Namespace string

	// PachVersion is the version of Pachyderm to be deployed. This can be a
	// specific version number (e.g. "2.8.1"), or "local", indicating the pachd
	// and worker images built locally and tagged ":local" should be used.
	PachVersion string

	// ObjectStorage indicates the kind of object storage pachd should be
	// configured to use. As this library is intended to start Pachyderm instances
	// for testing purposes, the valid options for this field are "minio"
	// (indicating minio should be deployed alongside Pachyderm and used as its
	// object storage provider) or "local", indicating Pachyderm's local storage
	// client and a local volume should be used. If unset, "minio" is the default.
	ObjectStorage string

	// DisableEnterprise will, if set, cause Pachyderm to be deployed in
	// community-edition mode. That is, normally helmlib deploys Pachyderm using
	// the host's Pachyderm Enterprise token (so that enterprise features are
	// enabled) and console and auth activated, but with this flag set, none of
	// those will happen.
	DisableEnterprise bool

	// DisableAuth will, if set, cause Pachyderm to be deployed with auth
	// disabled. That is, normally helmlib deploys Pachyderm using the host's
	// Pachyderm Enterprise token (so that enterprise features are enabled) and
	// auth and console all activated, but with this flag set, auth will not be
	// activated (enterprise and console both will).
	DisableAuth bool

	// DisableConsole will, if set, cause Pachyderm to be deployed with its
	// console disabled. That is, normally helmlib deploys Pachyderm using the
	// host's Pachyderm Enterprise token (so that enterprise features are enabled)
	// and auth and console all activated, but with this flag set, console will
	// not be activated (enterprise and auth both will).
	DisableConsole bool

	// EnableLoki will, if set, cause Loki to be deployed alongside Pachyderm, and
	// will configure Pachyderm to use Loki for log retention and retrieval.
	EnableLoki bool

	// ValueOverrides are manually-set helm values passed by the caller.
	ValueOverrides map[string]string

	// TODO(msteffen) Add back the fields below, if/when this is merged with
	// Minikubetestenv.
	// Determined will, if set, cause Determiend to be deployed alongside
	// Pachyderm.
	// Determined bool

	// PortOffset is a fixed constant added to the pachyderm-proxy service's
	// nodeport port. This is necessary when multiple Pachyderm clusers are
	// deployed into the same Kubernetes cluster (e.g. when running multiple tests
	// in parallel) because NodePorts are cluster-wide, so this allows each
	// Pachyderm cluster to have a unique NodePort.
	// NOTE: it might make more sense to declare port instead of offset
	// PortOffset  uint16

	// TODO(msteffen): there are a couple of places where we need to know the
	// external address of the pach cluster (the IP of the KiND node, basically).
	// The main one is for auth: we need to know where to redirect users (and what
	// host to accept) during the OAuth login flow. It's also needed to connect to
	// the cluster and health check it while waiting for the helm deploy to
	// finish.
	//
	// I'd like this library to be as general-purpose as possible, and I don't
	// know if this is the right way to do things (what if a LoadBalancer is
	// created and the PachAddress could/should be determined by inspecting the
	// LoadBalancer with the k8s client? Or it's dynamic in some other way.)
	// PachHost string

	// EnterpriseServer and EnterpriseMember can be used when 'Enterprise' is
	// false to test manual enterprise/auth activation.
	// TODO(msteffen) add back
	// EnterpriseMember bool
	// EnterpriseServer bool

	// TLS            bool
	// CertPool       *x509.CertPool
	// ValuesFiles    []string
}

// stack is sort of analogous to Go's builtin 'append' function for slices, but
// for to *helm.Options. The right argument ("top", for clarity) is "appended"
// to the left("bottom")--it overwrites any helm values in "bottom" that
// conflict, and adds any helm values that aren't set.
func Stack(first *helm.Options, rest ...*helm.Options) *helm.Options {
	out := &helm.Options{
		SetValues:    make(map[string]string),
		SetStrValues: make(map[string]string),
	}
	for _, src := range append([]*helm.Options{first}, rest...) {
		if src.KubectlOptions != nil {
			out.KubectlOptions = &k8s.KubectlOptions{Namespace: src.KubectlOptions.Namespace}
		}
		if src.Version != "" {
			out.Version = src.Version
		}
		for k, v := range src.SetValues {
			out.SetValues[k] = v
		}
		for k, v := range src.SetStrValues {
			out.SetStrValues[k] = v
		}
	}
	return out
}

func BaseOptions(namespace string) *helm.Options {
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

func WithPachd(image string) *helm.Options {
	return &helm.Options{
		SetValues: map[string]string{
			"pachd.image.tag":    image,
			"pachd.service.type": "ClusterIP",
			// For tests, traffic from outside the cluster is routed through the
			// proxy, but we bind the internal k8s service ports to the same numbers
			// for in-cluster traffic, like enterprise registration.
			"proxy.enabled": "true",
			// minikubetestenv uses LoadBalancer services if GOOS is 'darwin' or
			// 'windows', because minikubetestenv assumes Minikube will be the
			// kubernetes provider and Minikube provides 'minikube tunnel' on those
			// systems for connecting to LoadBalancer services. 'pachdev' uses KinD on
			// the other hand, which has no 'kind tunnel' command, so 'helmlib' justs
			// uses NodePort in all cases.
			"proxy.service.type":                  "NodePort",
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
			// Don't deploy Loki by default
			"pachd.lokiDeploy":  "false",
			"pachd.lokiLogging": "false",
			// Use minio as Pachyderm's object store by default
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

func WithoutProxy() *helm.Options {
	return &helm.Options{
		SetValues: map[string]string{
			"proxy.enabled": "false",
			// See comment above re. proxy.service.type for an explanation.
			"pachd.service.type": "NodePort",
		},
	}
}

func WithLocalStorage() *helm.Options {
	return &helm.Options{
		SetValues: map[string]string{
			"deployTarget":          "LOCAL",
			"pachd.storage.backend": "LOCAL",
			// Unset minio options
			"pachd.storage.minio.bucket":   "",
			"pachd.storage.minio.endpoint": "",
			"pachd.storage.minio.id":       "",
			"pachd.storage.minio.secret":   "",
		},
		SetStrValues: map[string]string{
			"pachd.storage.minio.signature": "",
			"pachd.storage.minio.secure":    "",
		},
	}
}

func WithLoki() *helm.Options {
	return &helm.Options{
		SetValues: map[string]string{
			"pachd.lokiDeploy":  "true",
			"pachd.lokiLogging": "true",
			// See note in 'withPachd' re. hard-coding ServiceType to NodePort.
			"loki-stack.loki.service.type": "NodePort",
			// TODO(msteffen) Is Loki's port being set explicitly just to avoid
			// NodePort collisions? I believe the complexity of this library could be
			// reduced if we instead hard-coded the Loki service type to NodePort
			// above, and changed our tests to access the Loki service via port
			// forwarding (this is also consistent with our usual cluster
			// mental-model, in which Loki is normally only accessed by pachd).
			// "loki-stack.loki.service.port":                                    fmt.Sprintf("%v", port+9),
			// "loki-stack.loki.service.nodePort":                                fmt.Sprintf("%v", port+9),
			// "loki-stack.loki.config.server.http_listen_port":                  fmt.Sprintf("%v", port+9),
			// "loki-stack.promtail.config.serverPort":                           fmt.Sprintf("%v", port+9),
			// "loki-stack.promtail.config.clients[0].url":                       fmt.Sprintf("http://%s-loki:%d/loki/api/v1/push", namespace, port+9),
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

// GetHelmValues generates a set of helm values for deploying Pachyderm into an
// existing kubernetes cluster. See
// src/internal/minikubetestenv/deploy.go:putRelease for the original code being
// ported here. Some of the main changes:
//   - helmlib factors out the helm-value specific parts of that from from the
//     parts that create kubernetes resources (such as the enterprise secret),
//     finding/using the right helm chart based on the Pachyderm version, etc.
func (opts *DeploymentOpts) GetHelmValues() *helm.Options {
	helmOpts := BaseOptions(opts.Namespace)

	// if opts.Enterprise {
	// 	// here, pachd address is used so that the enterprise server knows how to
	// 	// talk to pachd, and so it uses dex (which I guess is running in pachd)
	// 	// correctly -- I think this should use the internal pachyderm service
	// 	// TODO(msteffen) add back if supporting EnterpriseMember
	// 	// issuerPort := int(pachAddress.Port+opts.PortOffset) + 8
	// 	// if opts.EnterpriseMember { issuerPort = 31658 }
	// 	helmOpts = Stack(helmOpts, withEnterprise(pachAddress.Host, RootToken, issuerPort, int(pachAddress.Port+opts.PortOffset)+7))
	// }
	// TODO(msteffen): Determined deployment is not yet ported. See README.md
	// if opts.Determined { ... }

	// TODO(msteffen): EnterpriseServer and EnterpriseMember are not yet ported.
	// See README.md
	// if opts.EnterpriseServer { ...	} else {

	helmOpts = Stack(helmOpts, WithPachd(opts.PachVersion))
	if opts.ObjectStorage == "local" {
		helmOpts = Stack(helmOpts, WithLocalStorage())
	}
	// TODO(msteffen): See README.md about why withPort() has not been ported.
	// if opts.PortOffset != 0 { ... }
	// TODO(msteffen): See README.md about withEnterpriseServer()
	// if opts.EnterpriseMember { ... }
	if !opts.DisableConsole {
		helmOpts.SetValues["console.enabled"] = "true"
	}
	if opts.EnableLoki {
		helmOpts = Stack(helmOpts, WithLoki())
	}
	if opts.DisableAuth {
		helmOpts = Stack(helmOpts, &helm.Options{
			SetValues: map[string]string{
				"pachd.activateAuth": "false",
			},
		})
	}
	if !(opts.PachVersion == "" || strings.HasPrefix(opts.PachVersion, "2.3")) {
		helmOpts = Stack(helmOpts, WithoutProxy())
	}
	if opts.ValueOverrides != nil {
		helmOpts = Stack(helmOpts, &helm.Options{SetValues: opts.ValueOverrides})
	}
	// if opts.TLS {
	// 	pachAddress.Secured = true
	// }
	// helmOpts.ValuesFiles = opts.ValuesFiles
	return helmOpts
}

func pachClient(ctx context.Context, kubeClient *kubernetes.Clientset, namespace, authUser string, certpool *x509.CertPool) (*client.APIClient, error) {
	var c *client.APIClient

	// retry connecting if it doesn't immediately work
	if err := backoff.Retry(func() error {
		var err error
		opts := []client.Option{client.WithDialTimeout(10 * time.Second)}
		if certpool != nil {
			opts = append(opts, client.WithCertPool(certpool))
		}
		c, err = client.NewFromPachdAddress(pctx.TODO(), pachAddress, opts...)
		if err != nil {
			return errors.Wrapf(err, "failed to initialize pach client")
		}
		// Ensure that pachd is really ready to receive requests.
		if _, err := c.InspectCluster(); err != nil {
			return errors.Wrapf(grpcutil.ScrubGRPC(err), "failed to inspect cluster on port %v", pachAddress.Port)
		}
		return nil
	}, backoff.RetryEvery(3*time.Second).For(50*time.Second)); err != nil {
		return nil, errors.Wrapf(err, "failed to connect to pachd")
	}
	if authUser != "" {
		c = testutil.AuthenticateClient(t, c, authUser)
	}
	return c
}

// Reuse looks for an existing, previously-deployed Pachyderm cluster with the
// given DeploymentOpts and, if one is found, returns a pach client connected to
// that cluster. See 'Create()', 'Upgrade()' and 'Delete()'--like those
// operations, this uses the value in opts.PachVersion to decide what helm chart
// to use. Like those and like GetHelmValues, this contains code from
// `putRelease`.
//
// General note on structure here:
//   - there's a lot of logging in putRelease that happens while waiting e.g.
//     for pachd to come up. Rather than try to figure out how to log in this
//     library, I went with the pattern of breaking things up, so that
//     minikubetestenv would still have the for loop and the log line, but much of
//     the body of the for loop would be in here. E.g. we expect callers to call
//     Reuse() and then, if that errors, call Create(). This avoids teh need to
//     log "no cluster found, create cluster..." anywhere in this library, as that
//     will now naturally happen in the caller. Likewise, we added Status(), so
//     instead of having to log "waiting for cluster to come up..." every second,
//     teh caller can call "Status()" then log if it wants, then call `Status()`
//     again.
//
// TODO(msteffen) the enterprise secret *can* also be passed via helm value.
// Should this tool do that instead?
func Reuse(ctx context.Context, kubeClient *kubernetes.Clientset, opts *DeploymentOpts) (*client.APIClient, error) {
	previousOptsHash := getOptsConfigMapData(t, ctx, op.GetKubeClient(), namespace)
	pachdExists, err := checkForPachd(t, ctx, op.GetKubeClient(), namespace, opts.EnterpriseServer)
	require.NoError(t, err)
	if !pachdExists && previousOptsHash != nil {
		// In case somehow a config map got left without a corresponding installed
		// release. Cleanup *shouldn't* let this happen, but in case something
		// failed, check pachd for sanity.
		// delete config map
	}
	return pachClient(t, pachAddress, opts.AuthUser, namespace, opts.CertPool)
}

func Create(ctx context.Context, kubeClient *kubernetes.Clientset, opts *DeploymentOpts) (*client.APIClient, error) {
	if !opts.DisableEnterprise {
		if err := createSecretEnterpriseKeySecret(ctx, op.kubeClient, opts.Namespace); err != nil {
			return nil, err
		}
	}
	if err := helm.InstallE(t, helmOpts, chartPath, namespace); err != nil {
		require.NoErrorWithinTRetry(t,
			time.Minute,
			func() error {
				deleteRelease(t, context.Background(), namespace, op.GetKubeClient())
				return errors.EnsureStack(helm.InstallE(t, helmOpts, chartPath, namespace))
			})
	}
}

func Upgrade(ctx context.Context, kubeClient *kubernetes.Clientset, opts *DeploymentOpts) (*client.APIClient, error) {
	require.NoErrorWithinTRetry(t,
		time.Minute,
		func() error {
			return errors.EnsureStack(helm.UpgradeE(t, helmOpts, chartPath, namespace))
		})
}
func Delete(ctx context.Context, kubeClient *kubernetes.Clientset, opts *DeploymentOpts) (*client.APIClient, error) {
	options := &helm.Options{
		KubectlOptions: &k8s.KubectlOptions{Namespace: namespace},
	}
	mu.Lock()
	err := helm.DeleteE(t, options, namespace, true)
	// op.client = pachClient(t, pachAddress, opts.AuthUser, namespace)
	// Don't know how to do this part of what minikubetestenv does.
	// t.Cleanup(func() {
	// 	collectMinikubeCodeCoverage(t, pClient, opts.ValueOverrides)
	// })
	return op
}

type DeployStatus bool

const (
	NotReady DeployStatus = false
	Ready                 = true
)

func (op *DeployOperation) Status(ctx context.Context) error {
	// createOptsConfigMap(ctx, op.kubeClient, op.opts.Namespace, helmOpts, chartPath)

	// wait for pachd
	label := "suite=pachyderm,app=pachd"
	pachds, err := op.kubeClient.CoreV1().Pods(op.opts.Namespace).List(ctx, metav1.ListOptions{LabelSelector: label})
	if err != nil {
		return errors.Wrap(err, "error on pod list")
	}
	// var unacceptablePachds []string
	// var acceptablePachds []string
	var good int
	for _, p := range pachds.Items {
		if p.Status.Phase == v1.PodRunning && strings.HasSuffix(p.Spec.Containers[0].Image, ":"+opts.PachVersion) && p.Status.ContainerStatuses[0].Ready {
			good++
			// acceptablePachds = append(acceptablePachds, fmt.Sprintf("%v: image=%v status=%s", p.Name, p.Spec.Containers[0].Image, formatPodStatus(p.Status)))
		} // else {
		// unacceptablePachds = append(unacceptablePachds, fmt.Sprintf("%v: image=%v status=%s", p.Name, p.Spec.Containers[0].Image, formatPodStatus(p.Status)))
		// }
	}
	// if len(acceptablePachds) > 0 && (len(unacceptablePachds) == 0) {
	if good == len(pachds.Items) {
		return nil
	}
	return errors.Errorf("deployment in progress; pachd pods ready: %d / %d",
		good, len(pachds.Items))
	// return errors.Errorf("deployment in progress; pachds ready: %d / %d \nun-ready pachds: %v",
	// 	len(acceptablePachds), len(acceptablePachds)+len(unacceptablePachds),
	// 	strings.Join(unacceptablePachds, "; "),
	// )

	// if !opts.DisableLoki {
	// 	waitForLoki(t, pachAddress.Host, int(pachAddress.Port)+9)
	// }
	// waitForPgbouncer(ctx)
	// waitForPostgres(ctx)
	// if opts.Determined {
	// 	waitForDetermined(ctx)
	// }
	return nil
}

func createSecretEnterpriseKeySecret(ctx context.Context, kubeClient *kubernetes.Clientset, ns string) error {
	_, err := kubeClient.CoreV1().Secrets(ns).Create(ctx, &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: licenseKeySecretName},
		StringData: map[string]string{
			"enterprise-license-key": testutil.GetTestEnterpriseCode(t),
		},
	}, metav1.CreateOptions{})
	if err != nil && strings.Contains(err.Error(), "already exists") {
		return nil
	}
	return err
}
