package kindenv

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/promutil"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Kubeconfig is the path to a kubeconfig file.  It has methods on it for working on the cluster
// that is configured by that file.
type Kubeconfig string

// GetKubeconfig returns a Kubeconfig object for the cluster.  Call `Close()` when you're done to
// avoid filling up /tmp with junk.
func (c *Cluster) GetKubeconfig(ctx context.Context) (Kubeconfig, error) {
	f, err := os.CreateTemp("", "pachdev-kubeconfig-*")
	if err != nil {
		return "", errors.Wrap(err, "create tmpfile for config")
	}
	cfg, err := c.provider.KubeConfig(c.name, false)
	if err != nil {
		return "", errors.Wrap(err, "get kubeconfig")
	}
	if _, err := io.Copy(f, strings.NewReader(cfg)); err != nil {
		return "", errors.Wrap(err, "write kubeconfig")
	}
	if err := f.Close(); err != nil {
		return "", errors.Wrap(err, "close kubeconfig")
	}
	return Kubeconfig(f.Name()), nil
}

// Close removes the kubeconfig file.
func (k Kubeconfig) Close() error {
	if k != "" {
		return os.Remove(string(k))
	}
	return nil
}

// KubectlCommand returns an exec.Cmd that will run kubectl.
func (k Kubeconfig) KubectlCommand(ctx context.Context, args ...string) *exec.Cmd {
	path, ok := bazel.FindBinary("//tools/kubectl", "_kubectl")
	if !ok {
		log.Error(ctx, "binary not built with bazel; falling back to host kubectl")
		path = "kubectl"
	}
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "KUBECONFIG="+string(k))
	cmd.Args[0] = "kubectl"
	cmd.Stdout = log.WriterAt(pctx.Child(ctx, "kubectl.stdout"), log.InfoLevel)
	cmd.Stderr = log.WriterAt(pctx.Child(ctx, "kubectl.stderr"), log.InfoLevel)
	return cmd
}

// HelmCommand returns an exec.Cmd that will run helm.
func (k Kubeconfig) HelmCommand(ctx context.Context, args ...string) *exec.Cmd {
	path, ok := bazel.FindBinary("//tools/helm", "_helm")
	if !ok {
		log.Error(ctx, "binary not built with bazel; falling back to host helm")
		path = "helm"
	}
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "KUBECONFIG="+string(k))
	cmd.Args[0] = "helm"
	cmd.Stdout = log.WriterAt(pctx.Child(ctx, "helm.stdout"), log.InfoLevel)
	cmd.Stderr = log.WriterAt(pctx.Child(ctx, "helm.stderr"), log.InfoLevel)
	return cmd
}

// Client returns a kubernetes.Interface connected to the cluster referenced by k.
func (k Kubeconfig) Client() (kubernetes.Interface, error) {
	loadingRules := &clientcmd.ClientConfigLoadingRules{
		ExplicitPath: string(k),
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
