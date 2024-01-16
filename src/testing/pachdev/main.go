package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var verbose bool

func restartClusterCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restart cluster",
		Short: "Restart the entire kind cluster and redeploy Pachyderm",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Restarting the cluster...")
			// TODO: Implement 'restart cluster' logic
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
		},
	}
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "pachdev",
		Short: "A CLI tool for Pachyderm development",
	}

	rootCmd.AddCommand(restartClusterCmd())
	rootCmd.AddCommand(restartDeploymentCmd())
	rootCmd.AddCommand(restartPachdCmd())
	rootCmd.AddCommand(updateDeploymentCmd())
	rootCmd.AddCommand(updatePachdCmd())
	rootCmd.AddCommand(updateImagesCmd())
	rootCmd.AddCommand(printImagesCmd())
	rootCmd.AddCommand(printHelmCmd())
	rootCmd.AddCommand(printManifestCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
