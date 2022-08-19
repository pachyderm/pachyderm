package sandbox

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cluster/nodes"
	"sigs.k8s.io/kind/pkg/cluster/nodeutils"
	"sigs.k8s.io/kind/pkg/cmd"
	"sigs.k8s.io/kind/pkg/fs"

	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/exec"
)

/*
TODO
- deploy different versions of pachyderm
- better logging
- replicate sandbox environment in production?
- allow user to specify local chart and values for pachyderm developers to test latest changes
*/

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
	logger.V(0).Info("Running helm repo add and update ...")
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
	// home, err := os.UserHomeDir()
	// if err != nil {
	// 	return err
	// }
	// if err := f.WriteFile(filepath.Join(home, ".cache/helm"), 0644); err != nil {
	// 	return err
	// }

	// helm install
	release, chart := "pachyderm", "pach/pachyderm"
	client := action.NewInstall(actionConfig)
	client.ReleaseName = release
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

	logger.V(0).Info("Running helm install ...")
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

func imageID(name string) (string, error) {
	c := exec.Command("docker", "image", "inspect", "-f", "{{ .Id }}", name)
	data, err := c.Output()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func loadImage(path string, node nodes.Node) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	nodeutils.LoadImageArchive(node, f)
	return nil
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
			// kind config can be either a file or streamed from stdin
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
		Use:   "load-image <IMAGE> [IMAGE ...]",
		Short: "Loads local images to nodes",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("a list of image names is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				name = getClusterName()
			}
			provider := cluster.NewProvider(cluster.ProviderWithLogger(logger), cluster.ProviderWithDocker())
			var (
				imageNames []string
				imageIDs   []string
			)
			imageNames = args
			for _, imageName := range imageNames {
				imageID, err := imageID(imageName)
				if err != nil {
					return err
				}
				imageIDs = append(imageIDs, imageID)
			}
			nodeList, err := provider.ListInternalNodes(name)
			if err != nil {
				return err
			}
			// pick only the nodes that don't have the image
			var selectedNodes []nodes.Node
			for i, imageName := range imageNames {
				imageID := imageIDs[i]
				for _, node := range nodeList {
					id, err := nodeutils.ImageID(node, imageName)
					if err != nil || id != imageID {
						selectedNodes = append(selectedNodes, node)
					}
				}
			}
			if len(selectedNodes) == 0 {
				return nil
			}

			// Setup the tar path where the images will be saved
			dir, err := fs.TempDir("", "images-tar")
			if err != nil {
				return err
			}
			defer os.RemoveAll(dir)
			imagesTarPath := filepath.Join(dir, "images.dar")
			// save the images to a tar
			if err := exec.Command("docker", append([]string{
				"save",
				"-o",
				imagesTarPath},
				imageNames...)...).Run(); err != nil {
				return err
			}

			var fns []func() error
			for _, node := range selectedNodes {
				node := node // capture loop variable for closure
				fns = append(fns, func() error {
					return loadImage(imagesTarPath, node)
				})
			}
			// TODO execute fns
			g := new(errgroup.Group)
			for _, f := range fns {
				f := f // capture loop variable for closure
				g.Go(f)
			}
			if err := g.Wait(); err != nil {
				return err
			}
			return nil
		},
	}
	loadImage.Flags().StringVar(&name, "name", "", "the cluster name")

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
