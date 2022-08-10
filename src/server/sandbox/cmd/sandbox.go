package sandbox

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cmd"

	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// TODO
// logging
// multi node k8s?
// replicate sandbox environment in production?
// load local docker images into kind
// add a status command to check if there is a sandbox

const clusterName = "pach-sandbox"

var (
	logger       = cmd.NewLogger() // from kind
	helmSettings = cli.New()       // from helm
)

func getClusterName() string {
	if name := os.Getenv("PACH_CLUSTER_NAME"); name != "" {
		return name
	}
	return clusterName
}

func debug(format string, v ...interface{}) {
	if helmSettings.Debug {
		log.Output(2, fmt.Sprintf(format, v...))
	}
}

func helmDeploy() error {
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(helmSettings.RESTClientGetter(), helmSettings.Namespace(), os.Getenv("HELM_DRIVER"), debug); err != nil {
		return err
	}

	// helm repo add
	// helm repo update
	cfg := repo.Entry{
		Name: "pach",
		URL:  "https://helm.pachyderm.com",
	}
	r, err := repo.NewChartRepository(&cfg, getter.All(helmSettings))
	if err != nil {
		return err
	}
	_, err = r.DownloadIndexFile()
	if err != nil {
		return err
	}
	var f repo.File
	f.Update(&cfg)
	if err := f.WriteFile("~/.cache/helm", 0644); err != nil {
		return err
	}

	// helm install
	name, chart := "pachyderm", "pach/pachyderm"
	client := action.NewInstall(actionConfig)
	client.ReleaseName = name
	client.Namespace = helmSettings.Namespace()
	client.Timeout = time.Minute * 5
	client.Wait = true
	// client.Devel = true
	// client.Version = "2.3.0-rc.2"

	cp, err := client.ChartPathOptions.LocateChart(chart, helmSettings)
	if err != nil {
		return err
	}
	p := getter.All(helmSettings)
	valueOpts := &values.Options{
		StringValues: []string{""},
	}
	vals, err := valueOpts.MergeValues(p)
	if err != nil {
		return err
	}
	vals["deployTarget"] = "LOCAL"
	chartRequested, err := loader.Load(cp)
	if err != nil {
		return err
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	cSignal := make(chan os.Signal, 2)
	signal.Notify(cSignal, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-cSignal
		cancel()
	}()

	_, err = client.RunWithContext(ctx, chartRequested, vals)
	if err != nil {
		return err
	}
	return nil
}

func addContextToPach(kubeContextName string) error {
	cfg, err := config.Read(false, false)
	if err != nil {
		return err
	}
	kubeConfig, err := config.RawKubeConfig()
	if err != nil {
		return err
	}
	kubeContext := kubeConfig.Contexts[kubeContextName]
	if kubeContext == nil {
		return errors.Errorf("kubernetes context does not exist: %s", kubeContextName)
	}
	var context = config.Context{
		Source:      config.ContextSource_IMPORTED,
		ClusterName: kubeContext.Cluster,
		Namespace:   kubeContext.Namespace,
	}
	cfg.V2.Contexts[kubeContextName] = &context
	cfg.V2.ActiveContext = kubeContextName
	return cfg.Write()
}

func NewCommand() *cobra.Command {
	sandbox := &cobra.Command{
		Use:   "sandbox",
		Short: "Pachyderm local cluster util",
	}

	var (
		name       string
		kindConfig string
	)
	up := &cobra.Command{
		Use:   "up",
		Short: "Starts a local Pachyderm cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				name = getClusterName()
			}
			provider := cluster.NewProvider(
				cluster.ProviderWithLogger(logger),
				cluster.ProviderWithDocker())

			var withConfig cluster.CreateOption
			if kindConfig == "-" {
				raw, err := ioutil.ReadAll(os.Stdin)
				if err != nil {
					return err
				}
				withConfig = cluster.CreateWithRawConfig(raw)
			} else {
				withConfig = cluster.CreateWithConfigFile(kindConfig)
			}
			if err := provider.Create(
				name,
				withConfig,
				cluster.CreateWithWaitForReady(time.Minute)); err != nil {
				return err
			}
			logger.V(0).Info("Deploying Pachyderm ...")
			if err := helmDeploy(); err != nil {
				return err
			}
			// kind prefixes cluster name with "kind-"
			// configure pachctl to talk to the new k8s cluster
			return addContextToPach("kind-" + name)
		},
	}
	up.Flags().StringVar(&name, "name", "", "name of cluster")
	up.Flags().StringVar(&kindConfig, "config", "", "path to kind config file, or set --config=- to stream from stdin")

	// TODO
	loadImage := &cobra.Command{
		Use:   "load-image",
		Short: "Loads docker images from host to sandbox cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	destroy := &cobra.Command{
		Use:   "destroy",
		Short: "Removes the local Pachyderm cluster and config created by `sandbox up`",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				name = getClusterName()
			}
			provider := cluster.NewProvider(
				cluster.ProviderWithLogger(logger),
				cluster.ProviderWithDocker())
			logger.V(0).Infof("Deleting cluster %q ...", name)
			if err := provider.Delete(name, ""); err != nil {
				return err
			}
			// delete this context from pach config
			cfg, err := config.Read(false, false)
			if err != nil {
				return err
			}
			kubeContext := "kind-" + name
			if _, ok := cfg.V2.Contexts[kubeContext]; !ok {
				logger.V(0).Infof("context does not exist: %s", kubeContext)
				return nil
			}
			if cfg.V2.ActiveContext == kubeContext {
				cfg.V2.ActiveContext = "default"
			}
			delete(cfg.V2.Contexts, kubeContext)
			return cfg.Write()
		},
	}
	destroy.Flags().StringVar(&name, "name", "", "name of cluster")

	sandbox.AddCommand(up, loadImage, destroy)
	return sandbox
}
