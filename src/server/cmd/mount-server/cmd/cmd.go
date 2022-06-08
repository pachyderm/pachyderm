package cmd

import (
	"os"

	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	pfscmds "github.com/pachyderm/pachyderm/v2/src/server/pfs/cmds"
	"github.com/pachyderm/pachyderm/v2/src/server/pfs/fuse"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func MountServerCmd() *cobra.Command {
	var mountDir string
	rootCmd := &cobra.Command{
		Use:   os.Args[0],
		Short: "Start a mount server for controlling FUSE mounts via a local REST API.",
		Long:  "Starts a REST mount server, running in the foreground and logging to stdout.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {

			// Show info messages to user by default
			logrus.SetLevel(logrus.InfoLevel)

			serverOpts := &fuse.ServerOptions{
				MountDir: mountDir,
			}
			pfscmds.PrintWarning()
			return fuse.Server(serverOpts, nil)
		}),
	}
	rootCmd.Flags().StringVar(&mountDir, "mount-dir", "/pfs", "Target directory for mounts e.g /pfs")

	return rootCmd
}
