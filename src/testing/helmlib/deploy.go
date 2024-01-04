package helmlib

import (
	"bytes"
	"context"
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
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

const (
	helmChartPublishedPath = "pachyderm/pachyderm"

	MinioEndpoint = "minio.default.svc.cluster.local:9000"
	MinioBucket   = "pachyderm-test"
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
	// client and a local volume should be used.
	ObjectStorage string

	// DisableEnterprise will, if set, cause Pachyderm to be deployed using the
	// host's Pachyderm Enterprise token (so that enterprise features are enabled)
	// and auth activated
	DisableEnterprise bool

	// DisableAuth will, if set, cause Pachyderm to be deployed with auth disabled
	DisableAuth bool

	// DisableConsole will, if set, cause Pachyderm Console to be deployed alongside
	// Pachyderm.
	DisableConsole bool

	// DisableLoki will, if set, cause Pachyderm to be deployed without Loki
	DisableLoki bool

	// ValueOverrides are manually-set helm values passed by the caller.
	ValueOverrides map[string]string

	// TODO(msteffen) Add back the fields below, if/when this is merged with
	// Minikubetestenv.
	// Determined will, if set, cause Determiend to be deployed alongside
	// Pachyderm.
	// Determined bool

	// AuthUser string

	// TODO(msteffen) add back
	// CleanupAfter       bool
	// UseLeftoverCluster bool

	// PortOffset is a fixed constant added to the pachyderm-proxy service's
	// nodeport port. This is necessary when multiple Pachyderm clusers are
	// deployed into the same Kubernetes cluster (e.g. when running multiple tests
	// in parallel) because NodePorts are cluster-wide, so this allows each
	// Pachyderm cluster to have a unique NodePort.
	// NOTE: it might make more sense to declare port instead of offset
	// PortOffset  uint16

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
func stack(bottom, top *helm.Options) *helm.Options {
	out := &helm.Options{
		SetValues:    make(map[string]string),
		SetStrValues: make(map[string]string),
	}
	for _, src := range []*helm.Options{bottom, top} {
		if src.KubectlOptions != nil {
			out.KubectlOptions = &k8s.KubectlOptions{Namespace: src.KubectlOptions.Namespace}
		}
		for k, v := range src.SetValues {
			out.SetValues[k] = v
		}
		for k, v := range src.SetStrValues {
			out.SetStrValues[k] = v
		}
		if src.Version != "" {
			out.Version = src.Version
		}
	}
	return out
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
		},
	}
}

func withoutProxy() *helm.Options {
	return &helm.Options{
		SetValues: map[string]string{
			"proxy.enabled": "false",
			// See comment above re. proxy.service.type for an explanation.
			"pachd.service.type": "NodePort",
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

func withLocalStorage() *helm.Options {
	return &helm.Options{
		SetValues: map[string]string{
			"deployTarget": "LOCAL",
		},
	}
}

func withoutLokiOptions() *helm.Options {
	return &helm.Options{
		SetValues: map[string]string{
			"pachd.lokiDeploy":  "false",
			"pachd.lokiLogging": "false",
		},
	}
}

func withLokiOptions() *helm.Options {
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
	// TODO(msteffen) the use of pachAddress does not make sense to me.
	// Pachyderm hasn't been deployed at the start of this function, so what is
	// supposed to be in that variable?
	// Based on src/internal/minikubetestenv/deploy.go:GetPachAddress, it is:
	// - by default, grpcutil/addr.go:DefaultPachdAddress, i.e.
	//   grpc://0.0.0.0:30650
	// - If -testenv.host is set, overwrite the above with that flag
	// - if GOOS isn't 'darwin' or 'windows' (i.e. is "linux"):
	//   grpc://$(minikube ip):30650
	// - if -testenv.baseport is set, port is set to that value (overriding
	// pachAddress := op.PachClient().GetAddress() // what is this for?
	helmOpts := withBase(opts.Namespace)

	// if opts.Enterprise {
	// 	// here, pachd address is used so that the enterprise server knows how to
	// 	// talk to pachd, and so it uses dex (which I guess is running in pachd)
	// 	// correctly
	// 	// TODO(msteffen) add back if supporting EnterpriseMember
	// 	// issuerPort := int(pachAddress.Port+opts.PortOffset) + 8
	// 	// if opts.EnterpriseMember { issuerPort = 31658 }
	// 	helmOpts = stack(helmOpts, withEnterprise(pachAddress.Host, RootToken, issuerPort, int(pachAddress.Port+opts.PortOffset)+7))
	// }
	// TODO(msteffen): Determined deployment is not yet ported. See README.md
	// if opts.Determined { ... }

	// TODO(msteffen): EnterpriseServer and EnterpriseMember are not yet ported.
	// See README.md
	// if opts.EnterpriseServer { ...	} else {

	helmOpts = stack(helmOpts, withPachd(opts.PachVersion))
	if opts.ObjectStorage == "minio" {
		helmOpts = stack(helmOpts, withMinio())
	} else {
		helmOpts = stack(helmOpts, withLocalStorage())
	}
	// TODO(msteffen): See README.md about why withPort() has not been ported.
	// if opts.PortOffset != 0 { ... }
	// TODO(msteffen): See README.md about withEnterpriseServer()
	// if opts.EnterpriseMember { ... }
	if !opts.DisableConsole {
		helmOpts.SetValues["console.enabled"] = "true"
	}
	if opts.DisableLoki {
		helmOpts = stack(helmOpts, withoutLokiOptions())
	} else {
		helmOpts = stack(helmOpts, withLokiOptions())
	}
	if opts.DisableAuth {
		helmOpts = stack(helmOpts, &helm.Options{
			SetValues: map[string]string{
				"pachd.activateAuth": "false",
			},
		})
	}
	if !(opts.PachVersion == "" || strings.HasPrefix(opts.PachVersion, "2.3")) {
		helmOpts = stack(helmOpts, withoutProxy())
	}
	if opts.ValueOverrides != nil {
		helmOpts = stack(helmOpts, &helm.Options{SetValues: opts.ValueOverrides})
	}
	// if opts.TLS {
	// 	pachAddress.Secured = true
	// }
	// helmOpts.ValuesFiles = opts.ValuesFiles
	return helmOpts
}

// HelmOp indicates which specific operation Helm will be asked to perform when
// a new Pachyderm cluster is being brounght up. See below for the possible
// values.
type HelmOp int8

const (
	// HelmDetect which causes this library to detect whether Pachyderm is
	// installed and, if so, which version. Then, it will take the fastest path to
	// bringing the pach cluster into the target state (which may be to do
	// nothing).
	HelmDetect HelmOp = iota
	// HelmUpgrade directs a cluster deployment operation to use 'helm upgrade'
	HelmUpgrade
)

type DeployOperation struct {
	// opts descript the pachyderm deployment that is to be realized (by creation
	// or update).
	opts *DeploymentOpts

	// kubeClient is a k8s client connected to the cluster where Pachyderm should
	// be deployed.
	kubeClient *kubernetes.Clientset

	// helmOp is the operation (Upgrade, which always runs 'helm upgrade' and
	// therefore requires a Pachyderm deployment to already exist, or Detect,
	// which runs either 'helm create' or 'helm upgrade' as is appropriate.
	helmOp HelmOp
}

func NewDeployOperation(opts *DeploymentOpts, kubeClient *kubernetes.Clientset, helmOp HelmOp) *DeployOperation {
	return &DeployOperation{
		opts:       opts,
		kubeClient: kubeClient,
		helmOp:     helmOp,
	}
}

func (op *DeployOperation) waitForInstallFinished(ctx context.Context, namespace string) error {
	// createOptsConfigMap(ctx, op.kubeClient, namespace, helmOpts, chartPath)

	// wait for pachd
	label := "suite=pachyderm,app=pachd"
	if err := backoff.Retry(func() error {
		pachds, err := op.kubeClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: label})
		if err != nil {
			return errors.Wrap(err, "error on pod list")
		}
		// var unacceptablePachds []string
		// var acceptablePachds []string
		var good int
		for _, p := range pachds.Items {
			if p.Status.Phase == v1.PodRunning && strings.HasSuffix(p.Spec.Containers[0].Image, ":"+op.opts.PachVersion) && p.Status.ContainerStatuses[0].Ready {
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
	}, &backoff.ConstantBackOff{
		Interval:       2 * time.Second,
		MaxElapsedTime: 120 * time.Second,
	}); err != nil {
		return err
	}

	// if !op.opts.DisableLoki {
	// 	waitForLoki(t, pachAddress.Host, int(pachAddress.Port)+9)
	// }
	// waitForPgbouncer(ctx, op.kubeClient, namespace)
	// waitForPostgres(ctx, op.kubeClient, namespace)
	// if opts.Determined {
	// 	waitForDetermined(ctx, op.kubeClient, namespace)
	// }
	return nil
}

// HelmDeploy is actually executes a Helm create/update operation. See the
// comments in GetHelmValues, but this likewise contains code from `putRelease`.
// However, this focuses on the non-helm values parts of that process (which is
// the minority; most of putRelease is focused on constructing the right helm
// values). This includes:
//
//   - Getting the right helm chart (from helm itself, or locally if deploying
//     from a clone of the Pachyderm repo)
//
//   - Creating kubernetes resources in addition to the ones specified by the
//     helm chart (such as the Enterprise secret).
//
//     TODO(msteffen) the enterprise secret *can* also be passed via helm value.
//     Should this tool do that instead?
func (op *DeployOperation) HelmDeploy(ctx context.Context) (*client.APIClient, error) {
	if op.opts.Enterprise {
		if err := createSecretEnterpriseKeySecret(ctx, op.kubeClient, op.opts.Namespace); err != nil {
			return nil, err
		}
	}
	previousOptsHash := getOptsConfigMapData(t, ctx, op.GetKubeClient(), namespace)
	pachdExists, err := checkForPachd(t, ctx, op.GetKubeClient(), namespace, opts.EnterpriseServer)
	require.NoError(t, err)
	if helmOp == HelmUpgrade {
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
					deleteRelease(t, context.Background(), namespace, op.GetKubeClient())
					return errors.EnsureStack(helm.InstallE(t, helmOpts, chartPath, namespace))
				})
		}
		waitForInstallFinished()
	} else { // same config, no need to change anything
		t.Logf("Previous helmOpts matched the previous cluster config, no changes made to cluster in %v", namespace)
	}
	// pClient := pachClient(t, pachAddress, opts.AuthUser, namespace)
	// t.Cleanup(func() {
	// 	collectMinikubeCodeCoverage(t, pClient, opts.ValueOverrides)
	// })
	// return pClient
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
