package cmds

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/pachyderm/pachyderm/src/client/pkg/config"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/spf13/cobra"
)

const (
	listContextHeader = "ACTIVE\tNAME"
)

func readContext() (*config.Context, error) {
	var buf bytes.Buffer
	var decoder *json.Decoder
	var result config.Context

	contextReader := io.TeeReader(os.Stdin, &buf)
	fmt.Println("Reading from stdin.")
	decoder = json.NewDecoder(contextReader)

	if err := jsonpb.UnmarshalNext(decoder, &result); err != nil {
		if err == io.EOF {
			return nil, err
		}
		return nil, fmt.Errorf("malformed context: %s", err)
	}
	return &result, nil
}

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
		Run: cmdutil.Run(func(args []string) (retErr error) {
			cfg, err := config.ReadPachConfig()
			if err != nil {
				return err
			}
			fmt.Printf("%v\n", cfg.V3.Metrics)
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

			cfg, err := config.ReadPachConfig()
			if err != nil {
				return err
			}

			cfg.V3.Metrics = metrics
			return cfg.Write()
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(setMetrics, "config set metrics"))

	getContext := &cobra.Command{
		Short: "Gets the active Pachyderm context.",
		Long:  "Gets the active Pachyderm context.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) (retErr error) {
			context, err := config.ReadPachContext()
			if err != nil {
				return err
			}
			if err = marshaller.Marshal(os.Stdout, context); err != nil {
				return err
			}
			fmt.Println()
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(getContext, "config get context"))

	var overwrite bool
	setContext := &cobra.Command{
		Short: "Set the active Pachyderm context.",
		Long:  "Set the active Pachyderm context.",
		Run: cmdutil.RunFixedArgs(0, func(args []string) (retErr error) {
			context, err := readContext()
			if err != nil {
				return err
			}
			return context.Write()
		}),
	}
	setContext.Flags().BoolVar(&overwrite, "overwrite", false, "Overwrite a context if it already exists.")
	commands = append(commands, cmdutil.CreateAlias(setContext, "config set context"))

	var pachdAddress string
	updateContext := &cobra.Command{
		Short: "Updates a context.",
		Long:  "Updates an existing context config from a given name.",
		Run: cmdutil.RunCmdFixedArgs(0, func(cmd *cobra.Command, args []string) (retErr error) {
			context, err := config.ReadPachContext()
			if err != nil {
				return err
			}

			if cmd.Flags().Changed("pachd-address") {
				// Use this method since we want to differentiate between no
				// `pachd-address` flag being set (the value shouldn't be
				// changed) vs the flag being an empty string (meaning we want
				// to set the value to an empty string)
				context.PachdAddress = pachdAddress
			}

			return context.Write()
		}),
	}
	updateContext.Flags().StringVar(&pachdAddress, "pachd-address", "", "Set a new name pachd address.")
	commands = append(commands, cmdutil.CreateAlias(updateContext, "config update context"))

	configDocs := &cobra.Command{
		Short: "Manages the pachyderm config.",
		Long:  "Gets/sets pachyderm config values.",
	}
	commands = append(commands, cmdutil.CreateDocsAlias(configDocs, "config", "^pachctl config "))

	configGetRoot := &cobra.Command{
		Short: "Commands for getting pachyderm config values",
		Long:  "Commands for getting pachyderm config values",
	}
	commands = append(commands, cmdutil.CreateAlias(configGetRoot, "config get"))

	configSetRoot := &cobra.Command{
		Short: "Commands for setting pachyderm config values",
		Long:  "Commands for setting pachyderm config values",
	}
	commands = append(commands, cmdutil.CreateAlias(configSetRoot, "config set"))

	configUpdateRoot := &cobra.Command{
		Short: "Commands for updating pachyderm config values",
		Long:  "Commands for updating pachyderm config values",
	}
	commands = append(commands, cmdutil.CreateAlias(configUpdateRoot, "config update"))

	configDeleteRoot := &cobra.Command{
		Short: "Commands for deleting pachyderm config values",
		Long:  "Commands for deleting pachyderm config values",
	}
	commands = append(commands, cmdutil.CreateAlias(configDeleteRoot, "config delete"))

	configListRoot := &cobra.Command{
		Short: "Commands for listing pachyderm config values",
		Long:  "Commands for listing pachyderm config values",
	}
	commands = append(commands, cmdutil.CreateAlias(configListRoot, "config list"))

	return commands
}
