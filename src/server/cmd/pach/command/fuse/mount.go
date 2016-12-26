package fuse

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pfs/fuse"

	"github.com/urfave/cli"
)

func newMountCommand() cli.Command {
	return cli.Command{
		Name:        "mount",
		Aliases:     []string{"m"},
		Usage:       "Mount pfs locally. This command blocks.",
		ArgsUsage:   "path/to/mount/point",
		Description: "Mount pfs locally. This command blocks.",
		Action:      actMount,
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "debug, d", Usage: "Turn on debug messages."},
			cli.BoolFlag{Name: "all-commits, a", Usage: "Show archived and cancelled commits."},
			cli.IntFlag{Name: "file-shard, s", Usage: "file shard to read", Value: 0},
			cli.IntFlag{Name: "file-modulus, m", Usage: "file shard to read", Value: 1},
			cli.IntFlag{Name: "block-shard, b", Usage: "block shard to read", Value: 0},
			cli.IntFlag{Name: "block-modulus, n", Usage: "file shard to read", Value: 1},
		},
	}
}

func actMount(c *cli.Context) error {
	clnt, err := client.NewMetricsClientFromAddress(c.GlobalString("address"), c.GlobalBool("metrics"), "fuse")
	if err != nil {
		return err
	}
	go func() { clnt.KeepConnected(nil) }()
	mounter := fuse.NewMounter(c.GlobalString("address"), clnt)
	mountPoint := c.Args().First()
	ready := make(chan bool)
	go func() {
		<-ready
		fmt.Println("Filesystem mounted, CTRL-C to exit.")
	}()
	return mounter.Mount(mountPoint, shard(c), nil, ready, c.Bool("debug"), c.Bool("all-commits"), false)
}

func shard(c *cli.Context) *pfsclient.Shard {
	return &pfsclient.Shard{
		FileNumber:   uint64(c.Int("file-shard")),
		FileModulus:  uint64(c.Int("file-modulus")),
		BlockNumber:  uint64(c.Int("block-shard")),
		BlockModulus: uint64(c.Int("block-modulus")),
	}
}
