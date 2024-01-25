package pachdev

import (
	"fmt"
	"path"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/kindenv"
	"github.com/spf13/cobra"
)

const DefaultClusterName = "pach"

func DeleteClusterCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete-cluster [<name>]",
		Short: "Delete a local pachdev cluster",
		Long: `Delete a local pachdev cluster.

The cluster to delete is defined by the name, or your current kubernetes context if not set.  The
cluster will only be deleted if the cluster was created by pachdev.  'bazel run //tools/kind --
delete cluster --name=X' will delete any Kind cluster.  `,

		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (retErr error) {
			var name string
			if len(args) > 0 {
				name = args[0]
			}
			ctx := cmd.Context()
			cluster, err := kindenv.New(ctx, name)
			if err != nil {
				return errors.Wrap(err, "kindenv.New")
			}
			defer errors.Close(&retErr, cluster, "close cluster")
			if err := cluster.Delete(ctx); err != nil {
				return errors.Wrap(err, "cluster.Delete")
			}
			return nil
		},
	}
}

func CreateClusterCmd() *cobra.Command {
	var opts kindenv.CreateOpts
	cmd := &cobra.Command{
		Use:   "create-cluster [<name>]",
		Short: "Deploy a local pachdev cluster",
		Long: `Deploy a local pachdev cluster.

The rest of this script relies on having one of these clusters.  You can name it anything; if you
only have one, it will be automatically used.`,

		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (retErr error) {
			name := DefaultClusterName
			if len(args) > 0 {
				name = args[0]
			}
			ctx := cmd.Context()
			cluster, err := kindenv.New(ctx, name)
			if err != nil {
				return errors.Wrap(err, "kindenv.New")
			}
			defer errors.Close(&retErr, cluster, "close cluster")
			if err := cluster.Create(ctx, &opts); err != nil {
				return errors.Wrap(err, "kindenv.Create")
			}
			return nil
		},
	}
	cmd.Flags().Int32Var(&opts.TestNamespaceCount, "test-namespaces", 0, "The number of tests this cluster should support running in parallel.  Note: requires 10 * count free ports.")
	cmd.Flags().StringVar(&opts.ExternalHostname, "hostname", "127.0.0.1", "The externally-accessible hostname (or IP address) for the cluster.")
	cmd.Flags().StringVar(&opts.ImagePushPath, "push-path", "", "If set, push images to the cluster with this Skopeo destination.  If unset, a registry will be started and this value will be automatically configured.")
	cmd.Flags().BoolVar(&opts.BindHTTPPorts, "http", true, "If set, ports 80 and 443 will refer to the Pachyderm running in the default namespace in this cluster.  Only one cluster per machine can have this set.")
	cmd.Flags().Int32Var(&opts.StartingPort, "starting-port", -1, "If set, each namespace will get ports forwarded to the local machine starting with this port number.  If 0, a default will be selected.  If -1, then don't expose any ports.")
	return cmd
}

func LoadImageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "load-image [<cluster name>] <source> <name:tag>",
		Short: "Load a container image into the cluster, renaming it to <name:tag>",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) (retErr error) {
			var name, src, dst string
			switch len(args) {
			case 2:
				src, dst = args[0], args[1]
			case 3:
				name, src, dst = args[0], args[1], args[2]
			default:
				panic("impossible")
			}
			ctx := cmd.Context()
			cluster, err := kindenv.New(ctx, name)
			if err != nil {
				return errors.Wrap(err, "kindenv.New")
			}
			defer errors.Close(&retErr, cluster, "close cluster")
			cfg, err := cluster.GetConfig(ctx, "default")
			if err != nil {
				return errors.Wrap(err, "get cluster config")
			}
			if err := cluster.PushImage(ctx, src, dst); err != nil {
				return errors.Wrap(err, "push image")
			}
			fmt.Printf("%v\n", path.Join(cfg.ImagePullPath, dst))
			return nil
		},
	}
	return cmd
}

func PushPachydermCmd() *cobra.Command {
	var opts kindenv.HelmConfig
	cmd := &cobra.Command{
		Use:   "push [<cluster name>]",
		Short: "Push installs or upgrades Pachyderm so as to reflect any changes in your working directory",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (retErr error) {
			var name string
			if len(args) > 0 {
				name = args[0]
			}
			ctx := cmd.Context()
			cluster, err := kindenv.New(ctx, name)
			if err != nil {
				return errors.Wrap(err, "kindenv.New")
			}
			defer errors.Close(&retErr, cluster, "close cluster")
			if !opts.Diff {
				if err := cluster.PushPachyderm(ctx); err != nil {
					return errors.Wrap(err, "push pachyderm")
				}
			}
			if err := cluster.InstallPachyderm(ctx, &opts); err != nil {
				return errors.Wrap(err, "install pachyderm")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&opts.Namespace, "namespace", "default", "The Kubernetes namespace to install/upgrade.")
	cmd.Flags().BoolVar(&opts.Diff, "diff", false, "If set, instead of deploying, just print a diff between what would be deployed and what is currently deployed.")
	cmd.Flags().BoolVar(&opts.NoSwitchContext, "no-switch-context", false, "If set, don't switch to this Pachyderm context.")
	cmd.Flags().StringVar(&opts.ConsoleTag, "console", "", "If set, use this version of console instead of what's in the helm chart.")
	return cmd
}

//nolint:unused
func restartClusterCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restart cluster",
		Short: "Restart the entire kind cluster and redeploy Pachyderm",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Restarting the cluster...")
			// TODO: Implement 'restart cluster' logic
			panic("Not implemented")
		},
	}
}

//nolint:unused
func restartDeploymentCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restart pach-deployment",
		Short: "Undeploy and redeploy Pachyderm, but leave the cluster intact.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Restarting Pachyderm...")
			// TODO: Implement 'restart pach-deployment' logic
			panic("Not implemented")
		},
	}
}

//nolint:unused
func restartPachdCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restart pods",
		Short: "Restart all pachyderm pods, but leave all persistent cluster state.",
		Long: "Restart all pachyderm pods, leaving other cluster state intact. " +
			"This can be the fastest way to test small code changes--if you rebuild " +
			"e.g. the 'pachd:local' image and then restart all pachd pods, they'll " +
			"load your new code on startup.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Restarting Pachyderm...")
			// TODO: Implement 'restart pods' logic
			panic("Not implemented")
		},
	}
}

//nolint:unused
func updateDeploymentCmd() *cobra.Command {
	return &cobra.Command{
		Use: "update pach-deployment",
		Short: "Build & push worker and pachd images, and then undeploy and " +
			"redeploy Pachyderm",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Redeploying Pachyderm...")
			// TODO: Implement 'update pach-deployment' logic
			panic("Not implemented")
		},
	}
}

//nolint:unused
func updatePachdCmd() *cobra.Command {
	return &cobra.Command{
		Use: "update pods",
		Short: "Build & push worker and pachd images, and restart all pachd and " +
			"worker pods (but leave all cluster state intact)",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Restarting Pachyderm pods...")
			// TODO: Implement 'update pods' logic
			panic("Not implemented")
		},
	}
}

//nolint:unused
func updateImagesCmd() *cobra.Command {
	return &cobra.Command{
		Use: "update images",
		Short: "Build & push worker and pachd images, but don't restart existing " +
			"deployments (useful for testing)",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Updating images in kind...")
			// TODO: Implement 'update images' logic
			panic("Not implemented")
		},
	}
}

//nolint:unused
func printImagesCmd() *cobra.Command {
	return &cobra.Command{
		Use: "print images",
		Short: "Print the list of all images that this tool references when " +
			"deploying a Pachyderm test cluster with the given flags",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Listing images...")
			// TODO: Implement 'list images' logic
			panic("Not implemented")
		},
	}
}

//nolint:unused
func printHelmCmd() *cobra.Command {
	return &cobra.Command{
		Use: "print helm-cmd",
		Short: "Print the helm command (including all helm values) that this " +
			"tool uses when deploying a Pachyderm test cluster with the given flags",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Listing values...")
			// TODO: Implement 'list values' logic
			panic("Not implemented")
		},
	}
}

//nolint:unused
func printManifestCmd() *cobra.Command {
	return &cobra.Command{
		Use: "print manifest",
		Short: "Print the kubernetes manifest used to deploy pachyderm " +
			"Pachyderm test cluster with the given flags",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Printing manifest...")
			// TODO: Implement 'print manifest' logic
			panic("Not implemented")
		},
	}
}
