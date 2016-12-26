package fuse

import "github.com/urfave/cli"

// NewCommand returns mount command.
func NewCommand() cli.Command {
	return cli.Command{
		Name:  "fuse",
		Usage: "Mount or unmount PFS locally.",
		Subcommands: []cli.Command{
			newMountCommand(),
			newUnmountCommand(),
		},
	}
}
