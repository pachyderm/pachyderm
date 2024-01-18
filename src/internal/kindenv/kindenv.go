// Package kindenv manages Kind (github.com/kubernetes-sigs/kind) environments.
package kindenv

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/promutil"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
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
)

// CreateOpts specifies a Kind environment.
//
// Note: kind treats port numbers as int32, so we do too.  They are actually uint16s.
type CreateOpts struct {
	// TestNamespaceCount controls how many K8s tests can run concurrently.
	TestNamespaceCount int32
	// ExternalRegistry is the Skopeo path of the local container registry from the host
	// machine, usually something like `oci:/path/to/pach-registry`.
	ExternalRegistry string
	// BindHTTPPorts, if true, binds localhost:80 and localhost:443 to 30080 and 30443 in the
	// cluster; used for the install of Pachyderm in the "default" namespace.
	BindHTTPPorts bool
	// StartingPort is the port number that begins the exposed ports for this cluster.  Each
	// TestNamespace gets 10.
	StartingPort int32
}

func named(x string) string {
	if x == "pach" || x == "" {
		return "pach"
	}
	return "pach-" + x
}

// Cluster represents a Kind cluster.  It may not exist yet.
type Cluster struct {
	name     string
	provider *cluster.Provider
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
// the cluster in the current kubernetes context will be used.
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

// Create creates a new cluster.
func (c *Cluster) Create(ctx context.Context, opts *CreateOpts) error {
	// k8s annotations to be applied to the default namespace; this is how we transfer
	// configuration between tests/dev tools/etc.
	annotations := map[string]map[string]string{
		"default": {
			"dev.pachyderm.io/kindenv-version": strconv.Itoa(version),
			// Kind recommends this KEP, but it never got approved.  So we build our own
			// that doesn't involve parsing YAML.
			// https://github.com/kubernetes/enhancements/tree/master/keps/sig-cluster-lifecycle/generic/1755-communicating-a-local-registry
			"dev.pachyderm.io/registry": opts.ExternalRegistry,
		},
	}

	if opts.ExternalRegistry == "" {
		path, err := ensureRegistry(ctx)
		if err != nil {
			return errors.Wrap(err, "setup pach-registry container")
		}
		annotations["default"]["dev.pachyderm.io/registry"] = "oci:" + path
	}

	var ports []v1alpha4.PortMapping
	if opts.BindHTTPPorts {
		annotations["default"]["dev.pachyderm.io/standard-http-ports"] = "true"
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
	if opts.TestNamespaceCount > 0 && opts.StartingPort == 0 {
		opts.StartingPort = 30500
	}
	for i := int32(0); i < opts.TestNamespaceCount+1; i++ {
		var nsPorts []string
		for j := int32(0); j < 10; j++ {
			port := opts.StartingPort + i*10 + j
			nsPorts = append(nsPorts, strconv.Itoa(int(port)))
			ports = append(ports, v1alpha4.PortMapping{
				ContainerPort: port,
				HostPort:      port,
				Protocol:      v1alpha4.PortMappingProtocolTCP,
			})
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
		ns["dev.pachyderm.io/exposed-ports"] = strings.Join(nsPorts, ",")
	}
	config := &v1alpha4.Cluster{
		TypeMeta: v1alpha4.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "kind.x-k8s.io/v1alpha4",
		},
		ContainerdConfigPatches: []string{
			`[plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:5001"]` +
				"\n" +
				`  endpoint = ["http://pach-registry:5000"]`,
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

	if err := c.provider.Create(c.name, cluster.CreateWithNodeImage(nodeImage), cluster.CreateWithV1Alpha4Config(config)); err != nil {
		return errors.Wrap(err, "create cluster")
	}

	kube, err := c.KubeClient(ctx)
	if err != nil {
		return errors.Wrap(err, "get kube client")
	}

	// Patch default namespace.
	log.Info(ctx, "configuring default namespace")
	patch := map[string]any{"metadata": map[string]map[string]string{"annotations": annotations["default"]}}
	js, err := json.Marshal(patch)
	if err != nil {
		return errors.Wrapf(err, "marshal json for namespace patch")
	}
	if _, err := kube.CoreV1().Namespaces().Patch(ctx, "default", types.StrategicMergePatchType, js, v1.PatchOptions{
		FieldManager: "pachdev",
	}); err != nil {
		return errors.Wrap(err, "add default annotations")
	}
	delete(annotations, "default")

	for name, a := range annotations {
		log.Info(ctx, "configuring test-runner namespace", zap.String("namespace", name))
		ns := &corev1.Namespace{
			ObjectMeta: v1.ObjectMeta{
				Name:        name,
				Annotations: a,
			},
		}
		kube.CoreV1().Namespaces().Create(ctx, ns, v1.CreateOptions{
			FieldManager: "pachdev",
		})
	}

	return nil
}

// KubeClient gets a Kubernetes API client for the named cluster.
func (c *Cluster) KubeClient(ctx context.Context) (kubernetes.Interface, error) {
	f, err := os.CreateTemp("", "pachdev-kubeconfig-*")
	if err != nil {
		return nil, errors.Wrap(err, "create tmpfile for config")
	}
	cfg, err := c.provider.KubeConfig(c.name, false)
	if err != nil {
		return nil, errors.Wrap(err, "get kubeconfig")
	}
	if _, err := io.Copy(f, strings.NewReader(cfg)); err != nil {
		return nil, errors.Wrap(err, "write kubeconfig")
	}
	if err := f.Close(); err != nil {
		return nil, errors.Wrap(err, "close kubeconfig")
	}

	loadingRules := &clientcmd.ClientConfigLoadingRules{
		ExplicitPath: f.Name(),
	}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, errors.Wrap(err, "load kubeconfig")
	}
	config.WrapTransport = func(rt http.RoundTripper) http.RoundTripper {
		return promutil.InstrumentRoundTripper("kubernetes", rt)
	}
	st, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "new k8s clientset")
	}
	return st, nil
}

// Delete destroys the cluster.
func (c *Cluster) Delete(ctx context.Context) error {
	if err := c.provider.Delete(c.name, ""); err != nil {
		return errors.Wrap(err, "delete cluster")
	}
	return nil
}
