package cmd

import (
	"os"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	pfscmds "github.com/pachyderm/pachyderm/v2/src/server/pfs/cmds"
	"github.com/pachyderm/pachyderm/v2/src/server/pfs/fuse"
	"github.com/spf13/cobra"
)

func MountServerCmd() *cobra.Command {
	var mountDir string
	rootCmd := &cobra.Command{
		Use:   os.Args[0],
		Short: "Start a mount server for controlling FUSE mounts via a local REST API.",
		Long:  "Starts a REST mount server, running in the foreground and logging to stdout.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) error {
			serverOpts := &fuse.ServerOptions{
				MountDir: mountDir,
			}
			pfscmds.PrintWarning()
			c, err := client.NewOnUserMachine("user", client.WithDialTimeout(5*time.Second))
			if err != nil {
				return fuse.Server(serverOpts, nil)
			}
			return fuse.Server(serverOpts, c)
		}),
	}
	rootCmd.Flags().StringVar(&mountDir, "mount-dir", "/pfs", "Target directory for mounts e.g /pfs")

	return rootCmd
}
