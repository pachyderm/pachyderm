package main

import (
	"io/ioutil"
	"log"
	"os"
	"syscall"

	"google.golang.org/grpc/grpclog"

	"github.com/urfave/cli"
	"go.pedge.io/lion"

	"github.com/pachyderm/pachyderm/src/server/cmd/pach/command/branch"
	"github.com/pachyderm/pachyderm/src/server/cmd/pach/command/commit"
	"github.com/pachyderm/pachyderm/src/server/cmd/pach/command/deploy"
	"github.com/pachyderm/pachyderm/src/server/cmd/pach/command/file"
	"github.com/pachyderm/pachyderm/src/server/cmd/pach/command/fuse"
	"github.com/pachyderm/pachyderm/src/server/cmd/pach/command/job"
	"github.com/pachyderm/pachyderm/src/server/cmd/pach/command/misc"
	"github.com/pachyderm/pachyderm/src/server/cmd/pach/command/pipeline"
	"github.com/pachyderm/pachyderm/src/server/cmd/pach/command/repo"
	"github.com/pachyderm/pachyderm/src/server/cmd/pach/command/shell"
	"github.com/pachyderm/pachyderm/src/server/cmd/pach/command/version"
)

func main() {

	// Silence log from grpc and lion.
	grpclog.SetLogger(log.New(ioutil.Discard, "", 0))
	lion.SetLevel(lion.LevelNone)

	// Disable cli's default version flag -v / --version
	cli.VersionFlag = cli.BoolFlag{}

	app := cli.NewApp()
	app.Name = "pach"
	app.Usage = `Access the Pachyderm API.

  Environment variables:
    ADDRESS=<host>:<port>, the pachd server to connect to (e.g. 127.0.0.1:30650).
`
	app.Commands = []cli.Command{
		repo.NewCommand(),
		commit.NewCommand(),
		branch.NewCommand(),
		file.NewCommand(),
		fuse.NewCommand(),
		job.NewCommand(),
		pipeline.NewCommand(),
		deploy.NewCommand(),
		misc.NewCommand(),
		version.NewCommand(),
		shell.NewCommand(
			app,
			shell.WithPrompt("pach > "),
			shell.WithSignalMask(syscall.SIGINT),
			shell.WithExitCmds("q", "quit"),
		),
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "address, a",
			Value:  "0.0.0.0:30650",
			Usage:  "the pachd server to connect to",
			EnvVar: "ADDRESS",
		},
		cli.BoolFlag{
			Name:  "metrics, m",
			Usage: "Report user metrics for this command",
		},
		cli.BoolFlag{
			Name:  "verbose, v",
			Usage: "Output verbose logs",
		},
	}

	app.Run(os.Args)
}
