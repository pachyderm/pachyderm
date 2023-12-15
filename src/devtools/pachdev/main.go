package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	// logging library used by yqlib
	logging "gopkg.in/op/go-logging.v1"
)

var verbose bool

func restartClusterCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restart cluster",
		Short: "Restart the entire kube cluster and redeploy Pachyderm",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Restarting the cluster...")
			// TODO: Implement restart cluster logic
		},
	}
}

func restartDeploymentCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restart deployment",
		Short: "Undeploy and redeploy Pachyderm, but leave the kube cluster intact",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Restarting Pachyderm...")
			// TODO: Implement restart Pachyderm logic
		},
	}
}

func restartPachdCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restart pachd",
		Short: "Restart all pachd and worker pods, but leave all cluster state intact",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Restarting Pachyderm...")
			// TODO: Implement restart Pachyderm logic
		},
	}
}

func updateDeploymentCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update deployment",
		Short: "Build & push worker and pachd images, and undeploy and redeploy Pachyderm",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Restarting Pachyderm...")
			// TODO: Implement restart Pachyderm logic
		},
	}
}

func updatePachdCmd() *cobra.Command {
	return &cobra.Command{
		Use: "update pachd",
		Short: "Build & push worker and pachd images, and restart all pachd and " +
			"worker pods (but leave all cluster state intact)",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Restarting Pachyderm...")
			// TODO: Implement restart Pachyderm logic
		},
	}
}

func updateImagesCmd() *cobra.Command {
	return &cobra.Command{
		Use: "update images",
		Short: "Build & push worker and pachd images, but don't restart existing " +
			"deployments (useful for testing)",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Updating images...")
			// images := getImages()
			// TODO: Implement update images logic
		},
	}
}

func getImagesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "images",
		Short: "List all images that this tool references when deploying a Pachyderm test cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			op, err := NewDeployOp(verbose)
			if err != nil {
				return err
			}
			defer op.Cancel()
			images, err := op.getImages()
			if err != nil {
				return err
			}
			for _, i := range images {
				fmt.Printf("%v\n", i)
			}
			return nil
		},
	}
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "pachdev",
		Short: "A CLI tool for Pachyderm development",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Reach into yqlib and set the logging level. It's annoying that
			// this is necessary, but the default debug level of go-logging.v1
			// (which yqlib uses) appears to be "DEBUG", so without this code,
			// pachdev emits a lot of noisy logs eminating from yqlib.
			if verbose {
				logging.SetLevel(logging.DEBUG, "yq-lib")
			} else {
				logging.SetLevel(logging.WARNING, "yq-lib")
			}
		},
	}
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose mode")

	rootCmd.AddCommand(restartClusterCmd())
	rootCmd.AddCommand(restartDeploymentCmd())
	rootCmd.AddCommand(restartPachdCmd())
	rootCmd.AddCommand(updateDeploymentCmd())
	rootCmd.AddCommand(updatePachdCmd())
	rootCmd.AddCommand(updateImagesCmd())

	/* >>> */
	rootCmd.AddCommand(getImagesCmd())
	/* >>> */

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
