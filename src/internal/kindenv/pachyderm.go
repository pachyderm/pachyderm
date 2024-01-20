package kindenv

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	ci "github.com/pachyderm/pachyderm/v2/src/internal/middleware/logging/client"
	"github.com/pachyderm/pachyderm/v2/src/version"
	"github.com/pachyderm/pachyderm/v2/src/version/versionpb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

func getPachydermVersions() (string, string, error) {
	pachdVersion, err := readRunfile("_main/oci/pachd_image.json.sha256")
	if err != nil {
		return "", "", errors.Wrap(err, "get pachd image digest")
	}
	pachdVersion = bytes.TrimRight(pachdVersion[7:], "\n")
	workerVersion, err := readRunfile("_main/oci/worker_image.json.sha256")
	if err != nil {
		return "", "", errors.Wrap(err, "get worker image digest")
	}
	workerVersion = bytes.TrimRight(workerVersion[7:], "\n")
	return string(pachdVersion), string(workerVersion), nil
}

// PushPachyderm pushes the images necessary to run Pachyderm.
func (c *Cluster) PushPachyderm(ctx context.Context) error {
	log.Info(ctx, "locating images")
	pachdVersion, workerVersion, err := getPachydermVersions()
	if err != nil {
		return errors.Wrap(err, "get pachyderm version")
	}
	pachdImage, err := runfiles.Rlocation("_main/oci/pachd_image")
	if err != nil {
		return errors.Wrap(err, "find pachd image")
	}
	workerImage, err := runfiles.Rlocation("_main/oci/worker_image")
	if err != nil {
		return errors.Wrap(err, "find worker image")
	}

	cfg, err := c.GetConfig(ctx, "default")
	if err != nil {
		return errors.Wrap(err, "get cluster configuration")
	}

	log.Info(ctx, "pushing images")
	if err := SkopeoCommand(ctx, "copy", "oci:"+pachdImage, path.Join(cfg.ImagePushPath, "pachd:"+pachdVersion)).Run(); err != nil {
		return errors.Wrap(err, "copy pachd image to registry")
	}
	if err := SkopeoCommand(ctx, "copy", "oci:"+workerImage, path.Join(cfg.ImagePushPath, "worker:"+workerVersion)).Run(); err != nil {
		return errors.Wrap(err, "copy worker image to registry")
	}

	return nil
}

type HelmConfig struct {
	Namespace       string
	Diff            bool
	NoSwitchContext bool
}

func (c *Cluster) helmFlags(ctx context.Context, install *HelmConfig) ([]string, error) {
	pachdVersion, workerVersion, err := getPachydermVersions()
	if err != nil {
		return nil, errors.Wrap(err, "get pachyderm version")
	}

	cfg, err := c.GetConfig(ctx, install.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "get cluster config")
	}

	flags := []string{
		"pachyderm",
		// TODO(jrockway): If installing a pachyderm version that's not current, use that
		// version's helm chart.
		filepath.Join(os.Getenv("BUILD_WORKSPACE_DIRECTORY"), "etc/helm/pachyderm"),
		"--namespace",
		install.Namespace,
		"--set",
		strings.Join([]string{
			// Configure images.
			"pachd.image.repository=" + path.Join(cfg.ImagePullPath, "pachd"),
			"worker.image.repository=" + path.Join(cfg.ImagePullPath, "worker"),
			"pachd.image.tag=" + pachdVersion,
			"worker.image.tag=" + workerVersion,
			// Configure the external hostname.
			"proxy.host=" + cfg.Hostname,
		}, ","),
	}

	// Configure enterprise.
	if key := os.Getenv("ENT_ACT_CODE"); key != "" {
		flags = append(flags, "--set", "pachd.enterpriseLicenseKey="+key)
	} else {
		log.Info(ctx, "not activating enterprise, as $ENT_ACT_CODE is unset")
	}

	// Configure the proxy.
	proxyPort, err := c.AllocatePort(ctx, install.Namespace, pachydermProxy)
	if err != nil {
		return nil, errors.Wrap(err, "allocate proxy port")
	}
	proxyPortParts := strings.Split(proxyPort, ",")
	if len(proxyPortParts) == 1 {
		if cfg.TLS {
			flags = append(flags, "--set", "proxy.service.httpsNodePort="+proxyPortParts[0])
		} else {
			flags = append(flags, "--set", "proxy.service.httpNodePort="+proxyPortParts[0])
		}
	} else if len(proxyPortParts) > 1 {
		flags = append(flags, "--set", "proxy.service.httpNodePort="+proxyPortParts[0])
		flags = append(flags, "--set", "proxy.service.httpsNodePort="+proxyPortParts[1])
	}
	if cfg.TLS {
		flags = append(flags, "--set", "proxy.tls.enabled=true")
	}

	// Get the default values file.
	defaultValues, err := runfiles.Rlocation("_main/src/internal/kindenv/pachyderm-values.json")
	if err != nil {
		return nil, errors.Wrap(err, "get default values file")
	}
	flags = append(flags, "-f", defaultValues)

	// Look for ~/.config/pachdev/pachyderm-values.json and use if found.
	global, err := xdg.ConfigFile("pachdev/pachyderm-values.json")
	if err != nil {
		log.Info(ctx, "not using global value overrides; $XDG_CONFIG_HOME broken", zap.String("path", global), zap.Error(err))
	} else if _, err := os.Stat(global); err != nil {
		log.Debug(ctx, "not using global value overrides; file unreadable", zap.String("path", global))
	} else {
		log.Info(ctx, "using global value overrides", zap.String("path", global))
		flags = append(flags, "-f", global)
	}

	// Look for ~/.config/pachdev/<name>-pachyderm-values.json and use if found.
	local, err := xdg.ConfigFile(fmt.Sprintf("pachdev/%s-pachyderm-values.json", c.name))
	if err != nil {
		log.Info(ctx, "not using local value overrides; $XDG_CONFIG_HOME broken", zap.String("path", local), zap.Error(err))
	} else if _, err := os.Stat(local); err != nil {
		log.Debug(ctx, "not using local value overrides; file unreadable", zap.String("path", local))
	} else {
		log.Info(ctx, "using local value overrides", zap.String("path", local))
		flags = append(flags, "-f", local)
	}

	return flags, nil
}

// InstallPachyderm installs or upgrades Pachyderm.
func (c *Cluster) InstallPachyderm(ctx context.Context, install *HelmConfig) error {
	var flags []string
	if install.Diff {
		flags = append(flags, "diff", "upgrade")
	} else {
		flags = append(flags, "upgrade", "--install")
	}
	moreFlags, err := c.helmFlags(ctx, install)
	if err != nil {
		return errors.Wrap(err, "compute helm flags")
	}
	flags = append(flags, moreFlags...)

	k, err := c.GetKubeconfig(ctx)
	if err != nil {
		return errors.Wrap(err, "get kubeconfig")
	}

	// Helm upgrade.
	log.Info(ctx, "running helm upgrade --install")
	log.Debug(ctx, "helm flags", zap.Strings("flags", flags))
	cmd := k.HelmCommand(ctx, flags...)
	if install.Diff {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "helm %v", flags)
	}
	if install.Diff {
		return nil
	}

	// Configure pachctl now, in case the user presses C-c because they're tired of waiting for
	// image pulls.
	log.Info(ctx, "configuring pachctl")
	if err := c.ConfigurePachctl(ctx, install.Namespace, !install.NoSwitchContext); err != nil {
		log.Error(ctx, "unable to configure pachctl", zap.Error(err))
	}

	// Wait for the proxy to come up, so we know the network works for the next step.
	log.Info(ctx, "waiting for the proxy to come up")
	if err := k.KubectlCommand(ctx, "--namespace="+install.Namespace, "rollout", "status", "deployment", "pachyderm-proxy").Run(); err != nil {
		log.Error(ctx, "problem waiting for the proxy; continuing anyway", zap.Error(err))
	}

	// Now wait for pachd.
	if err := c.WaitForPachd(ctx, install.Namespace); err != nil {
		return errors.Wrap(err, "wait for pachd to be ready")
	}
	return nil
}

func (c *Cluster) PachdAddress(ctx context.Context, namespace string) (string, error) {
	cc, err := c.GetConfig(ctx, namespace)
	if err != nil {
		return "", errors.Wrap(err, "get config")
	}
	var grpcs string
	port, err := c.AllocatePort(ctx, namespace, pachydermProxy)
	if err != nil {
		return "", errors.Wrap(err, "get proxy port")
	}
	switch {
	case port == "30080,30443" && cc.TLS:
		port = "443"
		grpcs = "s"
	case port == "30080,30443" && !cc.TLS:
		port = "80"
	case cc.TLS:
		grpcs = "s"
	}
	return fmt.Sprintf("grpc%v://%v:%v", grpcs, cc.Hostname, port), nil
}

func (c *Cluster) ConfigurePachctl(ctx context.Context, namespace string, activate bool) error {
	pachd, err := c.PachdAddress(ctx, namespace)
	if err != nil {
		return errors.Wrap(err, "get pachd address")
	}
	cf, err := config.Read(false, false)
	if err != nil {
		return errors.Wrap(err, "read config")
	}
	context := c.name
	if namespace != "default" {
		context = c.name + "/" + namespace
	}
	cf.V2.Contexts[context] = &config.Context{
		PachdAddress: pachd,
		SessionToken: "iamroot",
	}
	if activate {
		cf.V2.ActiveContext = context
	}
	if err := cf.Write(); err != nil {
		return errors.Wrap(err, "write config")
	}
	log.Info(ctx, "configured pachctl context", zap.String("context", context))
	return nil
}

func (c *Cluster) WaitForPachd(ctx context.Context, namespace string) error {
	addr, err := c.PachdAddress(ctx, namespace)
	if err != nil {
		return errors.Wrap(err, "get pachd address")
	}

	opts := []grpc.DialOption{
		grpc.WithUnaryInterceptor(ci.LogUnary),
		grpc.WithStreamInterceptor(ci.LogStream),
	}
	u, err := url.Parse(addr)
	if err != nil {
		return errors.Wrap(err, "parse pachd address")
	}
	if u.Scheme == "grpc" {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		t := &tls.Config{
			InsecureSkipVerify: true,
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(t)))
	}

	log.Info(ctx, "waiting for pachd to be ready")
	if err := backoff.RetryUntilCancel(ctx, func() error {
		ctx, c := context.WithTimeout(ctx, 2*time.Second)
		defer c()
		cc, err := grpc.DialContext(ctx, "dns:///"+u.Host, opts...)
		if err != nil {
			return errors.Wrap(err, "dial")
		}
		vc := versionpb.NewAPIClient(cc)
		v, err := vc.GetVersion(ctx, &emptypb.Empty{})
		if err != nil {
			return errors.Wrap(err, "GetVersion")
		}
		got, want := v.Canonical(), version.Version.Canonical()
		if got != want {
			return errors.Errorf("running pachd has wrong version; got %v, want %v", got, want)
		}
		log.Info(ctx, "pachd ready", log.Proto("version", v))
		return nil
	}, backoff.NewConstantBackOff(time.Second), func(err error, d time.Duration) error {
		log.Info(ctx, "pachd not ready: "+err.Error())
		return nil
	}); err != nil {
		return errors.Wrap(err, "dial and check version")
	}
	return nil
}
