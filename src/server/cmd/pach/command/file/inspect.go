package file

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pfs/pretty"

	"github.com/urfave/cli"
)

func newInspectCommand() cli.Command {
	return cli.Command{
		Name:        "inspect",
		Aliases:     []string{"i"},
		Usage:       "Return info about a file.",
		ArgsUsage:   "repo-name commit-id path/to/file",
		Description: "Return info about a file.",
		Action:      actInspect,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "from, f",
				Usage: "only consider data written since this commit",
			},
			cli.BoolFlag{
				Name:  "full-file",
				Usage: "if there has been data since the from commit return the full file",
			},
			cli.IntFlag{Name: "file-shard, s", Usage: "file shard to read", Value: 0},
			cli.IntFlag{Name: "file-modulus, m", Usage: "file shard to read", Value: 1},
			cli.IntFlag{Name: "block-shard, b", Usage: "block shard to read", Value: 0},
			cli.IntFlag{Name: "block-modulus, n", Usage: "file shard to read", Value: 1},
		},
	}
}

func actInspect(c *cli.Context) error {
	if c.NArg() != 1 {
		return fmt.Errorf("invalid argument")
	}
	clnt, err := client.NewMetricsClientFromAddress(c.GlobalString("address"), c.GlobalBool("metrics"), "user")
	if err != nil {
		return err
	}
	args := c.Args()
	fileInfo, err := clnt.InspectFile(args[0], args[1], args[2], c.String("from"), c.Bool("full-file"), shard(c))
	if err != nil {
		return err
	}
	if fileInfo == nil {
		return fmt.Errorf("file %s not found", args[2])
	}
	return pretty.PrintDetailedFileInfo(fileInfo)
}

func shard(c *cli.Context) *pfsclient.Shard {
	return &pfsclient.Shard{
		FileNumber:   uint64(c.Int("file-shard")),
		FileModulus:  uint64(c.Int("file-modulus")),
		BlockNumber:  uint64(c.Int("block-shard")),
		BlockModulus: uint64(c.Int("block-modulus")),
	}
}
