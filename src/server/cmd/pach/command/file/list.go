package file

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/pachyderm/pachyderm/src/client"
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pfs/pretty"

	"github.com/urfave/cli"
)

func newListCommand() cli.Command {
	return cli.Command{
		Name:        "list",
		Aliases:     []string{"l"},
		Usage:       "Return the files in a directory.",
		ArgsUsage:   "repo-name commit-id path/to/dir",
		Description: "Return the files in a directory.",
		Action:      actList,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "from, f",
				Usage: "only consider data written since this commit",
			},
			cli.BoolFlag{
				Name:  "full-file",
				Usage: "if there has been data since the from commit return the full file",
			},
			cli.BoolFlag{
				Name:  "recurse, r",
				Usage: "if recurse is true, compute and display the sizes of directories",
			},
			cli.BoolFlag{
				Name:  "fast",
				Usage: "if fast is true, don't compute the sizes of files; this makes list-file faster",
			},
			cli.IntFlag{Name: "file-shard, s", Usage: "file shard to read", Value: 0},
			cli.IntFlag{Name: "file-modulus, m", Usage: "file shard to read", Value: 1},
			cli.IntFlag{Name: "block-shard, b", Usage: "block shard to read", Value: 0},
			cli.IntFlag{Name: "block-modulus, n", Usage: "file shard to read", Value: 1},
		},
	}
}

func actList(c *cli.Context) error {
	fast := c.Bool("fast")
	recurse := c.Bool("recurse")
	if fast && recurse {
		return fmt.Errorf("you may only provide either --fast or --recurse, but not both")
	}
	clnt, err := client.NewMetricsClientFromAddress(c.GlobalString("address"), c.GlobalBool("metrics"), "user")
	if err != nil {
		return err
	}
	var path string
	args := c.Args()
	if len(args) == 3 {
		path = args[2]
	}
	var fileInfos []*pfsclient.FileInfo
	if fast {
		fileInfos, err = clnt.ListFileFast(args[0], args[1], path, c.String("from"), c.Bool("full-file"), shard(c))
	} else {
		fileInfos, err = clnt.ListFile(args[0], args[1], path, c.String("from"), c.Bool("full-file"), shard(c), recurse)
	}
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)
	pretty.PrintFileInfoHeader(w)
	for _, fileInfo := range fileInfos {
		pretty.PrintFileInfo(w, fileInfo, recurse, fast)
	}
	return w.Flush()
}
