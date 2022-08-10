package sandbox

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cmd"
)

// TODO
// logging
// cleanup pach config
// multi node k8s?
// replicate sandbox environment in production?

var clusterName = "pachsandbox"
var logger = cmd.NewLogger()

func debug(format string, v ...interface{}) {
	log.Output(2, fmt.Sprintf(format, v...))
}

func deploy() error {
	settings := cli.New()
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), debug); err != nil {
		return err
	}
	// helm repo add
	cfg := repo.Entry{
		Name: "pach",
		URL:  "https://helm.pachyderm.com",
	}
	r, err := repo.NewChartRepository(&cfg, getter.All(settings))
	if err != nil {
		return err
	}
	_, err = r.DownloadIndexFile()
	if err != nil {
		return err
	}
	var f repo.File
	// helm repo update
	f.Update(&cfg)
	if err := f.WriteFile("~/.cache/helm", 0644); err != nil {
		return err
	}
	// helm install
	name, chart := "pachd", "pach/pachyderm"
	client := action.NewInstall(actionConfig)
	client.ReleaseName = name
	client.Namespace = settings.Namespace()
	client.Timeout = time.Minute * 5
	client.Wait = true

	cp, err := client.ChartPathOptions.LocateChart(chart, settings)
	if err != nil {
		return err
	}
	p := getter.All(settings)
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

func NewCommand() *cobra.Command {
	sandbox := &cobra.Command{
		Use:   "sandbox",
		Short: "Pachyderm local cluster util",
	}

	up := &cobra.Command{
		Use:   "up",
		Short: "starts a local Pachyderm cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cluster.NewProvider(
				cluster.ProviderWithLogger(logger),
				cluster.ProviderWithDocker()).Create(clusterName); err != nil {
				return err
			}
			logger.V(0).Info("Deploying Pachyderm")
			if err := deploy(); err != nil {
				return err
			}
			// configure pachctl to talk to the new k8s cluster
			cfg, err := config.Read(false, false)
			if err != nil {
				return err
			}
			kubeConfig, err := config.RawKubeConfig()
			if err != nil {
				return err
			}
			kubeContextName := "kind-" + clusterName
			kubeContext := kubeConfig.Contexts[kubeContextName]
			if kubeContext == nil {
				return fmt.Errorf("kubernetes context does not exist: %s", kubeContextName)
			}
			var context = config.Context{
				Source:      config.ContextSource_IMPORTED,
				ClusterName: kubeContext.Cluster,
				Namespace:   kubeContext.Namespace,
			}
			cfg.V2.Contexts[clusterName] = &context
			cfg.V2.ActiveContext = clusterName
			return cfg.Write()
		},
	}

	destroy := &cobra.Command{
		Use:   "destroy",
		Short: "deletes a local Pachyderm cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.V(0).Infof("Deleting cluster %q ...", clusterName)
			return cluster.NewProvider(
				cluster.ProviderWithLogger(logger),
				cluster.ProviderWithDocker()).Delete(clusterName, "")
		},
	}

	sandbox.AddCommand(up, destroy)
	return sandbox
}
