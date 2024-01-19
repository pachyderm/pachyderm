// Package kindenv manages Kind (github.com/kubernetes-sigs/kind) environments.
package kindenv

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	"sigs.k8s.io/kind/pkg/cluster"
)

const (
	// version is stored in the cluster and will inform developers that they should recreate
	// their kind cluster when the stored version doesn't match this.
	version = 1

	// nodeImage is the kind node image to use.  If you change this, also change the version
	// above.  Always include the sha256 checksum (due to quirks in Kind's release process,
	// according to their docs).
	nodeImage = "kindest/node:v1.29.0@sha256:eaa1450915475849a73a9227b8f201df25e55e268e5d619312131292e324d570"

	clusterRegistryPushKey = "dev.pachyderm.io/registry-push-path"
	clusterRegistryPullKey = "dev.pachyderm.io/registry-pull-path"
	clusterVersionKey      = "dev.pachyderm.io/kindenv-version"
	clusterHostnameKey     = "dev.pachyderm.io/hostname"
	httpPortsKey           = "dev.pachyderm.io/standard-http-ports"
	exposedPortsKey        = "dev.pachyderm.io/exposed-ports"
	fieldManager           = "pachdev"
)

const (
	// It is unlikely that localhost:5001 will work for all setups, like pushing from Mac ->
	// registry on Linux VM.  So we'll pretend this is configurable and then come up with a plan
	// for making it configurable.  The port has to be 5001 for compatability with "docker
	// push"; it's the magic number that tells it "hey don't check the TLS cert".
	//
	// It is unclear if this actually matters, though.  If pushing to
	// $XDG_STATE_DIR/pach-registry/registry with skopeo and loading container images as
	// "pach-registry:5000/whatever:tag", then you don't even need to know this hostname.
	registryHostname = "localhost"
)

// CreateOpts specifies a Kind environment.
//
// Note: kind treats port numbers as int32, so we do too.  They are actually uint16s.
type CreateOpts struct {
	// TestNamespaceCount controls how many K8s tests can run concurrently.
	TestNamespaceCount int32
	// ExternalHostname is the hostname for the cluster.
	ExternalHostname string
	// ImagePushPath is the Skopeo path of the local container registry from the host machine,
	// usually something like `oci:/path/to/pach-registry`.  If empty, a new registry will be
	// started and a correct push path automatically configured.
	ImagePushPath string
	// BindHTTPPorts, if true, binds <externalHost>:80 and <externalHost>:443 to 30080 and 30443
	// in the cluster; used for the install of Pachyderm in the "default" namespace.
	BindHTTPPorts bool
	// StartingPort is the port number that begins the exposed ports for this cluster.  Each
	// TestNamespace gets 10.  If -1, then no ports will be created.  If unset (0), then ports
	// start at 30500.
	StartingPort int32

	// options just for internal tests

	// Where to put the kubeconfig; if empty, then do the default (which is to edit
	// ~/.kube/config).
	kubeconfigPath string
	// If set, use a registry name other than "pach-registry".  Also don't bind it to
	// localhost:5001.
	registryName string
}

func named(x string) string {
	if x == "pach" || x == "" {
		return "pach"
	}
	return "pach-" + x
}

// Cluster represents a Kind cluster.  It may not exist yet.
type Cluster struct {
	name       string
	provider   *cluster.Provider
	kubeconfig Kubeconfig
}

func getKindClusterFromContext() (string, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	raw, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, nil).RawConfig()
	if err != nil {
		return "", errors.Wrap(err, "get raw config")
	}
	ctx := raw.CurrentContext
	switch {
	case ctx == "kind-pach":
		return "pach", nil
	case strings.HasPrefix(ctx, "kind-pach-"):
		return ctx[len("kind-pach-"):], nil
	}
	return "", errors.Wrapf(err, "current k8s context %q is not a pachdev environment")
}

// New creates a new Cluster object, suitable for manipulating the named cluster.  If name is empty,
// the cluster in the current kubernetes context will be used, if it's a pachdev cluster.
func New(ctx context.Context, name string) (*Cluster, error) {
	po, err := cluster.DetectNodeProvider()
	if err != nil {
		return nil, errors.Wrap(err, "detect kind node provider")
	}
	if po == nil {
		return nil, errors.New("kind could not detect docker or podman; install docker")
	}
	p := cluster.NewProvider(po, cluster.ProviderWithLogger(log.NewKindLogger(pctx.Child(ctx, "kind"))))
	cluster := &Cluster{provider: p}
	if name == "" {
		var err error
		name, err = getKindClusterFromContext()
		if err != nil {
			return nil, errors.Wrap(err, "get cluster name from k8s context")
		}
	}
	cluster.name = named(name)
	return cluster, nil
}

// Close cleans up temporary data associated with the Cluster object.
func (c *Cluster) Close() error {
	return errors.Wrap(c.kubeconfig.Close(), "close cached kubeconfig")
}

// ClusterConfig is the configuration of the attached cluster.
type ClusterConfig struct {
	// Version of kindenv that deployed the cluster.
	Version int
	// A skopeo url that pushes images to this cluster.
	ImagePushPath string
	// A docker prefix that pulls images pushed to this cluster; use in
	// pod.spec.containers.image to refer to images pushed to this cluster.
	ImagePullPath string
	// A hostname that serves the exposed ports of the cluster.
	Hostname string
}

func (c *Cluster) GetConfig(ctx context.Context) (*ClusterConfig, error) {
	kc, err := c.GetKubeconfig(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get kubeconfig")
	}
	kube, err := kc.Client()
	if err != nil {
		return nil, errors.Wrap(err, "get kube client")
	}
	ns, err := kube.CoreV1().Namespaces().Get(ctx, "default", v1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "get default namespace")
	}
	if ns.Annotations == nil {
		return nil, errors.New("no annotations on default namespace")
	}

	result := new(ClusterConfig)
	if version, ok := ns.Annotations[clusterVersionKey]; ok {
		v, err := strconv.Atoi(version)
		if err != nil {
			log.Error(ctx, "cluster version is unparseable", zap.Error(err))
		}
		result.Version = v
	}
	if result.Version == 0 {
		return nil, errors.New("cluster is not a pachdev environment")
	}
	if result.Version != version {
		log.Error(ctx, "your pachdev cluster is outdated; please delete and re-create it soon", zap.Int("your_version", result.Version), zap.Int("latest_version", version))
	}
	var errs error
	if path, ok := ns.Annotations[clusterRegistryPullKey]; ok {
		result.ImagePullPath = path
	}
	if result.ImagePullPath == "" {
		errors.JoinInto(&errs, errors.New("there is no way to pull images from your cluster"))
	}
	if path, ok := ns.Annotations[clusterRegistryPushKey]; ok {
		result.ImagePushPath = path
	}
	if result.ImagePushPath == "" {
		errors.JoinInto(&errs, errors.New("there is no way to push images to your cluster"))
	}
	if host, ok := ns.Annotations[clusterHostnameKey]; ok {
		result.Hostname = host
	}
	if result.Hostname == "" {
		errors.JoinInto(&errs, errors.New("there is no way to access your cluster from the host network"))
	}
	if errs != nil {
		return nil, errs
	}
	return result, nil
}

// Create creates a new cluster.
func (c *Cluster) Create(ctx context.Context, opts *CreateOpts) (retErr error) {
	ctx, done := log.SpanContext(ctx, "create")
	defer done(log.Errorp(&retErr))
	if opts == nil {
		return errors.New("nil CreateOpts")
	}

	if opts.ExternalHostname == "" {
		opts.ExternalHostname = "127.0.0.1"
	}

	if opts.StartingPort == 0 {
		opts.StartingPort = 30500
	}

	containerdPullOverride := true
	pullPath := registryHostname + ":5001"
	exposeRegistry := true
	registryName := "pach-registry"
	var ensuredRegistry bool
	if opts.ImagePushPath == "" {
		if n := opts.registryName; n != "" {
			registryName = n
			pullPath = n + ":5000"
			containerdPullOverride = false
			exposeRegistry = false
		}
		path, err := ensureRegistry(ctx, registryName, exposeRegistry)
		if err != nil {
			return errors.Wrap(err, "setup registry container")
		}
		ensuredRegistry = true
		opts.ImagePushPath = "oci:" + path
	}

	// k8s annotations to be applied to the default namespace; this is how we transfer
	// configuration between tests/dev tools/etc.
	annotations := map[string]map[string]string{
		"default": {
			clusterVersionKey:  strconv.Itoa(version),
			clusterHostnameKey: opts.ExternalHostname,

			// Kind recommends this KEP, but it never got approved.  So we build our own
			// that doesn't involve parsing YAML.
			// https://github.com/kubernetes/enhancements/tree/master/keps/sig-cluster-lifecycle/generic/1755-communicating-a-local-registry
			clusterRegistryPullKey: pullPath,
			clusterRegistryPushKey: opts.ImagePushPath,
		},
	}

	// Configure the Kind cluster.
	var ports []v1alpha4.PortMapping
	if opts.BindHTTPPorts {
		annotations["default"][httpPortsKey] = "true"
		ports = append(ports,
			v1alpha4.PortMapping{
				ContainerPort: 30080,
				HostPort:      80,
				Protocol:      v1alpha4.PortMappingProtocolTCP,
			},
			v1alpha4.PortMapping{
				ContainerPort: 30443,
				HostPort:      443,
				Protocol:      v1alpha4.PortMappingProtocolTCP,
			},
		)
	}
	for i := int32(0); i < opts.TestNamespaceCount+1; i++ {
		var nsPorts []string
		for j := int32(0); j < 10; j++ {
			if opts.StartingPort != -1 {
				port := opts.StartingPort + i*10 + j
				nsPorts = append(nsPorts, strconv.Itoa(int(port)))
				ports = append(ports, v1alpha4.PortMapping{
					ContainerPort: port,
					HostPort:      port,
					Protocol:      v1alpha4.PortMappingProtocolTCP,
				})
			}
		}
		var ns map[string]string
		if i == 0 {
			ns = annotations["default"]
		} else {
			x := make(map[string]string)
			name := "test-namespace-" + strconv.Itoa(int(i))
			annotations[name] = x
			ns = annotations[name]
		}
		ns[exposedPortsKey] = strings.Join(nsPorts, ",")
	}
	config := &v1alpha4.Cluster{
		TypeMeta: v1alpha4.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "kind.x-k8s.io/v1alpha4",
		},
		Nodes: []v1alpha4.Node{
			{
				Role: v1alpha4.ControlPlaneRole,
				KubeadmConfigPatches: []string{`
kind: InitConfiguration
nodeRegistration:
    kubeletExtraArgs:
        node-labels: "ingress-ready=true"`,
				},
				ExtraPortMappings: ports,
			},
		},
	}
	if containerdPullOverride {
		config.ContainerdConfigPatches = []string{
			fmt.Sprintf(`[plugins."io.containerd.grpc.v1.cri".registry.mirrors.%q]`, registryHostname+":5001") +
				"\n" +
				`  endpoint = ["http://pach-registry:5000"]`,
		}
	}

	// Create the Kind cluster and wait for it to be ready.
	co := []cluster.CreateOption{
		cluster.CreateWithNodeImage(nodeImage),
		cluster.CreateWithV1Alpha4Config(config),
		cluster.CreateWithDisplayUsage(false),
		cluster.CreateWithDisplaySalutation(false),
	}
	if p := opts.kubeconfigPath; p != "" {
		co = append(co, cluster.CreateWithKubeconfigPath(p))
		c.kubeconfig = Kubeconfig(p)
	}
	if err := c.provider.Create(c.name, co...); err != nil {
		return errors.Wrap(err, "create cluster")
	}

	// Link the registry to kind's network.  We do this here because kind creates a "kind"
	// network the first time you create a cluster; most people probably have this, but CI
	// doesn't.
	if ensuredRegistry {
		if err := connectRegistry(ctx, registryName); err != nil {
			return errors.Wrap(err, "connect registry to the kind network")
		}
	}

	// Configure the Kubernetes cluster.
	kc, err := c.GetKubeconfig(ctx)
	if err != nil {
		return errors.Wrap(err, "get kubeconfig")
	}
	kube, err := kc.Client()
	if err != nil {
		return errors.Wrap(err, "get kube client")
	}

	// Add annotations to the default namespace.
	log.Info(ctx, "configuring default namespace")
	patch := map[string]any{"metadata": map[string]map[string]string{"annotations": annotations["default"]}}
	js, err := json.Marshal(patch)
	if err != nil {
		return errors.Wrapf(err, "marshal json for namespace patch")
	}
	if _, err := kube.CoreV1().Namespaces().Patch(ctx, "default", types.StrategicMergePatchType, js, v1.PatchOptions{
		FieldManager: fieldManager,
	}); err != nil {
		return errors.Wrap(err, "add default annotations")
	}
	delete(annotations, "default")

	// Create the test-runner namespaces and configure them.
	for name, a := range annotations {
		log.Info(ctx, "configuring test-runner namespace", zap.String("namespace", name))
		ns := &corev1.Namespace{
			ObjectMeta: v1.ObjectMeta{
				Name:        name,
				Annotations: a,
			},
		}
		if _, err := kube.CoreV1().Namespaces().Create(ctx, ns, v1.CreateOptions{
			FieldManager: fieldManager,
		}); err != nil {
			return errors.Wrapf(err, "create namespace %v", ns)
		}
	}

	// Install minio.
	log.Info(ctx, "installing minio")
	minioYaml, err := runfiles.Rlocation("_main/src/internal/kindenv/minio.yaml")
	if err != nil {
		return errors.Wrap(err, "minio.yaml not in binary runfiles; build with bazel")
	}
	if err := kc.KubectlCommand(ctx, "apply", "-f", minioYaml).Run(); err != nil {
		return errors.Wrap(err, "kubectl apply -f minio.yaml")
	}

	// Tweak DNS for higher performance.
	log.Info(ctx, "tweaking DNS config")
	dnsConfig, err := runfiles.Rlocation("_main/src/internal/kindenv/coredns_configmap.yaml")
	if err != nil {
		return errors.Wrap(err, "coredns_configmap.yaml not in binary runfiles; build with bazel")
	}
	if err := kc.KubectlCommand(ctx, "apply", "-f", dnsConfig).Run(); err != nil {
		return errors.Wrap(err, "kubectl apply -f coredns_configmap.yaml")
	}

	// Install metrics-server.
	log.Info(ctx, "installing metrics-server")
	metricsServerTar, err := runfiles.Rlocation("com_github_kubernetes_sigs_metrics_server_helm_chart/file/metrics-server.tgz")
	if err != nil {
		return errors.Wrap(err, "metrics-server.tgz not in binary runfiles; build with bazel")
	}
	if err := kc.HelmCommand(ctx, "install", "--set", "args={--kubelet-insecure-tls}", "metrics-server", metricsServerTar, "--namespace", "kube-system").Run(); err != nil {
		return errors.Wrap(err, "helm install metrics-server")
	}

	return nil
}

// Delete destroys the cluster.
func (c *Cluster) Delete(ctx context.Context) error {
	if err := c.provider.Delete(c.name, ""); err != nil {
		return errors.Wrap(err, "delete cluster")
	}
	return nil
}
