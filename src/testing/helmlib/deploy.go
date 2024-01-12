package helmlib

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	helm "helm.sh/helm/v3/pkg/action"
	helmvals "helm.sh/helm/v3/pkg/cli/values"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

const (
	helmChartPublishedPath = "pachyderm/pachyderm"

	MinioEndpoint = "minio.default.svc.cluster.local:9000"
	MinioBucket   = "pachyderm-test"

	// licenseKeySecretName is the name of the k8s secret that's created to hold
	// the new cluster's enterprise license key
	licenseKeySecretName = "enterprise-license-key-secret"

	configMapName      = "test-pach-helm-opts"
	configMapHashField = "helm-opts-hash"

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

// Stack is sort of analogous to Go's builtin 'append' function for slices, but
// for to *helm.Options. The right argument ("uppers", for clarity) is "appended"
// to the left("bottom")--it overwrites any helm values in "bottom" that
// conflict, and adds any helm values that aren't set.
//
// N.B. this is similar to mergeMaps
// (https://github.com/helm/helm/blob/27921d062f520357303309ca959a129bc49a4abc/pkg/cli/values/options.go#L108)
// in the helm library, but that function is private.
func Stack(bottom map[string]interface{}, uppers ...map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for _, src := range append([]*helm.Options{bottom}, uppers...) {
		for k, srcV := range src {
			outV, ok := out[k]
			if !ok {
				out[k] = srcV
				continue
			}
			srcVMap, srcVIsMap := srcV.(map[string]interface{})
			outVMap, outVIsMap := outV.(map[string]interface{})
			if !srcVIsMap || !outVIsMap {
				out[k] = srcV
				continue
			}
			out[k] = Stack(outV, srcV)
		}
	}
	return out
}

func Stack2(bottom *helmvals.Options, uppers ...*helmvals.Options) *helmvals.Options {
	out := &helmvals.Options{}
	out := make(map[string]interface{})
	for _, src := range append([]*helm.Options{bottom}, uppers...) {
		for k, srcV := range src {
			outV, ok := out[k]
			if !ok {
				out[k] = srcV
				continue
			}
			srcVMap, srcVIsMap := srcV.(map[string]interface{})
			outVMap, outVIsMap := outV.(map[string]interface{})
			if !srcVIsMap || !outVIsMap {
				out[k] = srcV
				continue
			}
			out[k] = Stack(outV, srcV)
		}
	}
	return out
}
func BaseOptions(namespace string) map[string]interface{} {
	opts, err := &helmvals.Options{
		Values: map[string]string{
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
	}.MergeValues(nil)
	if err != nil {
		panic("error parsing BaseOptions values: " + err.Error())
	}
	return opts
}

func WithPachd(image string) map[string]interface{} {
	opts, err := &helmvals.Options{
		Values: map[string]string{
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
		StringValues: map[string]string{
			"pachd.storage.minio.signature": "",
			"pachd.storage.minio.secure":    "false",
		},
	}.MergeValues(nil)
	if err != nil {
		panic("error parsing WithPachd values: " + err.Error())
	}
	return opts
}

func WithoutProxy() map[string]interface{} {
	opts, err := &helmvals.Options{
		Values: map[string]string{
			"proxy.enabled": "false",
			// See comment above re. proxy.service.type for an explanation.
			"pachd.service.type": "NodePort",
		},
	}.MergeValues(nil)
	if err != nil {
		panic("error parsing WithoutProxy values: " + err.Error())
	}
}

func WithLocalStorage() map[string]interface{} {
	opts, err := &helmvals.Options{
		Values: map[string]string{
			"deployTarget":          "LOCAL",
			"pachd.storage.backend": "LOCAL",
			// Unset minio options
			"pachd.storage.minio.bucket":   "",
			"pachd.storage.minio.endpoint": "",
			"pachd.storage.minio.id":       "",
			"pachd.storage.minio.secret":   "",
		},
		StringValues: map[string]string{
			"pachd.storage.minio.signature": "",
			"pachd.storage.minio.secure":    "",
		},
	}.MergeValues(nil)
	if err != nil {
		panic("error parsing WithLocalStorage values: " + err.Error())
	}
}

func WithLoki() map[string]interface{} {
	opts, err := &helmvals.Options{
		Values: map[string]string{
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
	}.MergeValues(nil).MergeValues(nil)
	if err != nil {
		panic("error parsing WithLoki values: " + err.Error())
	}
}

// GetHelmValues generates a set of helm values for deploying Pachyderm into an
// existing kubernetes cluster. See
// src/internal/minikubetestenv/deploy.go:putRelease for the original code being
// ported here. Some of the main changes:
//   - helmlib factors out the helm-value specific parts of that from from the
//     parts that create kubernetes resources (such as the enterprise secret),
//     finding/using the right helm chart based on the Pachyderm version, etc.
func (opts *DeploymentOpts) GetHelmValues() map[string]interface{} {
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
		helmOpts = Stack(helmOpts, &helmvals.Options{
			Values: map[string]string{
				"pachd.activateAuth": "false",
			},
		})
	}
	if !(opts.PachVersion == "" || strings.HasPrefix(opts.PachVersion, "2.3")) {
		helmOpts = Stack(helmOpts, WithoutProxy())
	}
	if opts.ValueOverrides != nil {
		helmOpts = Stack(helmOpts, &helmvals.Options{Values: opts.ValueOverrides})
	}
	// if opts.TLS {
	// 	pachAddress.Secured = true
	// }
	// helmOpts.ValuesFiles = opts.ValuesFiles
	return helmOpts
}

func (opts *DeploymentOpts) GetHelmChartPath() string {
	return "etc/helm/pachyderm" // TODO(msteffen): fix
}

// getConfigMapOptsHash is copied from minikubetestenv/deploy.go. It reads teh
// ConfigMap that this library creates alongside each pachyderm deployment, for
// the purpose of determined whether that deployment can be reused. This version
// is a helper called by Reuse below
func (opts *DeploymentOpts) getConfigMapOptsHash(ctx context.Context, kubeClient *kubernetes.Clientset) []byte {
	configMap, err := kubeClient.CoreV1().ConfigMaps(opts.Namespace).Get(ctx, configMapName, metav1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		return []byte{}
	}
	return configMap.BinaryData[configMapHashField]
}

// deletes the existing configmap and writes the new hel opts to this one
func (opts *DeploymentOpts) createConfigMap(ctx context.Context, kubeClient *kubernetes.Clientset) error {
	// delete prior config map, if any
	err := kubeClient.CoreV1().ConfigMaps(opts.Namespace).Delete(ctx, configMapName, *metav1.NewDeleteOptions(0))
	if err != nil && !k8serrors.IsNotFound(err) {
		return errors.Wrap(err, "error deleting ConfigMap")
	}
	hash, err := opts.hash()
	if err != nil {
		return errors.Wrap(err, "error hashing helm chart values for ConfigMap")
	}
	_, err = kubeClient.CoreV1().ConfigMaps(opts.Namespace).Create(ctx, &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: opts.Namespace,
			Labels:    map[string]string{"suite": "pachyderm"},
		},
		BinaryData: map[string][]byte{
			configMapHashField: hash,
		},
	}, metav1.CreateOptions{})
	return err
}

func (opts *DeploymentOpts) hash() ([]byte, error) { // return string for consistency with configMap data
	// we don't need to deserialize, we just use this to check differences, so just save a hash
	// Note: lists in the helm values might be ordered differently, causing
	// functionally identical helm values to hash differently, but there are not
	// many list fields that we actively use that would be different in practice
	//
	optsHash := sha256.New()
	err := json.NewEncoder(optsHash).Encode(opts.GetHelmValues())
	if err != nil {
		// annoying that we have to handle this, but I don't see how to avoid it
		return nil, err
	}
	// Per https://pkg.go.dev/hash#Hash, hash.Write() never returns an error
	optsHash.Write([]byte(opts.GetHelmChartPath()))
	return optsHash.Sum(nil), nil
}

// Exists returns true if an existing, previously-deployed Pachyderm cluster is
// up in the given k8s cluster and was created with the same DeploymentOpts.
//
// General note on structure here:
//   - there's a lot of logging in putRelease that happens while waiting e.g.
//     for pachd to come up. Rather than try to figure out how to log in this
//     library, I went with the pattern of breaking things up, so that
//     minikubetestenv would still have the for loop and the log line, but much of
//     the body of the for loop would be in here. E.g. we expect callers to call
//     Reuse() and then, if that errors, call Create(). This avoids the need to
//     log "no cluster found, create cluster..." anywhere in this library, as that
//     will now naturally happen in the caller. Likewise, we added Status(), so
//     instead of having to log "waiting for cluster to come up..." every second,
//     teh caller can call "Status()" then log if it wants, then call `Status()`
//     again.
//
// TODO(msteffen) the enterprise secret *can* also be passed via helm value.
// Should this tool do that instead?
func (opts *DeploymentOpts) Exists(ctx context.Context, kubeClient *kubernetes.Clientset) (bool, error) {
	previousOptsHash := opts.getConfigMapOptsHash(ctx, kubeClient)
	status, err := opts.Status()
	if err != nil {
		return false, errors.Wrap(err, "error getting pachd status while attempting to reuse pachd cluster")
	}
	currentOptsHash, err := opts.hash()
	if err != nil {
		return false, errors.Wrap(err, "error getting helm values hash while attempting to reuse pachd cluster")
	}
	return status.PachdPodsCreated > 0 && bytes.Equal(previousOptsHash, currentOptsHash), nil
}

func (opts *DeploymentOpts) createSecretEnterpriseKeySecret(ctx context.Context, kubeClient *kubernetes.Clientset) error {
	// TODO(msteffen) this duplicates testutil.GetTestEnterpriseCode(). We can't
	// call that function from here, though, as 1. it's under src/internal, and 2.
	// it requires a *testing.T.
	enterpriseCode, exists := os.LookupEnv("ENT_ACT_CODE")
	if !exists {
		return errors.New("enterprise activation code not found; env var \"ENT_ACT_CODE\" is not set")
	}
	_, err := kubeClient.CoreV1().Secrets(opts.Namespace).Create(ctx, &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: licenseKeySecretName},
		StringData: map[string]string{
			"enterprise-license-key": enterpriseCode,
		},
	}, metav1.CreateOptions{})
	if err != nil && strings.Contains(err.Error(), "already exists") {
		return nil
	}
	return err
}

func helmLog(format string, v ...interface{}) {
	fmt.Printf("[helm/debug] "+format+"\n", v...)
}

func (opts *DeploymentOpts) Create(ctx context.Context, kubeClient *kubernetes.Clientset) error {
	if !opts.DisableEnterprise {
		if err := opts.createSecretEnterpriseKeySecret(ctx, kubeClient); err != nil {
			return err
		}
	}
	cfg := &helm.Configuration{}
	cfg.Init(nil, opts.Namespace, "" /* =	"secret" */, helmLog)
	install := helm.NewInstall()
	// TODO(msteffen) fix
	// if err := helm.InstallE(helmOpts, opts.GetHelmChartPath(), opts.Namespace); err != nil {
	// 	deleteRelease(t, context.Background(), opts.Namespace, op.GetKubeClient())
	// 	return errors.EnsureStack(helm.InstallE(t, helmOpts, opts.GetHelmChartPath(), opts.Namespace))
	// }
}

func (opts *DeploymentOpts) Upgrade(ctx context.Context, kubeClient *kubernetes.Clientset) error {
	// TODO(msteffen) fix
	// return errors.EnsureStack(helm.UpgradeE(t, helmOpts, chartPath, namespace))
}

func (opts *DeploymentOpts) Delete(ctx context.Context, kubeClient *kubernetes.Clientset) (retErr error) {
	// TODO(msteffen) fix
	// retErr = helm.DeleteE(t, options, namespace, true)
	retErr = kubeClient.CoreV1().ConfigMaps(namespace).DeleteCollection(ctx, *metav1.NewDeleteOptions(0), metav1.ListOptions{LabelSelector: "suite=pachyderm"})
	retErr = kubeClient.CoreV1().PersistentVolumeClaims(namespace).DeleteCollection(ctx, *metav1.NewDeleteOptions(0), metav1.ListOptions{LabelSelector: "suite=pachyderm"})
	return
}

type StatusInfo struct {
	PachdPodsCreated int
	PachdPodsReady   int
	PachPvcs         int
}

func (opts *DeploymentOpts) Status() (StatusInfo, error) {
	result := &StatusInfo{}
	label := "app=pachd"
	if opts.EnterpriseServer {
		label = "app=pach-enterprise"
	}
	pachds, err := kubeClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: label})
	if !k8serrors.IsNotFound(err) && err != nil {
		return false, errors.Wrap(err, "error listing pachd pods")
	} else if err == nil {
		for _, p := range pachds.Items {
			if !strings.HasSuffix(p.Spec.Containers[0].Image, ":"+version) {
				continue
			}
			result.PachdPodsCreated++
			if p.Status.Phase == v1.PodRunning && p.Status.ContainerStatuses[0].Ready {
				result.PachdPodsReady++
			}
		}
	}
	pachPvcs, err := kubeClient.CoreV1().PersistentVolumeClaims(namespace).List(ctx, metav1.ListOptions{LabelSelector: "suite=pachyderm"})
	if !k8serrors.IsNotFound(err) && err != nil {
		return errors.Wrap(err, "error listing Pachyderm PVCs")
	} else if err == nil {
		result.PachPvcs = len(pachPvcs.Items)
	}
	return result
}
