package deploy

import (
	"bytes"
	"fmt"
	"os"

	"github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/server/pkg/deploy/assets"

	"github.com/urfave/cli"

	"go.pedge.io/pkg/exec"
)

// NewCommand returns cli command for deploying a pachyderm cluster.
func NewCommand() cli.Command {
	return cli.Command{
		Name:        "deploy",
		Aliases:     []string{"d"},
		Usage:       "Deploy a Pachyderm cluster.",
		ArgsUsage:   "deploy amazon|google|microsoft|basic",
		Description: "Deploy a Pachyderm cluster.",
		Before:      configOpts,
		Subcommands: []cli.Command{
			newLocalCommand(),
			newAmazonCommand(),
			newGoogleCommand(),
			newMicrosoftCommand(),
		},
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "dev, d",
				Usage: "Don't use a specific version of pachyderm/pachd.",
			},
			cli.IntFlag{
				Name:  "shards",
				Usage: "Number of Pachd nodes (stateless Pachyderm API servers).",
				Value: 1,
			},
			cli.IntFlag{
				Name: "rethink-shards",
				Usage: "Number of RethinkDB shards (for pfs metadata storage) if " +
					"--deploy-rethink-as-stateful-set is used.",
				Value: 1,
			},
			cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Don't actually deploy pachyderm to Kubernetes, instead just print the manifest.",
			},
			cli.StringFlag{
				Name: "rethinkdb-cache-size",
				Usage: "Size of in-memory cache to use for Pachyderm's RethinkDB instance, " +
					"e.g. \"2G\". Size is specified in bytes, with allowed SI suffixes (M, K, G, Mi, Ki, Gi, etc).",
				Value: "768M",
			},
			cli.StringFlag{
				Name:  "log-level",
				Usage: "The level of log messages to print options are, from least to most verbose: \"error\", \"info\", \"debug\".",
				Value: "info",
			},
			cli.BoolFlag{
				Name: "deploy-rethink-as-rc",
				Usage: "Defunct flag (does nothing). The default behavior since " +
					"Pachyderm 1.3.2 is to manage RethinkDB with a Kubernetes Replication Controller",
			},
			cli.BoolFlag{
				Name: "deploy-rethink-as-stateful-set",
				Usage: "Deploy RethinkDB as a multi-node cluster " +
					"controlled by kubernetes StatefulSet, instead of a single-node instance controlled by a Kubernetes Replication Controller. Note that both " +
					"your local kubectl binary and the kubernetes server must be at least version 1.5.",
			},
		},
	}
}

var opts *assets.AssetOpts

func configOpts(c *cli.Context) error {
	if c.Bool("deploy-rethink-as-rc") && c.Bool("deploy-rethink-as-stateful-set") {
		return fmt.Errorf("Error: pachctl deploy received contradictory flags: " +
			"--deploy-rethink-as-rc and --deploy-rethink-as-stateful-set")
	}
	if c.Bool("deploy-rethink-as-rc") {
		fmt.Fprintf(os.Stderr, "Warning: --deploy-rethink-as-rc is no Descriptioner "+
			"necessary (and is ignored). The default behavior since Pachyderm "+
			"1.3.2 is to manage RethinkDB with a Kubernetes Replication Controller. "+
			"This flag will be removed by Pachyderm's 1.4 release, so please remove "+
			"it from your scripts. Also see --deploy-rethink-as-stateful-set.\n")
	}
	if !c.Bool("deploy-rethink-as-stateful-set") && c.Int("rethink-shards") > 1 {
		return fmt.Errorf("Error: --deploy-rethink-as-stateful-set was not set, " +
			"but --rethink-shards was set to value >1. Since 1.3.2, 'pachctl deploy' " +
			"deploys RethinkDB as a single-node instance by default, unless " +
			"--deploy-rethink-as-stateful-set is set. Please set that flag if you " +
			"wish to deploy RethinkDB as a multi-node cluster.")
	}
	opts = &assets.AssetOpts{
		PachdShards:                uint64(c.Int("shards")),
		RethinkShards:              uint64(c.Int("rethink-shards")),
		RethinkdbCacheSize:         c.String("rethinkdb-cache-size"),
		DeployRethinkAsStatefulSet: c.Bool("deploy-rethink-as-stateful-set"),
		Version:                    version.PrettyPrintVersion(version.Version),
		LogLevel:                   c.String("log-level"),
		Metrics:                    c.GlobalBool("metrics"),
	}
	return nil
}

func maybeKcCreate(c *cli.Context, manifest *bytes.Buffer) error {
	if c.Bool("dry-run") {
		_, err := os.Stdout.Write(manifest.Bytes())
		return err
	}
	return pkgexec.RunIO(pkgexec.IO{Stdin: manifest, Stdout: os.Stdout, Stderr: os.Stderr}, "kubectl", "create", "-f", "-")
}
