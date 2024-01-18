package pachdev

import (
	"fmt"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/kindenv"
	"github.com/spf13/cobra"
)

const DefaultClusterName = "pach"

func DeleteClusterCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete-cluster [<name>]",
		Short: "Deploy a local Kubernetes cluster",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var name string
			if len(args) > 0 {
				name = args[0]
			}
			ctx := cmd.Context()
			cluster, err := kindenv.New(ctx, name)
			if err != nil {
				return errors.Wrap(err, "kindenv.New")
			}
			if err := cluster.Delete(ctx); err != nil {
				return errors.Wrap(err, "cluster.Delete")
			}
			return nil
		},
	}
}

func CreateClusterCmd() *cobra.Command {
	var registry string
	cmd := &cobra.Command{
		Use:   "create-cluster [<name>]",
		Short: "Deploy a local Kubernetes cluster",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := DefaultClusterName
			if len(args) > 0 {
				name = args[0]
			}
			ctx := cmd.Context()
			cluster, err := kindenv.New(ctx, name)
			if err != nil {
				return errors.Wrap(err, "kindenv.New")
			}
			if err := cluster.Create(ctx, &kindenv.CreateOpts{
				TestNamespaceCount: 3,
				ExternalRegistry:   registry,
				BindHTTPPorts:      true,
				StartingPort:       30600,
			}); err != nil {
				return errors.Wrap(err, "kindenv.Create")
			}
			return nil
		},
	}
	return cmd
}

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
