package cmds

import (
	"errors"
	"fmt"
	"os"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"

	"github.com/golang/snappy"
	"github.com/spf13/cobra"
)

// Cmds returns a slice containing admin commands.
func Cmds() []*cobra.Command {
	var commands []*cobra.Command

	var enable bool
	var disable bool
	metrics := &cobra.Command{
		Short: "Gets or sets whether metrics are enabled.",
		Long:  "Gets or sets whether metrics are enabled.",
		Example: `
# Disable metrics:
$ {{alias}} --disable

# Get whether metrics are enabled:
$ {{alias}}

# Enable metrics:
$ {{alias}} --enable`,
		Run: cmdutil.RunFixedArgs(0, func(args []string) (retErr error) {
			if enable && disable {
				return errors.New("cannot set both `--enable` and `--disable`")
			}

			cfg, err := config.Read()
			if err != nil {
				return err
			}

			if enable {
				cfg.NoMetrics = false
				return cfg.Write()
			} else if disable {
				cfg.NoMetrics = true
				return cfg.Write()
			}

			fmt.Printf("%v\n", !cfg.NoMetrics)
			return nil
		}),
	}
	metrics.Flags().BoolVar(&enable, "enable", false, "enable metrics")
	metrics.Flags().BoolVar(&disable, "disable", false, "disable metrics")
	commands = append(commands, cmdutil.CreateAlias(metrics, "metrics"))

	return commands
}
