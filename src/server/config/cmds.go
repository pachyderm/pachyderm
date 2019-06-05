package cmds

import (
	"errors"
	"fmt"
	"os"

	"github.com/pachyderm/pachyderm/src/client/pkg/config"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/spf13/cobra"
)

// Cmds returns a slice containing admin commands.
func Cmds() []*cobra.Command {
	marshaller := &jsonpb.Marshaler{
		Indent:   "  ",
		OrigName: true,
	}

	var commands []*cobra.Command

	getMetrics := &cobra.Command{
		Short: "Gets whether metrics are enabled.",
		Long:  "Gets whether metrics are enabled.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) (retErr error) {
			cfg, err := config.Read()
			if err != nil {
				return err
			}
			fmt.Printf("%v\n", !cfg.V2.NoMetrics)
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(getMetrics, "config get metrics"))

	setMetrics := &cobra.Command{
		Short: "Sets whether metrics are enabled.",
		Long:  "Sets whether metrics are enabled.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) (retErr error) {
			metrics := true
			if args[0] == "false" {
				metrics = false
			} else if args[0] != "true" {
				return errors.New("invalid argument; use either `true` or `false`")
			}

			cfg, err := config.Read()
			if err != nil {
				return err
			}

			cfg.V2.NoMetrics = !metrics
			return cfg.Write()
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(setMetrics, "config set metrics"))

	useContext := &cobra.Command{
		Short: "Sets the currently active context.",
		Long:  "Sets the currently active context.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) (retErr error) {
			cfg, err := config.Read()
			if err != nil {
				return err
			}
			if _, ok := cfg.V2.Contexts[args[0]]; !ok {
				return fmt.Errorf("context does not exist: %s", args[0])
			}
			cfg.V2.ActiveContext = args[0]
			return cfg.Write()
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(useContext, "config use context"))

	var active bool
	getContext := &cobra.Command{
		Short: "Gets a context.",
		Long:  "Gets the config of a context by its name, or the current active one.",
		Run: cmdutil.RunBoundedArgs(0, 1, func(args []string) (retErr error) {
			if active && len(args) == 1 {
				return errors.New("cannot get both the active context, and a specific one by its name")
			}

			cfg, err := config.Read()
			if err != nil {
				return err
			}

			var name string
			if active {
				name = cfg.V2.ActiveContext
			} else {
				name = args[0]
			}

			context, ok := cfg.V2.Contexts[name]
			if !ok {
				return fmt.Errorf("context does not exist: %s", name)
			}

			return marshaller.Marshal(os.Stdout, context)
		}),
	}
	getContext.Flags().BoolVar(&active, "active", false, "Get the active context.")
	commands = append(commands, cmdutil.CreateAlias(getContext, "config get context"))

	return commands
}
