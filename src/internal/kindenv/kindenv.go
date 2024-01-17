// Package kindenv manages Kind (github.com/kubernetes-sigs/kind) environments.
package kindenv

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
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
	// Name is the name of the cluster and Kubernetes context. "kind-pach-" is prepended unless
	// the name is empty, in which case "kind-pach" is the name of the cluster.
	Name string
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

// KindName returns the name that Kind uses to refer to this cluster.
func (o *CreateOpts) KindName() string {
	if o.Name == "" {
		return "pach"
	}
	return "pach-" + o.Name
}

// Create creates a new cluster.
func Create(ctx context.Context, opts *CreateOpts) error {
	po, err := cluster.DetectNodeProvider()
	if err != nil {
		return errors.Wrap(err, "detect kind node provider")
	}
	if po == nil {
		return errors.New("kind could not detect docker or podman; install docker")
	}

	// k8s annotations to be applied to the default namespace; this is how we transfer
	// configuration between tests/dev tools/etc.
	defaultAnnotations := map[string]string{
		"io.pachyderm/kindenv-version": strconv.Itoa(version),
	}

	var ports []v1alpha4.PortMapping
	if opts.BindHTTPPorts {
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
	for i := int32(0); i < opts.TestNamespaceCount; i++ {
		for j := int32(0); j < 10; j++ {
			port := opts.StartingPort + i*10 + j
			ports = append(ports, v1alpha4.PortMapping{
				ContainerPort: port,
				HostPort:      port,
				Protocol:      v1alpha4.PortMappingProtocolTCP,
			})
		}
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

	p := cluster.NewProvider(po) // TODO: add logger
	name := opts.KindName()
	if err := p.Create(name, cluster.CreateWithNodeImage(nodeImage), cluster.CreateWithV1Alpha4Config(config)); err != nil {
		return errors.Wrap(err, "create cluster")
	}
	cfg, err := p.KubeConfig(name, false)
	if err != nil {
		return errors.Wrap(err, "get config")
	}
	fmt.Println(defaultAnnotations)
	fmt.Printf("\n%s\n", cfg)
	return nil
}
