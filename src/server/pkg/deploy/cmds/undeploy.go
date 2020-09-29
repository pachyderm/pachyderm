package cmds

import (
	"fmt"
	"strings"

	"github.com/pachyderm/pachyderm/src/client/pkg/config"
	"github.com/pachyderm/pachyderm/src/client/pkg/helm"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func CreateUndeployCmd() *cobra.Command {
	var all bool
	var includingMetadata bool
	var includingIDE bool
	var namespace string
	undeploy := &cobra.Command{
		Short: "Tear down a deployed Pachyderm cluster.",
		Long:  "Tear down a deployed Pachyderm cluster.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			// TODO(ys): remove the `--namespace` flag here eventually
			if namespace != "" {
				fmt.Printf("WARNING: The `--namespace` flag is deprecated and will be removed in a future version. Please set the namespace in the pachyderm context instead: pachctl config update context `pachctl config get active-context` --namespace '%s'\n", namespace)
			}
			// TODO(ys): remove the `--all` flag here eventually
			if all {
				fmt.Printf("WARNING: The `--all` flag is deprecated and will be removed in a future version. Please use `--metadata` instead.\n")
				includingMetadata = true
			}

			if includingMetadata {
				fmt.Printf(`
You are going to delete persistent volumes where metadata is stored. If your
persistent volumes were dynamically provisioned (i.e. if you used the
"--dynamic-etcd-nodes" flag), the underlying volumes will be removed, making
metadata such as repos, commits, pipelines, and jobs unrecoverable. If your
persistent volume was manually provisioned (i.e. if you used the
"--static-etcd-volume" flag), the underlying volume will not be removed.
`)
			}

			if ok, err := cmdutil.InteractiveConfirm(); err != nil {
				return err
			} else if !ok {
				return nil
			}

			cfg, err := config.Read(false)
			if err != nil {
				return err
			}
			_, activeContext, err := cfg.ActiveContext(true)
			if err != nil {
				return err
			}

			if namespace == "" {
				namespace = activeContext.Namespace
			}

			assets := []string{
				"service",
				"replicationcontroller",
				"deployment",
				"serviceaccount",
				"secret",
				"statefulset",
				"clusterrole",
				"clusterrolebinding",
			}
			if includingMetadata {
				assets = append(assets, []string{
					"storageclass",
					"persistentvolumeclaim",
					"persistentvolume",
				}...)
			}
			if err := kubectl(nil, activeContext, "delete", strings.Join(assets, ","), "-l", "suite=pachyderm", "--namespace", namespace); err != nil {
				return err
			}

			if includingIDE {
				// remove IDE
				if err = helm.Destroy(activeContext, "pachyderm-ide", namespace); err != nil {
					log.Errorf("failed to delete helm installation: %v", err)
				}
				ideAssets := []string{
					"replicaset",
					"deployment",
					"service",
					"pod",
				}
				if err = kubectl(nil, activeContext, "delete", strings.Join(ideAssets, ","), "-l", "app=jupyterhub", "--namespace", namespace); err != nil {
					return err
				}
			}

			// remove the context from the config
			kubeConfig, err := config.RawKubeConfig()
			if err != nil {
				return err
			}
			kubeContext := kubeConfig.Contexts[kubeConfig.CurrentContext]
			if kubeContext != nil {
				cfg, err := config.Read(true)
				if err != nil {
					return err
				}
				ctx := &config.Context{
					ClusterName: kubeContext.Cluster,
					AuthInfo:    kubeContext.AuthInfo,
					Namespace:   namespace,
				}

				// remove _all_ contexts associated with this
				// deployment
				configUpdated := false
				for {
					contextName, _ := findEquivalentContext(cfg, ctx)
					if contextName == "" {
						break
					}
					configUpdated = true
					delete(cfg.V2.Contexts, contextName)
					if contextName == cfg.V2.ActiveContext {
						cfg.V2.ActiveContext = ""
					}
				}
				if configUpdated {
					if err = cfg.Write(); err != nil {
						return err
					}
				}
			}

			return nil
		}),
	}
	undeploy.Flags().BoolVarP(&all, "all", "a", false, "DEPRECATED: Use \"--metadata\" instead.")
	undeploy.Flags().BoolVarP(&includingMetadata, "metadata", "", false, `
Delete persistent volumes where metadata is stored. If your persistent volumes
were dynamically provisioned (i.e. if you used the "--dynamic-etcd-nodes"
flag), the underlying volumes will be removed, making metadata such as repos,
commits, pipelines, and jobs unrecoverable. If your persistent volume was
manually provisioned (i.e. if you used the "--static-etcd-volume" flag), the
underlying volume will not be removed.`)
	undeploy.Flags().BoolVarP(&includingIDE, "ide", "", false, "Delete the Pachyderm IDE deployment if it exists.")
	undeploy.Flags().StringVar(&namespace, "namespace", "", "Kubernetes namespace to undeploy Pachyderm from.")
	return undeploy
}
